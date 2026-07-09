package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {
	numWorkers := 50
	stats, err := AnalyzeLog("access.log", numWorkers)
	if err != nil {
		log.Fatalf("error analyzing logs: %v", err)
	}

	fmt.Println("stats: ", stats)
}

type LineResult struct {
	Level string // INFO, WARN, ERROR
	Count int
}

func AnalyzeLog(filename string, numWorkers int) (map[string]int, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("err opening log file: %v", err)
	}
	defer f.Close()

	jobsCh := make(chan []byte)
	resCh := make(chan string, 100)

	var wg sync.WaitGroup
	// 1. job sender goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()

		scanner := bufio.NewReader(f)
		for {
			ln, err := scanner.ReadBytes('\n')
			if err != nil && err != io.EOF {
				break
			}

			if err == io.EOF {
				break
			}

			jobsCh <- ln
		}

		close(jobsCh)
	}()

	done := make(chan struct{})

	// 2. workers
	wg.Add(1)
	go func() {
		var workersWg sync.WaitGroup
		defer wg.Done()

		for range numWorkers {
			workersWg.Add(1)

			go func() {
				defer workersWg.Done()

				for job := range jobsCh {
					resCh <- extractLevel(job)
				}
			}()
		}

		workersWg.Wait()
		done <- struct{}{}
	}()

	go func() {
		<-done
		close(resCh)
	}()

	// 3. aggregator goroutine -> metrics
	ans := make(map[string]int)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for res := range resCh {
			ans[res]++
		}
	}()

	wg.Wait()

	return ans, nil
}

func extractLevel(bt []byte) string {
	start := strings.IndexByte(string(bt), '[')
	end := strings.IndexByte(string(bt), ']')

	if start == -1 || end == -1 || end < start {
		return ""
	}

	return string(bt[start+1 : end])
}
