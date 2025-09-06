package main

import (
	"fmt"
	"time"
)

// ISSUE 1: String concatenation in loop (HIGH impact)
func buildMessage(users []string) string {
	result := ""
	for _, user := range users {
		result += "Hello " + user + "!\n"  // AST will detect this accurately
	}
	return result
}

// ISSUE 2: Slice without capacity (MEDIUM impact)  
func collectData() []int {
	data := make([]int)  // AST detects: make([]int) with no capacity
	
	for i := 0; i < 10000; i++ {
		data = append(data, i*i)
	}
	return data
}

// ISSUE 3: Potential goroutine leak (MEDIUM impact)
func startWorker() {
	go func() {
		for {  // Infinite loop detected by AST
			// No exit condition - potential leak!
			time.Sleep(1 * time.Second)
			fmt.Println("Working...")
		}
	}()
}

// GOOD: No issues here
func efficientBuild(users []string) string {
	var builder strings.Builder  // Good: using strings.Builder
	for _, user := range users {
		builder.WriteString("Hello ")
		builder.WriteString(user)
		builder.WriteString("!\n")
	}
	return builder.String()
}

// GOOD: Slice with proper capacity
func efficientCollect() []int {
	data := make([]int, 0, 10000)  // Good: capacity specified
	
	for i := 0; i < 10000; i++ {
		data = append(data, i*i)
	}
	return data
}

func main() {
	users := []string{"Alice", "Bob", "Charlie"}
	
	// This will trigger the issues
	msg := buildMessage(users)
	fmt.Println(msg)
	
	data := collectData()
	fmt.Println("Data size:", len(data))
	
	startWorker()
}
