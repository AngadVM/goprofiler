package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type loopConditionPattern struct{}

func (p loopConditionPattern) Name() string { return "Function Call in Loop Condition" }
func (p loopConditionPattern) Description() string {
	return "Function calls in loop conditions are evaluated every iteration"
}
func (p loopConditionPattern) Impact() string { return "medium" }

func (p loopConditionPattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
	return func(fset *token.FileSet, file *ast.File) []types.Issue {
		var issues []types.Issue

		ast.Inspect(file, func(n ast.Node) bool {
			if forStmt, ok := n.(*ast.ForStmt); ok && forStmt.Cond != nil {
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

						if !isOptimizedBuiltin(funcName) {
							issues = append(issues, types.Issue{
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
}

func init() {
	RegisterASTPattern(loopConditionPattern{})
}
