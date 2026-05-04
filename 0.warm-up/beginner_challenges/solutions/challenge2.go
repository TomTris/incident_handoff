package main

import "fmt"

type Customer struct {
	Name string
	Age  int
}

func (c Customer) CanByBeer() bool {
	return c.Age >= 18
}

func main() {
	customer := []Customer{
		{Name: "Tom", Age: 20},
		{Name: "Anna", Age: 16},
		{Name: "Leo", Age: 18},
		{Name: "Kim", Age: 15},
	}

	for _, v := range customer {
		fmt.Println(v.Name, "-", v.CanByBeer())
	}
}
