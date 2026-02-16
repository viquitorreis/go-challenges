package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddAndMatch(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	ob.AddOrder(NewOrder("b1", "alice", Bid, 103, 5))
	ob.AddOrder(NewOrder("a1", "bob", Ask, 103, 3))

	trades := ob.Match()
	assert.Len(t, trades, 1)
	assert.Equal(t, 3, trades[0].Quantity)
	assert.Equal(t, 2, ob.BidDepth()) // 5 - 3 = 2 restantes
	assert.Equal(t, 0, ob.AskDepth())
}

func TestNoMatch(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	ob.AddOrder(NewOrder("b1", "alice", Bid, 100, 5))
	ob.AddOrder(NewOrder("a1", "bob", Ask, 105, 3))

	trades := ob.Match()
	assert.Empty(t, trades)
}

func TestCancel(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	ob.AddOrder(NewOrder("b1", "alice", Bid, 100, 5))
	assert.True(t, ob.Cancel("b1"))
	assert.Equal(t, 0, ob.BidDepth())
	assert.False(t, ob.Cancel("b1")) // já removida
}

func TestPriceTimePriority(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	// dois bids no mesmo preço — b1 chegou primeiro
	ob.AddOrder(NewOrder("b1", "alice", Bid, 103, 2))
	ob.AddOrder(NewOrder("b2", "bob", Bid, 103, 2))
	ob.AddOrder(NewOrder("a1", "carol", Ask, 103, 2))

	trades := ob.Match()
	assert.Len(t, trades, 1)
	assert.Equal(t, "b1", trades[0].BidOrderID) // b1 tem prioridade
}

func TestConcurrentAddOrders(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ob.AddOrder(NewOrder(
				fmt.Sprintf("b%d", n), "user", Bid, 100+n, 1,
			))
		}(i)
	}
	wg.Wait()
	assert.Equal(t, 100, ob.BidDepth())
}
