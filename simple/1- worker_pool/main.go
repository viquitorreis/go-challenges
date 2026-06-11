package main

import (
	"fmt"
	"sync"
)

func main() {
	jobs := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

	fmt.Println(workerPool(jobs))
}

func workerPool(jobs []int) []int {
	jobsCh, resCh := make(chan int), make(chan int)
	var wg sync.WaitGroup

	// 1. sender goroutine -> envia os jobs para jobsCh para processar em paralelo
	go func() {
		for _, job := range jobs {
			jobsCh <- job
		}

		// depois de processar tudo em loop sync, tem que fechar os jobs ch para Não dar deadlock
		close(jobsCh)
	}()

	// 2. workers goroutines -> processamento dos jobs em paralelo
	for range 3 {
		wg.Add(1)
		go multiply(jobsCh, resCh, &wg)
	}

	// 3. espera os jobs serem processados em uma goroutine para não travar nada
	go func() {
		wg.Wait()
		close(resCh)
	}()

	ans := []int{}
	for res := range resCh {
		fmt.Println("processed: ", res)
		ans = append(ans, res)
	}

	return ans
}

func multiply(jobsCh, resCh chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobsCh {
		resCh <- job * 2
	}
}
