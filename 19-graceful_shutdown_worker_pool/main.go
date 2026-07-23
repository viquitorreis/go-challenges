package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	pool := NewWorkerPool(3)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)

	// simula jobs chegando continuamente
	go func() {
		for i := 1; i <= 20; i++ {
			err := pool.Submit(Job{
				ID:      i,
				Payload: fmt.Sprintf("task-%d", i),
			})

			if err != nil {
				log.Printf("job %d rejected: %v", i, err)
				break
			}

			time.Sleep(300 * time.Millisecond)
		}
	}()

	// graceful shutdown no sinal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	log.Println("shutdown signal received")
	pool.Shutdown()
	log.Println("shutdown complete")
}

type Job struct {
	ID      int
	Payload string
}

type WorkerPool struct {
	Workers  int
	jobsChan chan Job

	shutdown bool
	wg       sync.WaitGroup
	mu       sync.Mutex
	once     sync.Once
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		Workers:  numWorkers,
		jobsChan: make(chan Job),
		wg:       sync.WaitGroup{},
		mu:       sync.Mutex{},
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := range wp.Workers {
		wp.wg.Add(1)

		go func(id int) {
			defer wp.wg.Done()

			for {
				select {
				case job, ok := <-wp.jobsChan:
					if !ok {
						return
					}

					log.Printf("job %d started (worker %d)", job.ID, id)
					time.Sleep(1 * time.Second) // mock trabalho
					log.Printf("job %d finished (worker %d)", job.ID, id)
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
}

func (wp *WorkerPool) Submit(job Job) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	switch {
	case wp.shutdown:
		return fmt.Errorf("pool is shutting down")
	default:
		wp.jobsChan <- job
	}

	return nil
}

func (wp *WorkerPool) Shutdown() {
	wp.once.Do(func() {
		wp.mu.Lock()
		wp.shutdown = true
		wp.mu.Unlock()
		close(wp.jobsChan)
		wp.wg.Wait()
	})
}
