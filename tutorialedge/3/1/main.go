package main

import (
	"fmt"
	"sort"
)

// Flight - a struct that
// contains information about flights
type Flight struct {
	Origin      string
	Destination string
	Price       int
}

// SortByPrice sorts flights from highest to lowest
func SortByPrice(flights []Flight) []Flight {
	sort.Slice(flights, func(i, j int) bool {
		return flights[i].Price > flights[j].Price
	})

	return flights
}

func printFlights(flights []Flight) {
	for _, flight := range flights {
		fmt.Printf("Origin: %s, Destination: %s, Price: %d", flight.Origin, flight.Destination, flight.Price)
	}
}

func main() {
	// an empty slice of flights
	var flights []Flight = []Flight{
		{
			Origin:      "Austin",
			Destination: "California",
			Price:       69,
		},
		{
			Origin:      "Abu Dhabi",
			Destination: "Qatar",
			Price:       23,
		},
		{
			Origin:      "Uberlandia",
			Destination: "Rome",
			Price:       2026,
		},
	}

	sortedList := SortByPrice(flights)
	printFlights(sortedList)
}
