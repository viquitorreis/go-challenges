# Exchange Order Book

**.:. Note**: A better version of this project was made on challenge 25, and evolved on a separate project. See [Distributed Matching Engine](https://github.com/viquitorreis/distributed_matching_engine) for full details of the implementation.

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

## Benchmark: Cancel Performance

Measured against the [rewritten version using a skip list and doubly linked
list](https://github.com/viquitorreis/distributed-matching-engine), which documents
the full analysis.

Run with `go test -bench=. -benchmem -benchtime=100x`.

**Note**: benchmarks were performed on a amd64 12th Gen Intel(R) Core(TM) i5-1235U 32GB RAM machine

**Cancelling a single order in the middle of a large price level:**

| Level depth | ns/op | B/op | allocs/op |
|---|---|---|---|
| 10 | 1,342 | 192 | 2 |
| 100 | 3,806 | 2,112 | 5 |
| 1,000 | 21,912 | 17,472 | 8 |
| 10,000 | 127,790 | 310,336 | 15 |

**Cancelling N orders, each in its own price level (each cancellation empties
a level):**

| Distinct levels | ns/op (total for N cancels) | B/op | allocs/op |
|---|---|---|---|
| 10 | 4,049 | 1,120 | 40 |
| 100 | 153,935 | 116,994 | 765 |
| 1,000 | 8,683,372 | 11,651,964 | 11,117 |
| 10,000 | 978,893,452 | 1,570,093,820 | 167,667 |

At 10,000 distinct price levels, cancelling all of them takes close to 1
second with this heap-based implementation, since every cancellation that
empties a level triggers a full heap rebuild. See the rewritten version for
the fix and the full breakdown.