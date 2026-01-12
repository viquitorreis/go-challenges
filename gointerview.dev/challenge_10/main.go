// Package challenge10 contains the solution for Challenge 10.
package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	// Add any necessary imports here
)

func main() {
	rect, _ := NewRectangle(5, 3)
	circle, _ := NewCircle(4)
	triangle, _ := NewTriangle(3, 4, 5)

	triangle.Area()

	calculator := NewShapeCalculator()
	shapes := []Shape{rect, circle, triangle}

	totalArea := calculator.TotalArea(shapes)
	fmt.Printf("Total area: %.2f\n", totalArea)

	sortedShapes := calculator.SortByArea(shapes, true)
	for _, s := range sortedShapes {
		calculator.PrintProperties(s)
	}

	largest := calculator.LargestShape(shapes)
	fmt.Printf("Largest shape: %s with area %.2f\n", largest, largest.Area())

	fmt.Println(rect.String())
}

var (
	ErrInvalidDimension = errors.New("only positive values are allowed")
)

// Shape interface defines methods that all shapes must implement
type Shape interface {
	Area() float64
	Perimeter() float64
	fmt.Stringer // Includes String() string method
}

// Rectangle represents a four-sided shape with perpendicular sides
type Rectangle struct {
	Width  float64
	Height float64
}

// NewRectangle creates a new Rectangle with validation
func NewRectangle(width, height float64) (*Rectangle, error) {
	// TODO: Implement validation and construction
	if width <= 0 || height <= 0 {
		return nil, ErrInvalidDimension
	}

	return &Rectangle{
		Width:  width,
		Height: height,
	}, nil
}

func (r *Rectangle) Area() float64 {
	return r.Height * r.Width
}

func (r *Rectangle) Perimeter() float64 {
	return (2 * r.Height) + (2 * r.Width)
}

// String returns a string representation of the rectangle
func (r *Rectangle) String() string {
	return fmt.Sprintf("Rectangle. width: %.2f height %.2f", r.Width, r.Height)
}

// Circle represents a perfectly round shape
type Circle struct {
	Radius float64
}

func NewCircle(radius float64) (*Circle, error) {
	if radius <= 0 {
		return nil, ErrInvalidDimension
	}

	return &Circle{
		Radius: radius,
	}, nil
}

func (c *Circle) Area() float64 {
	return (c.Radius * c.Radius) * math.Pi
}

func (c *Circle) Perimeter() float64 {
	return math.Pi * (c.Radius + c.Radius)
}

func (c *Circle) String() string {
	return fmt.Sprintf("Circle. radius: %.2f", c.Radius)
}

// Triangle represents a three-sided polygon
type Triangle struct {
	SideA float64
	SideB float64
	SideC float64
}

func NewTriangle(a, b, c float64) (*Triangle, error) {
	if a+b <= c {
		return nil, ErrInvalidDimension
	}

	if a <= 0 || b <= 0 || c <= 0 {
		return nil, ErrInvalidDimension
	}

	return &Triangle{
		SideA: a,
		SideB: b,
		SideC: c,
	}, nil
}

func (t *Triangle) Area() float64 {
	semiperimeter := (t.SideA + t.SideB + t.SideC) / 2
	return math.Sqrt(semiperimeter * ((semiperimeter - t.SideA) * (semiperimeter - t.SideB) * (semiperimeter - t.SideC)))
}

func (t *Triangle) Perimeter() float64 {
	return t.SideA + t.SideB + t.SideC
}

func (t *Triangle) String() string {
	return fmt.Sprintf("Triangle. Sides: %.2f x %.2f x %.2f", t.SideA, t.SideB, t.SideC)
}

// ShapeCalculator provides utility functions for shapes
type ShapeCalculator struct{}

// NewShapeCalculator creates a new ShapeCalculator
func NewShapeCalculator() *ShapeCalculator {
	return &ShapeCalculator{}
}

// PrintProperties prints the properties of a shape
func (sc *ShapeCalculator) PrintProperties(s Shape) {
	// TODO: Implement printing shape properties
	fmt.Println(s.Area())
	fmt.Println(s.Perimeter())
}

// TotalArea calculates the sum of areas of all shapes
func (sc *ShapeCalculator) TotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, shape := range shapes {
		total += shape.Area()
	}

	return total
}

// LargestShape finds the shape with the largest area
func (sc *ShapeCalculator) LargestShape(shapes []Shape) Shape {
	sort.Slice(shapes, func(i, j int) bool {
		return shapes[i].Area() > shapes[j].Area()
	})

	return shapes[0]
}

// SortByArea sorts shapes by area in ascending or descending order
func (sc *ShapeCalculator) SortByArea(shapes []Shape, ascending bool) []Shape {
	sort.Slice(shapes, func(i, j int) bool {
		if ascending {
			return shapes[i].Area() < shapes[j].Area()
		} else {
			return shapes[i].Area() > shapes[j].Area()
		}
	})

	return shapes
}
