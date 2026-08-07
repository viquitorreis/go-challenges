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
	assert.Equal(t, 2, ob.BidDepth()) // 5 - 3 = 2 left
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
	assert.False(t, ob.Cancel("b1")) // already removed
}

func TestPriceTimePriority(t *testing.T) {
	ob := NewOrderBook("BTC-USD")
	// both bids have the same price, but b1 arrived first
	ob.AddOrder(NewOrder("b1", "alice", Bid, 103, 2))
	ob.AddOrder(NewOrder("b2", "bob", Bid, 103, 2))
	ob.AddOrder(NewOrder("a1", "carol", Ask, 103, 2))

	trades := ob.Match()
	assert.Len(t, trades, 1)
	assert.Equal(t, "b1", trades[0].BidOrderID) // b1 have priority over b2
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

// BenchmarkCancelMiddleOfLevel bench test is same scanario from the old (heap)
// with direct comparison. Here the expectation is O(1) per Cancel, regardless of
// N, since the doubly linked list allows removing the node from the middle
// without scanning the rest of the queue.
func BenchmarkCancelMiddleOfLevel(b *testing.B) {
	sizes := []int{10, 100, 1_000, 10_000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("depth_%d", n), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ob := NewOrderBook("BTC-USD")
				ids := make([]string, n)
				for j := 0; j < n; j++ {
					id := fmt.Sprintf("o%d", j)
					ids[j] = id
					ob.AddOrder(NewOrder(id, "user", Bid, 100, 1))
				}
				middleID := ids[n/2]
				b.StartTimer()

				ob.Cancel(middleID)
			}
		})
	}
}

// BenchmarkCancelManyDistinctLevels bench test is same scanario from the old (heap)
// to direct comparison. Here the expectation is O(log n) per Cancel, not O(n log n),
//
//	since the skip list removes a price level pointwise without reconstructing the entire structure.
func BenchmarkCancelManyDistinctLevels(b *testing.B) {
	sizes := []int{10, 100, 1_000, 10_000}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("levels_%d", n), func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				b.StopTimer()
				ob := NewOrderBook("BTC-USD")
				ids := make([]string, n)
				for j := 0; j < n; j++ {
					id := fmt.Sprintf("o%d", j)
					ids[j] = id
					ob.AddOrder(NewOrder(id, "user", Bid, 100+j, 1))
				}
				b.StartTimer()

				for _, id := range ids {
					ob.Cancel(id)
				}
			}
		})
	}
}
