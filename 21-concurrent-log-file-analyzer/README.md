# Concurrent Log File Analyzer

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency / I/O
**Estimated time:** ~35-40 minutes

## What it is

A log file analyzer that reads a (potentially large) log file line by line and counts entries per level (INFO/WARN/ERROR), using one sequential reader, N parsing workers, and one aggregator goroutine.

## What you'll learn

- Why reading a text file is inherently sequential (you can't jump into the middle of a file without already knowing where lines start), while per-line parsing is exactly where concurrency adds value once the per-line work gets expensive.
- The reader -> N workers -> 1 aggregator shape as a general template for line-oriented log/data processing without loading the whole file into memory.
- Why a single aggregator goroutine writing to a `map[string]int` avoids needing a mutex on that map, even though multiple workers are producing results concurrently.

## What's implemented

- `AnalyzeLog(filename string, numWorkers int) (map[string]int, error)` as the public entry point.
- A single goroutine reading the file with `bufio.Scanner`, sending each line into a jobs channel.
- `extractLevel(bt []byte) string` parsing the level out of a line (text between `[` and `]`).
- `numWorkers` workers consuming lines and parsing them concurrently.
- A single aggregating goroutine collecting parsed results into the final `map[string]int`.
- `createLogs()` for generating a sample `access.log` file to test against.
- Malformed lines (missing the expected `[LEVEL]` format) are skipped rather than crashing the program.

## Design decisions

- Only one goroutine ever writes to the result map, which is what makes it safe without a mutex, the concurrency is pushed into the parsing stage, not the aggregation stage.
- Channels are closed in a specific order: the reader closes the jobs channel when the file ends, and the workers close the results channel (via a `WaitGroup`) once every worker has finished, not before.

## How to run

```bash
go run .
```
