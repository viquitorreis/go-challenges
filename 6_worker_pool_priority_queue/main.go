package main

import (
	"container/heap"
	"fmt"
	"log"
	"sync"
	"time"
)

/*
TODO - PASSOS PARA IMPLEMENTAR WORKER POOL COM PRIORITY QUEUE

1. IMPLEMENTAR A PRIORITY QUEUE THREAD-SAFE
   - Usar container/heap do Go para manter jobs ordenados
   - Proteger a heap com mutex (múltiplos producers e consumers)
   - Implementar Push e Pop que sejam thread-safe

2. CRIAR OS WORKERS
   - Pool de N workers rodando em goroutines
   - Cada worker pega job de maior prioridade (menor número = maior prioridade)
   - Workers devem parar gracefully quando receber sinal de shutdown

3. IMPLEMENTAR BACKPRESSURE
   - Decidir: bloquear quando queue está cheia? Dropar jobs? Queue unbounded?
   - Se bounded: usar channel ou semaphore para limitar tamanho
   - Sinalizar para producer quando não conseguir adicionar job

4. ADICIONAR MÉTRICAS E OBSERVABILIDADE
   - Quantos jobs estão na queue?
   - Quantos jobs foram processados?
   - Quantos jobs foram dropados (se aplicável)?

5. GRACEFUL SHUTDOWN
   - Workers precisam terminar jobs atuais antes de parar
   - Drenar a queue (processar jobs restantes) ou cancelar?
   - WaitGroup para esperar workers terminarem
*/

func main() {
	fmt.Println("=== Worker Pool com Priority Queue ===")

	// Simular processador de jobs (pode ser qualquer coisa)
	processor := func(job *Job) error {
		priorityName := map[int]string{
			PriorityCritical: "CRITICAL",
			PriorityHigh:     "HIGH",
			PriorityNormal:   "NORMAL",
			PriorityLow:      "LOW",
		}

		log.Printf("[Worker] Processing job %d [%s]: %s\n",
			job.ID, priorityName[job.Priority], job.Payload)

		// Simular trabalho
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Criar worker pool: 3 workers, queue max 10 jobs
	pool := NewWorkerPool(3, 10, processor)
	pool.Start()

	// Submeter jobs com diferentes prioridades
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
		log.Printf("submitting job: %d - %d: %s\n", job.ID, job.Priority, job.Payload)
		if err := pool.Submit(job); err != nil {
			fmt.Printf("Failed to submit job %d: %v\n", job.ID, err)
		}
	}

	// Esperar um pouco para processar
	time.Sleep(2 * time.Second)

	// Mostrar estatísticas
	fmt.Printf("\n=== Stats ===\n%+v\n", pool.Stats())

	// Graceful shutdown
	fmt.Println("\nShutting down...")
	pool.Shutdown()
	fmt.Println("All workers stopped. Done!")
}

const (
	PriorityCritical = iota // Password reset, payment confirmation
	PriorityHigh            // Welcome emails, notifications
	PriorityNormal          // Newsletter, analytics
	PriorityLow             // Cleanup tasks, logs aggregation
)

type Job struct {
	ID       int
	Priority int
	Payload  string
}

type JobHeap []*Job

func (h JobHeap) Len() int {
	return len(h)
}

// prioridade -> menor prioridade primeiro
func (h JobHeap) Less(i, j int) bool {
	return h[i].Priority < h[j].Priority
}

func (h JobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *JobHeap) Push(x any) {
	job := x.(*Job)
	*h = append(*h, job)
}

func (h *JobHeap) Pop() any {
	old := *h
	n := len(old)
	job := old[n-1]
	*h = old[:n-1]
	return job
}

type PriorityQueue struct {
	JobHeap   *JobHeap
	maxSize   int
	enqueued  int
	processed int
	dropped   int

	isShuttingDown bool

	cond *sync.Cond
	mu   sync.Mutex
}

func NewPriorityQueue(maxSize int) *PriorityQueue {
	pq := &PriorityQueue{
		JobHeap: &JobHeap{},
		maxSize: maxSize,
	}
	heap.Init(pq.JobHeap)
	pq.cond = sync.NewCond(&pq.mu)
	return pq
}

func (pq *PriorityQueue) Enqueue(job *Job) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.isShuttingDown {
		return fmt.Errorf("can't enqueue, its shutting down")
	}

	if pq.maxSize > 0 && pq.JobHeap.Len() >= pq.maxSize {
		pq.dropped++
		return fmt.Errorf("queue is full")
	}

	heap.Push(pq.JobHeap, job)
	pq.enqueued++
	pq.cond.Signal()

	return nil
}

// Dequeue remove e retorna job de maior prioridade
// Bloqueia se queue estiver vazia até ter job ou shutdown
func (pq *PriorityQueue) Dequeue() (*Job, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for pq.JobHeap.Len() == 0 && !pq.isShuttingDown {
		pq.cond.Wait()
	}

	if pq.isShuttingDown && pq.JobHeap.Len() == 0 {
		return nil, false
	}

	job := heap.Pop(pq.JobHeap).(*Job)
	pq.processed++

	return job, true
}

func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	return len(*pq.JobHeap)
}

// Shutdown sinaliza que não aceita mais jobs
func (pq *PriorityQueue) Shutdown() {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	pq.isShuttingDown = true
	pq.cond.Broadcast()
}

// WorkerPool gerencia pool de workers processando jobs
type WorkerPool struct {
	pq         *PriorityQueue
	numWorkers int

	processFunc func(*Job) error

	wg sync.WaitGroup
}

// NewWorkerPool cria pool com N workers
func NewWorkerPool(numWorkers int, queueSize int, processor func(*Job) error) *WorkerPool {
	return &WorkerPool{
		pq:          NewPriorityQueue(queueSize),
		numWorkers:  numWorkers,
		processFunc: processor,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker é o loop de cada worker
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	for {
		job, processed := wp.pq.Dequeue()
		if !processed {
			return
		}

		if err := wp.processFunc(job); err != nil {
			fmt.Printf("Worker %d: error processing job %d: %v\n", id, job.ID, err)
		}
	}
}

// Submit adiciona job para processamento
func (wp *WorkerPool) Submit(job *Job) error {
	return wp.pq.Enqueue(job)
}

// Shutdown para de aceitar jobs e espera workers terminarem
func (wp *WorkerPool) Shutdown() {
	wp.pq.Shutdown()
	wp.wg.Wait()
}

// Stats retorna estatísticas do pool
func (wp *WorkerPool) Stats() map[string]int {
	// TODO: implementar
	// Retornar métricas úteis: jobs enqueued, processed, dropped, queue size, etc
	stats := map[string]int{
		"dropped:":   wp.pq.dropped,
		"enqueued:":  wp.pq.enqueued,
		"processed:": wp.pq.processed,
	}

	return stats
}
