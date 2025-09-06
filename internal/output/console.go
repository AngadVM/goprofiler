package output

import (
	"fmt"

	"github.com/AngadVM/goprofiler/internal/analyzer"
)

// Formatter handles different output formats
type Formatter struct {
	format string
}

// NewFormatter creates a new output formatter
func NewFormatter(format string) *Formatter {
	return &Formatter{format: format}
}

// PrintResults outputs the analysis results in the specified format
func (f *Formatter) PrintResults(results []analyzer.AnalysisResult, verbose bool) error {
	switch f.format {
	case "json":
		return f.printJSON(results, verbose)
	default:
		return f.printConsole(results, verbose)
	}
}

// printConsole outputs results in a human-readable console format
func (f *Formatter) printConsole(results []analyzer.AnalysisResult, verbose bool) error {
	totalIssues := 0
	highImpact := 0
	mediumImpact := 0
	lowImpact := 0

	for _, result := range results {
		totalIssues += len(result.Issues)
		for _, issue := range result.Issues {
			switch issue.Impact {
			case "high":
				highImpact++
			case "medium":
				mediumImpact++
			case "low":
				lowImpact++
			}
		}
	}

	// Summary
	fmt.Printf("📊 Analysis Summary:\n")
	fmt.Printf("   Files analyzed: %d\n", len(results))
	fmt.Printf("   Total issues: %d\n", totalIssues)
	fmt.Printf("   High impact: %d | Medium: %d | Low: %d\n\n", 
		highImpact, mediumImpact, lowImpact)

	if totalIssues == 0 {
		fmt.Println("✅ No performance issues detected!")
		return nil
	}

	// Show issues by file
	for _, result := range results {
		if len(result.Issues) == 0 {
			continue
		}

		fmt.Printf("📁 %s\n", result.FilePath)
		for _, issue := range result.Issues {
			f.printIssue(issue, verbose)
		}
		fmt.Println()
	}

	return nil
}

// printIssue formats and prints a single issue
func (f *Formatter) printIssue(issue analyzer.Issue, verbose bool) {
	var icon string
	switch issue.Impact {
	case "high":
		icon = "[!]"
	case "medium":
		icon = "[*]"
	case "low":
		icon = "[i]"
	default:
		icon = "[-]"
	}

	fmt.Printf("   %s Line %d: %s (%s impact)\n", 
		icon, issue.Line, issue.Title, issue.Impact)

	if verbose {
		fmt.Printf("       %s\n", issue.Description)
		if issue.Suggestion != "" {
			fmt.Printf("       Suggestion: %s\n", issue.Suggestion)
		}
	}
}

// printJSON outputs results in JSON format (placeholder for now)
func (f *Formatter) printJSON(results []analyzer.AnalysisResult, verbose bool) error {
	fmt.Println("JSON output not implemented yet")
	return nil
}
