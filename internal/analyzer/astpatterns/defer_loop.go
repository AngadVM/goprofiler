package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type deferLoopPattern struct{}

func (p deferLoopPattern) Name() string { return "Defer in Loop" }
func (p deferLoopPattern) Description() string {
	return "Defer statements inside loops can cause memory buildup"
}
func (p deferLoopPattern) Impact() string { return "high" }

func (p deferLoopPattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
	return func(fset *token.FileSet, file *ast.File) []types.Issue {
		var issues []types.Issue

		ast.Inspect(file, func(n ast.Node) bool {
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

			ast.Inspect(loopBody, func(n ast.Node) bool {
				if deferStmt, ok := n.(*ast.DeferStmt); ok {
					pos := fset.Position(deferStmt.Pos())

					var funcName string
					switch call := deferStmt.Call.Fun.(type) {
					case *ast.Ident:
						funcName = call.Name
					case *ast.SelectorExpr:
						funcName = getVarName(call)
					default:
						funcName = "function"
					}

					issues = append(issues, types.Issue{
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
}

func init() {
	RegisterASTPattern(deferLoopPattern{})
}
