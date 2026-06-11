package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	result := run(ctx)
	fmt.Println(result)
}

func run(ctx context.Context) []int {
	jobsCh := make(chan int)
	filteredCh := make(chan int)
	resCh := make(chan int)
	doubles := []int{}

	var wg sync.WaitGroup

	wg.Add(1)
	go createNums(ctx, 10, &wg, jobsCh)

	wg.Add(1)
	go filterNums(ctx, &wg, jobsCh, filteredCh)

	wg.Add(1)
	go doubleNums(ctx, &wg, filteredCh, resCh)

	go func() {
		wg.Wait()
		close(resCh)
	}()

	for res := range resCh {
		doubles = append(doubles, res)
	}

	return doubles
}

func createNums(ctx context.Context, n int, wg *sync.WaitGroup, jobsCh chan int) {
	defer wg.Done()
	defer close(jobsCh)

	for i := range n {
		select {
		case <-ctx.Done():
			fmt.Println("context expired on createNums")
			return
		case jobsCh <- i:
		}
	}

}

func filterNums(ctx context.Context, wg *sync.WaitGroup, jobsCh, filteredCh chan int) {
	defer wg.Done()
	defer close(filteredCh)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("context expired on filterNums")
			return
		case job, ok := <-jobsCh:
			if !ok {
				return
			}

			if job%2 == 0 {
				filteredCh <- job
			}
		}
	}
}

func doubleNums(ctx context.Context, wg *sync.WaitGroup, filteredCh, resCh chan int) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("context expired on doubleNums")
			return
		case filtered, ok := <-filteredCh:
			if !ok {
				return
			}

			resCh <- filtered * 2
		}
	}
}
