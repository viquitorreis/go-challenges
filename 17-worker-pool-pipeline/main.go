package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logsCh := generateLogs(ctx, 50)
	errorsCh := filterErrors(ctx, logsCh, 4)
	resultsCh := sink(ctx, errorsCh, 5) // 5/sec

	count := 0
	for res := range resultsCh {
		fmt.Println(res)
		count++
	}

	fmt.Println("total processed:", count)
}

type LogEntry struct {
	Source    string
	Level     string // "info", "warn", "error"
	Message   string
	Timestamp time.Time
}

type SinkResult struct {
	Entry  LogEntry
	SentAt time.Time
	Err    error
}

func generateLogs(ctx context.Context, n int) <-chan LogEntry {
	logs := []string{"info", "warn", "error"}
	logsCh := make(chan LogEntry, n)

	for range n {
		select {
		case <-ctx.Done():
			logsCh <- LogEntry{
				Source:    "generator",
				Level:     "error",
				Message:   "context deadline exceeded",
				Timestamp: time.Now(),
			}

			return logsCh
		default:
			log := logs[rand.IntN(3)]
			fmt.Println("sending log", log)
			logsCh <- LogEntry{
				Source:    "generator",
				Level:     log,
				Message:   "hello world",
				Timestamp: time.Now(),
			}
		}
	}

	close(logsCh)
	return logsCh
}

func filterErrors(ctx context.Context, in <-chan LogEntry, numWorkers int) <-chan LogEntry {
	log.Println("called filterErrors")

	errorsCh := make(chan LogEntry)
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case entry, ok := <-in:
					if !ok {
						return
					}

					if strings.Contains(entry.Level, "error") {
						// learn
						select {
						case errorsCh <- entry:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(errorsCh)
	}()

	return errorsCh
}

func sink(ctx context.Context, in <-chan LogEntry, ratePerSecond int) <-chan SinkResult {
	resultsCh := make(chan SinkResult)

	go func() {
		defer close(resultsCh)
		ticker := time.NewTicker(time.Second / time.Duration(ratePerSecond))
		defer ticker.Stop()

		for log := range in {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// mock sending...
				time.Sleep(time.Millisecond * 10)
				select {
				case resultsCh <- SinkResult{
					Entry:  log,
					SentAt: time.Now(),
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return resultsCh
}
