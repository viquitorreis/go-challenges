# Fan-out/Fan-in Rate Limiter

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency
**Estimated time:** ~30 minutes

## What it is

A simulated concurrent API request processor: N worker goroutines process requests in parallel (fan-out), but all of them share a single global rate limit, and their results converge back into one channel (fan-in).

## What you'll learn

- The distinction between a **per-worker** rate limit and a **global** rate limit shared across all workers, and why a shared `time.Ticker` is what makes the limit actually global.
- Combining fan-out (multiple workers pulling from a shared job source) with fan-in (all results converging into one channel) in a single function.

## What's implemented

- `ProcessRequests(requests []Request, numWorkers int, ratePerSecond int) []Result` as the single public entry point.
- `numWorkers` workers processing requests concurrently, gated by one shared rate limiter (not one limiter per worker).
- Mocked processing (`time.Sleep(50ms)`, always succeeds) so the focus stays on the concurrency pattern rather than real I/O.

## Design decisions

- The rate limiter is a single shared `time.Ticker` (or equivalent) that every worker reads from before processing a request, rather than each worker having its own ticker, which is what makes the limit global instead of `numWorkers * ratePerSecond`.
- Fan-in and fan-out are combined in one function rather than split across separate exported helpers, since the challenge's scope is this specific combined pattern.

## How to run

```bash
go run .
```
