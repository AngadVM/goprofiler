package main

import (
	"fmt"
	"strings"
)

// Example function with performance issues
func processUsers(users []string) string {
	// ISSUE 1: String concatenation in loop (HIGH impact)
	result := ""
	for _, user := range users {
		result += user + ","  // This is slow!
	}
	return result
}

func createSlice() []int {
	// ISSUE 2: Slice without capacity hint (MEDIUM impact)  
	data := make([]int, 0)  // Should specify capacity if known
	
	for i := 0; i < 1000; i++ {
		data = append(data, i)
	}
	return data
}

func betterProcessUsers(users []string) string {
	// GOOD: Using strings.Builder
	var builder strings.Builder
	for _, user := range users {
		builder.WriteString(user)
		builder.WriteString(",")
	}
	return builder.String()
}

func main() {
	users := []string{"Alice", "Bob", "Charlie"}
	
	// Slow version
	slow := processUsers(users)
	fmt.Println("Slow result:", slow)
	
	// Fast version  
	fast := betterProcessUsers(users)
	fmt.Println("Fast result:", fast)
	
	// Another issue
	data := createSlice()
	fmt.Println("Data length:", len(data))
}
