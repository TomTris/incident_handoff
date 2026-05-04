package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Describe() string
}

type Circle struct {
	Radius float64
}

type Rectangle struct {
	Width, Height float64
}

type Triangle struct {
	Base, Height float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Describe() string {
	return fmt.Sprintf("Circle r=%.1f", c.Radius)
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Describe() string {
	return fmt.Sprintf("Rectangle %.1f x %.1f", r.Width, r.Height)
}

func (t Triangle) Area() float64 {
	return t.Base * t.Height / 2
}

func (t Triangle) Describe() string {
	return fmt.Sprintf("Triangle b=%.1f, h=%.1f", t.Base, t.Height)
}

func printArea(s Shape) {
	extra := ""
	switch s.(type) {
	case Circle:
		extra = "(is a Circle) "
	}
	fmt.Printf("%s %s→ area: %.2f\n", s.Describe(), extra, s.Area())
}

func totalArea(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func main() {
	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 3, Height: 4},
		Triangle{Base: 6, Height: 8},
	}
	for _, s := range shapes {
		printArea(s)
	}
	fmt.Printf("Total area: %.2f\n", totalArea(shapes))
}
