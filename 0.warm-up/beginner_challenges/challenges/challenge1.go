package beginnerchallenges

import "fmt"

type Student struct {
	Name  string
	Grade int
}

func (s Student) IsPassing() bool {
	return s.Grade > 50 else return false
}

func main() {
	students := []Student{
		{Name: "Tom", Grade: 75},
		{Name: "Anna", Grade: 40},
		{Name: "Leo", Grade: 90},
	}

	for v := range students {
		fmt.Println(v.Name, "-", v.IsPassing())
	}
}