package main

import (
	"fmt"
	"math"
	"sort"
)

func main() {
	arr := []int{4, 3, 1, 5, 10, 20, -30}
	fmt.Println(LargestSmallest(arr))
	fmt.Println(LargestSmallest(arr))
}

func LargestSmallest(arr []int) (int, int) {
	min, max := math.MaxInt, math.MinInt

	for left, right := 0, len(arr)-1; left <= len(arr) && right >= 0; left, right = left+1, right-1 {
		if arr[left] < min {
			min = arr[left]
		}

		if arr[right] > max {
			max = arr[right]
		}
	}

	return min, max
}

func LasgestSmallestNaive(arr []int) (int, int) {
	sort.Ints(arr)
	return arr[0], arr[len(arr)-1]
}
