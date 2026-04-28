package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type regexpCompilePattern struct{}

func (p regexpCompilePattern) Name() string { return "Regexp Compiled in Function" }
func (p regexpCompilePattern) Description() string {
	return "regexp.MustCompile called inside a function body compiles on every call"
}
func (p regexpCompilePattern) Impact() string { return "high" }

func (p regexpCompilePattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
	return func(fset *token.FileSet, file *ast.File) []types.Issue {
		var issues []types.Issue

		ast.Inspect(file, func(n ast.Node) bool {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok || funcDecl.Body == nil {
				return true
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "regexp" {
					if sel.Sel.Name == "MustCompile" {
						pos := fset.Position(callExpr.Pos())
						issues = append(issues, types.Issue{
							Line:        pos.Line,
							Title:       "Regexp compiled inside function",
							Description: "regexp.MustCompile() is called inside function '" + funcDecl.Name.Name + "'. This recompiles the regex every time the function is called.",
							Suggestion:  "Move regexp.MustCompile() to package level or use sync.Once for one-time compilation",
							Impact:      "high",
							Type:        "compilation",
						})
					}
				}

				return true
			})

			return true
		})

		return issues
	}
}

func init() {
	RegisterASTPattern(regexpCompilePattern{})
}
