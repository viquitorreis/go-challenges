package main

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestJobHeapBasic testa operações básicas da heap
func TestJobHeapBasic(t *testing.T) {
	h := &JobHeap{}
	heap.Init(h)

	jobs := []*Job{
		{ID: 1, Priority: PriorityNormal},
		{ID: 2, Priority: PriorityCritical},
		{ID: 3, Priority: PriorityLow},
		{ID: 4, Priority: PriorityHigh},
	}

	// Push jobs
	for _, job := range jobs {
		heap.Push(h, job)
	}

	// Pop deve retornar em ordem de prioridade
	expected := []int{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for i, exp := range expected {
		job := heap.Pop(h).(*Job)
		if job.Priority != exp {
			t.Errorf("Pop %d: expected priority %d, got %d", i, exp, job.Priority)
		}
	}
}

// TestPriorityQueueThreadSafety testa acesso concorrente
func TestPriorityQueueThreadSafety(t *testing.T) {
	pq := NewPriorityQueue(100)

	var wg sync.WaitGroup
	numProducers := 10
	numConsumers := 5
	jobsPerProducer := 20

	// Producers
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < jobsPerProducer; j++ {
				job := &Job{
					ID:       id*jobsPerProducer + j,
					Priority: j % 4, // Variar prioridades
					Payload:  "test",
				}
				pq.Enqueue(job)
			}
		}(i)
	}

	// Consumers
	consumed := int32(0)
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				job, ok := pq.Dequeue()
				if !ok {
					return
				}
				if job != nil {
					atomic.AddInt32(&consumed, 1)
				}
			}
		}()
	}

	// Esperar producers terminarem
	time.Sleep(100 * time.Millisecond)

	// Shutdown
	pq.Shutdown()
	wg.Wait()

	expected := int32(numProducers * jobsPerProducer)
	if consumed != expected {
		t.Errorf("Expected %d jobs consumed, got %d", expected, consumed)
	}
}

// TestBoundedQueue testa comportamento quando queue enche
func TestBoundedQueue(t *testing.T) {
	maxSize := 5
	pq := NewPriorityQueue(maxSize)

	// Encher queue
	for i := 0; i < maxSize; i++ {
		err := pq.Enqueue(&Job{ID: i, Priority: PriorityNormal})
		if err != nil {
			t.Fatalf("Failed to enqueue job %d: %v", i, err)
		}
	}

	// Próximo enqueue deve falhar ou bloquear dependendo da implementação
	err := pq.Enqueue(&Job{ID: 999, Priority: PriorityCritical})
	if err == nil && pq.Len() > maxSize {
		t.Error("Queue exceeded max size without error")
	}
}

// TestWorkerPoolProcessing testa que workers processam jobs
func TestWorkerPoolProcessing(t *testing.T) {
	processed := int32(0)
	processor := func(job *Job) error {
		atomic.AddInt32(&processed, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(3, 20, processor)
	pool.Start()

	// Submit jobs
	numJobs := 15
	for i := 0; i < numJobs; i++ {
		pool.Submit(&Job{
			ID:       i,
			Priority: i % 4,
			Payload:  "test",
		})
	}

	// Esperar processar
	time.Sleep(500 * time.Millisecond)
	pool.Shutdown()

	if processed != int32(numJobs) {
		t.Errorf("Expected %d jobs processed, got %d", numJobs, processed)
	}
}

// TestPriorityOrdering testa que jobs de alta prioridade são processados primeiro
func TestPriorityOrdering(t *testing.T) {
	var mu sync.Mutex
	var processOrder []int

	processor := func(job *Job) error {
		mu.Lock()
		processOrder = append(processOrder, job.ID)
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(1, 10, processor) // 1 worker para ordem determinística
	pool.Start()

	// Submit em ordem aleatória mas IDs baixos = prioridade alta
	jobs := []*Job{
		{ID: 3, Priority: PriorityLow},
		{ID: 1, Priority: PriorityCritical},
		{ID: 4, Priority: PriorityLow},
		{ID: 2, Priority: PriorityHigh},
	}

	for _, job := range jobs {
		pool.Submit(job)
	}

	time.Sleep(500 * time.Millisecond)
	pool.Shutdown()

	// Verificar que processou em ordem de prioridade
	if len(processOrder) != 4 {
		t.Fatalf("Expected 4 jobs processed, got %d", len(processOrder))
	}

	// ID 1 (Critical) deve vir antes de ID 2 (High)
	// ID 2 (High) deve vir antes de IDs 3,4 (Low)
	if processOrder[0] != 1 || processOrder[1] != 2 {
		t.Errorf("Wrong processing order: %v", processOrder)
	}
}
