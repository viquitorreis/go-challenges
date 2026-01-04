package challenges

import "testing"

func TestFindMax(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"empty slice", []int{}, 0},
		{"single element", []int{5}, 5},
		{"multiple elements", []int{1, 5, 3, 9, 2}, 9},
		{"negative numbers", []int{-5, -1, -10, -3}, -1},
		{"mixed numbers", []int{-5, 10, -3, 8}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindMax(tt.input)
			if result != tt.expected {
				t.Errorf("FindMax(%v) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindMin(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"empty slice", []int{}, 0},
		{"single element", []int{5}, 5},
		{"multiple elements", []int{1, 5, 3, 9, 2}, 1},
		{"negative numbers", []int{-5, -1, -10, -3}, -10},
		{"mixed numbers", []int{-5, 10, -3, 8}, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindMin(tt.input)
			if result != tt.expected {
				t.Errorf("FindMin(%v) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"empty slice", []int{}, 0},
		{"single element", []int{5}, 5},
		{"multiple elements", []int{1, 2, 3, 4, 5}, 15},
		{"negative numbers", []int{-1, -2, -3}, -6},
		{"mixed numbers", []int{-5, 10, -3, 8}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sum(tt.input)
			if result != tt.expected {
				t.Errorf("Sum(%v) = %d; expected %d", tt.input, result, tt.expected)
			}
		})
	}
}
