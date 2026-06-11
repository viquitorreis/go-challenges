##Problem: LRU Cache

Implement a Least Recently Used (LRU) cache with the following operations:

- Get(key int) int - return the value if key exists, otherwise -1
- Put(key, value int) - insert or update the key. If capacity is exceeded, evict the least recently used key.

Both operations must run in O(1).

```go
type LRUCache struct {
    // your fields
}

func NewLRUCache(capacity int) LRUCache {
    // your code
}

func (c *LRUCache) Get(key int) int {
    // your code
}

func (c *LRUCache) Put(key, value int) {
    // your code
}

cache := NewLRUCache(2)
cache.Put(1, 1)
cache.Put(2, 2)
cache.Get(1)    // returns 1
cache.Put(3, 3) // evicts key 2
cache.Get(2)    // returns -1 (evicted)
cache.Get(3)    // returns 3
```

Hint if you need it: O(1) for both ops requires two data structures working together. Think about what each one gives you.