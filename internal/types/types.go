package types

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

// BenchmarkSuite - a group of benchmarks for a single file
type BenchmarkSuite struct {
	OriginalFile        string
	BenchmarkFile       string
	TestCases           []TestCase
	ExpectedImprovement string
}

// TestCase - a single benchmark test case with original and optimized functions
type TestCase struct {
	Name          string
	OriginalFunc  string
	OptimizedFunc string
	Description   string
	ExpectedGain  string
}
