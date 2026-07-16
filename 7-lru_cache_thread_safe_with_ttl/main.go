package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

/*
TODO - PASSOS PARA IMPLEMENTAR LRU CACHE COM TTL

1. CRIAR A DOUBLY LINKED LIST
   - Node com ponteiros para prev e next
   - Métodos: addToFront, remove, moveToFront
   - Manter ponteiros head e tail atualizados

2. CRIAR O CACHE
   - HashMap (map) para O(1) lookup: key -> node
   - Linked list para rastrear ordem de uso
   - Fields: capacity, size, mutex

3. IMPLEMENTAR Get()
   - Verificar se key existe no map
   - Verificar se expirou (comparar timestamp com TTL)
   - Se válido: mover node para frente da lista, retornar valor
   - Se expirado: remover do cache, retornar miss

4. IMPLEMENTAR Set()
   - Se key já existe: atualizar valor, mover para frente, atualizar timestamp
   - Se não existe: criar node, adicionar no map e na frente da lista
   - Se cache cheio: remover node do final da lista antes de adicionar novo
   - Sempre atualizar timestamp com time.Now()

5. IMPLEMENTAR CLEANUP AUTOMÁTICO
   - Goroutine em background roda periodicamente
   - Varre lista do final (itens mais antigos)
   - Remove itens expirados
   - Para quando encontrar item não expirado (otimização)

6. GRACEFUL SHUTDOWN
   - Sinalizar cleanup goroutine para parar
   - Esperar goroutine terminar
*/

func main() {
	fmt.Println("=== LRU Cache com TTL ===")

	// Cache: max 5 items, TTL de 2 segundos
	cache := NewLRUCache(5, 2*time.Second)
	defer cache.Close()

	// Adicionar alguns items
	fmt.Println("Adding items...")
	cache.Set("user:1", "Alice")
	cache.Set("user:2", "Bob")
	cache.Set("user:3", "Charlie")
	cache.Set("user:4", "Diana")
	cache.Set("user:5", "Eve")

	// Tentar get
	if val, ok := cache.Get("user:1"); ok {
		fmt.Printf("Got user:1 = %v\n", val)
	}

	// Adicionar 6º item (deve evict LRU)
	fmt.Println("\nAdding 6th item (cache full, should evict LRU)...")
	cache.Set("user:6", "Frank")

	// user:2 deve ter sido evicted (era o LRU, user:1 foi acessado)
	if _, ok := cache.Get("user:2"); !ok {
		fmt.Println("user:2 was evicted (LRU)")
	}

	// Testar TTL
	fmt.Println("\nWaiting for TTL expiration (2.5s)...")
	time.Sleep(2500 * time.Millisecond)

	// Todos os items devem ter expirado
	if _, ok := cache.Get("user:1"); !ok {
		fmt.Println("user:1 expired (TTL)")
	}

	// Adicionar novos items
	cache.Set("user:7", "Grace")
	cache.Set("user:8", "Henry")

	fmt.Printf("\nFinal cache size: %d\n", cache.Size())
	fmt.Println("Done!")
}

// CacheItem representa um item no cache
type CacheItem struct {
	key       string
	value     any
	timestamp time.Time
	prev      *CacheItem
	next      *CacheItem
}

type LRUCache struct {
	capacity int
	cache    map[string]*CacheItem
	ttl      time.Duration
	shutdown chan struct{}

	head *CacheItem
	tail *CacheItem

	mu sync.Mutex
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	lruCache := &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*CacheItem),
		ttl:      ttl,
		shutdown: make(chan struct{}),
	}

	go lruCache.cleanup()

	return lruCache
}

func (c *LRUCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	node, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	if time.Since(node.timestamp) > c.ttl {
		c.remove(node)
		delete(c.cache, key)
		return nil, false
	}

	c.moveToFront(node)

	return node.value, true
}

func (c *LRUCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, ok := c.cache[key]; ok {
		node.value = value
		node.timestamp = time.Now()
		c.moveToFront(node)
		return
	}

	node := &CacheItem{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}

	if len(c.cache) >= c.capacity {
		delete(c.cache, c.tail.key)
		c.remove(c.tail)
	}

	c.cache[key] = node
	c.addToFront(node)
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if val, ok := c.cache[key]; ok {
		if val == c.head {
			c.head = c.head.next
		}

		if val == c.tail {
			c.tail = c.tail.prev
		}

		delete(c.cache, key)
		c.remove(val)
	}
}

func (c *LRUCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cache)
}

func (c *LRUCache) addToFront(node *CacheItem) {
	node.prev = nil
	node.next = c.head // prioximo é o head atual

	if c.head != nil {
		c.head.prev = node // head antigo aponta para trás para novo node
	} else {
		c.tail = node // lista estava vazia entao o node tambem vira a tail
	}

	c.head = node
}

func (c *LRUCache) remove(node *CacheItem) {
	if node.prev != nil {
		node.prev.next = node.next // aponta o proximo do anterior para o proximo do node atual
	} else {
		c.head = node.next // se nao tem previous era head, então o head vai ser so o proximo
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		c.tail = node.prev // node era tail
	}

	node.prev = nil
	node.next = nil
}

func (c *LRUCache) moveToFront(node *CacheItem) {
	if node == c.head {
		return
	}

	c.remove(node)
	c.addToFront(node)
}

func (c *LRUCache) cleanup() {
	ticker := time.NewTicker(time.Duration(100 * time.Millisecond))
	defer ticker.Stop()

	for {
		select {
		case tik := <-ticker.C:
			log.Println("cleanup: ", tik)
			c.mu.Lock()
			curr := c.tail

			for curr != nil {
				if time.Since(curr.timestamp) > c.ttl {
					toRemove := curr
					curr = curr.prev
					c.remove(toRemove)
					delete(c.cache, toRemove.key)
				} else {
					break
				}
			}

			c.mu.Unlock()

		case <-c.shutdown:
			return
		}
	}

}

// Close para cleanup goroutine
func (c *LRUCache) Close() {
	// TODO: implementar
	// Enviar sinal para shutdown channel
	c.shutdown <- struct{}{}
}
