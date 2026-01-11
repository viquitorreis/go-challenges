package main

import "fmt"

func main() {
	ch := make(chan string)
	go func() {
		ch <- "Hello, World!"
	}()

	res := <-ch
	fmt.Println(res)
}
