package main

import (
	"container/list"
	"log/slog"
	"math/rand"
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
	Price    int
	Orders   *list.List
	Side     Side
	TotalQty int // sum of Quantity of all active orders on that level
}

// OrderBook maintains bids and asks for a single ticker
type OrderBook struct {
	Symbol string

	// price -> all bids for the price (people trying to buy on that price)
	bidsIndex *SkipList
	// price -> all asks for the price (people trying to sell on that price)
	asksIndex *SkipList

	tracker map[string]*list.Element
	mu      sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:    symbol,
		bidsIndex: NewSkipList(16, Bid, 0.5, rand.New(rand.NewSource(42))),
		asksIndex: NewSkipList(16, Ask, 0.5, rand.New(rand.NewSource(42))),
		tracker:   make(map[string]*list.Element),
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
		var level *PriceLevel

		if v, ok := ob.bidsIndex.Search(o.Price); ok {
			level = v.(*PriceLevel)
		} else {
			level = &PriceLevel{
				Price:  o.Price,
				Orders: list.New(),
				Side:   Bid,
			}

			ob.bidsIndex.Insert(o.Price, level)
		}

		// push back because its a FIFO
		el := level.Orders.PushBack(o)

		level.TotalQty += o.Quantity

		ob.tracker[o.ID] = el
	case Ask:
		var level *PriceLevel

		if v, ok := ob.asksIndex.Search(o.Price); ok {
			level = v.(*PriceLevel)
		} else {
			level = &PriceLevel{
				Price:  o.Price,
				Orders: list.New(),
				Side:   Ask,
			}

			ob.asksIndex.Insert(o.Price, level)
		}

		// push back because its a FIFO
		el := level.Orders.PushBack(o)

		level.TotalQty += o.Quantity

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
	for {
		bidLevel, hasBid := ob.bidsIndex.Front()
		askLevel, hasAsk := ob.asksIndex.Front()
		if !hasBid || !hasAsk || bidLevel.Price < askLevel.Price {
			break
		}

		bidElem := bidLevel.Orders.Front()
		askElem := askLevel.Orders.Front()

		bestBid := bidElem.Value.(*Order)
		bestAsk := askElem.Value.(*Order)

		orderQuantity := min(bestBid.Quantity, bestAsk.Quantity)
		bestBid.Quantity -= orderQuantity
		bestAsk.Quantity -= orderQuantity

		bidLevel.TotalQty -= orderQuantity
		askLevel.TotalQty -= orderQuantity

		if bestBid.Quantity == 0 {
			bidLevel.Orders.Remove(bidElem)
			delete(ob.tracker, bestBid.ID)
		}

		if bestAsk.Quantity == 0 {
			askLevel.Orders.Remove(askElem)
			delete(ob.tracker, bestAsk.ID)
		}

		if bidLevel.Orders.Len() == 0 {
			ob.bidsIndex.Delete(bidLevel.Price)
		}

		if askLevel.Orders.Len() == 0 {
			ob.asksIndex.Delete(askLevel.Price)
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

	var index *SkipList
	if order.Side == Bid {
		index = ob.bidsIndex
	} else {
		index = ob.asksIndex
	}

	v, ok := index.Search(order.Price)
	if !ok {
		return false
	}

	level := v.(*PriceLevel)
	level.Orders.Remove(elem)
	level.TotalQty -= order.Quantity
	delete(ob.tracker, orderID)

	if level.Orders.Len() == 0 {
		index.Delete(order.Price)
	}

	return true
}

func (ob *OrderBook) BidDepth() int {
	count := 0

	ob.mu.RLock()
	for _, level := range ob.bidsIndex.All() {
		count += level.TotalQty
	}
	ob.mu.RUnlock()

	return count
}

func (ob *OrderBook) AskDepth() int {
	count := 0

	ob.mu.RLock()
	for _, level := range ob.asksIndex.All() {
		count += level.TotalQty
	}
	ob.mu.RUnlock()

	return count
}
