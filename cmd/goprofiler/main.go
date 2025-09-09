package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/AngadVM/goprofiler/internal/analyzer"
	"github.com/AngadVM/goprofiler/internal/benchmark"
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
				Name:    "benchmark",
				Aliases: []string{"b"},
				Usage:   "Generate and run benchmarks for detected issues",
				Action:  benchmarkCommand,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "generate-only",
						Usage: "Only generate benchmark files without running them",
					},
					&cli.BoolFlag{
						Name:  "cleanup",
						Value: true,
						Usage: "Remove generated benchmark files after running",
					},
					&cli.StringFlag{
						Name:  "output",
						Value: "console",
						Usage: "Output format: console, json",
					},
					&cli.StringFlag{
						Name:  "save-results",
						Usage: "Save benchmark results to file (e.g., results.json)",
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

	fmt.Printf(" GoProfiler - Analyzing: %s\n", target)
	fmt.Println("=" + repeatString("=", 40))

	// Use the analyzer package
	a := analyzer.New()
	results, err := a.AnalyzePath(target)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Using the output package
	formatter := output.NewFormatter(outputFormat)
	return formatter.PrintResults(results, verbose)
}

func benchmarkCommand(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("please provide a Go file or directory to benchmark")
	}

	target := ctx.Args().Get(0)
	generateOnly := ctx.Bool("generate-only")
	cleanup := ctx.Bool("cleanup")
	outputFormat := ctx.String("output")
	saveResults := ctx.String("save-results")

	fmt.Printf("--> GoProfiler - Benchmarking: %s\n", target)
	fmt.Println("=" + repeatString("=", 50))

	// Step 1: Analyze for performance issues
	fmt.Println(" [*] Analyzing code for performance issues...")
	a := analyzer.New()
	analysisResults, err := a.AnalyzePath(target)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Count total issues
	totalIssues := 0
	for _, result := range analysisResults {
		totalIssues += len(result.Issues)
	}

	if totalIssues == 0 {
		fmt.Println(" No performance issues found - nothing to benchmark!")
		return nil
	}

	fmt.Printf("   Found %d potential performance issues\n\n", totalIssues)

	// Step 2: Generate benchmarks
	fmt.Println(" [*] Generating benchmark files...")
	generator := benchmark.NewBenchmarkGenerator()
	benchmarkSuites, err := generator.GenerateBenchmarks(analysisResults)
	if err != nil {
		return fmt.Errorf("benchmark generation failed: %w", err)
	}

	if len(benchmarkSuites) == 0 {
		fmt.Println(" No benchmarks could be generated from the detected issues. This may be due to unsupported issue types.")
		return nil
	}

	fmt.Printf("   Generated %d benchmark files for supported issues\n", len(benchmarkSuites))
	for _, suite := range benchmarkSuites {
		fmt.Printf("  (#) %s\n", suite.BenchmarkFile)
	}

	if generateOnly {
		fmt.Println("\n Benchmark generation complete! Use --generate-only=false to run them.")
		return nil
	}

	// Step 3: Run benchmarks
	fmt.Println("\n [*] Running performance benchmarks...")
	fmt.Println("   This may take a few moments...")

	runner := benchmark.NewBenchmarkRunner("./benchmark-results")
	comparisonResults, err := runner.RunBenchmarks(benchmarkSuites)
	if err != nil {
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	// Step 4: Display results
	fmt.Println("\n [*] Performance Results")
	if outputFormat == "json" {
		err = printBenchmarkJSON(comparisonResults)
		if err != nil {
			return fmt.Errorf("failed to print JSON results: %w", err)
		}
	} else {
		runner.PrintResults(comparisonResults)
	}

	// Step 5: Save results if requested
	if saveResults != "" {
		fmt.Printf("\n -> Saving results to %s\n", saveResults)
		if err := runner.SaveResults(comparisonResults, saveResults); err != nil {
			return fmt.Errorf("failed to save results: %w", err)
		}
	}

	// Step 6: Cleanup if requested
	if cleanup {
		fmt.Println(" [*] Cleaning up temporary benchmark files...")
		if err := runner.CleanupBenchmarkFiles(benchmarkSuites); err != nil {
			fmt.Printf("Warning: Cleanup failed: %v\n", err)
		}
	}

	return nil
}

func checkCommand(ctx *cli.Context) error {
	if ctx.NArg() == 0 {
		return fmt.Errorf("please provide a Go file to check")
	}

	target := ctx.Args().Get(0)
	fmt.Printf("[CHECK] Quick analysis: %s\n", target)

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
		fmt.Printf("[!] Found %d high-impact performance issues\n", highImpactIssues)
		fmt.Println("    Run 'goprofiler analyze' for details")
	} else {
		fmt.Println("[OK] No critical performance issues detected")
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

func printBenchmarkJSON(results []benchmark.ComparisonResult) error {
	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	fmt.Println(string(jsonData))
	return nil
}
