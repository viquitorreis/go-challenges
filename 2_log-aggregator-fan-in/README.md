# Log Aggregator (Fan-in)

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency
**Estimated time:** ~1 hour

## What it is

A log aggregator that collects entries from multiple independent producers (simulating services like `api`, `database`, `cache`) into a single ordered result, using the fan-in pattern (N producers to 1 consumer).

## What you'll learn

- **Fan-in (N to 1)**: reading from multiple channels concurrently through a shared "bridge" channel instead of polling each one sequentially.
- Why `select` with a `default` case turns into busy-waiting, and why plain `for range` on a channel is the idiomatic way to block efficiently.
- The critical `sync.WaitGroup` rule: every `Add` must happen before the corresponding `Wait` call, or the internal counter races.
- Layered graceful shutdown: producers close their channels first, then the bridge closes, then the aggregator goroutine returns.
- Single-writer pattern: when only one goroutine writes to a slice, no mutex is needed.

## What's implemented

- `Register(logChan <-chan LogEntry)` - registers a producer's channel to be aggregated.
- `Start()` - launches the internal fan-in goroutine that merges every registered channel into one.
- `Stop() []LogEntry` - waits for all producers and the aggregator to finish, then returns everything collected.
- Tests cover single producer, multiple concurrent producers, graceful shutdown, zero producers, mixed log levels, and a concurrency stress test.

## Design decisions

- A `done` channel signals the consumer goroutine has fully drained the bridge before `Stop()` returns, avoiding races on the internal log slice.
- The bridge channel is created inside `Start()`, not the constructor, so the fan-in goroutine can't close it before any producer registers.
- Ownership of channels is respected: each producer closes its own channel; the aggregator never closes a channel it didn't create.

## How to run

```bash
go run .
go test ./...
```
