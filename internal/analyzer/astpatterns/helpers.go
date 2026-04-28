package astpatterns

import (
	"go/ast"
	"strings"
)

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
