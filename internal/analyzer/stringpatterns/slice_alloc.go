package stringpatterns

import (
	"strings"

	"github.com/AngadVM/goprofiler/internal/types"
)

type sliceAllocPattern struct{}

func (p sliceAllocPattern) Name() string        { return "Empty Slice Allocation" }
func (p sliceAllocPattern) Description() string { return "Slice allocated without capacity hint" }
func (p sliceAllocPattern) Impact() string      { return "medium" }

func (p sliceAllocPattern) Detector() func(source string) []types.Issue {
	return func(source string) []types.Issue {
		var issues []types.Issue
		lines := strings.Split(source, "\n")

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)

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
}

func init() {
	RegisterPattern(sliceAllocPattern{})
}
