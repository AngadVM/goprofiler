package benchmark

import (
    "fmt"
    "go/ast"
    "go/format"  
    "go/parser"
    "go/token"
    "strings"
    
    "github.com/AngadVM/goprofiler/internal/types"
)

// extracting code blocks from Go source files
type CodeExtractor struct {
	fileSet *token.FileSet
}

// representing code from src file 
type ExtractedCode struct {
	OriginalFunc	*ast.FuncDecl
	ContainingFunc 	*ast.FuncDecl
	IssueLocation 	token.Pos
	SourceCode		string
	FuncName		string
	Parameters		[]*ast.Field
	ReturnType		ast.Expr
}

// creates new code extractor 
func NewCodeExtractor() *CodeExtractor {
	return &CodeExtractor{
		fileSet: token.NewFileSet(),
	}
}


// extract code block with issues 
func (ce *CodeExtractor) ExtractCodeForIssue(filePath string, issue types.Issue) (*ExtractedCode, error) {
	// parse src file 
	src, err := parser.ParseFile(ce.fileSet, filePath, nil, parser.ParseComments)

	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// finding func containing issue line 
	containingFunc := ce.findFunctionContainingLine(src, issue.Line)
	if containingFunc == nil {
		return nil, fmt.Errorf("could not find function containing line %d", issue.Line)
	}

	// convert func back to source code 
	funcSource, err := ce.astToSource(containingFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert AST to source: %w", err)
	}

	return &ExtractedCode{
		OriginalFunc: containingFunc,
		ContainingFunc: containingFunc,
		SourceCode: funcSource,
		FuncName: containingFunc.Name.Name,
		Parameters: containingFunc.Type.Params.List,
		ReturnType: ce.getReturnType(containingFunc),
	}, nil
}

// search for the fuction which contains given line no.
func (ce *CodeExtractor) findFunctionContainingLine(file *ast.File, targetLine int) *ast.FuncDecl {
	var result *ast.FuncDecl

	ast.Inspect(file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			start := ce.fileSet.Position(funcDecl.Pos()).Line
			end := ce.fileSet.Position(funcDecl.End()).Line

			if targetLine >= start && targetLine <= end {
				result = funcDecl
				return false // stop searching when found
			}
		}
		return true
	})

	return result
}

// convert AST node to source code
func (ce *CodeExtractor) astToSource(node ast.Node) (string, error) {
	var buf strings.Builder
	err := format.Node(&buf, ce.fileSet, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// return type from a function declaration
func (ce *CodeExtractor) getReturnType(funcDecl *ast.FuncDecl) ast.Expr {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) == 0 {
		return nil 
	}
	return funcDecl.Type.Results.List[0].Type
}
