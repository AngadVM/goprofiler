package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/AngadVM/goprofiler/internal/analyzer/astpatterns"
	"github.com/AngadVM/goprofiler/internal/analyzer/stringpatterns"
	"github.com/AngadVM/goprofiler/internal/types"
)

// handles performance analysis of Go code
type Analyzer struct {
	stringAnalyzer *stringpatterns.StringAnalyzer
	astAnalyzer    *astpatterns.ASTAnalyzer
}

// creates a new analyzer with both string and AST patterns
func New() *Analyzer {
	return &Analyzer{
		stringAnalyzer: stringpatterns.NewStringAnalyzer(),
		astAnalyzer:    astpatterns.NewASTAnalyzer(),
	}
}

//  analyzes a file or directory for performance issues
func (a *Analyzer) AnalyzePath(target string) ([]types.AnalysisResult, error) {
	var results []types.AnalysisResult

	// check if target is a file/dir
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		// analyze all .go files in dir
		err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".go" {
				result, err := a.analyzeFile(path)
				if err != nil {
					return err
				}
				results = append(results, result)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// analyze single file
		result, err := a.analyzeFile(target)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzes a single Go file using both string and AST patterns
func (a *Analyzer) analyzeFile(filePath string) (types.AnalysisResult, error) {
	var allIssues []types.Issue

	// run string-based patterns
	stringIssues, err := a.stringAnalyzer.AnalyzeFile(filePath)
	if err != nil {
		return types.AnalysisResult{}, err
	}
	allIssues = append(allIssues, stringIssues...)

	// run AST-based patterns
	astIssues, err := a.astAnalyzer.AnalyzeFile(filePath)
	if err != nil {
		
		fmt.Printf("Warning: AST analysis for %s failed: %v\n", filePath, err)
	} else {
		allIssues = append(allIssues, astIssues...)
	}

	return types.AnalysisResult{
		FilePath: filePath,
		Issues:   allIssues,
	}, nil
}
