# Event Bus (Fan-out)

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency
**Estimated time:** ~1 hour

## What it is

A simple in-memory pub/sub event bus: publishers send string events, and every subscriber that registered gets its own read-only channel and receives every event, concurrently and independently of the others.

## What you'll learn

- **Fan-out**: one event distributed to N independent subscribers.
- The classic **goroutine-closure-over-loop-variable** bug and how to avoid it.
- Guarding a shared subscriber map with `sync.RWMutex`.
- Buffered vs. unbuffered channels and the throughput/backpressure trade-off that comes with each.

## What's implemented

- `Subscribe(name string) <-chan string` - registers a subscriber and returns a dedicated, read-only, buffered channel.
- `Publish(event string)` - sends an event to every registered subscriber's channel.
- `Close()` - closes every subscriber channel and clears the internal map.
- If there are no subscribers when `Publish` runs, the event is simply dropped (no error, no blocking) - this is documented as acceptable behavior for the challenge.

## Design decisions

- Each subscriber gets its **own buffered channel** (capacity 10) instead of sharing one - a slow consumer only affects itself, not the others.
- The bus is exposed as an `EventBus` interface with a private `eventBus` struct implementation, keeping the concurrency details out of the public API.
- A single `sync.RWMutex` protects the subscriber map; reads (`Publish` iterating subscribers) could use `RLock`, writes (`Subscribe`/`Close`) use `Lock`.

## How to run

```bash
go run .
go test ./...
```
