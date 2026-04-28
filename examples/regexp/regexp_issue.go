package main

import (
	"fmt"
	"regexp"
)

func validateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func validatePhone(phone string) bool {
	re := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return re.MatchString(phone)
}

func processData() {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString("abc123def456", -1)
	fmt.Println(matches)
}

func main() {
	fmt.Println(validateEmail("test@example.com"))
	fmt.Println(validatePhone("+1234567890"))
	processData()
}
