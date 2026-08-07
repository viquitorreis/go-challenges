package main

import "fmt"

func main() {
	ob := NewOrderBook("BTC-USD")

	ob.AddOrder(NewOrder("o1", "alice", Bid, 102, 5))
	ob.AddOrder(NewOrder("o2", "bob", Ask, 103, 3))
	ob.AddOrder(NewOrder("o3", "carol", Bid, 103, 2)) // must match o2

	trades := ob.Match()
	for _, t := range trades {
		fmt.Printf("TRADE: %s x %s @ %d qty %d\n",
			t.BidOrderID, t.AskOrderID, t.Price, t.Quantity)
	}

	ob.Cancel("o1")
	// fmt.Printf("Book depth — bids: %d asks: %d\n",
	// 	ob.BidDepth(), ob.AskDepth())
}
