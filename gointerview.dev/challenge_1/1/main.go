package main

import "fmt"

func main() {
	var (
		a, b int
	)
	_, err := fmt.Scanf("%d %d\n", &a, &b)
	if err != nil {
		fmt.Println("Error reading input: ", err)
		return
	}

	res := Sum(a, b)
	fmt.Println(res)
}

// https://app.gointerview.dev/challenge/1
func Sum(a, b int) int {
	return a + b
}
