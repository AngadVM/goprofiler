package main

import (
	"fmt"
	"os"
)

func ProblematicFunction() {
	result := ""
	for i := 0; i < 1000; i++ {
		result += "Hello "
	}
	fmt.Println(result)

	items := []string{}
	for i := 0; i < 100; i++ {
		items = append(items, fmt.Sprintf("item%d", i))
	}

	userScores := map[string]int{
		"alice": 95,
		"bob":   87,
		"carol": 92,
	}
	users := []string{"alice", "bob", "carol"}
	for _, user := range users {
		if userScores[user] > 90 {
			fmt.Printf("%s has high score\n", user)
			bonus := userScores[user] * 10
			fmt.Printf("Bonus: %d\n", bonus)
		}
	}

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

func main() {
	ProblematicFunction()
}
