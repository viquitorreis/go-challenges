package main

import "fmt"

func main() {
	fmt.Println("Odd or Even Factors")

	numFactors := OddEvenFactors(23)
	fmt.Println(numFactors) // "even"

	numFactors = OddEvenFactors(36)
	fmt.Println(numFactors) // "odd"
}

func OddEvenFactors(num int) string {
	if num%2 != 0 {
		return "even"
	}

	return "odd"
}
