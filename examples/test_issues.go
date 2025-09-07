package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ISSUE 1: String concatenation in loop (HIGH impact) - String-based pattern
func slowStringBuilding(items []string) string {
	result := ""
	for _, item := range items {
		result += "Item: " + item + "\n"  // This will be detected by string pattern
	}
	return result
}

// ISSUE 2: Slice without capacity (MEDIUM impact) - String-based pattern
func inefficientSliceGrowth() []int {
	data := make([]int)  // Missing capacity - will be detected
	
	for i := 0; i < 10000; i++ {
		data = append(data, i*i)
	}
	return data
}

// ISSUE 3: Multiple map lookups in loop (HIGH impact) - AST-based pattern
func duplicateMapAccess() {
	userScores := map[string]int{
		"alice":   95,
		"bob":     87,
		"charlie": 92,
		"diana":   98,
	}
	
	users := []string{"alice", "bob", "charlie", "diana"}
	
	for _, user := range users {
		if userScores[user] > 90 {           // First lookup
			fmt.Printf("User %s has high score\n", user)
			bonus := userScores[user] * 0.1   // Second lookup - AST will detect this
			fmt.Printf("Bonus for %s: %.2f\n", user, bonus)
		}
	}
}

// ISSUE 4: Defer in loop (HIGH impact) - AST-based pattern  
func deferInLoop() {
	filenames := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"}
	
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			continue
		}
		defer file.Close()  // AST will detect this as problematic
		
		// Simulate file processing
		data := make([]byte, 1024)
		file.Read(data)
		fmt.Printf("Processed %s\n", filename)
	}
}

// ISSUE 5: Function call in loop condition (MEDIUM impact) - AST-based pattern
func expensiveLoopCondition(items []string) {
	for i := 0; i < len(items); i++ {  // len() called every iteration - AST detects this
		fmt.Printf("Processing item %d: %s\n", i, items[i])
		time.Sleep(1 * time.Millisecond) // Simulate work
	}
}

// ISSUE 6: Interface conversion in loop (MEDIUM impact) - AST-based pattern
func interfaceConversionsInLoop() {
	items := []interface{}{"hello", 42, 3.14, true, "world"}
	
	for _, item := range items {
		str := fmt.Sprintf("%v", item)  // Interface conversion - AST detects this
		fmt.Printf("String representation: %s\n", str)
		
		// Type assertion
		if s, ok := item.(string); ok {  // AST will detect type assertion in loop
			fmt.Printf("It's a string: %s\n", strings.ToUpper(s))
		}
	}
}

// GOOD EXAMPLES - These should NOT trigger warnings

// GOOD: Efficient string building
func efficientStringBuilding(items []string) string {
	var builder strings.Builder
	for _, item := range items {
		builder.WriteString("Item: ")
		builder.WriteString(item)
		builder.WriteString("\n")
	}
	return builder.String()
}

// GOOD: Slice with proper capacity
func efficientSliceGrowth() []int {
	data := make([]int, 0, 10000)  // Capacity specified
	
	for i := 0; i < 10000; i++ {
		data = append(data, i*i)
	}
	return data
}

// GOOD: Single map lookup
func singleMapAccess() {
	userScores := map[string]int{
		"alice":   95,
		"bob":     87,
		"charlie": 92,
		"diana":   98,
	}
	
	users := []string{"alice", "bob", "charlie", "diana"}
	
	for _, user := range users {
		score, exists := userScores[user]  // Single lookup
		if exists && score > 90 {
			fmt.Printf("User %s has high score: %d\n", user, score)
			bonus := score * 0.1  // Using the variable, not another lookup
			fmt.Printf("Bonus for %s: %.2f\n", user, bonus)
		}
	}
}

// GOOD: Defer outside loop
func properDeferUsage() {
	filenames := []string{"file1.txt", "file2.txt", "file3.txt", "file4.txt"}
	
	for _, filename := range filenames {
		func() {
			file, err := os.Open(filename)
			if err != nil {
				return
			}
			defer file.Close()  // Defer in anonymous function is OK
			
			// Process file
			data := make([]byte, 1024)
			file.Read(data)
			fmt.Printf("Processed %s\n", filename)
		}()
	}
}

// GOOD: Store length outside loop
func efficientLoopCondition(items []string) {
	length := len(items)  // Called once outside loop
	for i := 0; i < length; i++ {
		fmt.Printf("Processing item %d: %s\n", i, items[i])
		time.Sleep(1 * time.Millisecond)
	}
}

func main() {
	fmt.Println("=== Testing Performance Issues ===")
	
	testItems := []string{"apple", "banana", "cherry", "date", "elderberry"}
	
	// These will trigger performance warnings
	fmt.Println("\n--- Problematic Functions ---")
	slow := slowStringBuilding(testItems)
	fmt.Printf("Slow result length: %d\n", len(slow))
	
	inefficientData := inefficientSliceGrowth()
	fmt.Printf("Inefficient data length: %d\n", len(inefficientData))
	
	duplicateMapAccess()
	deferInLoop()
	expensiveLoopCondition(testItems)
	interfaceConversionsInLoop()
	
	// These should be efficient
	fmt.Println("\n--- Efficient Functions ---")
	fast := efficientStringBuilding(testItems)
	fmt.Printf("Fast result length: %d\n", len(fast))
	
	efficientData := efficientSliceGrowth()
	fmt.Printf("Efficient data length: %d\n", len(efficientData))
	
	singleMapAccess()
	properDeferUsage()
	efficientLoopCondition(testItems)
}
