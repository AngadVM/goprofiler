# Contributing

Welcome! This guide explains how to add new performance detection patterns to goprofiler.

## Architecture Overview

goprofiler has two types of pattern detectors:

1. **String Patterns** (`internal/analyzer/stringpatterns/`) - Simple regex/text-based detection
2. **AST Patterns** (`internal/analyzer/astpatterns/`) - Deep analysis using Go's AST

## Adding a String Pattern

String patterns analyze source code as plain text. Use these for simple pattern matching.

### Steps

1. Create a new file in `internal/analyzer/stringpatterns/`:

```go
// my_pattern.go
package stringpatterns

import (
    "strings"
    "github.com/AngadVM/goprofiler/internal/types"
)

type myPattern struct{}

func (p myPattern) Name() string        { return "My Pattern Name" }
func (p myPattern) Description() string { return "What this pattern detects" }
func (p myPattern) Impact() string     { return "high" | "medium" | "low" }

func (p myPattern) Detector() func(source string) []types.Issue {
    return func(source string) []types.Issue {
        var issues []types.Issue
        lines := strings.Split(source, "\n")
        
        // Your detection logic here
        for i, line := range lines {
            if strings.Contains(line, "pattern to detect") {
                issues = append(issues, types.Issue{
                    Line:        i + 1,
                    Title:       "Issue Title",
                    Description: "Description of the issue",
                    Suggestion:  "How to fix it",
                    Impact:      "high",
                    Type:        "category",
                })
            }
        }
        
        return issues
    }
}

func init() {
    RegisterPattern(myPattern{})
}
```

2. That's it! The pattern is auto-registered via `init()`

## Adding an AST Pattern

AST patterns use Go's Abstract Syntax Tree for deeper analysis. Use these when you need to understand code structure.

### Steps

1. Create a new file in `internal/analyzer/astpatterns/`:

```go
// my_ast_pattern.go
package astpatterns

import (
    "go/ast"
    "go/token"
    "github.com/AngadVM/goprofiler/internal/types"
)

type myASTPattern struct{}

func (p myASTPattern) Name() string        { return "My AST Pattern" }
func (p myASTPattern) Description() string { return "What this pattern detects" }
func (p myASTPattern) Impact() string     { return "high" | "medium" | "low" }

func (p myASTPattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
    return func(fset *token.FileSet, file *ast.File) []types.Issue {
        var issues []types.Issue
        
        // Use ast.Inspect to traverse the AST
        ast.Inspect(file, func(n ast.Node) bool {
            // Your detection logic here
            // Return true to continue traversal, false to stop
            
            // Example: detect for loops
            if forStmt, ok := n.(*ast.ForStmt); ok {
                // Analyze the loop
                pos := fset.Position(forStmt.Pos())
                issues = append(issues, types.Issue{
                    Line:        pos.Line,
                    Title:       "Issue Title",
                    Description: "Description",
                    Suggestion:  "How to fix",
                    Impact:      "high",
                    Type:        "category",
                })
            }
            
            return true
        })
        
        return issues
    }
}

func init() {
    RegisterASTPattern(myASTPattern{})
}
```

2. Done! The pattern is auto-registered

## Pattern Guidelines

### Impact Levels

- **high**: Significant performance impact (e.g., string concat in loop)
- **medium**: Moderate impact (e.g., unnecessary allocations)
- **low**: Minor impact (e.g., minor inefficiencies)

### Issue Types

Common types:
- `allocation` - Memory allocation issues
- `loop` - Loop-related issues  
- `lookup` - Map/array access issues
- `memory` - Memory management issues
- `conversion` - Type conversion issues
- `performance` - General performance issues

## Testing

Add tests for your pattern in the same package:

```go
func TestMyPattern(t *testing.T) {
    patterns := GetPatterns() // for AST or GetPatterns() for string
    var pattern // find your pattern
    
    detector := pattern.Detector()
    issues := detector(source)
    
    if len(issues) != 1 {
        t.Errorf("expected 1 issue, got %d", len(issues))
    }
}
```

## Running Tests

```bash
go test ./...
```

## Code Style

- Use meaningful variable names
- Add comments for complex logic
- Keep patterns focused on a single issue type
- Follow existing code conventions
