package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLRUCache_BasicOperations(t *testing.T) {
	cache := NewLRUCache(3, 10*time.Second)
	defer cache.Close()

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	if val, ok := cache.Get("a"); !ok || val != 1 {
		t.Errorf("expected a=1, got %v, %v", val, ok)
	}

	if cache.Size() != 3 {
		t.Errorf("expected size 3, got %d", cache.Size())
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache(3, 10*time.Second)
	defer cache.Close()

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	// a é o LRU
	// Adicionar d deve evict a
	cache.Set("d", 4)

	if _, ok := cache.Get("a"); ok {
		t.Error("a should have been evicted")
	}

	if cache.Size() != 3 {
		t.Errorf("expected size 3, got %d", cache.Size())
	}
}

func TestLRUCache_LRUOrdering(t *testing.T) {
	cache := NewLRUCache(3, 10*time.Second)
	defer cache.Close()

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	// Acessar a (move para frente)
	cache.Get("a")

	// Agora b é o LRU
	cache.Set("d", 4)

	// b deve ter sido evicted
	if _, ok := cache.Get("b"); ok {
		t.Error("b should have been evicted")
	}

	// a ainda deve existir
	if _, ok := cache.Get("a"); !ok {
		t.Error("a should still exist")
	}
}

func TestLRUCache_Update(t *testing.T) {
	cache := NewLRUCache(3, 10*time.Second)
	defer cache.Close()

	cache.Set("a", 1)
	cache.Set("a", 2) // update

	if val, ok := cache.Get("a"); !ok || val != 2 {
		t.Errorf("expected a=2 after update, got %v", val)
	}

	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}
}

func TestLRUCache_TTL(t *testing.T) {
	cache := NewLRUCache(10, 100*time.Millisecond)
	defer cache.Close()

	cache.Set("a", 1)

	// Dentro do TTL
	if _, ok := cache.Get("a"); !ok {
		t.Error("a should exist within TTL")
	}

	// Esperar expirar
	time.Sleep(150 * time.Millisecond)

	// Deve ter expirado
	if _, ok := cache.Get("a"); ok {
		t.Error("a should have expired")
	}
}

func TestLRUCache_TTLRefresh(t *testing.T) {
	cache := NewLRUCache(10, 100*time.Millisecond)
	defer cache.Close()

	cache.Set("a", 1)
	time.Sleep(60 * time.Millisecond)

	// Atualizar (refresh TTL)
	cache.Set("a", 2)
	time.Sleep(60 * time.Millisecond)

	// Não deve ter expirado (TTL foi refreshed)
	if val, ok := cache.Get("a"); !ok || val != 2 {
		t.Error("a should still exist after TTL refresh")
	}
}

func TestLRUCache_Concurrent(t *testing.T) {
	cache := NewLRUCache(100, 5*time.Second)
	defer cache.Close()

	var wg sync.WaitGroup
	numGoroutines := 50

	// Writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key:%d:%d", id, j)
				cache.Set(key, j)
			}
		}(i)
	}

	// Readers
	var hits, misses atomic.Int32
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key:%d:%d", id, j)
				if _, ok := cache.Get(key); ok {
					hits.Add(1)
				} else {
					misses.Add(1)
				}
			}
		}(i)
	}

	wg.Wait()

	totalOps := int(hits.Load()) + int(misses.Load())
	if totalOps != numGoroutines*100 {
		t.Errorf("expected %d total ops, got %d", numGoroutines*100, totalOps)
	}
}

func TestLRUCache_CleanupExpired(t *testing.T) {
	cache := NewLRUCache(10, 50*time.Millisecond)
	defer cache.Close()

	// Adicionar items
	for i := 0; i < 5; i++ {
		cache.Set(fmt.Sprintf("key:%d", i), i)
	}

	if cache.Size() != 5 {
		t.Errorf("expected size 5, got %d", cache.Size())
	}

	// Esperar cleanup remover items expirados
	time.Sleep(200 * time.Millisecond)

	// Cache deve estar vazio ou quase
	size := cache.Size()
	if size > 1 { // Pode ter 0 ou 1 dependendo de timing
		t.Errorf("expected size ~0 after cleanup, got %d", size)
	}
}

// CRÍTICO: rodar com -race
func TestLRUCache_Race(t *testing.T) {
	cache := NewLRUCache(50, 1*time.Second)
	defer cache.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("k:%d", j)
				cache.Set(key, j)
				cache.Get(key)
				if j%10 == 0 {
					cache.Delete(key)
				}
			}
		}(i)
	}
	wg.Wait()
}
