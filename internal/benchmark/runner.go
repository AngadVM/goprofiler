package benchmark

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/AngadVM/goprofiler/internal/types"
)

type BenchmarkRunner struct {
	outputDir string
}

type BenchmarkResult struct {
	Name        string  `json:"name"`
	Iterations  int64   `json:"iterations"`
	NsPerOp     float64 `json:"ns_per_op"`
	AllocsPerOp int64   `json:"allocs_per_op"`
	BytesPerOp  int64   `json:"bytes_per_op"`
}

type ComparisonResult struct {
	TestCase        string             `json:"test_case"`
	OriginalResult  BenchmarkResult    `json:"original"`
	OptimizedResult BenchmarkResult    `json:"optimized"`
	Improvement     ImprovementMetrics `json:"improvement"`
}

type ImprovementMetrics struct {
	SpeedupFactor     float64 `json:"speedup_factor"`
	SpeedupPercentage float64 `json:"speedup_percentage"`
	AllocReduction    float64 `json:"alloc_reduction_percentage"`
	MemoryReduction   float64 `json:"memory_reduction_percentage"`
}

func NewBenchmarkRunner(outputDir string) *BenchmarkRunner {
	return &BenchmarkRunner{
		outputDir: outputDir,
	}
}

// executes all generated benchmarks and returns comparison results
func (br *BenchmarkRunner) RunBenchmarks(suites []types.BenchmarkSuite) ([]ComparisonResult, error) {
	var allResults []ComparisonResult

	for _, suite := range suites {
		fmt.Printf("   Running benchmarks for %s...\n", filepath.Base(suite.OriginalFile))
		results, err := br.runSuite(suite)
		if err != nil {
			fmt.Printf("   Warning: Failed to run benchmarks for %s: %v\n", suite.OriginalFile, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// executes a single benchmark suite in a temporary directory
func (br *BenchmarkRunner) runSuite(suite types.BenchmarkSuite) ([]ComparisonResult, error) {
	var results []ComparisonResult

	// create a temporary directory for the benchmark
	tempDir, err := os.MkdirTemp("", "goprofiler-bench-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Ensure cleanup

	// Copy the generated benchmark file to the temporary directory
	benchFileBasename := filepath.Base(suite.BenchmarkFile)
	newBenchPath := filepath.Join(tempDir, benchFileBasename)
	
	// Use copy-and-delete to handle cross-device moves
	sourceFile, err := os.Open(suite.BenchmarkFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open source benchmark file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(newBenchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination benchmark file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to copy benchmark file: %w", err)
	}
	
	// remove the original file
	if err := os.Remove(suite.BenchmarkFile); err != nil {
		fmt.Printf("Warning: Failed to remove original benchmark file: %v\n", err)
	}

	suite.BenchmarkFile = newBenchPath

	// Copy the original file to the temporary directory
	originalFileBasename := filepath.Base(suite.OriginalFile)
	newOriginalPath := filepath.Join(tempDir, originalFileBasename)
	originalContent, err := os.ReadFile(suite.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read original file: %w", err)
	}
	if err := os.WriteFile(newOriginalPath, originalContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to copy original file: %w", err)
	}
	
	// Create a temporary go.mod file in the temporary directory
	modContent := `module goprofiler_bench_temp

go 1.21
`
	modPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(modContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write temporary go.mod: %w", err)
	}

	// Change to the temporary directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	if err := os.Chdir(tempDir); err != nil {
		return nil, fmt.Errorf("failed to change directory: %w", err)
	}

	// Run 'go mod tidy' to resolve dependencies
	fmt.Printf("   Resolving dependencies...\n")
	cmdTidy := exec.Command("go", "mod", "tidy")
	cmdTidy.Stdout = os.Stdout
	cmdTidy.Stderr = os.Stderr
	if err := cmdTidy.Run(); err != nil {
		return nil, fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Run the benchmark
	cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "-count=1", "./"+benchFileBasename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("benchmark execution failed: %w\nOutput: %s", err, string(output))
	}

	// Parse and return results
	benchResults, err := br.parseBenchmarkOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse benchmark output: %w", err)
	}

	// Create comparison results
	testCaseMap := map[string]string{
		"BenchmarkStringConcatOriginal":   "StringConcat",
		"BenchmarkStringConcatOptimized":  "StringConcat",
		"BenchmarkSliceAllocOriginal":     "SliceAlloc",
		"BenchmarkSliceAllocOptimized":    "SliceAlloc", 
		"BenchmarkMapLookupOriginal":      "MapLookup",
		"BenchmarkMapLookupOptimized":     "MapLookup",
	}

	// Group results by test case
	testCases := make(map[string]map[string]BenchmarkResult)
	for benchName, result := range benchResults {
		if testCase, exists := testCaseMap[benchName]; exists {
			if testCases[testCase] == nil {
				testCases[testCase] = make(map[string]BenchmarkResult)
			}
			
			if strings.Contains(benchName, "Original") {
				testCases[testCase]["original"] = result
			} else if strings.Contains(benchName, "Optimized") {
				testCases[testCase]["optimized"] = result
			}
		}
	}

	// Create comparison results
	for testCase, resultMap := range testCases {
		originalResult, hasOriginal := resultMap["original"]
		optimizedResult, hasOptimized := resultMap["optimized"]

		if hasOriginal && hasOptimized {
			improvement := br.calculateImprovement(originalResult, optimizedResult)
			results = append(results, ComparisonResult{
				TestCase:        testCase,
				OriginalResult:  originalResult,
				OptimizedResult: optimizedResult,
				Improvement:     improvement,
			})
		}
	}

	return results, nil
}

func (br *BenchmarkRunner) hasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func (br *BenchmarkRunner) createTempGoMod(dir string) error {
	modContent := `module benchmark_temp

go 1.21
`
	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modContent), 0644)
}

func (br *BenchmarkRunner) cleanupTempGoMod(dir string) {
	os.Remove(filepath.Join(dir, "go.mod"))
	os.Remove(filepath.Join(dir, "go.sum"))
}

func (br *BenchmarkRunner) parseBenchmarkOutput(output string) (map[string]BenchmarkResult, error) {
	results := make(map[string]BenchmarkResult)

	// Updated regex to handle different benchmark output formats
	benchmarkRegex := regexp.MustCompile(`^(Benchmark\w+)(?:-\d+)?\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := benchmarkRegex.FindStringSubmatch(line)

		if len(matches) >= 4 {
			name := matches[1]
			iterations, _ := strconv.ParseInt(matches[2], 10, 64)
			nsPerOp, _ := strconv.ParseFloat(matches[3], 64)

			result := BenchmarkResult{
				Name:       name,
				Iterations: iterations,
				NsPerOp:    nsPerOp,
			}

			// Parse optional memory metrics
			if len(matches) >= 5 && matches[4] != "" {
				result.BytesPerOp, _ = strconv.ParseInt(matches[4], 10, 64)
			}
			if len(matches) >= 6 && matches[5] != "" {
				result.AllocsPerOp, _ = strconv.ParseInt(matches[5], 10, 64)
			}

			results[name] = result
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no benchmark results found in output: %s", output)
	}

	return results, nil
}

func (br *BenchmarkRunner) calculateImprovement(original, optimized BenchmarkResult) ImprovementMetrics {
	var improvement ImprovementMetrics

	// Calculate speed improvement
	if optimized.NsPerOp > 0 && original.NsPerOp > 0 {
		improvement.SpeedupFactor = original.NsPerOp / optimized.NsPerOp
		improvement.SpeedupPercentage = ((original.NsPerOp - optimized.NsPerOp) / original.NsPerOp) * 100
	}

	// Calculate allocation reduction
	if original.AllocsPerOp > 0 {
		improvement.AllocReduction = ((float64(original.AllocsPerOp - optimized.AllocsPerOp)) / float64(original.AllocsPerOp)) * 100
	}

	// Calculate memory reduction
	if original.BytesPerOp > 0 {
		improvement.MemoryReduction = ((float64(original.BytesPerOp - optimized.BytesPerOp)) / float64(original.BytesPerOp)) * 100
	}

	return improvement
}

// saves benchmark results to JSON file for later analysis
func (br *BenchmarkRunner) SaveResults(results []ComparisonResult, filename string) error {
	if br.outputDir != "" {
		err := os.MkdirAll(br.outputDir, 0755)
		if err != nil {
			return err
		}
		filename = filepath.Join(br.outputDir, filename)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// displays benchmark comparison results in a formatted table
func (br *BenchmarkRunner) PrintResults(results []ComparisonResult) {
	if len(results) == 0 {
		fmt.Println("   No benchmark results to display")
		return
	}

	fmt.Printf("\nBenchmark Results Summary\n")
	fmt.Printf("=" + strings.Repeat("=", 50) + "\n\n")

	totalTests := len(results)
	significantImprovements := 0

	for _, result := range results {
		// Consider significant if speedup > 20% or alloc reduction > 30%
		if result.Improvement.SpeedupPercentage > 20 || result.Improvement.AllocReduction > 30 {
			significantImprovements++
		}

		fmt.Printf("Test Case: %s\n", result.TestCase)
		fmt.Printf("   Original:  %8.0f ns/op | %4d allocs/op | %6d B/op\n", 
			result.OriginalResult.NsPerOp, result.OriginalResult.AllocsPerOp, result.OriginalResult.BytesPerOp)
		fmt.Printf("   Optimized: %8.0f ns/op | %4d allocs/op | %6d B/op\n", 
			result.OptimizedResult.NsPerOp, result.OptimizedResult.AllocsPerOp, result.OptimizedResult.BytesPerOp)

		
		speedIcon := br.getImprovementIcon(result.Improvement.SpeedupPercentage)
		allocIcon := br.getImprovementIcon(result.Improvement.AllocReduction)

		fmt.Printf("   %s Speed:  %.1fx faster (%.1f%% improvement)\n", 
			speedIcon, result.Improvement.SpeedupFactor, result.Improvement.SpeedupPercentage)
		fmt.Printf("   %s Memory: %.1f%% fewer allocations, %.1f%% less memory\n\n", 
			allocIcon, result.Improvement.AllocReduction, result.Improvement.MemoryReduction)
	}

	// Summary statistics
	improvementRate := float64(significantImprovements) / float64(totalTests) * 100
	fmt.Printf("Summary: %d/%d tests showed significant improvements (%.1f%%)\n", 
		significantImprovements, totalTests, improvementRate)

	if significantImprovements > 0 {
		fmt.Printf("Average gains: %.1fx faster with %.1f%% fewer allocations\n", 
			br.calculateAverageSpeedup(results), br.calculateAverageAllocReduction(results))
	}
}

// returns a suitable character based on improvement percentage
func (br *BenchmarkRunner) getImprovementIcon(percentage float64) string {
	switch {
	case percentage >= 50:
		return "++" 
	case percentage >= 20:
		return "+" 
	case percentage >= 5:
		return "~"
	case percentage > 0:
		return "•"
	default:
		return "x" 
	}
}

func (br *BenchmarkRunner) calculateAverageSpeedup(results []ComparisonResult) float64 {
	if len(results) == 0 {
		return 0
	}

	total := 0.0
	for _, result := range results {
		total += result.Improvement.SpeedupFactor
	}
	return total / float64(len(results))
}

func (br *BenchmarkRunner) calculateAverageAllocReduction(results []ComparisonResult) float64 {
	if len(results) == 0 {
		return 0
	}

	total := 0.0
	for _, result := range results {
		total += result.Improvement.AllocReduction
	}
	return total / float64(len(results))
}

// removes generated benchmark files
func (br *BenchmarkRunner) CleanupBenchmarkFiles(suites []types.BenchmarkSuite) error {
	for _, suite := range suites {
		// Add a check to see if the file exists before trying to remove it
		if _, err := os.Stat(suite.BenchmarkFile); err == nil {
			if err := os.Remove(suite.BenchmarkFile); err != nil {
				fmt.Printf("   Warning: Could not remove %s: %v\n", suite.BenchmarkFile, err)
			}
		}
	}
	return nil
}
