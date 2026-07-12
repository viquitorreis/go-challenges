# LRU Cache with TTL

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Data Structures / Concurrency
**Estimated time:** ~1.5 hours

## What it is

A production-style in-memory cache combining LRU (least recently used) eviction with per-item TTL (time to live) expiration, safe for concurrent reads and writes.

## What you'll learn

- Combining a hashmap with a doubly linked list to get O(1) `Get`/`Set`/eviction, the classic LRU implementation.
- The trade-off between lazy TTL expiration (checked on access) and active cleanup (a background goroutine sweeping periodically) - and why this implementation does both.
- Choosing between a single `sync.Mutex`, `sync.RWMutex`, or sharding for a cache under concurrent load.

## What's implemented

- `NewLRUCache(capacity int, ttl time.Duration) *LRUCache`.
- `Get(key string) (any, bool)` - returns the value if present and not expired, and moves it to the front (most recently used).
- `Set(key string, value any)` - inserts or updates, evicting the least recently used item when capacity is exceeded.
- `Delete(key string)` and `Size() int`.
- Internal doubly linked list helpers: `addToFront`, `remove`, `moveToFront`.
- `cleanup()` running in the background to actively purge expired items, plus `Close()` to stop it.
- Tests cover basic operations, eviction, LRU ordering, updates, TTL expiry, TTL refresh on access, concurrent access, background cleanup, and `-race`.

## Design decisions

- The linked list tracks recency order in O(1) per operation; the hashmap gives O(1) key lookup - neither alone is enough for true O(1) LRU.
- TTL cleanup is both lazy (checked on `Get`) and active (background goroutine), so expired entries don't linger in memory indefinitely even if nobody reads them.

## How to run

```bash
go run .
go test -race ./...
```
