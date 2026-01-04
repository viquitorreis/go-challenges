# Go Challenges

A collection of common programming challenges implemented in Go, designed to help developers practice and improve their Go programming skills.

## Overview

This repository contains implementations of various coding challenges organized by category:
- **String Manipulation**: String reversal, palindrome checking
- **Array Operations**: Finding max/min values, sum calculations
- **Algorithms**: Fibonacci numbers, factorials, prime number checking

## Installation

```bash
go get github.com/viquitorreis/go-challenges
```

## Usage

Import the package in your Go code:

```go
import "github.com/viquitorreis/go-challenges"
```

### String Challenges

```go
// Reverse a string
reversed := challenges.ReverseString("hello")
// Output: "olleh"

// Check if a string is a palindrome
isPalin := challenges.IsPalindrome("racecar")
// Output: true
```

### Array Challenges

```go
// Find maximum value in a slice
max := challenges.FindMax([]int{1, 5, 3, 9, 2})
// Output: 9

// Find minimum value in a slice
min := challenges.FindMin([]int{1, 5, 3, 9, 2})
// Output: 1

// Calculate sum of slice elements
sum := challenges.Sum([]int{1, 2, 3, 4, 5})
// Output: 15
```

### Algorithm Challenges

```go
// Calculate Fibonacci number (recursive)
fib := challenges.Fibonacci(10)
// Output: 55

// Calculate Fibonacci number (iterative - more efficient)
fib := challenges.FibonacciIterative(15)
// Output: 610

// Calculate factorial
fact := challenges.Factorial(5)
// Output: 120

// Check if a number is prime
isPrime := challenges.IsPrime(17)
// Output: true
```

## Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run tests for a specific file:

```bash
go test -v strings_test.go strings.go
```

## Example

See `examples/demo.go` for a complete example. Run it with:

```bash
cd examples && go run demo.go
```

Output:
```
=== Go Challenges Demo ===

String Challenges:
ReverseString('hello'): olleh
IsPalindrome('racecar'): true
IsPalindrome('hello'): false

Array Challenges:
FindMax([1 5 3 9 2]): 9
FindMin([1 5 3 9 2]): 1
Sum([1 5 3 9 2]): 20

Algorithm Challenges:
Fibonacci(10): 55
FibonacciIterative(15): 610
Factorial(5): 120
IsPrime(17): true
IsPrime(20): false
```

## Project Structure

```
.
├── algorithms.go       # Algorithm challenges (Fibonacci, Factorial, Prime)
├── algorithms_test.go  # Tests for algorithm challenges
├── arrays.go           # Array manipulation challenges
├── arrays_test.go      # Tests for array challenges
├── strings.go          # String manipulation challenges
├── strings_test.go     # Tests for string challenges
├── examples/
│   └── demo.go        # Demo showcasing all challenges
├── go.mod              # Go module file
└── README.md           # This file
```

## Contributing

Feel free to add more challenges! When contributing:
1. Add your implementation to the appropriate file (or create a new one)
2. Write comprehensive tests
3. Update this README with usage examples

## License

MIT