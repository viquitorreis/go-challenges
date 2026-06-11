package main

import (
	"fmt"
	"sync"
)

func main() {
	numWorkers := 3
	jobs := []int{1, 2, 3, 4, 5}

	fmt.Println(processJobs(jobs, numWorkers))
}

func processJobs(jobs []int, numWorkers int) []int {
	jobsCh := make(chan int)
	resultCh := make(chan int)

	// 1. jobs goroutine
	go func() {
		for i := range jobs {
			jobsCh <- i
		}

		close(jobsCh)
	}()

	// 2. pool of workers
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Add(1)
		go func() {
			for job := range jobsCh {
				resultCh <- job * job
			}

			wg.Done()
		}()

	}

	res := []int{}

	go func() {
		// 4. close result ch
		wg.Wait()
		close(resultCh)
	}()

	// 3. read result
	for job := range resultCh {
		fmt.Printf("received job: %d\n", job)
		res = append(res, job)
	}

	return res
}
