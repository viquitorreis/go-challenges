package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("Smallest Difference Challenge")

	original := "gophers"
	doubled := DoubleChars(original)
	fmt.Println(doubled) // ggoopphheerrss
}

func DoubleChars(original string) string {
	var s strings.Builder

	for i := 0; i < len(original); i++ {
		s.WriteByte(original[i])
		s.WriteByte(original[i])
	}

	return s.String()
}
