# Exchange Order Book

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Data Structures
**Estimated time:** ~1.5 hours

## What it is

A price-time priority matching engine for a single ticker: bids and asks are kept in separate heaps, and orders match whenever the best bid price is at least as high as the best ask price.

## What you'll learn

- Modeling an order book with two heaps: a max-heap for bids, a min-heap for asks, both ordered by price and, for ties, by arrival time (price-time priority).
- The matching condition `max(bids) >= min(asks)` and how partial fills work when order quantities don't match exactly.
- Implementing `heap.Interface` twice for two different orderings over structurally similar data.

## What's implemented

- `bidsHeap` and `asksHeap`, both implementing `heap.Interface` with opposite price ordering.
- `NewOrderBook(symbol string) *OrderBook`, `NewOrder(id, userID string, side Side, price, qty int) *Order`.
- `AddOrder(o *Order)` inserting into the correct side's heap.
- `Match() []Trade` executing all currently possible matches and returning the resulting trades.
- `Cancel(orderID string) bool`, `BidDepth() int`, `AskDepth() int`.
- Tests cover adding and matching, no-match scenarios, cancellation, price-time priority ordering, and concurrent order submission.

## Design decisions

- Bids and asks use two separate heap types instead of one generic heap with a runtime side check, keeping the ordering comparator simple and correct for each side.
- Matching is a separate, explicit `Match()` call rather than matching inline on every `AddOrder`, which keeps the insertion path cheap and makes matching behavior easy to test in isolation.

## How to run

```bash
make execute
# or
go run .
go test ./...
```
