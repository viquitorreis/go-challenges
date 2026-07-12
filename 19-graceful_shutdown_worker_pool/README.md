# Graceful Shutdown Worker Pool

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency
**Estimated time:** ~45 minutes

## What it is

A worker pool that processes jobs from a queue continuously and, on `SIGTERM`/`SIGINT`, stops accepting new jobs, finishes the jobs already in flight, and exits cleanly, all without relying on a framework like `http.Server.Shutdown()` to do the draining for you.

## What you'll learn

- Implementing graceful shutdown by hand: stop accepting new work, let in-flight work finish, then exit, in that specific order.
- Coordinating shutdown across multiple worker goroutines with `context` cancellation and a `sync.WaitGroup`.

## What's implemented

- `NewWorkerPool(numWorkers int) *WorkerPool`.
- `Start(ctx context.Context)` launching the worker goroutines.
- `Submit(job Job) error` for enqueueing new jobs.
- `Shutdown()` draining in-flight jobs before returning.

## Design decisions

- Shutdown is manual (no `http.Server.Shutdown()` shortcut available here, since this isn't an HTTP server): `Shutdown()` explicitly stops intake first and only then waits for workers to drain their current job.

## How to run

```bash
go run ./1
```

Note: the source lives in the `1/` subfolder rather than directly in this challenge's root; run command adjusted accordingly.
