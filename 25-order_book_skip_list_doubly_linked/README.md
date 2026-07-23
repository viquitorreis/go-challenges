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

## Benchmark: Cancel Performance vs. the Original Heap-Based Implementation

Compared against the [original heap + slice implementation](https://github.com/viquitorreis/go-challenges/tree/main/11-exchange_order_book),
using the same two scenarios, same sizes, same methodology
(`go test -bench=. -benchmem -benchtime=100x`).

**Note**: benchmarks were performed on a amd64 12th Gen Intel(R) Core(TM) i5-1235U 32GB RAM machine

### Scenario A: cancelling a single order in the middle of a large price level

This targets the doubly linked list rewrite directly: a slice can't remove a
middle element without scanning and rebuilding it, a linked list can, given a
pointer to the node.

| Level depth | Old (ns/op) | New (ns/op) | Speedup |
|---|---|---|---|
| 10 | 1,342 | 179.8 | 7.5x |
| 100 | 3,806 | 639.5 | 6.0x |
| 1,000 | 21,912 | 712.5 | 30.8x |
| 10,000 | 127,790 | 1,112 | **114.9x** |

The old implementation grows close to linearly with level depth, matching
the expected O(n) cost of rebuilding the slice on every cancellation. The
new one stays close to flat. The small growth it does show (180ns to
1,112ns) isn't algorithmic: `list.Remove` is legit O(1) operation regardless
of list size. It's more likely that is GC overhead correlating with
total live heap size at larger scales, not to the cancel operation itself
doing more work. I would like to address this explicitly since it's a common source of
confusion when reading wall clock benchmarks: flat algorithmic complexity
doesn't always mean flat wall clock time once GC enters the picture.

Allocations confirm the same story: the old implementation allocates more per
cancel as the level grows (2 to 15 allocs/op, from `append` reallocating the
rebuilt slice), while the new implementation allocates **zero** per cancel at
every size, since `list.Remove` never allocates.

### Scenario B: cancelling N orders, each in its own price level

This targets the skip list rewrite directly: a heap can't remove an arbitrary
price level without a full rebuild, a skip list can, in O(log n).

Raw numbers represent the total cost of cancelling all N orders in one pass:

| Distinct levels | Old total (ns) | New total (ns) | Speedup |
|---|---|---|---|
| 10 | 4,049 | 3,280 | 1.2x |
| 100 | 153,935 | 33,494 | 4.6x |
| 1,000 | 8,683,372 | 309,355 | 28.1x |
| 10,000 | 978,893,452 | 4,054,022 | **241.4x** |

Normalized to cost per individual cancel (total ÷ N), the difference in
growth pattern is clearer:

| Distinct levels | Old (ns/cancel) | New (ns/cancel) |
|---|---|---|
| 10 | 404.9 | 328.0 |
| 100 | 1,539.4 | 334.9 |
| 1,000 | 8,683.4 | 309.4 |
| 10,000 | 97,889.3 | 405.4 |

The old implementation's per-cancel cost grows roughly 240x from N=10 to
N=10,000, consistent with each cancellation triggering an O(n) heap rebuild
(so the total cost of cancelling N distinct levels approaches O(n²)). The new
implementation's per-cancel cost stays essentially flat (~300-400ns)
regardless of N, consistent with O(log n) per deletion, where the log factor
is negligible at these scales.

At 10,000 price levels, cancelling all of them takes about 979ms with the
old implementation and about 4ms with the new one. That gap widens, not
narrows, as N grows further, since the two implementations are on different
complexity classes (O(n²) vs. O(n log n) for cancelling all N levels), not
just different constants.

Allocations tell a matching story: the old implementation's allocations per
cancel grow with N (4 to 16.8 allocs/cancel), from the heap's backing array
being reallocated on rebuild. The new implementation holds constant at
**exactly 1 allocation per cancel**, regardless of N, coming from the fixed-size
`predecessors` slice (`make([]*SkipListNode, maxLevel)`) allocated on every
`Delete` call, independent of how many elements are actually in the skip
list.

### Takeaway

Both rewritten tracks hold up under measurement, not just under theoretical
complexity analysis. Track B (doubly linked list) delivers consistent,
allocation-free O(1) cancellation regardless of price level depth. Track A
(skip list) turns what was an increasingly expensive O(n²) total cost for
churny order books (many price levels, frequent cancellations) into a
near-linear O(n log n) cost, a 241x improvement at 10,000 price levels that
keeps growing with scale.

## Load Test Stage 1: Single Node implementation, Concurrency Under Contention

Simulates a realistic order flow (70% cancel, 20% add, 10% match) across
increasing levels of concurrent goroutines, measuring throughput and
per-operation latency.

Run with `go test -run TestLoad -v ./...`.

| Workers | Total ops | Throughput | p50 | p99 |
|---|---|---|---|---|
| 1 | 500 | 2,301,867 ops/s | 248ns | 2.5µs |
| 10 | 5,000 | 1,207,425 ops/s | 384ns | 134µs |
| 50 | 25,000 | 1,055,594 ops/s | 476ns | 1.1ms |
| 100 | 50,000 | 1,137,860 ops/s | 438ns | 2.2ms |

### What's actually happening: lock contention

The `OrderBook` uses a single `sync.RWMutex` guarding `AddOrder`, `Cancel`,
and `Match`. Every operation, no matter how small, has to wait its turn to
acquire that lock before doing any real work.

With 1 worker, there's no one to wait for, so latency stays low and
consistent. As soon as more goroutines compete for the same lock, two
things happen at once:

- **Throughput drops immediately** (1 -> 10 workers: from 2.3M to 1.2M
  ops/s), because CPU time that used to go toward useful work now goes
  toward goroutines sitting idle, waiting for the lock
- **p99 grows far faster than p50** (2.5µs -> 2.2ms, roughly 900x, while
  p50 only moves from 248ns to 438ns). Most operations still complete
  quickly once they get the lock. But a growing share of them get stuck
  waiting behind everyone else in line, and those are the ones that blow
  up the tail latency

This is the general shape lock contention takes: the median barely moves,
while the tail gets dramatically worse. A dashboard showing only average
latency would completely miss this.

### Why throughput flattens instead of climbing

Past 10 workers, throughput stays roughly flat (~1.0-1.1M ops/s) instead
of continuing to fall or rise. That's the mutex fully saturated: the
system is already spending as much time serializing access to the lock as
it's going to. Adding more goroutines beyond that point doesn't help or
hurt much, it just makes the queue behind the lock longer, which is
exactly what the growing p99 shows.

### Something worth being honest about

`p50=248ns` for an operation that includes a skip list lookup and a linked
list mutation looks suspiciously fast. Likely explanation: with 70% of
operations being `Cancel` on a randomly picked ID out of only 1,000 seeded
orders, a meaningful share of those calls hit `if !exists { return false }`
immediately, without doing real work, because another goroutine already
cancelled that ID first. That doesn't invalidate the contention pattern
(the lock queue is real either way), but it does mean the raw throughput
number is somewhat inflated. A more accurate benchmark would guarantee
each cancel targets an order that's actually still live.

### Takeaway

The data structure rewrite (skip list + doubly linked list) did its job:
individual operations are fast. The current ceiling isn't algorithmic
complexity anymore, it's the single mutex serializing all access. That's
the concrete argument for the planned move toward either sharding the
lock (e.g. per price-level locks) or an actor-style design (a single
goroutine owning book state, everyone else talking to it over channels,
no lock at all) as this evolves into the distributed matching engine.