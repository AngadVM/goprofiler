package stringpatterns

import (
	"os"
	"strings"
	"github.com/AngadVM/goprofiler/internal/types"
)

// handles string-based performance analysis
type StringAnalyzer struct {
	patterns []Pattern
}

// defines a performance pattern to detect (string-based)
type Pattern struct {
	Name        string
	Description string
	Impact      string
	Detector    func(string) []types.Issue
}

// create a new analyzer with string patterns
func NewStringAnalyzer() *StringAnalyzer {
	return &StringAnalyzer{
		patterns: getDefaultPatterns(),
	}
}

// analyzes a single Go file using string patterns
func (a *StringAnalyzer) AnalyzeFile(filePath string) ([]types.Issue, error) {
	var allIssues []types.Issue

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	source := string(content)

	for _, pattern := range a.patterns {
		issues := pattern.Detector(source)
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}


// returns the built-in string-based performance patterns
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

// finds string concatenation in loops
func detectStringConcatenation(source string) []types.Issue {
	var issues []types.Issue
	lines := strings.Split(source, "\n")
	
	inLoop := false
	
	for i, line := range lines {
		// look for loop keywords
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "for ") || 
		   strings.Contains(trimmedLine, " for ") {
			inLoop = true
		}
		
		// checking for string concatenation in loops
		if inLoop && strings.Contains(line, "+=") && strings.Contains(line, "\"") {
			issues = append(issues, types.Issue{
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

// finds slices allocated without capacity hints
func detectSliceAllocation(source string) []types.Issue {
	var issues []types.Issue
	lines := strings.Split(source, "\n")
	
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		
		// Look for make([]Type) without capacity
		if strings.Contains(trimmedLine, "make([]") && 
		   strings.Contains(trimmedLine, ")") &&
		   !strings.Contains(trimmedLine, ",") {
			
			issues = append(issues, types.Issue{
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
