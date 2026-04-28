package stringpatterns

import (
	"os"

	"github.com/AngadVM/goprofiler/internal/types"
)

type StringAnalyzer struct {
	patterns []Pattern
}

func NewStringAnalyzer() *StringAnalyzer {
	return &StringAnalyzer{
		patterns: GetPatterns(),
	}
}

func (a *StringAnalyzer) AnalyzeFile(filePath string) ([]types.Issue, error) {
	var allIssues []types.Issue

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	source := string(content)

	for _, pattern := range a.patterns {
		issues := pattern.Detector()(source)
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}
