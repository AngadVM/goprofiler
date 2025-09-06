package main

import (
	"fmt"
	"os"
)

// This function has multiple performance issues that GoProfiler should detect
func ProblematicFunction() {
	// Issue 1: String concatenation in loop (string-based pattern)
	result := ""
	for i := 0; i < 1000; i++ {
		result += "Hello "  // Should trigger warning
	}
	fmt.Println(result)

	// Issue 2: Slice without capacity hint (string-based pattern)
	items := make([]string)  // Should trigger warning
	for i := 0; i < 100; i++ {
		items = append(items, fmt.Sprintf("item%d", i))
	}

	// Issue 3: Multiple map lookups in loop (AST-based pattern)
	userScores := map[string]int{
		"alice": 95,
		"bob":   87,
		"carol": 92,
	}
	
	users := []string{"alice", "bob", "carol"}
	for _, user := range users {
		if userScores[user] > 90 {           // First lookup
			fmt.Printf("%s has high score\n", user)
			bonus := userScores[user] * 0.1   // Second lookup - should trigger warning
			fmt.Printf("Bonus: %.1f\n", bonus)
		}
	}

	// Issue 4: Defer in loop (AST-based pattern)
	filenames := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			continue
		}
		defer file.Close()  // Should trigger warning - defers accumulate!
		
		// Simulate file processing
		data := make([]byte, 1024)
		file.Read(data)
		fmt.Printf("Processed %s\n", filename)
	}
}

// Another function with defer in regular for loop
func AnotherProblematicFunction() {
	// Issue 5: Another defer in loop case
	for i := 0; i < 5; i++ {
		file, err := os.Create(fmt.Sprintf("temp%d.txt", i))
		if err != nil {
			continue
		}
		defer file.Close()  // Should also trigger warning
		
		file.WriteString("test data")
	}
}
