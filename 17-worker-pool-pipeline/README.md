# Worker Pool + Pipeline (Log Processing)

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency / Pipeline
**Estimated time:** ~40 minutes

## What it is

A 3-stage pipeline simulating log processing: generate mock log entries, filter down to errors using a worker pool, then push them through a rate-limited sink (like an external API with a requests/second limit).

## What you'll learn

- Combining a **pipeline** (sequential stages) with a **worker pool inside one stage** (fan-out for filtering, fan-in back into a single channel).
- The trickiest part of a fan-in stage: multiple workers writing to the same output channel, and making sure it's only closed after every single worker has finished, not after the first one.
- Rate limiting an output stage with `time.Ticker`/`time.Sleep` while still respecting `context` cancellation.

## What's implemented

- `generateLogs(ctx context.Context, n int) <-chan LogEntry` producing `n` mock entries with randomized levels, respecting `ctx.Done()`.
- `filterErrors(ctx context.Context, in <-chan LogEntry, numWorkers int) <-chan LogEntry` running `numWorkers` filtering goroutines that fan their output back into a single channel.
- `sink(ctx context.Context, in <-chan LogEntry, ratePerSecond int) <-chan SinkResult` processing at most `ratePerSecond` items per second and returning results on another channel.

## Design decisions

- The fan-in channel from `filterErrors` is closed by a coordinating goroutine that waits on a `sync.WaitGroup` for all filter workers, not by any individual worker, avoiding a "send on closed channel" panic.
- `context` is threaded through all three stage functions so cancellation stops the pipeline at any point, not just at the generator.

## How to run

```bash
go run .
```
