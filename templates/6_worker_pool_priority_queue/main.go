package main

import (
	"fmt"
	"time"
)

/*
TODO - PASSOS PARA IMPLEMENTAR WORKER POOL COM PRIORITY QUEUE

1. IMPLEMENTAR JobHeap COM heap.Interface
  - Len(), Less(), Swap() para ordenação
  - Push() e Pop() para container/heap
  - Menor priority number = maior prioridade (vem primeiro)

2. CRIAR PriorityQueue THREAD-SAFE
  - Envolver JobHeap com mutex
  - Usar sync.Cond para acordar workers quando job chegar
  - Implementar Enqueue (adiciona job, sinaliza workers)
  - Implementar Dequeue (bloqueia se vazio, acorda em shutdown)

3. IMPLEMENTAR WorkerPool
  - Criar N workers em goroutines
  - Cada worker loop: Dequeue -> Processa -> Repete
  - WaitGroup para coordenar shutdown

4. GRACEFUL SHUTDOWN
  - Parar de aceitar novos jobs
  - Broadcast para acordar todos os workers bloqueados
  - Esperar workers terminarem jobs atuais
*/
func main() {
	fmt.Println("=== Worker Pool com Priority Queue ===\n")

	processor := func(job *Job) error {
		priorityName := map[int]string{
			PriorityCritical: "CRITICAL",
			PriorityHigh:     "HIGH",
			PriorityNormal:   "NORMAL",
			PriorityLow:      "LOW",
		}
		fmt.Printf("[Worker] Processing job %d [%s]: %s\n",
			job.ID, priorityName[job.Priority], job.Payload)
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(3, 10, processor)
	pool.Start()

	jobs := []*Job{
		{ID: 1, Priority: PriorityNormal, Payload: "Send newsletter"},
		{ID: 2, Priority: PriorityCritical, Payload: "Password reset email"},
		{ID: 3, Priority: PriorityLow, Payload: "Cleanup old logs"},
		{ID: 4, Priority: PriorityHigh, Payload: "Welcome email"},
		{ID: 5, Priority: PriorityCritical, Payload: "Payment confirmation"},
		{ID: 6, Priority: PriorityNormal, Payload: "Weekly report"},
		{ID: 7, Priority: PriorityLow, Payload: "Aggregate metrics"},
		{ID: 8, Priority: PriorityHigh, Payload: "Push notification"},
	}

	fmt.Println("Submitting jobs...")
	for _, job := range jobs {
		if err := pool.Submit(job); err != nil {
			fmt.Printf("Failed to submit job %d: %v\n", job.ID, err)
		}
	}

	time.Sleep(2 * time.Second)

	enq, proc, drop, qs := pool.Stats()
	fmt.Printf("\n=== Stats ===\n")
	fmt.Printf("Enqueued: %d, Processed: %d, Dropped: %d, Queue Size: %d\n", enq, proc, drop, qs)

	fmt.Println("\nShutting down...")
	pool.Shutdown()
	fmt.Println("All workers stopped. Done!")
}

const (
	PriorityCritical = 0 // Password reset, payment confirmation
	PriorityHigh     = 1 // Welcome emails, notifications
	PriorityNormal   = 2 // Newsletter, analytics
	PriorityLow      = 3 // Cleanup tasks, logs aggregation
)

type Job struct {
	ID       int
	Priority int
	Payload  string
}

type JobHeap []*Job

type PriorityQueue struct {
}

func NewPriorityQueue(maxSize int) *PriorityQueue {

}

func (pq *PriorityQueue) Enqueue(job *Job) error {

}

func (pq *PriorityQueue) Dequeue() (*Job, bool) {

}

func (pq *PriorityQueue) Shutdown() {

}

func (pq *PriorityQueue) Stats() (enqueued, processed, dropped, queueSize int) {

}

type WorkerPool struct {
}

func NewWorkerPool(numWorkers, queueSize int, processor func(*Job) error) *WorkerPool {

}

func (wp *WorkerPool) Start() {

}

func (wp *WorkerPool) worker(id int) {

}

func (wp *WorkerPool) Submit(job *Job) error {
}

func (wp *WorkerPool) Shutdown() {
}

func (wp *WorkerPool) Stats() (enqueued, processed, dropped, queueSize int) {
}
