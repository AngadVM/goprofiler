# goprofiler

A CLI tool that analyzes Go source code to detect performance bottlenecks and suggests optimizations.

## Features

- **Static Analysis**: Detects common performance anti-patterns
- **AST-based Detection**: Uses Go's AST for accurate pattern matching
- **Benchmark Generation**: Creates benchmarks comparing original vs optimized code
- **Multiple Output Formats**: Console (colored) and JSON output

## Detected Issues

| Issue | Impact | Description |
|-------|--------|-------------|
| String concatenation in loop | High | Using `+=` for strings in loops |
| Slice without capacity | Medium | Missing capacity hint in `make()` |
| Multiple map lookups | High | Same key accessed multiple times in loop |
| Defer in loop | High | Defers accumulate causing memory buildup |
| Function call in loop condition | Medium | Function called on every iteration |
| Interface conversion in loop | Medium | Type assertions inside loops |

## Installation

### From Source

```bash
git clone https://github.com/AngadVM/goprofiler.git
cd goprofiler
go install
```

### Using Go Install

```bash
go install github.com/AngadVM/goprofiler@latest
```

### Download Binary

Download the appropriate binary from the [releases page](https://github.com/AngadVM/goprofiler/releases).

## Usage

### Quick Check

Quick scan for critical issues:

```bash
goprofiler check ./path/to/code
```

### Analyze

Full analysis with detailed output:

```bash
# Basic analysis
goprofiler analyze ./path/to/code

# Verbose output with suggestions
goprofiler analyze ./path/to/code --verbose

# JSON output (for CI/automation)
goprofiler analyze ./path/to/code --output json

# Analyze a single file
goprofiler analyze ./path/to/file.go
```

### Benchmark

Generate and run benchmarks comparing original vs optimized implementations:

```bash
# Analyze and run benchmarks
goprofiler benchmark ./path/to/code

# Only generate benchmark files (without running)
goprofiler benchmark ./path/to/code --generate-only

# Keep benchmark files after running
goprofiler benchmark ./path/to/code --cleanup=false

# Save results to file
goprofiler benchmark ./path/to/code --save-results results.json
```

### Shortcuts

```bash
goprofiler c ./path/to/code  # check
goprofiler a ./path/to/code  # analyze  
goprofiler b ./path/to/code  # benchmark
```

## Examples

### Analyze a directory

```bash
$ goprofiler analyze ./examples/simple
 GoProfiler - Analyzing: ./examples/simple
========================================
[*] Analysis Summary:
    Files analyzed: 1
    Total issues: 2
    High impact: 1 | Medium: 1 | Low: 0

>> ./examples/simple/slow_code.go
   [!] Line 8: String concatenation in loop (high impact)
       Suggestion: Use strings.Builder for better performance
   [*] Line 14: Slice allocated without capacity (medium impact)
       Suggestion: Use make([]Type, 0, capacity) if you know the expected size
```

### Quick check

```bash
$ goprofiler check ./examples/problematic
[CHECK] Quick analysis: ./examples/problematic
[!] Found 4 high-impact performance issues
    Run 'goprofiler analyze' for details
```

## Project Structure

```
goprofiler/
├── cmd/goprofiler/      # CLI entry point
├── internal/
│   ├── analyzer/        # Core analysis engine
│   │   ├── astpatterns/ # AST-based pattern detection
│   │   └── stringpatterns/ # Regex-based pattern detection
│   ├── benchmark/       # Benchmark generation and execution
│   ├── output/         # Output formatting
│   └── types/          # Shared types
├── examples/           # Example code with known issues
└── .github/            # CI/CD workflows
```

## Development

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build -o goprofiler ./cmd/goprofiler
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.