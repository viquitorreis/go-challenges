# Skip List

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Data Structures / Concurrency
**Estimated time:** ~1.5 hours

## What it is

A thread-safe probabilistic skip list, the structure behind Redis sorted sets (`ZADD`, `ZRANGE`, `ZRANK`) and the storage engines of LevelDB/RocksDB: multiple stacked linked lists where each node randomly "levels up" to give O(log n) amortized search, insert, and delete.

## What you'll learn

- Why probabilistic balancing (coin-flip level assignment) is simpler to implement and to make concurrent than a deterministically balanced BST.
- The **update array pattern**: walking every level top to bottom, recording the last node "to the left" of the insertion point at each level, then rewiring pointers using those predecessors.
- Horizontal locking (across list levels) as a different concurrency shape than the trie's vertical (per-node, parent-to-child) locking.

## What's implemented

- `NewSkipList(maxLevel int, p float64, rng *rand.Rand) *SkipList`.
- `Insert(score int, value any)`, `Search(score int) (any, bool)`, `Delete(score int) bool`.
- `RangeSearch(min, max int) []any` and `Size() int`.
- `randomLevel()` implementing the coin-flip level assignment with probability `p`.
- Tests cover insert/search, updating an existing score, delete, range search (including an empty list), insertion order, and concurrent inserts/reads-and-writes/deletes.

## Design decisions

- The update array is computed once per insert/delete by scanning from the top level down, avoiding repeated full-list traversals per level.
- `p = 0.5` is used for level probability, the standard choice that keeps roughly half the nodes at each successive level.

## How to run

```bash
make execute
# or
go run .
go test ./...
```
