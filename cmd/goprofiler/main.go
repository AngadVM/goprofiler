package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AngadVM/goprofiler/internal/analyzer"
	"github.com/AngadVM/goprofiler/internal/output"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "goprofiler",
		Usage: "Analyze and optimize Go code performance",
		Commands: []*cli.Command{
			{
				Name:    "analyze",
				Aliases: []string{"a"},
				Usage:   "Analyze Go source code for performance issues",
				Action:  analyzeCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "output",
						Value: "console",
						Usage: "Output format: console, json",
					},
					&cli.BoolFlag{
						Name:  "verbose",
						Usage: "Show detailed analysis",
					},
				},
			},
			{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "Quick performance check",
				Action:  checkCommand,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func analyzeCommand(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("please provide a Go file or directory to analyze")
	}

	target := ctx.Args().Get(0)
	verbose := ctx.Bool("verbose")
	outputFormat := ctx.String("output")

	fmt.Printf("🚀 GoProfiler - Analyzing: %s\n", target)
	fmt.Println("=" + repeatString("=", 40))

	// Use the analyzer package
	a := analyzer.New()
	results, err := a.AnalyzePath(target)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Use the output package
	formatter := output.NewFormatter(outputFormat)
	return formatter.PrintResults(results, verbose)
}

func checkCommand(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("please provide a Go file to check")
	}

	target := ctx.Args().Get(0)
	fmt.Printf("🔍 Quick check: %s\n", target)

	a := analyzer.New()
	results, err := a.AnalyzePath(target)
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	// Show only high-impact issues for quick check
	highImpactIssues := 0
	for _, result := range results {
		for _, issue := range result.Issues {
			if issue.Impact == "high" {
				highImpactIssues++
			}
		}
	}

	if highImpactIssues > 0 {
		fmt.Printf("⚠️  Found %d high-impact performance issues\n", highImpactIssues)
		fmt.Println("Run 'goprofiler analyze' for details")
	} else {
		fmt.Println("✅ No critical performance issues detected")
	}

	return nil
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
