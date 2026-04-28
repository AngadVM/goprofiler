package astpatterns

import (
	"go/parser"
	"go/token"

	"github.com/AngadVM/goprofiler/internal/types"
)

type ASTAnalyzer struct {
	fset     *token.FileSet
	patterns []ASTPattern
}

func NewASTAnalyzer() *ASTAnalyzer {
	return &ASTAnalyzer{
		fset:     token.NewFileSet(),
		patterns: GetASTPatterns(),
	}
}

func (a *ASTAnalyzer) AnalyzeFile(filePath string) ([]types.Issue, error) {
	src, err := parser.ParseFile(a.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var allIssues []types.Issue

	for _, pattern := range a.patterns {
		issues := pattern.Detector()(a.fset, src)
		allIssues = append(allIssues, issues...)
	}

	return allIssues, nil
}
