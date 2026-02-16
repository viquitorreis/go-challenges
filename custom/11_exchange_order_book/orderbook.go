package main

import (
	"container/heap"
	"log/slog"
	"sync"
	"time"
)

type Side int

const (
	Bid Side = iota
	Ask
)

// Order representa uma ordem individual no book
type Order struct {
	ID        string
	UserID    string
	Side      Side
	Price     int // em centavos para evitar float
	Quantity  int
	Timestamp int64 // unix nano — usado para price-time priority
}

// Trade representa uma execução — quando bid e ask se encontram
type Trade struct {
	BidOrderID string
	AskOrderID string
	Price      int
	Quantity   int
}

// PriceLevel agrupa todas as ordens num mesmo preço (fila FIFO)
type PriceLevel struct {
	Price  int
	Orders []*Order
}

// OrderBook mantém bids e asks para um **único ticker**
type OrderBook struct {
	Symbol string

	//
	bids     map[int][]*Order
	bidsHeap *bidsHeap

	asks     map[int][]*Order
	asksHeap *asksHeap

	tracker map[string]*Order

	mu sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	bh := &bidsHeap{}
	heap.Init(bh)

	ah := &asksHeap{}
	heap.Init(ah)

	return &OrderBook{
		Symbol:   symbol,
		bidsHeap: bh,
		bids:     make(map[int][]*Order),
		asksHeap: ah,
		asks:     make(map[int][]*Order),
		tracker:  make(map[string]*Order),
	}
}

func NewOrder(id, userID string, side Side, price, qty int) *Order {
	return &Order{
		ID:        id,
		UserID:    userID,
		Side:      side,
		Price:     price,
		Quantity:  qty,
		Timestamp: time.Now().Unix(),
	}
}

// AddOrder adiciona uma ordem ao lado correto do book
func (ob *OrderBook) AddOrder(o *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	switch o.Side {
	case Bid:
		// so enviamos para heap se for um preço novo, não por ordem nova
		if _, exists := ob.bids[o.Price]; !exists {
			heap.Push(ob.bidsHeap, o.Price)
		}
		ob.bids[o.Price] = append(ob.bids[o.Price], o)

	case Ask:
		if _, exists := ob.asks[o.Price]; !exists {
			heap.Push(ob.asksHeap, o.Price)
		}
		ob.asks[o.Price] = append(ob.asks[o.Price], o)

	default:
		slog.Warn("wrong type of side for order", "side", o.Side)
	}

	ob.tracker[o.ID] = o
}

func (ob *OrderBook) Match() []Trade {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	res := []Trade{}
	for ob.asksHeap.Len() > 0 && ob.bidsHeap.Len() > 0 && (*ob.bidsHeap)[0] >= (*ob.asksHeap)[0] {
		bestBidPrice := (*ob.bidsHeap)[0]
		bestAskPrice := (*ob.asksHeap)[0]

		if len(ob.bids[bestBidPrice]) == 0 {
			heap.Pop(ob.bidsHeap)
			continue
		}
		if len(ob.asks[bestAskPrice]) == 0 {
			heap.Pop(ob.asksHeap)
			continue
		}

		bestBid := ob.bids[(*ob.bidsHeap)[0]][0]
		bestAsk := ob.asks[(*ob.asksHeap)[0]][0]

		orderQuantity := min(bestBid.Quantity, bestAsk.Quantity)
		bestBid.Quantity -= orderQuantity
		bestAsk.Quantity -= orderQuantity

		if bestBid.Quantity == 0 {
			ob.bids[bestBid.Price] = ob.bids[bestBid.Price][1:]
		}

		if bestAsk.Quantity == 0 {
			ob.asks[bestAsk.Price] = ob.asks[bestAsk.Price][1:]
		}

		// quando a fila de um nível de preço fica vazia precisamos remover do map e fazer pop
		// se nao o loop fica tentando acessar o nível infinitamente
		if len(ob.bids[bestBid.Price]) == 0 {
			delete(ob.bids, bestBid.Price)
			delete(ob.tracker, bestBid.ID)
			heap.Pop(ob.bidsHeap)
		}

		if len(ob.asks[bestAsk.Price]) == 0 {
			delete(ob.asks, bestAsk.Price)
			delete(ob.tracker, bestAsk.ID)
			heap.Pop(ob.asksHeap)
		}

		res = append(res, Trade{
			BidOrderID: bestBid.ID,
			AskOrderID: bestAsk.ID,
			Price:      bestAsk.Price,
			Quantity:   orderQuantity,
		})

	}
	return res
}

func (ob *OrderBook) Cancel(orderID string) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, exists := ob.tracker[orderID]
	if !exists {
		return false
	}

	var orders map[int][]*Order
	if order.Side == Bid {
		orders = ob.bids
	} else {
		orders = ob.asks
	}

	level := orders[order.Price]
	newLevelOrders := []*Order{}
	for _, ord := range level {
		if ord.ID != orderID {
			newLevelOrders = append(newLevelOrders, ord)
		}
	}

	if len(newLevelOrders) == 0 {
		delete(orders, order.Price)

		// filter heap
		if order.Side == Bid {
			newHeap := bidsHeap{}
			for _, p := range *ob.bidsHeap {
				if p != order.Price {
					newHeap = append(newHeap, p)
				}
			}

			ob.bidsHeap = &newHeap
			heap.Init(ob.bidsHeap)
		} else {
			newHeap := asksHeap{}
			for _, p := range *ob.asksHeap {
				if p != order.Price {
					newHeap = append(newHeap, p)
				}
			}

			ob.asksHeap = &newHeap
			heap.Init(ob.asksHeap)
		}
	} else {
		// order ainda existe nesse preço, apenas atualiza o slice das orders
		orders[order.Price] = newLevelOrders
	}

	delete(ob.tracker, orderID)

	return true
}

func (ob *OrderBook) BidDepth() int {
	count := 0

	ob.mu.RLock()
	for _, orders := range ob.bids {
		for _, o := range orders {
			count += o.Quantity
		}
	}
	ob.mu.RUnlock()

	return count
}

func (ob *OrderBook) AskDepth() int {
	count := 0

	ob.mu.RLock()
	for _, orders := range ob.asks {
		for _, o := range orders {
			count += o.Quantity
		}
	}
	ob.mu.RUnlock()

	return count
}
