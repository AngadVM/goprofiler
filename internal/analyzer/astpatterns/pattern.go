package astpatterns

import (
	"go/ast"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type ASTPattern interface {
	Name() string
	Description() string
	Impact() string
	Detector() func(*token.FileSet, *ast.File) []types.Issue
}

var registeredASTPatterns []ASTPattern

func RegisterASTPattern(p ASTPattern) {
	registeredASTPatterns = append(registeredASTPatterns, p)
}

func GetASTPatterns() []ASTPattern {
	return registeredASTPatterns
}
