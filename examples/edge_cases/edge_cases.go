package main

import "fmt"

func badStringConcat() string {
	s := ""
	for i := 0; i < 1000; i++ {
		s += "a"
	}
	return s
}

func badSliceAlloc() []int {
	data := []int{}
	for i := 0; i < 10000; i++ {
		data = append(data, i)
	}
	return data
}

func badMapLookup(scores map[string]int, name string) int {
	if _, ok := scores[name]; ok {
		return scores[name]
	}
	return 0
}

func main() {
	fmt.Println(badStringConcat())
	fmt.Println(len(badSliceAlloc()))
	testScores := map[string]int{"test": 100}
	fmt.Println(badMapLookup(testScores, "test"))
}
