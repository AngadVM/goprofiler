package benchmark

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/AngadVM/goprofiler/internal/analyzer"
)

type BenchmarkGenerator struct {
	fileSet *token.FileSet
}

type BenchmarkSuite struct {
	OriginalFile		string
	BenchmarkFile		string
	TestCases			[]TestCase 
	ExpectedImprovement	string
}

type TestCase struct {
	Name 			string
	OriginalFunc	string
	OptimizedFunc	string
	Description		string
	ExpectedGain	string
}

func NewBenchmarkGenerator() *BenchmarkGenerator {
	return &BenchmarkGenerator{
		fileSet: token.NewFileSet(),
	}
}

// creates benchmark files for detected performance issues
func (bg *BenchmarkGenerator) GenerateBenchmarks (results []analyzer.AnalysisResult) ([]BenchmarkSuite, error) {
	var suites []BenchmarkSuite

	for _, result := range results {
		if len(result.Issues) == 0 {
			continue 
		}

		suite, err := bg.generateSuiteForFile(result)
		if err != nil {
			continue // skipping files which can'tbe processed
		}

		suites = append(suites, suite)
	}

	return suites, nil
}

func (bg *BenchmarkGenerator) generateSuiteForFile(result analyzer.AnalysisResult) (BenchmarkSuite, error) {
	suite:= BenchmarkSuite{
		OriginalFile: result.FilePath,
		TestCases: []TestCase{},
	}

	// parse the original file
	node, err := parser.ParseFile(bg.fileSet, result.FilePath, nil, parser.ParseComments)
	if err != nil {
		return suite, err
	}
	
	// Generate test cases for each issue
	for _, issue := range result.Issues {
		testCase, err := bg.generateTestCase(node, issue)
		if err != nil {
			continue // skipping issues we can't benchmark 
		}

		suite.TestCases = append(suite.TestCases, testCase)
	}

	// Generate the benchmark file 
	benchmarkFile, err := bg.createBenchmarkFile(suite, node)
	if err != nil {
		return suite, err
	}

	suite.BenchmarkFile = benchmarkFile
	return suite, nil
}

func (bg *BenchmarkGenerator) generateTestCase(node *ast.File, issue analyzer.Issue) (TestCase, error) {
	var testCase TestCase

	switch issue.Type {
	case "allocation":
		if strings.Contains(issue.Title, "String concatenation")  {
			return bg.generateStringConcatBenchmark(node, issue)
		} else if strings.Contains(issue.Title, "Slice allocated") {
			return bg.generateSliceBenchmark(node, issue)
		}
	case "goroutine":
		return bg.generateGoroutineBenchmark(node, issue)
	}

	return testCase, fmt.Errorf("unsupported issue type for benchmarking")
}


func (bg *BenchmarkGenerator) generateStringConcatBenchmark(node *ast.File, issue analyzer.Issue) (TestCase, error) {
	// finding problematic func
	funcNode := bg.findFunctionContainingLine(node, issue.Line)
	if funcNode == nil {
		return TestCase{}, fmt.Errorf("could not find function")
	}

	originalFunc := bg.extractFunctionCode(funcNode)
	optimizedFunc := bg.optimizeStringConcatenation(funcNode)

	return TestCase{
		Name: 	fmt.Sprintf("StringConcat_%s", &funcNode.Name.Name),
		OriginalFunc: originalFunc,
		OptimizedFunc: optimizedFunc,
		Description: "String concatenation optimization using strings.Builder",
		ExpectedGain: "3-5x performance improvement, 80 percent fewer allocations",
	}, nil
}


func (bg *BenchmarkGenerator) generateSliceBenchmark(node *ast.File, issue analyzer.Issue) (TestCase, error) {
	funcNode := bg.findFunctionContainingLine(node, issue.Line)
	if funcNode == nil {
		return TestCase{}, fmt.Errorf("could not find function")
	}

	originalFunc := bg.extractFunctionCode(funcNode)
	optimizedFunc := bg.optimizeSliceAllocation(funcNode)

	return TestCase{
		Name: fmt.Sprintf("SliceAlloc_%s", &funcNode.Name.Name),
		OriginalFunc: originalFunc,
		OptimizedFunc: optimizedFunc,
		Description: "Slice allocation with proper capacity",
		ExpectedGain: "2-3x performance improvement, 60 percent fewer allocations",
	}, nil
}


func (bg *BenchmarkGenerator) generateGoroutineBenchmark(node *ast.File, issue analyzer.Issue) (TestCase, error) {
	return TestCase{
		Name: "GoroutineLeak_Detection",
		Description: "Goroutine leak detection benchmark",
		ExpectedGain: "Memory leak prevention",
	}, nil
}

// helper functions

func (bg *BenchmarkGenerator) findFunctionContainingLine(node *ast.File, line int) *ast.FuncDecl {
	var targetFunc *ast.FuncDecl

	ast.Inspect(node, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			pos := bg.fileSet.Position(funcDecl.Pos())
			end := bg.fileSet.Position(funcDecl.End())

			if pos.Line <= line && line <= end.Line {
				targetFunc = funcDecl
				return false
			}
		}
		return true
	})

	return targetFunc
}

func (bg *BenchmarkGenerator) extractFunctionCode(funcDecl *ast.FuncDecl) string {
	// converting AST back to source code 
	var buf strings.Builder
	format.Node(&buf, bg.fileSet, funcDecl)
	return buf.String()
}

func (bg *BenchmarkGenerator) optimizeStringConcatenation(funcDel *ast.FuncDecl) string {
	// create optimized version 
	funcName := funcDel.Name.Name + "Optimized"

	// extract func signature
	params := bg.extractParameters(funcDel)
	returnType := bg.extractReturnType(funcDel)

	template := `func %s(%s) %s{
		var builder strings.Builder
		for _, item := range items {
			builder.WriteString(item)
			builder.WriteString(",")
		}
		return builder.String()
	}`

	return fmt.Sprintf(template, funcName, params, returnType)
}

func (bg *BenchmarkGenerator) optimizeSliceAllocation(funcDecl *ast.FuncDecl) string {
	funcName := funcDecl.Name.Name + "Optimized"
	params := bg.extractParameters(funcDecl)
	returnType := bg.extractReturnType(funcDecl)
	
	template := `func %s(%s) %s {
	data := make([]int, 0, 10000)  // Pre-allocate with capacity
	for i := 0; i < 10000; i++ {
		data = append(data, i)
	}
	return data
}`

	return fmt.Sprintf(template, funcName, params, returnType)
}

func (bg *BenchmarkGenerator) extractParameters(funcDecl *ast.FuncDecl) string {
	if funcDecl.Type.Params == nil {
		return ""
	}

	var params []string
	for _, field := range funcDecl.Type.Params.List {
		var typeStr strings.Builder
		format.Node(&typeStr, bg.fileSet, field.Type)

		for _, name := range field.Names {
			params = append(params, name.Name+" "+typeStr.String())
		}
	}

	return strings.Join(params, ", ")
}

func (bg *BenchmarkGenerator) extractReturnType(funcDecl *ast.FuncDecl) string {
	if funcDecl.Type.Results == nil {
		return ""
	}

	var returnTypes []string
	for _, field := range funcDecl.Type.Results.List {
		var typeStr strings.Builder
		format.Node(&typeStr, bg.fileSet, field.Type)
		returnTypes = append(returnTypes, typeStr.String())
	}

	return strings.Join(returnTypes, ", ")
}

func (bg *BenchmarkGenerator) createBenchmarkFile(suite BenchmarkSuite, originalNode *ast.File) (string, error) {
	packageName := originalNode.Name.Name 

	// generate benchmark file content 
	content := bg.generateBenchmarkContent(suite, packageName)

	// create benchmark file path
	dir := filepath.Dir(suite.OriginalFile)
	filename := strings.TrimSuffix(filepath.Base(suite.OriginalFile), ".go")
	benchmarkPath := filepath.Join(dir, filename)

	// write benchmark file 
	err := os.WriteFile(benchmarkPath, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return benchmarkPath, nil
}

func (bg *BenchmarkGenerator) generateBenchmarkContent(suite BenchmarkSuite, packageName string) string {
	var content strings.Builder
	
	// Package declaration and imports
	content.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	content.WriteString("import (\n")
	content.WriteString("\t\"strings\"\n")
	content.WriteString("\t\"testing\"\n")
	content.WriteString(")\n\n")
	
	// Test data
	content.WriteString("var (\n")
	content.WriteString("\ttestUsers = []string{\"Alice\", \"Bob\", \"Charlie\", \"David\", \"Eve\"}\n")
	content.WriteString("\ttestSize = 1000\n")
	content.WriteString(")\n\n")
	
	// Generate benchmark functions for each test case
	for _, testCase := range suite.TestCases {
		content.WriteString(bg.generateBenchmarkFunction(testCase))
		content.WriteString("\n\n")
	}
	
	return content.String()
}

func (bg *BenchmarkGenerator) generateBenchmarkFunction(testCase TestCase) string {
	var content strings.Builder
	
	// Original benchmark
	content.WriteString(fmt.Sprintf("func Benchmark%sOriginal(b *testing.B) {\n", testCase.Name))
	content.WriteString("\tb.ResetTimer()\n")
	content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
	content.WriteString("\t\t_ = processUsers(testUsers)  // Original function\n")
	content.WriteString("\t}\n")
	content.WriteString("}\n\n")
	
	// Optimized benchmark  
	content.WriteString(fmt.Sprintf("func Benchmark%sOptimized(b *testing.B) {\n", testCase.Name))
	content.WriteString("\tb.ResetTimer()\n")
	content.WriteString("\tfor i := 0; i < b.N; i++ {\n")
	content.WriteString("\t\t_ = processUsersOptimized(testUsers)  // Optimized function\n")
	content.WriteString("\t}\n")
	content.WriteString("}")
	
	return content.String()
}
