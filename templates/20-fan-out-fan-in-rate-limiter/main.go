package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	reqs := []Request{}
	for i := range 20 {
		reqs = append(reqs, Request{ID: i})
	}

	numWorkers := 5

	results := ProcessRequests(reqs, numWorkers, 20)

	fmt.Println("results", results)
}

type Request struct {
	ID int
}

type Result struct {
	RequestID int
	Success   bool
	Duration  time.Duration
}

func ProcessRequests(requests []Request, numWorkers int, ratePerSecond int) []Result {
	// rps
	ticker := time.NewTicker(time.Second / time.Duration(ratePerSecond))
	defer ticker.Stop()

	in := make(chan Request, len(requests))
	out := make(chan Result)

	// fan -> 1 generator
	for i := range requests {
		in <- Request{ID: i}
	}
	// fechamos pois só vamos consumir do in, não vamos mandar mais
	close(in)

	var wg sync.WaitGroup

	// OUT -> N workers processando
	for range numWorkers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for req := range in {
				// cada worker vai bloquear até o ticker emitir o próximo ticker
				// quando o ticker ticka, vai colocar só um valor no channel (unbuffered channel)
				// então só um dos 5 workers que estão esperando vão conseguir consumir o valor
				// por isso é um rate limiter global
				<-ticker.C
				start := time.Now()
				time.Sleep(time.Millisecond * 50) // simula trabalho

				out <- Result{
					RequestID: req.ID,
					Success:   true,
					Duration:  time.Since(start),
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	var results []Result
	for r := range out {
		results = append(results, r)
	}

	return results
}
