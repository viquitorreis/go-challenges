package main

import (
	"fmt"
	challenges "github.com/viquitorreis/go-challenges"
)

func main() {
	fmt.Println("=== Go Challenges Demo ===")
	fmt.Println()

	// String challenges
	fmt.Println("String Challenges:")
	fmt.Printf("ReverseString('hello'): %s\n", challenges.ReverseString("hello"))
	fmt.Printf("IsPalindrome('racecar'): %v\n", challenges.IsPalindrome("racecar"))
	fmt.Printf("IsPalindrome('hello'): %v\n", challenges.IsPalindrome("hello"))
	fmt.Println()

	// Array challenges
	fmt.Println("Array Challenges:")
	nums := []int{1, 5, 3, 9, 2}
	fmt.Printf("FindMax(%v): %d\n", nums, challenges.FindMax(nums))
	fmt.Printf("FindMin(%v): %d\n", nums, challenges.FindMin(nums))
	fmt.Printf("Sum(%v): %d\n", nums, challenges.Sum(nums))
	fmt.Println()

	// Algorithm challenges
	fmt.Println("Algorithm Challenges:")
	fmt.Printf("Fibonacci(10): %d\n", challenges.Fibonacci(10))
	fmt.Printf("FibonacciIterative(15): %d\n", challenges.FibonacciIterative(15))
	fmt.Printf("Factorial(5): %d\n", challenges.Factorial(5))
	fmt.Printf("IsPrime(17): %v\n", challenges.IsPrime(17))
	fmt.Printf("IsPrime(20): %v\n", challenges.IsPrime(20))
}
