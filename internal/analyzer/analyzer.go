package analyzer

import (
	"os"
	"path/filepath"
)

// AnalysisResult - the analysis of a single file
type AnalysisResult struct {
	FilePath string
	Issues   []Issue
}

// Issue - performance issue found in code
type Issue struct {
	Line        int
	Title       string
	Description string
	Suggestion  string
	Impact      string // high, medium, low
	Type        string // allocation, loop, io, etc.
}

// Analyzer - handles performance analysis of Go code
type Analyzer struct {
	patterns []Pattern
}

// Pattern - defines a performance pattern to detect
type Pattern struct {
	Name        string
	Description string
	Impact      string
	Detector    func(string) []Issue
}

// New creates a new analyzer with default patterns
func New() *Analyzer {
	return &Analyzer{
		patterns: getDefaultPatterns(),
	}
}

// AnalyzePath - analyzes a file or directory for performance issues
func (a *Analyzer) AnalyzePath(target string) ([]AnalysisResult, error) {
	var results []AnalysisResult

	// Check if target is a file or directory
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		// Analyze all .go files in directory
		err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".go" {
				result, err := a.analyzeFile(path)
				if err != nil {
					return err
				}
				results = append(results, result)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Analyze single file
		result, err := a.analyzeFile(target)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzeFile analyzes a single Go file
func (a *Analyzer) analyzeFile(filePath string) (AnalysisResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return AnalysisResult{}, err
	}

	var allIssues []Issue
	source := string(content)

	// Run all pattern detectors
	for _, pattern := range a.patterns {
		issues := pattern.Detector(source)
		allIssues = append(allIssues, issues...)
	}

	return AnalysisResult{
		FilePath: filePath,
		Issues:   allIssues,
	}, nil
}
