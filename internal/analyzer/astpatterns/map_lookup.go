package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type mapLookupPattern struct{}

func (p mapLookupPattern) Name() string { return "Inefficient Map Access in Loop" }
func (p mapLookupPattern) Description() string {
	return "Multiple map lookups with same key in loop body"
}
func (p mapLookupPattern) Impact() string { return "high" }

func (p mapLookupPattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
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

			mapAccesses := make(map[string][]token.Pos)

			ast.Inspect(loopBody, func(n ast.Node) bool {
				indexExpr, ok := n.(*ast.IndexExpr)
				if !ok {
					return true
				}

				var mapName, keyName string

				if ident, ok := indexExpr.X.(*ast.Ident); ok {
					mapName = ident.Name
				} else if selectorExpr, ok := indexExpr.X.(*ast.SelectorExpr); ok {
					mapName = getVarName(selectorExpr)
				}

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

					if len(mapAccesses[key]) == 2 {
						pos := fset.Position(indexExpr.Pos())
						issues = append(issues, types.Issue{
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
}

func init() {
	RegisterASTPattern(mapLookupPattern{})
}
