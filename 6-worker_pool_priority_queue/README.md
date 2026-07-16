# Worker Pool with Priority Queue

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency / Data Structures
**Estimated time:** ~1.5 hours

## What it is

A job processing system where a fixed pool of workers always picks the highest-priority job available first, similar in spirit to Sidekiq or Celery, with multiple producers submitting jobs and multiple workers consuming them concurrently.

## What you'll learn

- Wrapping Go's `container/heap` (not thread-safe by itself) with a mutex without turning every operation into a bottleneck.
- Using `sync.Cond` so idle workers block efficiently on an empty queue instead of busy-looping or polling with `time.Sleep`.
- Backpressure: what to do when jobs arrive faster than workers can process them, instead of letting the queue grow unbounded.

## What's implemented

- `JobHeap` implementing `heap.Interface` (`Len`, `Less`, `Swap`, `Push`, `Pop`), ordered so a lower priority number is dequeued first.
- `PriorityQueue` wrapping the heap with a mutex: `Enqueue`, `Dequeue`, `Len`, `Shutdown`.
- `WorkerPool` with a configurable number of workers and a `processor func(*Job) error` callback: `Start`, `Submit`, `Shutdown`, `Stats`.
- `Enqueue` returns an error when the bounded queue is full instead of growing it unbounded, giving callers explicit backpressure.
- Tests cover basic heap ordering, thread safety under concurrent enqueue/dequeue, the bounded queue rejecting when full, job processing, and priority ordering under load.

## Design decisions

- The queue enforces a maximum size (`maxSize`) and rejects new jobs past that point rather than allowing unbounded memory growth - an explicit backpressure choice over silently dropping or blocking forever.
- `sync.Cond` wakes worker goroutines exactly when a job becomes available or shutdown is triggered, avoiding both busy-waiting and polling latency.

## How to run

```bash
go run .
go test ./...
```
