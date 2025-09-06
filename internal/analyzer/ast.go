package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ASTAnalyzer - handles AST-based performance analysis
type ASTAnalyzer struct {
	fset     *token.FileSet
	patterns []ASTPattern
}

// ASTPattern - defines AST-based performance patterns
type ASTPattern struct {
	Name        string
	Description string
	Impact      string
	Detector    func(*token.FileSet, *ast.File) []Issue
}

// NewASTAnalyzer creates an analyzer with AST-based patterns
func NewASTAnalyzer() *ASTAnalyzer {
	return &ASTAnalyzer{
		fset:     token.NewFileSet(),
		patterns: getASTPatterns(),
	}
}

// getASTPatterns returns sophisticated AST-based performance patterns
func getASTPatterns() []ASTPattern {
	return []ASTPattern{
		{
			Name:        "Inefficient Map Access in Loop",
			Description: "Multiple map lookups with same key in loop body",
			Impact:      "high",
			Detector:    detectInefficientMapAccess,
		},
		{
			Name:        "Defer in Loop",
			Description: "Defer statements inside loops can cause memory buildup",
			Impact:      "high",
			Detector:    detectDeferInLoop,
		},
		{
			Name:        "Interface Conversion in Loop",
			Description: "Type assertions and interface conversions inside loops",
			Impact:      "medium",
			Detector:    detectInterfaceConversionInLoop,
		},
		{
			Name:        "Function Call in Loop Condition",
			Description: "Function calls in loop conditions are evaluated every iteration",
			Impact:      "medium",
			Detector:    detectFunctionCallInLoopCondition,
		},
	}
}

// AnalyzeFileAST analyzes a single Go file using AST patterns
func (a *ASTAnalyzer) AnalyzeFileAST(filePath string) (AnalysisResult, error) {
	src, err := parser.ParseFile(a.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return AnalysisResult{}, err
	}

	var allIssues []Issue

	// Run all AST pattern detectors
	for _, pattern := range a.patterns {
		issues := pattern.Detector(a.fset, src)
		allIssues = append(allIssues, issues...)
	}

	return AnalysisResult{
		FilePath: filePath,
		Issues:   allIssues,
	}, nil
}

// detectInefficientMapAccess finds multiple map accesses with same key in loops
func detectInefficientMapAccess(fset *token.FileSet, file *ast.File) []Issue {
	var issues []Issue
	
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for for loops (both regular and range)
		var loopBody *ast.BlockStmt
		var loopType string
		
		if forStmt, ok := n.(*ast.ForStmt); ok {
			loopBody = forStmt.Body
			loopType = "for"
		} else if rangeStmt, ok := n.(*ast.RangeStmt); ok {
			loopBody = rangeStmt.Body
			loopType = "range"
		} else {
			return true
		}

		// Track map accesses in loop body
		mapAccesses := make(map[string][]token.Pos)
		
		ast.Inspect(loopBody, func(n ast.Node) bool {
			indexExpr, ok := n.(*ast.IndexExpr)
			if !ok {
				return true
			}

			// Check if this is a map access
			var mapName, keyName string
			
			// Get map name
			if ident, ok := indexExpr.X.(*ast.Ident); ok {
				mapName = ident.Name
			} else if selectorExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
				mapName = getVarName(selectorExpr)
			}
			
			// Get key name
			if keyIdent, ok := indexExpr.Index.(*ast.Ident); ok {
				keyName = keyIdent.Name
			} else if keyLit, ok := indexExpr.Index.(*ast.BasicLit); ok {
				keyName = keyLit.Value
			} else if keySelector, ok := indexExpr.Index.(*ast.SelectorExpr); ok {
				keyName = getVarName(keySelector)
			}
			
			if mapName != "" && keyName != "" {
				key := mapName + "[" + keyName + "]"
				mapAccesses[key] = append(mapAccesses[key], indexExpr.Pos())
				
				// If we've seen this key access more than once, flag it
				if len(mapAccesses[key]) == 2 { // Only report once per duplicate
					pos := fset.Position(indexExpr.Pos())
					issues = append(issues, Issue{
						Line:        pos.Line,
						Title:       "Multiple map lookups with same key",
						Description: "Map key '" + key + "' is accessed multiple times in " + loopType + " loop",
						Suggestion:  "Store the value in a variable: val, exists := " + key + "; if exists { ... }",
						Impact:      "high",
						Type:        "lookup",
					})
				}
			}
			
			return true
		})
		
		return true
	})
	
	return issues
}

// detectDeferInLoop finds defer statements inside loops
func detectDeferInLoop(fset *token.FileSet, file *ast.File) []Issue {
	var issues []Issue
	
	ast.Inspect(file, func(n ast.Node) bool {
		var loopBody *ast.BlockStmt
		var loopType string
		
		// Handle both for loops and range loops
		if forStmt, ok := n.(*ast.ForStmt); ok {
			loopBody = forStmt.Body
			loopType = "for"
		} else if rangeStmt, ok := n.(*ast.RangeStmt); ok {
			loopBody = rangeStmt.Body
			loopType = "range"
		} else {
			return true
		}

		// Look for defer statements in the loop body
		ast.Inspect(loopBody, func(n ast.Node) bool {
			if deferStmt, ok := n.(*ast.DeferStmt); ok {
				pos := fset.Position(deferStmt.Pos())
				
				// Get function name being deferred
				var funcName string
				switch call := deferStmt.Call.Fun.(type) {
				case *ast.Ident:
					funcName = call.Name
				case *ast.SelectorExpr:
					funcName = getVarName(call)
				default:
					funcName = "function"
				}
				
				issues = append(issues, Issue{
					Line:        pos.Line,
					Title:       "Defer statement in " + loopType + " loop",
					Description: "defer " + funcName + "() in loops causes deferred calls to accumulate until function returns",
					Suggestion:  "Move defer outside loop, use explicit cleanup, or wrap in anonymous function",
					Impact:      "high",
					Type:        "memory",
				})
			}
			return true
		})
		
		return true
	})
	
	return issues
}

// detectInterfaceConversionInLoop finds type assertions and interface conversions in loops
func detectInterfaceConversionInLoop(fset *token.FileSet, file *ast.File) []Issue {
	var issues []Issue
	
	ast.Inspect(file, func(n ast.Node) bool {
		// Look for loops
		var loopBody *ast.BlockStmt
		var loopType string
		
		if forStmt, ok := n.(*ast.ForStmt); ok {
			loopBody = forStmt.Body
			loopType = "for"
		} else if rangeStmt, ok := n.(*ast.RangeStmt); ok {
			loopBody = rangeStmt.Body  
			loopType = "range"
		} else {
			return true
		}

		// Look for type assertions in loop body
		ast.Inspect(loopBody, func(n ast.Node) bool {
			switch expr := n.(type) {
			case *ast.TypeAssertExpr:
				pos := fset.Position(expr.Pos())
				var varName string
				if ident, ok := expr.X.(*ast.Ident); ok {
					varName = ident.Name
				} else {
					varName = "value"
				}
				
				typeName := getTypeName(expr.Type)
				
				issues = append(issues, Issue{
					Line:        pos.Line,
					Title:       "Type assertion in " + loopType + " loop",
					Description: "Type assertion " + varName + ".(" + typeName + ") inside loop may impact performance",
					Suggestion:  "Consider doing type assertion outside loop if possible",
					Impact:      "medium",
					Type:        "conversion",
				})
				
			case *ast.CallExpr:
				// Check for interface{} conversions
				if ident, ok := expr.Fun.(*ast.Ident); ok {
					// Common interface conversion patterns
					if isInterfaceConversion(ident.Name) {
						pos := fset.Position(expr.Pos())
						issues = append(issues, Issue{
							Line:        pos.Line,
							Title:       "Interface conversion in " + loopType + " loop",
							Description: "Interface conversion " + ident.Name + "() inside loop may cause allocations",
							Suggestion:  "Consider moving conversion outside loop or using more specific types",
							Impact:      "medium",
							Type:        "conversion",
						})
					}
				}
			}
			return true
		})
		
		return true
	})
	
	return issues
}

// detectFunctionCallInLoopCondition finds function calls in loop conditions
func detectFunctionCallInLoopCondition(fset *token.FileSet, file *ast.File) []Issue {
	var issues []Issue
	
	ast.Inspect(file, func(n ast.Node) bool {
		if forStmt, ok := n.(*ast.ForStmt); ok && forStmt.Cond != nil {
			// Check if condition contains function calls
			ast.Inspect(forStmt.Cond, func(n ast.Node) bool {
				if callExpr, ok := n.(*ast.CallExpr); ok {
					pos := fset.Position(callExpr.Pos())
					
					var funcName string
					switch fun := callExpr.Fun.(type) {
					case *ast.Ident:
						funcName = fun.Name
					case *ast.SelectorExpr:
						funcName = getVarName(fun)
					default:
						funcName = "function"
					}
					
					// Skip built-in functions that are typically optimized
					if !isOptimizedBuiltin(funcName) {
						issues = append(issues, Issue{
							Line:        pos.Line,
							Title:       "Function call in loop condition",
							Description: "Function " + funcName + "() is called on every loop iteration",
							Suggestion:  "Store result in variable before loop: limit := " + funcName + "()",
							Impact:      "medium",
							Type:        "performance",
						})
					}
				}
				return true
			})
		}
		return true
	})
	
	return issues
}

// Helper functions

// getVarName extracts variable name from expression
func getVarName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return getVarName(e.X) + "." + e.Sel.Name
	case *ast.IndexExpr:
		return getVarName(e.X) + "[...]"
	default:
		return "expr"
	}
}

// getTypeName extracts type name from type expression
func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return getVarName(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + getTypeName(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeName(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeName(t.Key) + "]" + getTypeName(t.Value)
	default:
		return "interface{}"
	}
}

// isInterfaceConversion checks if function name indicates interface conversion
func isInterfaceConversion(funcName string) bool {
	conversions := []string{
		"interface{}", "any", "fmt.Sprint", "fmt.Sprintf", 
		"reflect.ValueOf", "json.Marshal",
	}
	
	for _, conv := range conversions {
		if strings.Contains(funcName, conv) {
			return true
		}
	}
	return false
}

// isOptimizedBuiltin checks if function is a built-in that's typically optimized
func isOptimizedBuiltin(funcName string) bool {
	builtins := []string{
		"len", "cap", "make", "new", "append", "copy", "delete",
		"real", "imag", "complex", "close", "panic", "recover",
	}
	
	for _, builtin := range builtins {
		if funcName == builtin {
			return true
		}
	}
	return false
}
