package challenges

// FindMax returns the maximum value in a slice of integers
func FindMax(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, num := range nums {
		if num > max {
			max = num
		}
	}
	return max
}

// FindMin returns the minimum value in a slice of integers
func FindMin(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	min := nums[0]
	for _, num := range nums {
		if num < min {
			min = num
		}
	}
	return min
}

// Sum returns the sum of all integers in a slice
func Sum(nums []int) int {
	sum := 0
	for _, num := range nums {
		sum += num
	}
	return sum
}
