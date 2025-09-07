package main

import (
	"fmt"
	"strings"
)

// Inefficient String Concatenation: This should be flagged.
// The += operator in a loop is a common performance issue.
func badStringConcat() string {
	s := ""
	for i := 0; i < 1000; i++ {
		s += "a"
	}
	return s
}

// Correct String Concatenation: This should NOT be flagged.
// This example uses strings.Builder, which is the recommended approach.
func goodStringConcat() string {
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString("a")
	}
	return sb.String()
}

// Inefficient Slice Allocation: This should be flagged.
// The slice is created with zero capacity, causing multiple re-allocations during append.
func badSliceAlloc() []int {
	data := []int{}
	for i := 0; i < 10000; i++ {
		data = append(data, i)
	}
	return data
}

// Correct Slice Allocation: This should NOT be flagged.
// The slice is created with a pre-allocated capacity, preventing re-allocations.
func goodSliceAlloc() []int {
	data := make([]int, 0, 10000)
	for i := 0; i < 10000; i++ {
		data = append(data, i)
	}
	return data
}

// Redundant Map Lookup: This should be flagged.
// The map lookup is done twice: once in the 'if' condition and again to use the value.
func badMapLookup(scores map[string]int, name string) int {
	if _, ok := scores[name]; ok {
		return scores[name]
	}
	return 0
}

// Correct Map Lookup: This should NOT be flagged.
// The map lookup is done only once, and the value is reused.
func goodMapLookup(scores map[string]int, name string) int {
	if score, ok := scores[name]; ok {
		return score
	}
	return 0
}

func main() {
	// Calling the functions to make them part of the executable.
	fmt.Println(badStringConcat())
	fmt.Println(goodStringConcat())
	fmt.Println(len(badSliceAlloc()))
	fmt.Println(len(goodSliceAlloc()))
	
	testScores := map[string]int{
		"test": 100,
	}
	fmt.Println(badMapLookup(testScores, "test"))
	fmt.Println(goodMapLookup(testScores, "test"))
}
