package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLoad simulates realistic traffic: 70% cancel, 20% add, 10% match,
// varying the number of concurrent goroutines, measuring throughput and
// latency p50/p99 per individual operation.
func TestLoad(t *testing.T) {
	concurrencyLevels := []int{1, 10, 50, 100}
	opsPerGoroutine := 500

	for _, workers := range concurrencyLevels {
		t.Run(fmt.Sprintf("workers_%d", workers), func(t *testing.T) {
			ob := NewOrderBook("BTC-USD")

			// pre populate with orders to have something to cancel from the start
			var idCounter int64
			for i := 0; i < 1000; i++ {
				id := fmt.Sprintf("seed-%d", atomic.AddInt64(&idCounter, 1))
				ob.AddOrder(NewOrder(id, "user", Bid, 100+rand.Intn(50), 1))
			}

			var wg sync.WaitGroup
			var mu sync.Mutex
			var latencies []time.Duration

			start := time.Now()

			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					r := rand.New(rand.NewSource(time.Now().UnixNano()))

					for i := 0; i < opsPerGoroutine; i++ {
						op := r.Float64()
						opStart := time.Now()

						switch {
						case op < 0.70: // cancel
							id := fmt.Sprintf("seed-%d", r.Intn(1000)+1)
							ob.Cancel(id)
						case op < 0.90: // add
							id := fmt.Sprintf("w-%d", atomic.AddInt64(&idCounter, 1))
							side := Bid
							if r.Intn(2) == 0 {
								side = Ask
							}
							ob.AddOrder(NewOrder(id, "user", side, 100+r.Intn(50), 1))
						default: // match
							ob.Match()
						}

						elapsed := time.Since(opStart)
						mu.Lock()
						latencies = append(latencies, elapsed)
						mu.Unlock()
					}
				}()
			}

			wg.Wait()
			totalElapsed := time.Since(start)

			sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
			totalOps := workers * opsPerGoroutine
			throughput := float64(totalOps) / totalElapsed.Seconds()

			p50 := latencies[len(latencies)*50/100]
			p99 := latencies[len(latencies)*99/100]

			t.Logf("workers=%d total_ops=%d elapsed=%v throughput=%.0f ops/s p50=%v p99=%v",
				workers, totalOps, totalElapsed, throughput, p50, p99)
		})
	}
}
