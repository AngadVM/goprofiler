package analyzer

import (
	"strings"
)

// getDefaultPatterns returns the built-in performance patterns
func getDefaultPatterns() []Pattern {
	return []Pattern{
		{
			Name:        "String Concatenation in Loop",
			Description: "String concatenation with += in loops is inefficient",
			Impact:      "high",
			Detector:    detectStringConcatenation,
		},
		{
			Name:        "Empty Slice Allocation",
			Description: "Slice allocated without capacity hint",
			Impact:      "medium",
			Detector:    detectSliceAllocation,
		},
	}
}

// detectStringConcatenation finds string concatenation in loops
func detectStringConcatenation(source string) []Issue {
	var issues []Issue
	lines := strings.Split(source, "\n")
	
	inLoop := false
	
	for i, line := range lines {
		// Look for loop keywords
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "for ") || 
		   strings.Contains(trimmedLine, " for ") {
			inLoop = true
		}
		
		// Checking for string concatenation in loops
		if inLoop && strings.Contains(line, "+=") && strings.Contains(line, "\"") {
			issues = append(issues, Issue{
				Line:        i + 1,
				Title:       "String concatenation in loop",
				Description: "Using += for string concatenation in loops is inefficient",
				Suggestion:  "Use strings.Builder for better performance",
				Impact:      "high",
				Type:        "allocation",
			})
		}
		
		// end of block detection
		if inLoop && strings.TrimSpace(line) == "}" {
			inLoop = false
		}
	}
	
	return issues
}

// detectSliceAllocation finds slices allocated without capacity hints
func detectSliceAllocation(source string) []Issue {
	var issues []Issue
	lines := strings.Split(source, "\n")
	
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Look for make([]Type) without capacity
		if strings.Contains(trimmedLine, "make([]") && 
		   strings.Contains(trimmedLine, ")") &&
		   !strings.Contains(trimmedLine, ",") {
			
			issues = append(issues, Issue{
				Line:        i + 1,
				Title:       "Slice allocated without capacity",
				Description: "Consider providing capacity hint to avoid reallocations",
				Suggestion:  "Use make([]Type, 0, capacity) if you know the expected size",
				Impact:      "medium",
				Type:        "allocation",
			})
		}
	}
	
	return issues
}
