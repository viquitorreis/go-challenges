# Thread-Safe Trie

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Data Structures / Concurrency
**Estimated time:** ~1 hour

## What it is

An autocomplete-style trie (like the kind behind GitHub's repository search-as-you-type) that supports concurrent inserts, exact-match search, prefix search, and autocomplete, without a single global lock turning every operation into a bottleneck.

## What you'll learn

- Fine-grained, per-node ("hand-over-hand") locking: acquiring and releasing each node's mutex as you descend the tree, instead of locking the whole structure.
- Why this matters for high-throughput concurrent prefix lookups (the same problem HTTP routers like Gin/Echo, DNS servers, and IP routing tables solve).
- Snapshotting a node's children under lock before recursing, so traversal doesn't hold a parent's lock while walking its subtree.

## What's implemented

- `NewTrie() *Trie`.
- `Insert(word string)`, `Search(word string) bool`, `StartsWith(prefix string) bool`.
- `AutoComplete(prefix string, limit int) []string`, returning up to `limit` matches (or all of them when `limit` is 0).
- Tests cover basic insert/search, prefix search, autocomplete (with and without a limit), non-existent prefixes, an empty trie, single-character words, overlapping words, concurrent inserts/reads/mixed operations, unicode characters, a larger dataset, and benchmarks for insert/search/autocomplete/concurrent inserts.

## Design decisions

- Each `TrieNode` holds its own `sync.RWMutex`, so lookups on unrelated prefixes (e.g. "react" vs. "golang") never contend for the same lock.
- `collectWords` snapshots a node's children under a brief lock before recursing into each child, rather than holding the parent lock for the whole recursive walk.

## How to run

```bash
go run .
go test ./...
go test -bench=. ./...
```
