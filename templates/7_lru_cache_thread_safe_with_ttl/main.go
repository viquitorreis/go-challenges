package main

import (
	"fmt"
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
	// TODO: adicionar ponteiros prev e next para doubly linked list
}

// LRUCache é um cache thread-safe com política LRU e TTL
type LRUCache struct {
	// TODO: adicionar campos necessários
	// - capacity (max items)
	// - map[string]*CacheItem para O(1) lookup
	// - head e tail da linked list
	// - mutex para thread safety
	// - ttl (time to live)
	// - channel para shutdown do cleanup
}

// NewLRUCache cria novo cache
// capacity: número máximo de items
// ttl: quanto tempo items vivem antes de expirar
func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	// TODO: implementar
	// Inicializar map, linked list vazia (head/tail nil)
	// Iniciar goroutine de cleanup
	return nil
}

// Get busca valor por key
// Retorna (value, true) se encontrou e não expirou
// Retorna (nil, false) se não encontrou ou expirou
func (c *LRUCache) Get(key string) (any, bool) {
	// TODO: implementar
	// 1. Lock mutex
	// 2. Verificar se key existe no map
	// 3. Verificar se expirou (time.Since(timestamp) > ttl)
	// 4. Se expirou: remover do cache, retornar false
	// 5. Se válido: mover para frente da lista, retornar value
	// 6. Unlock mutex
	return nil, false
}

// Set adiciona ou atualiza item no cache
func (c *LRUCache) Set(key string, value any) {
	// TODO: implementar
	// 1. Lock mutex
	// 2. Se key já existe: atualizar valor, mover para frente, atualizar timestamp
	// 3. Se não existe:
	//    a. Se cache cheio: remover item do final (LRU)
	//    b. Criar novo node
	//    c. Adicionar no map e na frente da lista
	//    d. Atualizar timestamp
	// 4. Unlock mutex
}

// Delete remove item do cache
func (c *LRUCache) Delete(key string) {
	// TODO: implementar
	// Lock, remover do map e da lista, unlock
}

// Size retorna número de items no cache
func (c *LRUCache) Size() int {
	// TODO: implementar com lock
	return 0
}

// addToFront adiciona node no início da lista
func (c *LRUCache) addToFront(node *CacheItem) {
	// TODO: implementar
	// Atualizar ponteiros head/tail
}

// remove retira node da lista
func (c *LRUCache) remove(node *CacheItem) {
	// TODO: implementar
	// Atualizar ponteiros dos vizinhos e head/tail se necessário
}

// moveToFront move node existente para início
func (c *LRUCache) moveToFront(node *CacheItem) {
	// TODO: implementar
	// Remover da posição atual e adicionar na frente
}

// removeLRU remove item menos recentemente usado (final da lista)
func (c *LRUCache) removeLRU() {
	// TODO: implementar
	// Remover tail, atualizar map
}

// cleanup goroutine que remove items expirados periodicamente
func (c *LRUCache) cleanup() {
	// TODO: implementar
	// ticker := time.NewTicker(interval)
	// Loop: a cada tick, varrer lista do final
	// Remover items expirados
	// Parar quando receber sinal de shutdown
}

// Close para cleanup goroutine
func (c *LRUCache) Close() {
	// TODO: implementar
	// Enviar sinal para shutdown channel
}
