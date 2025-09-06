package benchmark

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)


type BenchmarkRunner struct {
	outputDir string
}

type BenchmarkResult struct {
	Name          string  `json:"name"`
	Iterations    int64   `json:"iterations"`
	NsPerOp       float64 `json:"ns_per_op"`
	AllocsPerOp   int64   `json:"allocs_per_op"`
	BytesPerOp    int64   `json:"bytes_per_op"`
	MemAllocsPerOp int64  `json:"mem_allocs_per_op"`
}

type ComparisonResult struct {
	TestCase          string             `json:"test_case"`
	OriginalResult    BenchmarkResult    `json:"original"`
	OptimizedResult   BenchmarkResult    `json:"optimized"`
	Improvement       ImprovementMetrics `json:"improvement"`
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
func (br *BenchmarkRunner) RunBenchmarks(suites []BenchmarkSuite) ([]ComparisonResult, error) {
	var allResults []ComparisonResult

	for _, suite := range suites {
		results, err := br.runSuite(suite)
		if err != nil {
			fmt.Printf("Warning: Faled to run benchmarks for %s: %v\n", suite.OriginalFile, err)
			continue 
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func (br *BenchmarkRunner) runSuite(suite BenchmarkSuite) ([]ComparisonResult, error) {
	var results []ComparisonResult

	// change to dir containing benchmark file 
	benchDir := filepath.Dir(suite.BenchmarkFile)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err := os.Chdir(benchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to change directory: %w", err)
	}
	
	// run benchmarks using go test 
	benchFile := filepath.Base(suite.BenchmarkFile)
	cmd := exec.Command("go", "test", "-bench=.", "-benchmem","-count=3", benchFile)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("benchmark execution failed: %w", err)
	}

	// parse benchmark results
	benchResults, err := br.parseBenchmarkOutput(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse benchmark ouput: %w", err)
	}
	
	// comparison results
	for _, testCase := range suite.TestCases {
		originalKey := testCase.Name + "Original"
		optimizedKey := testCase.Name + "Optimized"

		originalResult, hasOriginal := benchResults[originalKey]
		optimizedResult, hasOptimized := benchResults[optimizedKey]

		if hasOriginal && hasOptimized {
			comparison := br.calculateImprovement(originalResult, optimizedResult)
			results = append(results, ComparisonResult{
				TestCase: testCase.Name,
				OriginalResult: originalResult,
				OptimizedResult: optimizedResult,
				Improvement: comparison,
			})
		}
	}

	return results, nil
}

func (br *BenchmarkRunner) parseBenchmarkOutput(output string) (map[string]BenchmarkResult, error) {
	results := make(map[string]BenchmarkResult)
	
	// regex to parse benchmark lines
	benchmarkRegex := regexp.MustCompile(`^(Benchmark\w+)-\d+\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)
	
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := benchmarkRegex.FindStringSubmatch(line)
		
		if len(matches) >= 4 {
			name := matches[1]
			iterations, _ := strconv.ParseInt(matches[2], 10, 64)
			nsPerOp, _ := strconv.ParseFloat(matches[3], 64)
			
			result := BenchmarkResult{
				Name:        name,
				Iterations:  iterations,
				NsPerOp:     nsPerOp,
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
	
	return results, nil
}
func (br *BenchmarkRunner) calculateImprovement(original, optimized BenchmarkResult) ImprovementMetrics {
	var improvement ImprovementMetrics

	// calc speed improvement 
	if optimized.NsPerOp > 0 {
		improvement.SpeedupFactor = original.NsPerOp / optimized.NsPerOp
		improvement.SpeedupPercentage = ((original.NsPerOp- optimized.NsPerOp)/ original.NsPerOp) * 100
	}

	// calc allocation reduction
	if original.AllocsPerOp > 0 {
		improvement.AllocReduction = ((float64(original.AllocsPerOp - optimized.AllocsPerOp)) / float64(original.AllocsPerOp)) * 100
	}

	// calc memory reduction 
	if original.BytesPerOp > 0 {
		improvement.MemoryReduction = ((float64(original.BytesPerOp - optimized.BytesPerOp)) / float64(original.BytesPerOp)) * 100
	}

	return improvement
}

// save benchmark results to JSON file for later analysis
func (br *BenchmarkRunner) SaveResults(results []ComparisonResult, filename string) error {
	if br.outputDir != "" {
		err := os.MkdirAll(br.outputDir, 0755)
		if err != nil {
			return err
		}
		filename = filepath.Join(br.outputDir, filename)
	}

	data, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// displays benchmark comparison results in a formatted table
func (br *BenchmarkRunner) PrintResults(results []ComparisonResult) {
	if len(results) == 0 {
		fmt.Println("No benchmark results to display")
		return 
	}

	fmt.Printf("\n Benchmark Results Summary\n")
	fmt.Printf("=" + strings.Repeat("=", 60) + "\n\n")

	totalTests := len(results)
	significantImprovements := 0 

	for _, result := range results {
		// speedup 20% > or alloc reduction >30%
		if result.Improvement.SpeedupPercentage > 20 || result.Improvement.AllocReduction > 30 {
			significantImprovements++
		}

		fmt.Printf(" Test Case: %s\n", result.TestCase)
		fmt.Printf(" Original: %8.0f ns/op | %4d allocs/op | %6d B/op\n", result.OriginalResult.NsPerOp,result.OriginalResult.AllocsPerOp,result.OriginalResult.BytesPerOp)
		fmt.Printf(" Optimized: %8.0f ns/op | %4d allocs/op | %6d B/op\n", result.OptimizedResult.NsPerOp, result.OptimizedResult.AllocsPerOp, result.OriginalResult.BytesPerOp)

		// Color-code improvements 
		speedIcon := br.getImprovementIcon(result.Improvement.SpeedupPercentage)
		allocIcon := br.getImprovementIcon(result.Improvement.AllocReduction)

		fmt.Printf(" %s Speed:	%5.1fx faster (%.1%% improvement)\n",speedIcon, result.Improvement.SpeedupFactor, result.Improvement.SpeedupPercentage)
		fmt.Printf(" %s Memory:	%.1f%% fewer allocations, %.1f%% less memory\n\n", allocIcon, result.Improvement.AllocReduction, result.Improvement.AllocReduction)
	}

	// Summary statistics 
	fmt.Printf(" Summary: %d/%d tests showed significant improvements (%.1f%%)\n", significantImprovements, totalTests, float64(significantImprovements)/float64(totalTests)*100)

	if significantImprovements > 0 {
		fmt.Printf(" Average performance gain: %.1fx faster with %.1f%% fewer allocations\n", br.calculateAverageSpeedup(results), br.calculateAverageAllocReduction(results))
	}
}

func (br *BenchmarkRunner) getImprovementIcon(percentage float64) string {
	switch {
	case percentage >= 50:
		return "[+++]"
	case percentage >= 20:
		return "[++]"
	case percentage >= 5:
		return "[+]"
	case percentage > 0:
		return "[~]"
	default:
		return "[-]"
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
func (br *BenchmarkRunner) CleanupBenchmarkFiles(suites []BenchmarkSuite) error {
	for _, suite := range suites {
		if err := os.Remove(suite.BenchmarkFile); err != nil {
			fmt.Printf("Warning: Could not remove %s: %v\n", suite.BenchmarkFile, err)
		}
	}
	return nil
}
