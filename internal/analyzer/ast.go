package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type ASTAnalyzer struct {
	fileSet *token.FileSet
	issues []Issue
}

// creates a new AST analyzer
func NewASTAnalyzer() *ASTAnalyzer {
	return &ASTAnalyzer{
		fileSet: token.NewFileSet(),
		issues: []Issue{},
	}
}

// analyzes a Go file using AST parsing
func (a *ASTAnalyzer) AnalyzeFile(filepath string) ([]Issue, error) {
	// Reset issues for the file 
	a.issues = []Issue{}

	// Parse the Go file into AST
	node, err := parser.ParseFile(a.fileSet, filepath, nil,parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(node, a.InspectNode)

	return a.issues, nil
}

// called for every node in the AST
func (a *ASTAnalyzer) InspectNode(node ast.Node) bool {
	if node == nil {
		return false
	}

	// Checking diff types of AST nodes for issues
	a.checkForLoopStringConcatenation(node)
	a.checkForSliceAllocation(node)
	a.checkForGoroutineLeaks(node)

	return true // continue traversing
}

// detects string concatenation 
func (a *ASTAnalyzer) checkForLoopStringConcatenation(node ast.Node) {

	forStmt, ok := node.(*ast.ForStmt)
	if!ok {
		return 
	}

	ast.Inspect(forStmt.Body, func(n ast.Node) bool {
		assignStmt, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		// Check for += operator
		if assignStmt.Tok == token.ADD_ASSIGN{
			// Checking if rhs involves strings
			if a.containsStringOperation(assignStmt.Rhs){
				pos := a.fileSet.Position(assignStmt.Pos())
				a.issues = append(a.issues, Issue{
					Line: pos.Line,
					Title: "String concatenation in loop",
					Description: "Using += for string concatenation in loops creates many temporary strings",
					Suggestion: "Use strings.Builder for 3x better performance: var b strings.Builder; b.WriteString(...)",
					Impact: "high",
					Type: "allocation",
				})
			}
		}

		return true
	})
}

// detects inefficient slice allocations
func (a *ASTAnalyzer) checkForSliceAllocation(node ast.Node) {
	assignStmt, ok := node.(*ast.AssignStmt)
	if !ok {
		return 
	}

	for _, rhs := range assignStmt.Rhs {
		callExpr, ok := rhs.(*ast.CallExpr)
		if !ok {
			continue 
		}

		// Check if it's a make() call 
		if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "make" {
			if len(callExpr.Args) >= 1 {
				// check if first arg is slice type
				if a.isSliceType(callExpr.Args[0]) {
					// If only one arg(type), suggest adding capacity
					if len(callExpr.Args) == 1 {
						pos := a.fileSet.Position(callExpr.Pos())
						a.issues = append(a.issues, Issue{
							Line: pos.Line,
							Title: "Slice allocated without capacity hint",
							Description: "Slice will be reallocated and copied as it grows",
							Suggestion: "If you know expected size, use: make([]Type, 0, expectedCapacity)",
							Impact: "medium",
							Type: "allocation",
						})
					}
				}
			}
		}
	}
}


// detects potential goroutines leaks 
func (a *ASTAnalyzer) checkForGoroutineLeaks(node ast.Node) {
	// check for go statements
	goStmt, ok := node.(*ast.GoStmt)
	if !ok {
		return 
	}

	// Checking if it is a function call that might run forever
	if callExpr, ok := goStmt.Call.Fun.(*ast.FuncLit); ok {
		// Look for infinite loops or channels without proper cleanup
		hasInfiniteLoop := false
		hasChannelOperation := false

		ast.Inspect(callExpr.Body, func(n ast.Node) bool {
			// Check for 'for' loops without clear exit condition
			if forStmt, ok := n.(*ast.ForStmt); ok {
				if forStmt.Cond == nil {
					hasInfiniteLoop = true
				}
			}

			// Check for channel operations
			if _, ok := n.(*ast.SendStmt); ok {
				hasChannelOperation = true
			}
			return true
		})

		if hasInfiniteLoop && !hasChannelOperation {
			pos := a.fileSet.Position(goStmt.Pos())
			a.issues = append(a.issues, Issue{
				Line: pos.Line,
				Title: "Potential goroutine leaks",
				Description: "Goroutine with infinite loop and no channel communication may leak",
				Suggestion: "Add proper exit condition or context cancellation",
				Impact: "medium",
				Type: "goroutine",
			})
		}
	}
}

// Helper functions

func (a *ASTAnalyzer) containsStringOperation(expressions []ast.Expr) bool {
	for _, expr := range expressions {
		switch e := expr.(type) {
		case *ast.BasicLit:
			// String literal like "hello"
			if e.Kind.String() == "STRING" {
				return true
			}
		case *ast.BinaryExpr:
			// Binary exp like a+b 
			if e.Op.String() == "+" {
				return true
			}
		case *ast.Ident:
			return true
		}
	}
	return false
}

func (a *ASTAnalyzer) isSliceType(expr ast.Expr) bool {
	arrayType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}

	return arrayType.Len == nil 
}

// returns human readable name for the pattern type
func (a *ASTAnalyzer) GetPatternName(issueType string) string {
	patterns := map[string]string{
		"allocation": "Memory Allocation",
		"goroutine": "Concurrency Issue",
		"loop": "Loop Optimization",
		"io": "I/O Efficiency",
	}

	if name, exists := patterns[issueType]; exists {
		return name
	}

	return strings.Title(issueType)
}
