package main

import (
	"fmt"
	"math/rand"
)

func main() {
	sl := NewSkipList(4, 0.5, rand.New(rand.NewSource(42)))

	// Insert
	sl.Insert(10, "ten")
	sl.Insert(50, "fifty")
	sl.Insert(20, "twenty")
	sl.Insert(30, "thirty")
	sl.Insert(70, "seventy")

	// Search
	if val, ok := sl.Search(30); ok {
		fmt.Printf("Found 30: %v\n", val) // "thirty"
	}

	// RangeSearch: retorna todos os valores com score entre min e max (inclusive)
	results := sl.RangeSearch(15, 55)
	fmt.Printf("Range [15, 55]: %v\n", results) // [twenty thirty fifty]

	// Delete
	sl.Delete(20)
	if _, ok := sl.Search(20); !ok {
		fmt.Println("20 deleted successfully")
	}

	// Stats (útil para debugging)
	fmt.Printf("Size: %d\n", sl.Size())
}
