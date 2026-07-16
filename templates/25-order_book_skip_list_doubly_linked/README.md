# 25 — Order Book Rewrite: Skip List + Doubly Linked Cancellation

**Category:** Data Structures & Systems Design
**Time:** 3h+ (can be split into 2 sessions, see "Suggested sequence" below)
**Builds on:** the original heap-based order book (challenge 1). If you haven't done
that one yet, start there. This challenge assumes you already have a working matching
engine and are rewriting its internal data structures, not its behavior.

## How to work through this challenge

1. Take the original order book implementation (challenge 1, heap plus
   `map[price][]*Order`)
2. Copy it into a new folder. Keep the original untouched as a behavioral reference,
   you'll use it to confirm the rewrite doesn't change observable results, only
   performance and latency characteristics
3. Rewrite the internal structure following the requirements below
4. At the end, both versions must produce the exact same results for the same
   inputs. The rewrite changes the data structures, not the business logic

## Problem overview

An order book built on `container/heap` plus a `map[price][]*Order` has two
structural weaknesses that only show up under real volume:

1. **A heap doesn't support arbitrary removal.** It only knows how to pop the top.
   Cancelling the last order at a price level forces a full heap rebuild
   (O(n log n)), or requires tolerating stale entries that need to be filtered out
   later.
2. **A slice doesn't support O(1) removal from the middle.** Cancelling a specific
   order inside a price level, even when that level doesn't empty out, requires
   scanning the whole slice to rebuild it without that element: O(level size),
   every single time. This matters a lot in production matching engines, where
   cancels typically make up 70 to 80 percent of order book traffic.

Since cancellation is usually the most frequent operation in a real order book
(more common than fills), these two points tend to dominate system cost under load,
even when the matching logic itself is fast.

## Required work

### Track A: price level structure
- Replace the heap plus map with a **skip list** ordered by price (bids in
  descending order, best price first; asks in ascending order)
- Insertion, removal, and best-price lookup must be O(log n) expected
- The skip list needs to support **ordered iteration from the top** (required for
  any future market depth feature, see Bonus)

### Track B: internal structure of each price level
- Each `PriceLevel` stops holding `Orders []*Order` and instead holds a
  **doubly linked list** (Go's `container/list`, or your own implementation)
- The order tracker (ID to order lookup) stops storing just `*Order` and also
  stores the list node (`*list.Element` or equivalent), allowing a specific order
  to be removed in **O(1)**, regardless of its position within the level

### Behavior preservation
- `AddOrder`, `Match`, `Cancel`, `BidDepth`, `AskDepth` keep the same public
  signatures and the same observable behavior as the original order book
- Price-time priority (FIFO within each price level) must remain correct. The
  linked list preserves insertion order naturally, just confirm your
  implementation doesn't invert this anywhere
- Thread-safety is preserved (same level of protection against concurrent access
  as the original)

## Design decisions: record your choices and why

Unlike most earlier challenges, this one doesn't have a single "correct" answer
for some structural decisions. Document these in your solution's README:

**1. Reuse an existing skip list vs. write a new one.** If you already have a
generic skip list from another challenge, does it support a custom comparator
(ascending for asks, descending for bids)? Is it worth adapting, or is a leaner
purpose-built version better here?

**2. Layered locking.** If your skip list has its own internal lock (thread-safe
by itself) and the `OrderBook` also holds a mutex protecting everything, you have
two stacked locks per operation, which is redundant and a potential source of
unnecessary contention. Decision to record: skip list with an internal lock
(safer in isolation, slower combined), or a lock-free skip list relying entirely
on the `OrderBook`'s mutex (faster, but only correct if **every** access truly
goes through the `OrderBook`)?

**3. Aggregated quantity per level.** Is it worth having `PriceLevel` maintain an
incremental `TotalQty` (added to and subtracted from on every insert, cancel, and
match), turning `BidDepth`/`AskDepth` from "sum every individual order" into "sum
the handful of price levels"? This doesn't change the public signature, only the
internal cost of the calculation, a straightforward implementation-cost-versus-
performance tradeoff.

## Reference comparison: Skip List vs. Indexed Heap

If you decide to go with an indexed heap instead of a skip list (a valid variation
of this challenge, adjust the requirements accordingly):

| Criterion | Skip List | Indexed Heap |
|---|---|---|
| Best price (peek) | O(1) | O(1) |
| Insert/remove level (with a reference in hand) | O(log n) expected | O(log n) guaranteed |
| Top-N ordered levels (book depth) | O(k), natural | Poor, requires destructive pops or an O(n log n) sort |
| Worst-case guarantee | Probabilistic | Deterministic |
| Memory overhead | Higher (multiple pointers per node) | Lower |
| Implementation bug risk | Low-medium | Medium-high (the index map must stay perfectly in sync on every swap) |

The factor that matters most for most cases: if your order book will eventually
need to publish market depth (top-N levels, not just the best price), a common
requirement in any real trading system, a skip list pays for itself by supporting
this natively. An indexed heap is competitive if the only operation that matters
is always the isolated best price.

## Bonus (if time allows)

- A `TopLevels(n int) []PriceLevel` method: uses the skip list's ordered
  iteration from Track A to return the N best levels on each side, essentially
  for free
- A benchmark comparing the original implementation (heap) against the new one
  (skip list), specifically in a scenario with **many interleaved cancellations
  and few new orders**, since that's the scenario that most exposes the
  difference
- `go test -race` on both versions, to confirm the rewrite didn't introduce any
  new race condition

## What will be evaluated

1. Whether Track A and Track B were solved **independently and testable** each on
   their own, or ended up coupled in a way that makes it hard to tell which one
   introduced a bug, if one shows up
2. Whether observable behavior (match results, depth, execution order) remains
   identical to the original order book across the same test scenarios
3. Whether the design decisions (locking, custom vs. reused structure, quantity
   aggregation) were made consciously and documented, not just "whatever compiled
   first"

## Suggested sequence (if split into 2 sessions, ~1h each)

**Session 1: Track B in isolation.** Swap only the internal structure of each
`PriceLevel` (slice to linked list), keeping the heap and map exactly as in the
original for now. Testable on its own: cancelling an order in the middle of a
level is already O(1), without touching anything else.

**Session 2: Track A on top of the already-stable Session 1.** With `PriceLevel`
already working, swap the heap and map for the skip list. The change stays
isolated to the "how do I find/insert/remove a price level" layer, without mixing
in the level's internal logic, which was already validated in the previous
session.

At the end: run the same example scenario from the original order book's
reference `main.go` (including at least one match and one cancel) against both
versions and confirm the result is identical.