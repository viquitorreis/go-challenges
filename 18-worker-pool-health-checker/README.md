# Worker Pool Health Checker

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency
**Estimated time:** ~30 minutes

## What it is

A fixed-size worker pool that checks the "health" of a list of URLs concurrently, applying a per-URL timeout even though the underlying check function doesn't accept a `context.Context`.

## What you'll learn

- Applying a timeout from the outside a function that has no `context` support: wrapping the call and racing it against `time.After` with `select`, instead of modifying the function itself.
- Collecting results from N workers into a single slice safely, without requiring result order to match input order.
- Sizing a worker pool independently from the number of jobs, and making sure every job is accounted for in the output.

## What's implemented

- `CheckURLs(urls []string, numWorkers int, timeout time.Duration) []CheckResult` as the public entry point.
- `doWork(ctx context.Context, started time.Time, jobsCh chan string, resCh chan CheckResult)` run by each of the `numWorkers` goroutines.
- `mockCheck(url string) (bool, error)` simulating an HTTP check with a random delay and a deterministic failure for any URL containing `"bad"`.
- The per-URL timeout is applied externally, since `mockCheck` doesn't accept a context; a URL that exceeds `timeout` comes back as `Healthy: false` with a deadline-exceeded error.

## Design decisions

- `sync.WaitGroup` plus a result channel is used instead of an errgroup or third-party library, per the challenge's own constraint of sticking to the standard library.
- Timeout enforcement lives at the call-site (`doWork`), not inside `mockCheck`, mirroring the real-world situation of wrapping a timeout around code you don't control.

## How to run

```bash
go run .
```
