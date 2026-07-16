package main

import (
	"container/heap"
	"container/list"
	"log/slog"
	"sync"
	"time"
)

type Side int

const (
	Bid Side = iota
	Ask
)

// Order represents an individual order in the book
type Order struct {
	ID        string
	UserID    string
	Side      Side
	Price     int // in cents to avoid float
	Quantity  int
	Timestamp int64 // unix nano, used for price-time priority
}

// Trade represents a trade execution when a bid and ask match
type Trade struct {
	BidOrderID string
	AskOrderID string
	Price      int
	Quantity   int
}

// PriceLevel groups all orders at the same price (FIFO queue)
type PriceLevel struct {
	Price  int
	Orders *list.List
	Side   Side
}

// OrderBook maintains bids and asks for a single ticker
type OrderBook struct {
	Symbol string

	// price -> all bids for the price (people trying to buy on that price)
	bids     map[int]*PriceLevel
	bidsHeap *bidsHeap

	// price -> all asks for the price (people trying to sell on that price)
	asks     map[int]*PriceLevel
	asksHeap *asksHeap

	tracker map[string]*list.Element

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
		bids:     map[int]*PriceLevel{},
		asksHeap: ah,
		asks:     map[int]*PriceLevel{},
		tracker:  make(map[string]*list.Element),
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

// AddOrder adds a new order at the correct book
func (ob *OrderBook) AddOrder(o *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	switch o.Side {
	case Bid:
		level, exists := ob.bids[o.Price]
		if !exists {
			level = &PriceLevel{
				Price:  o.Price,
				Orders: list.New(),
				Side:   Bid,
			}

			ob.bids[o.Price] = level

			heap.Push(ob.bidsHeap, o.Price)
		}

		// push back because its a FIFO
		el := level.Orders.PushBack(o)

		ob.tracker[o.ID] = el
	case Ask:
		level, exists := ob.asks[o.Price]
		if !exists {
			level = &PriceLevel{
				Price:  o.Price,
				Orders: list.New(),
				Side:   Ask,
			}

			ob.asks[o.Price] = level

			heap.Push(ob.asksHeap, o.Price)
		}

		// push back because its a FIFO
		el := level.Orders.PushBack(o)

		ob.tracker[o.ID] = el

	default:
		slog.Warn("wrong type of side for order", "side", o.Side)
	}
}

func (ob *OrderBook) Match() []Trade {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	res := []Trade{}
	// keep filling the order, always starting from the cheapest ask
	for ob.asksHeap.Len() > 0 && ob.bidsHeap.Len() > 0 && (*ob.bidsHeap)[0] >= (*ob.asksHeap)[0] {
		bestBidPrice := (*ob.bidsHeap)[0]
		bestAskPrice := (*ob.asksHeap)[0]

		bidLevel := ob.bids[bestBidPrice]
		askLevel := ob.asks[bestAskPrice]

		if bidLevel == nil || bidLevel.Orders.Len() == 0 {
			heap.Pop(ob.bidsHeap)
			continue
		}

		if askLevel == nil || askLevel.Orders.Len() == 0 {
			heap.Pop(ob.asksHeap)
			continue
		}

		bidElem := bidLevel.Orders.Front()
		askElem := askLevel.Orders.Front()

		bestBid := bidElem.Value.(*Order)
		bestAsk := askElem.Value.(*Order)

		orderQuantity := min(bestBid.Quantity, bestAsk.Quantity)
		bestBid.Quantity -= orderQuantity
		bestAsk.Quantity -= orderQuantity

		if bestBid.Quantity == 0 {
			bidLevel.Orders.Remove(bidElem)
			delete(ob.tracker, bestBid.ID)
		}

		if bestAsk.Quantity == 0 {
			askLevel.Orders.Remove(askElem)
			delete(ob.tracker, bestAsk.ID)
		}

		if bidLevel.Orders.Len() == 0 {
			delete(ob.bids, bestBidPrice)
			heap.Pop(ob.bidsHeap)
		}

		if askLevel.Orders.Len() == 0 {
			delete(ob.asks, bestAskPrice)
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

	elem, exists := ob.tracker[orderID]
	if !exists {
		return false
	}

	order := elem.Value.(*Order)

	var levels map[int]*PriceLevel
	if order.Side == Bid {
		levels = ob.bids
	} else {
		levels = ob.asks
	}

	level := levels[order.Price]
	level.Orders.Remove(elem)
	delete(ob.tracker, orderID)

	if level.Orders.Len() == 0 {
		delete(levels, order.Price)

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
	}

	return true
}

func (ob *OrderBook) BidDepth() int {
	count := 0

	ob.mu.RLock()
	for _, level := range ob.bids {
		for e := level.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*Order)
			count += o.Quantity
		}
	}
	ob.mu.RUnlock()

	return count
}

func (ob *OrderBook) AskDepth() int {
	count := 0

	ob.mu.RLock()
	for _, level := range ob.asks {
		for e := level.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*Order)
			count += o.Quantity
		}
	}
	ob.mu.RUnlock()

	return count
}
