package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	// Criar rate limiter: 10 tokens no máximo, adiciona 5 tokens a cada 100ms
	// Isso dá ~50 requests/segundo no steady state
	limiter := NewTokenBucket(10, 5, 100*time.Millisecond)

	fmt.Println("Starting rate limiter test...")
	fmt.Printf("Bucket: capacity=%d, refill rate=%d tokens per 100ms\n", 10, 5)

	accepts := 0
	rejecteds := 0

	var wg sync.WaitGroup

	for range 50 {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(25)))
		wg.Go(func() {
			if limiter.Allow() {
				accepts++
			} else {
				rejecteds++
			}
		})
	}

	wg.Wait()

	fmt.Printf("accepts: %d - rejecteds: %d\n", accepts, rejecteds)
}

type TokenBucket struct {
	tokens         uint64
	cap            uint64
	refillRate     int
	refillInterval time.Duration
	lastRefill     time.Time

	mu sync.Mutex
}

func NewTokenBucket(capacity uint64, refillRate int, refillInterval time.Duration) *TokenBucket {
	// TODO: inicializar bucket começando cheio (todos os tokens disponíveis)
	return &TokenBucket{
		tokens:         capacity,
		cap:            capacity,
		refillRate:     refillRate,
		refillInterval: refillInterval,
		lastRefill:     time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// refill é um método helper para calcular e adicionar tokens
// Retorna quantos tokens foram adicionados
func (tb *TokenBucket) refill() {
	elapsed := time.Since(tb.lastRefill)
	intervals := int64(elapsed / tb.refillInterval)

	if intervals <= 0 {
		return
	}

	tokensToAdd := uint64(intervals) * uint64(tb.refillRate)

	tb.tokens = min(tb.tokens+tokensToAdd, tb.cap)

	tb.lastRefill = time.Now()
}

func (tb *TokenBucket) Stats() (available, capacity uint64) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens, tb.cap
}
