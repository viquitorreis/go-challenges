package challenges

import "testing"

func TestFibonacci(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"fibonacci 0", 0, 0},
		{"fibonacci 1", 1, 1},
		{"fibonacci 2", 2, 1},
		{"fibonacci 5", 5, 5},
		{"fibonacci 10", 10, 55},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fibonacci(tt.input)
			if result != tt.expected {
				t.Errorf("Fibonacci(%d) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFibonacciIterative(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"fibonacci 0", 0, 0},
		{"fibonacci 1", 1, 1},
		{"fibonacci 2", 2, 1},
		{"fibonacci 5", 5, 5},
		{"fibonacci 10", 10, 55},
		{"fibonacci 15", 15, 610},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FibonacciIterative(tt.input)
			if result != tt.expected {
				t.Errorf("FibonacciIterative(%d) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFactorial(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"factorial 0", 0, 1},
		{"factorial 1", 1, 1},
		{"factorial 5", 5, 120},
		{"factorial 7", 7, 5040},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Factorial(tt.input)
			if result != tt.expected {
				t.Errorf("Factorial(%d) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPrime(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected bool
	}{
		{"0 is not prime", 0, false},
		{"1 is not prime", 1, false},
		{"2 is prime", 2, true},
		{"3 is prime", 3, true},
		{"4 is not prime", 4, false},
		{"17 is prime", 17, true},
		{"20 is not prime", 20, false},
		{"97 is prime", 97, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrime(tt.input)
			if result != tt.expected {
				t.Errorf("IsPrime(%d) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
