package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/AngadVM/goprofiler/internal/types"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m" // High impact
	ColorYellow = "\033[33m" // Medium impact
	ColorBlue   = "\033[34m" // Low impact
	ColorGreen  = "\033[32m" // Success/info
	ColorBold   = "\033[1m"  // Bold text
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
func (f *Formatter) PrintResults(results []types.AnalysisResult, verbose bool) error {
	switch f.format {
	case "json":
		return f.printJSON(results, verbose)
	default:
		return f.printConsole(results, verbose)
	}
}

// printConsole outputs results in a human-readable console format
func (f *Formatter) printConsole(results []types.AnalysisResult, verbose bool) error {
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

	// Summary with colors
	fmt.Printf("%s[*] Analysis Summary:%s\n", ColorBold, ColorReset)
	fmt.Printf("    Files analyzed: %d\n", len(results))
	fmt.Printf("    Total issues: %d\n", totalIssues)
	fmt.Printf("    %sHigh impact: %d%s | %sMedium: %d%s | %sLow: %d%s\n\n",
		ColorRed, highImpact, ColorReset,
		ColorYellow, mediumImpact, ColorReset,
		ColorBlue, lowImpact, ColorReset)

	if totalIssues == 0 {
		fmt.Printf("%s[+] No performance issues detected!%s\n", ColorGreen, ColorReset)
		return nil
	}

	// Show issues by file
	for _, result := range results {
		if len(result.Issues) == 0 {
			continue
		}

		fmt.Printf("%s>> %s%s\n", ColorBold, result.FilePath, ColorReset)
		for _, issue := range result.Issues {
			f.printIssue(issue, verbose)
		}
		fmt.Println()
	}

	return nil
}

// printIssue formats and prints a single issue with colors
func (f *Formatter) printIssue(issue types.Issue, verbose bool) {
	var icon string
	var color string

	switch issue.Impact {
	case "high":
		icon = "[!]"
		color = ColorRed
	case "medium":
		icon = "[*]"
		color = ColorYellow
	case "low":
		icon = "[i]"
		color = ColorBlue
	default:
		icon = "[-]"
		color = ColorReset
	}

	fmt.Printf("   %s%s Line %d: %s (%s impact)%s\n",
		color, icon, issue.Line, issue.Title, issue.Impact, ColorReset)

	if verbose {
		fmt.Printf("       %s\n", issue.Description)
		if issue.Suggestion != "" {
			fmt.Printf("       %sSuggestion:%s %s\n", ColorGreen, ColorReset, issue.Suggestion)
		}
	}
}

// printJSON outputs results in JSON format
func (f *Formatter) printJSON(results []types.AnalysisResult, verbose bool) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}
