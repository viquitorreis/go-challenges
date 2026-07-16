# Rate Limiter (Token Bucket)

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency
**Estimated time:** ~1 hour

## What it is

A thread-safe rate limiter based on the token bucket algorithm: the bucket holds up to N tokens, refills at a fixed rate, and each request consumes one token or gets rejected if none are available.

## What you'll learn

- The token bucket algorithm and how it differs from leaky bucket (token bucket allows bursts up to capacity; leaky bucket enforces a constant drain rate).
- Modeling bucket state without storing an array of tokens: just tracking a count and a last-refill timestamp.
- Lazy refill: computing how many tokens should exist based on elapsed time, instead of running a background goroutine on a timer for every tick.
- Where to put the mutex so concurrent `Allow()` calls stay correct without becoming a bottleneck.

## What's implemented

- `NewTokenBucket(capacity uint64, refillRate int, refillInterval time.Duration) *TokenBucket`.
- `Allow() bool` - attempts to consume one token, refilling first if enough time has elapsed.
- `Stats() (available, capacity uint64)` for introspection.
- Tests cover initial token count, consuming all tokens, refill behavior over time, concurrent access, rate-limiting correctness under load, and `-race` checks.

## Design decisions

- Refill is computed lazily inside `Allow()`/`refill()` based on elapsed time since the last refill, avoiding a dedicated ticking goroutine per bucket.
- A single mutex guards both the token count and the last-refill timestamp, since both are read and updated together on every `Allow()` call.

## How to run

```bash
go run .
go test -race ./...
```
