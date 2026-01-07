package main

import "fmt"

type Developer struct {
	Name string
	Age  int
}

func FilterUnique(developers []Developer) []string {
	// hash set
	set := make(map[string]Developer)

	unique := []string{}
	for i := range developers {
		if _, ok := set[developers[i].Name]; !ok {
			set[developers[i].Name] = developers[i]
			unique = append(unique, developers[i].Name)
		}
	}

	return unique
}

func main() {
	fmt.Println("Filter Unique Challenge")

	developers := []Developer{
		Developer{Name: "Elliot"},
		Developer{Name: "Alan"},
		Developer{Name: "Jennifer"},
		Developer{Name: "Graham"},
		Developer{Name: "Paul"},
		Developer{Name: "Alan"},
	}

	fmt.Println(FilterUnique(developers))
}
