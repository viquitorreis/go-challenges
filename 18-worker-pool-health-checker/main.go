package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func main() {
	urls := []string{"good1.com", "bad1.com", "good2.com", "bad2.com", "good3.com"}
	results := CheckURLs(urls, 2, 250*time.Millisecond)
	fmt.Println(results)
}

type CheckResult struct {
	URL      string
	Healthy  bool
	Err      error
	Duration time.Duration
}

func CheckURLs(urls []string, numWorkers int, timeout time.Duration) []CheckResult {
	jobsCh := make(chan string)
	resCh := make(chan CheckResult)
	results := []CheckResult{}
	now := time.Now()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var wg sync.WaitGroup

	// 1. sender goroutine -> job dispatcher goroutine
	wg.Go(func() {
		for _, url := range urls {
			select {
			case <-ctxTimeout.Done():
				fmt.Println("ctx timed out")
				return
			case jobsCh <- url:
			}
		}

		close(jobsCh)
	})

	// 2. worker pools with numWorkers
	wg.Go(func() {
		var workersWG sync.WaitGroup
		for range numWorkers {
			workersWG.Go(func() {
				doWork(ctxTimeout, now, jobsCh, resCh)
			})
		}

		workersWG.Wait()
		close(resCh)
	})

	wg.Go(func() {
		for res := range resCh {
			fmt.Println("res:", res)
			results = append(results, res)
		}
	})

	wg.Wait()

	return results
}

func mockCheck(url string) (bool, error) {
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	if strings.Contains(url, "bad") {
		return false, errors.New("connection refused")
	}

	return true, nil
}

func doWork(ctx context.Context, started time.Time, jobsCh chan string, resCh chan CheckResult) {
	for url := range jobsCh {
		select {
		case <-ctx.Done():
			fmt.Println("ctx expired on do work. Returning")
			return
		default:
			isValid, err := mockCheck(url)

			resCh <- CheckResult{
				URL:      url,
				Healthy:  isValid,
				Err:      err,
				Duration: time.Duration(time.Since(started).Microseconds()),
			}
		}
	}
}
