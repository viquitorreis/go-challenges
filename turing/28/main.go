package main

import "fmt"

func main() {
	sum := 0
	for i := 1; i <= 100; i++ {
		if i%2 == 0 {
			sum += i
		}
	}

	fmt.Println("sum: ", sum)
}
