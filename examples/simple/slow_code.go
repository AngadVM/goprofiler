package main

import (
	"fmt"
	"strings"
)

func processUsers(users []string) string {
	result := ""
	for _, user := range users {
		result += user + ","
	}
	return result
}

func createSlice() []int {
	data := []int{}
	for i := 0; i < 1000; i++ {
		data = append(data, i)
	}
	return data
}

func betterProcessUsers(users []string) string {
	var builder strings.Builder
	for _, user := range users {
		builder.WriteString(user)
		builder.WriteString(",")
	}
	return builder.String()
}

func main() {
	users := []string{"Alice", "Bob", "Charlie"}
	slow := processUsers(users)
	fmt.Println("Slow result:", slow)
	fast := betterProcessUsers(users)
	fmt.Println("Fast result:", fast)
	data := createSlice()
	fmt.Println("Data length:", len(data))
}
