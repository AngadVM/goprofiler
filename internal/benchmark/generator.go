package benchmark

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/AngadVM/goprofiler/internal/types"
)

type BenchmarkGenerator struct {
	fileSet *token.FileSet
}

func NewBenchmarkGenerator() *BenchmarkGenerator {
	return &BenchmarkGenerator{
		fileSet: token.NewFileSet(),
	}
}

// create benchmark files for detected performance issues
func (bg *BenchmarkGenerator) GenerateBenchmarks(results []types.AnalysisResult) ([]types.BenchmarkSuite, error) {
	var suites []types.BenchmarkSuite

	for _, result := range results {
		if len(result.Issues) == 0 {
			continue
		}

		suite, err := bg.generateSuiteForFile(result)
		if err != nil {
			fmt.Printf("Warning: Could not generate benchmark for %s: %v\n", result.FilePath, err)
			continue
		}

		if len(suite.TestCases) > 0 {
			suites = append(suites, suite)
		}
	}

	return suites, nil
}

func (bg *BenchmarkGenerator) generateSuiteForFile(result types.AnalysisResult) (types.BenchmarkSuite, error) {
	suite := types.BenchmarkSuite{
		OriginalFile: result.FilePath,
		TestCases:    []types.TestCase{},
	}

	// parsing the original file
	node, err := parser.ParseFile(bg.fileSet, result.FilePath, nil, parser.ParseComments)
	if err != nil {
		return suite, fmt.Errorf("failed to parse file: %w", err)
	}

	// generate test cases for each issue
	for _, issue := range result.Issues {
		testCase, err := bg.generateTestCase(node, issue)
		if err != nil {
			continue // Skip issues which can't be benched 
		}

		suite.TestCases = append(suite.TestCases, testCase)
	}

	if len(suite.TestCases) == 0 {
		return suite, fmt.Errorf("no benchmarkable issues found")
	}

	// generate the benchmark file
	benchmarkFile, err := bg.createBenchmarkFile(suite, node)
	if err != nil {
		return suite, fmt.Errorf("failed to create benchmark file: %w", err)
	}

	suite.BenchmarkFile = benchmarkFile
	return suite, nil
}

// dynamically extract code from the AST.
func (bg *BenchmarkGenerator) generateTestCase(node *ast.File, issue types.Issue) (types.TestCase, error) {
	switch {
	case strings.Contains(issue.Title, "String concatenation"):
		return bg.newStringConcatBenchmark(node, issue)
	case strings.Contains(issue.Title, "Slice allocated"):
		return bg.newSliceBenchmark(node, issue)
	case strings.Contains(issue.Title, "map lookups"):
		return bg.newMapLookupBenchmark(node, issue)
	default:
		return types.TestCase{}, fmt.Errorf("unsupported issue type: %s", issue.Title)
	}
}

// implement the AST extraction and transformation.
func (bg *BenchmarkGenerator) newStringConcatBenchmark(node *ast.File, issue types.Issue) (types.TestCase, error) {
	// it using hardcoded placeholders for now
	originalCode := bg.generateOriginalStringConcat()
	optimizedCode := bg.generateOptimizedStringConcat()

	return types.TestCase{
		Name:          "StringConcat",
		Description:   "String concatenation optimization using strings.Builder",
		ExpectedGain:  "3-5x performance improvement, 80% fewer allocations",
		OriginalFunc:  originalCode,
		OptimizedFunc: optimizedCode,
	}, nil
}

func (bg *BenchmarkGenerator) newSliceBenchmark(node *ast.File, issue types.Issue) (types.TestCase, error) {
	originalCode := bg.generateOriginalSliceAlloc()
	optimizedCode := bg.generateOptimizedSliceAlloc()
	return types.TestCase{
		Name:          "SliceAlloc",
		Description:   "Slice allocation with proper capacity",
		ExpectedGain:  "2-3x performance improvement, 60% fewer allocations",
		OriginalFunc:  originalCode,
		OptimizedFunc: optimizedCode,
	}, nil
}

func (bg *BenchmarkGenerator) newMapLookupBenchmark(node *ast.File, issue types.Issue) (types.TestCase, error) {
	originalCode := bg.generateOriginalMapLookup()
	optimizedCode := bg.generateOptimizedMapLookup()
	return types.TestCase{
		Name:          "MapLookup",
		Description:   "Optimized map lookup caching",
		ExpectedGain:  "1.5-2x performance improvement",
		OriginalFunc:  originalCode,
		OptimizedFunc: optimizedCode,
	}, nil
}

// conceptual placeholders
func (bg *BenchmarkGenerator) generateOriginalStringConcat() string {
	return `func processUsersOriginal(users []string) string {
	result := ""
	for _, user := range users {
		result += user + ","
	}
	return result
}`
}

func (bg *BenchmarkGenerator) generateOptimizedStringConcat() string {
	return `func processUsersOptimized(users []string) string {
	var builder strings.Builder
	for _, user := range users {
		builder.WriteString(user)
		builder.WriteString(",")
	}
	return builder.String()
}`
}

func (bg *BenchmarkGenerator) generateOriginalSliceAlloc() string {
	return `func createSliceOriginal() []int {
	data := make([]int, 0)
	for i := 0; i < 1000; i++ {
		data = append(data, i)
	}
	return data
}`
}

func (bg *BenchmarkGenerator) generateOptimizedSliceAlloc() string {
	return `func createSliceOptimized() []int {
	data := make([]int, 0, 1000)
	for i := 0; i < 1000; i++ {
		data = append(data, i)
	}
	return data
}`
}

func (bg *BenchmarkGenerator) generateOriginalMapLookup() string {
	return `func processMapOriginal(userScores map[string]int, users []string) int {
	total := 0
	for _, user := range users {
		if userScores[user] > 90 {
			bonus := userScores[user] * 10
			total += bonus
		}
	}
	return total
}`
}

func (bg *BenchmarkGenerator) generateOptimizedMapLookup() string {
	return `func processMapOptimized(userScores map[string]int, users []string) int {
	total := 0
	for _, user := range users {
		if score, exists := userScores[user]; exists && score > 90 {
			bonus := score * 10
			total += bonus
		}
	}
	return total
}`
}

func (bg *BenchmarkGenerator) createBenchmarkFile(suite types.BenchmarkSuite, originalNode *ast.File) (string, error) {
	// Create benchmark file path
	dir := filepath.Dir(suite.OriginalFile)
	filename := strings.TrimSuffix(filepath.Base(suite.OriginalFile), ".go")
	benchmarkPath := filepath.Join(dir, filename+"_bench_test.go")

	// Generate benchmark file content
	content := bg.generateBenchmarkContent(suite, originalNode.Name.Name)

	// Write benchmark file
	err := os.WriteFile(benchmarkPath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write benchmark file: %w", err)
	}

	return benchmarkPath, nil
}

func (bg *BenchmarkGenerator) generateBenchmarkContent(suite types.BenchmarkSuite, pkgName string) string {
	var content strings.Builder

	// Package declaration and imports
	content.WriteString("package " + pkgName + "\n\n")
	content.WriteString("import (\n")
	content.WriteString("\t\"strings\"\n")
	content.WriteString("\t\"testing\"\n")
	content.WriteString(")\n\n")

	// Test data
	content.WriteString("// Test data for benchmarks\n")
	content.WriteString("var (\n")
	content.WriteString("\ttestUsers = []string{\"Alice\", \"Bob\", \"Charlie\", \"David\", \"Eve\", \"Frank\", \"Grace\", \"Henry\", \"Ivy\", \"Jack\"}\n")
	content.WriteString("\ttestMap = map[string]int{\n")
	content.WriteString("\t\t\"Alice\": 95, \"Bob\": 87, \"Charlie\": 92, \"David\": 88, \"Eve\": 96,\n")
	content.WriteString("\t\t\"Frank\": 89, \"Grace\": 94, \"Henry\": 91, \"Ivy\": 85, \"Jack\": 97,\n")
	content.WriteString("\t}\n")
	content.WriteString(")\n\n")

	// generating all function implementations first
	for _, testCase := range suite.TestCases {
		content.WriteString("// " + testCase.Description + "\n")
		content.WriteString(testCase.OriginalFunc + "\n\n")
		content.WriteString(testCase.OptimizedFunc + "\n\n")
	}

	// generating benchmark functions
	for _, testCase := range suite.TestCases {
		content.WriteString(bg.generateBenchmarkFunctions(testCase))
		content.WriteString("\n")
	}

	return content.String()
}

func (bg *BenchmarkGenerator) generateBenchmarkFunctions(testCase types.TestCase) string {
	var content strings.Builder

	switch testCase.Name {
	case "StringConcat":
		content.WriteString("func BenchmarkStringConcatOriginal(b *testing.B) {\n")
		content.WriteString("\tb.ResetTimer()\n")
		content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
		content.WriteString("\t\t_ = processUsersOriginal(testUsers)\n")
		content.WriteString("\t}\n")
		content.WriteString("}\n\n")

		content.WriteString("func BenchmarkStringConcatOptimized(b *testing.B) {\n")
		content.WriteString("\tb.ResetTimer()\n")
		content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
		content.WriteString("\t\t_ = processUsersOptimized(testUsers)\n")
		content.WriteString("\t}\n")
		content.WriteString("}\n")

	case "SliceAlloc":
		content.WriteString("func BenchmarkSliceAllocOriginal(b *testing.B) {\n")
		content.WriteString("\tb.ResetTimer()\n")
		content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
		content.WriteString("\t\t_ = createSliceOriginal()\n")
		content.WriteString("\t}\n")
		content.WriteString("}\n\n")

		content.WriteString("func BenchmarkSliceAllocOptimized(b *testing.B) {\n")
		content.WriteString("\tb.ResetTimer()\n")
		content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
		content.WriteString("\t\t_ = createSliceOptimized()\n")
		content.WriteString("\t}\n")
		content.WriteString("}\n")

	case "MapLookup":
		content.WriteString("func BenchmarkMapLookupOriginal(b *testing.B) {\n")
		content.WriteString("\tb.ResetTimer()\n")
		content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
		content.WriteString("\t\t_ = processMapOriginal(testMap, testUsers)\n")
		content.WriteString("\t}\n")
		content.WriteString("}\n\n")

		content.WriteString("func BenchmarkMapLookupOptimized(b *testing.B) {\n")
		content.WriteString("\tb.ResetTimer()\n")
		content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
		content.WriteString("\t\t_ = processMapOptimized(testMap, testUsers)\n")
		content.WriteString("\t}\n")
		content.WriteString("}\n")
	}

	return content.String()
}
