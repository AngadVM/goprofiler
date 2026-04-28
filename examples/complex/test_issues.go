package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func slowStringBuilding(items []string) string {
	result := ""
	for _, item := range items {
		result += "Item: " + item + "\n"
	}
	return result
}

func inefficientSliceGrowth() []int {
	data := []int{}
	for i := 0; i < 10000; i++ {
		data = append(data, i*i)
	}
	return data
}

func duplicateMapAccess() {
	userScores := map[string]int{
		"alice": 95, "bob": 87, "charlie": 92, "diana": 98,
	}
	users := []string{"alice", "bob", "charlie", "diana"}
	for _, user := range users {
		if userScores[user] > 90 {
			fmt.Printf("User %s has high score\n", user)
			bonus := userScores[user] * 10
			fmt.Printf("Bonus for %s: %d\n", user, bonus)
		}
	}
}

func deferInLoop() {
	filenames := []string{"file1.txt"}
	for _, filename := range filenames {
		file, _ := os.Open(filename)
		if file != nil {
			defer file.Close()
			data := make([]byte, 1024)
			file.Read(data)
			fmt.Printf("Processed %s\n", filename)
		}
	}
}

func expensiveLoopCondition(items []string) {
	for i := 0; i < len(items); i++ {
		fmt.Printf("Processing item %d: %s\n", i, items[i])
		time.Sleep(1 * time.Millisecond)
	}
}

func interfaceConversionsInLoop() {
	items := []interface{}{"hello", 42, "world"}
	for _, item := range items {
		str := fmt.Sprintf("%v", item)
		fmt.Printf("String: %s\n", str)
		if s, ok := item.(string); ok {
			fmt.Printf("It's a string: %s\n", strings.ToUpper(s))
		}
	}
}

func main() {
	fmt.Println("=== Testing Performance Issues ===")
	testItems := []string{"apple", "banana", "cherry"}
	slow := slowStringBuilding(testItems)
	fmt.Printf("Slow result length: %d\n", len(slow))
	duplicateMapAccess()
	deferInLoop()
	expensiveLoopCondition(testItems)
}
