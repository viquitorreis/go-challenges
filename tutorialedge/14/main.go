package main

import "fmt"

func main() {
	fmt.Println("Check Leap Year")

	// year := 1700
	fmt.Println(CheckLeapYear(2020))
}

func CheckLeapYear(year int) bool {
	rem := 0
	if year%4 == 0 {
		if year%100 == 0 {
			fmt.Println(year % 400)
			if year%400 != 0 {
				return false
			}
		}

		rem = year % 4
	}

	return rem == 0
}
