package main

import "fmt"

// Flight struct which contains
// the origin, destination and price of a flight
type Flight struct {
	Origin      string
	Destination string
	Price       int
}

// IsSubset checks to see if the first set of
// flights is a subset of the second set of flights.
func IsSubset(first, second []Flight) bool {
	set := make(map[string]Flight)

	for i := range second {
		set[fmt.Sprintf("%s%s%d", second[i].Origin, second[i].Destination, second[i].Price)] = second[i]
	}

	for i := range first {
		if _, ok := set[fmt.Sprintf("%s%s%d", first[i].Origin, first[i].Destination, first[i].Price)]; !ok {
			return false
		}
	}

	return true
}

func main() {
	fmt.Println("Sets and Subsets Challenge")
	firstFlights := []Flight{
		{Origin: "GLA", Destination: "CDG", Price: 1000},
		{Origin: "GLA", Destination: "JFK", Price: 5000},
		{Origin: "GLA", Destination: "SNG", Price: 3000},
	}

	secondFlights := []Flight{
		{Origin: "GLA", Destination: "CDG", Price: 1000},
		{Origin: "GLA", Destination: "JFK", Price: 5000},
		{Origin: "GLA", Destination: "SNG", Price: 3000},
		{Origin: "GLA", Destination: "AMS", Price: 500},
	}

	subset := IsSubset(firstFlights, secondFlights)
	fmt.Println(subset)
}
