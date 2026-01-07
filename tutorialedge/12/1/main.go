package main

import (
	"fmt"
	"math"
)

type MyInt int

func (n *MyInt) IsArmstrong() bool {
	sum := 0

	rem := *n

	if rem >= 100 {
		cents := (rem / 100)
		sum += int(math.Pow(float64(cents), 3))
		rem = rem - (100 * cents)
	}

	if rem >= 10 {
		d := (rem / 10)
		sum += int(math.Pow(float64(d), 3))
		rem = rem - (10 * d)
	}

	if rem != 0 {
		sum += int(math.Pow(float64(rem), 3))
		rem -= rem
	}

	return *n == MyInt(sum)
}

func main() {
	fmt.Println("Armstrong Numbers")

	var num1 MyInt = 371
	fmt.Println(num1.IsArmstrong())
}
