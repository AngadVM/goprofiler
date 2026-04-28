package stringpatterns

import (
	"strings"

	"github.com/AngadVM/goprofiler/internal/types"
)

type stringConcatPattern struct{}

func (p stringConcatPattern) Name() string { return "String Concatenation in Loop" }
func (p stringConcatPattern) Description() string {
	return "String concatenation with += in loops is inefficient"
}
func (p stringConcatPattern) Impact() string { return "high" }

func (p stringConcatPattern) Detector() func(source string) []types.Issue {
	return func(source string) []types.Issue {
		var issues []types.Issue
		lines := strings.Split(source, "\n")

		inLoop := false

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "for ") ||
				strings.Contains(trimmedLine, " for ") {
				inLoop = true
			}

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

			if inLoop && strings.TrimSpace(line) == "}" {
				inLoop = false
			}
		}

		return issues
	}
}

func init() {
	RegisterPattern(stringConcatPattern{})
}
