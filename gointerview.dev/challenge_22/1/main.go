package main

import "fmt"

// https://app.gointerview.dev/challenge/22
func main() {
	denominations := []int{5, 10, 25}
	println(MinCoins(63, denominations))
	fmt.Printf("%v\n", CoinCombination(3, denominations))
}

func MinCoins(amount int, denominations []int) int {
	coinsUsed := 0

	for i := len(denominations) - 1; i >= 0; i-- {
		coinsUsed += amount / denominations[i]
		amount = amount % denominations[i]
	}

	if coinsUsed >= 0 && amount == 0 {
		return coinsUsed
	}

	return -1
}

func CoinCombination(amount int, denominations []int) map[int]int {
	combination := make(map[int]int)
	coinsUsed := 0

	for i := len(denominations) - 1; i >= 0; i-- {
		coinsAmount := amount / denominations[i]
		coinsUsed += coinsAmount
		if coinsAmount != 0 {
			combination[denominations[i]] = amount / denominations[i]
		}

		amount = amount % denominations[i]

		if amount == 0 {
			break
		}
	}

	return combination
}
