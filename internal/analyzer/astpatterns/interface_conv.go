package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type interfaceConvPattern struct{}

func (p interfaceConvPattern) Name() string { return "Interface Conversion in Loop" }
func (p interfaceConvPattern) Description() string {
	return "Type assertions and interface conversions inside loops"
}
func (p interfaceConvPattern) Impact() string { return "medium" }

func (p interfaceConvPattern) Detector() func(*token.FileSet, *ast.File) []types.Issue {
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

					issues = append(issues, types.Issue{
						Line:        pos.Line,
						Title:       "Type assertion in " + loopType + " loop",
						Description: "Type assertion " + varName + ".(" + typeName + ") inside loop may impact performance",
						Suggestion:  "Consider doing type assertion outside loop if possible",
						Impact:      "medium",
						Type:        "conversion",
					})

				case *ast.CallExpr:
					if ident, ok := expr.Fun.(*ast.Ident); ok {
						if isInterfaceConversion(ident.Name) {
							pos := fset.Position(expr.Pos())
							issues = append(issues, types.Issue{
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
}

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

func init() {
	RegisterASTPattern(interfaceConvPattern{})
}
