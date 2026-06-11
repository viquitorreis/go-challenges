## Worker Pool

Implement a worker pool in Go. Given a slice of jobs (integers), process them concurrently using exactly 3 workers. Each worker computes the square of the job value. Collect all results and return them.

```go
// Each job is an int, each result is an int (the square)
func workerPool(jobs []int) []int {
    // your code
}
```

```
Input:  [1, 2, 3, 4, 5]
Output: [1, 4, 9, 16, 25] (order may vary)
```

Requirements:

- Exactly 3 goroutines as workers
- Use channels to distribute jobs and collect results
- No race conditions
- Main goroutine collects all results before returning

## Modelo mental Worker Pool

```
sender goroutine:   sends all jobs -> closes jobsCh
workers:            range jobsCh -> work -> send to resCh -> exit when jobsCh closed
closer goroutine:   wg.Wait() -> closes resCh
main:               range resCh -> collects -> exits when resCh closed
```

-> Every channel must be closed by **whoever sends into it**. Never by the receiver.