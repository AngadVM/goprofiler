package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type fmtSprintfPattern struct{}

func (p fmtSprintfPattern) Name() string { return "fmt.Sprintf in Loop" }
func (p fmtSprintfPattern) Description() string {
	return "fmt.Sprintf called inside a loop to build strings causes repeated allocations"
}
func (p fmtSprintfPattern) Impact() string { return "high" }

func (p fmtSprintfPattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
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
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "fmt" && sel.Sel.Name == "Sprintf" {
					pos := fset.Position(callExpr.Pos())
					issues = append(issues, types.Issue{
						Line:        pos.Line,
						Title:       "fmt.Sprintf in " + loopType + " loop",
						Description: "fmt.Sprintf() inside a loop allocates a new string on every iteration",
						Suggestion:  "Use strings.Builder with fmt.Fprintf or strconv.Append* to build the string incrementally",
						Impact:      "high",
						Type:        "allocation",
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
	RegisterASTPattern(fmtSprintfPattern{})
}
