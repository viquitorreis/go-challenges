package main

import (
	"fmt"
)

func main() {
	cache := NewLRUCache(2)
	cache.Put(1, 1)
	cache.Put(2, 2)

	fmt.Println(cache.Get(1)) // returns 1
	cache.Put(3, 3)           // evicts key 2
	fmt.Println(cache.Get(2)) // returns -1 (evicted)
	fmt.Println(cache.Get(3)) // returns 3

}

type Node struct {
	key, val   int
	prev, next *Node
}

type LRUCache struct {
	capacity   int
	cache      map[int]*Node
	head, tail *Node
}

func NewLRUCache(capacity int) LRUCache {
	head := &Node{}
	tail := &Node{}
	head.next = tail
	tail.prev = head

	return LRUCache{
		capacity: capacity,
		cache:    make(map[int]*Node),
		head:     head,
		tail:     tail,
	}
}

func (c *LRUCache) removeNode(node *Node) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (c *LRUCache) insertFront(node *Node) {
	node.next = c.head.next
	node.prev = c.head

	c.head.next.prev = node
	c.head.next = node
}

func (c *LRUCache) Get(key int) int {
	if val, ok := (*c).cache[key]; ok {
		c.removeNode(val)
		c.insertFront(val) // marca como usado recentemente

		return val.val
	}

	return -1
}

func (c *LRUCache) Put(key, value int) {
	if val, ok := (*c).cache[key]; ok {
		val.val = value
		c.removeNode(val)
		c.insertFront(val) // marca como usado recentemente
	} else {
		if len(c.cache) >= c.capacity {
			lru := c.tail.prev // ultimo node real
			c.removeNode(lru)
			delete(c.cache, lru.key)
		}

		node := &Node{
			key: key,
			val: value,
		}

		c.insertFront(node)
		c.cache[node.key] = node
	}
}
