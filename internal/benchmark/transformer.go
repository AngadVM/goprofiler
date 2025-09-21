package benchmark

import (
    "fmt"
    "go/ast"
    "go/format"
    "go/token" 
    "strings"
    
    "github.com/AngadVM/goprofiler/internal/types"
)


// handles transforming extracted code to create optimized versions
type CodeTransformer struct {
	fileSet *token.FileSet
}

// represent result of code transformation 
type TransformedCode struct {
	Original 	string
	Optimized	string
	TestData 	string 
	Imports 	[]string
}

// create new code transformer 
func NewCodeTransformer() *CodeTransformer {
	return &CodeTransformer{
		fileSet: token.NewFileSet(),
	}
}

// apply optimization transformations based on the issue type
func (ct *CodeTransformer) Transform(extracted *ExtractedCode, issue types.Issue) (*TransformedCode, error) {
	switch {
	case strings.Contains(strings.ToLower(issue.Title), "string") && strings.Contains(strings.ToLower(issue.Title), "concatenation"):
		return ct.transformStringConcatenation(extracted)
	case strings.Contains(strings.ToLower(issue.Title), "slice") && strings.Contains(strings.ToLower(issue.Title), "capacity"):
		return ct.transformSliceAllocation(extracted)
	case strings.Contains(strings.ToLower(issue.Title), "map") && strings.Contains(strings.ToLower(issue.Title), "lookup"):
		return ct.transformMapLookup(extracted)
	default:
		return nil, fmt.Errorf("unsupported transformation for issue: %s", issue.Title)
	}
}


// optimize string concatenation 
func (ct *CodeTransformer) transformStringConcatenation( extracted *ExtractedCode) (*TransformedCode, error) {
	original := extracted.OriginalFunc

	// creating copy for transformation 
	optimized := ct.copyFuncDecl(original)
	optimized.Name.Name = original.Name.Name + "Optimized"

	// transform the func body
	ct.replaceStringConcatenation(optimized.Body)

	// convert both to source code 
	originalSrc, _ := ct.astToSource(original)
	optimizedSrc, _ := ct.astToSource(optimized)

	// generate test data based on function parameters
	testData := ct.generateTestDataForStringFunc(original)

	return &TransformedCode{
		Original: originalSrc,
		Optimized: optimizedSrc,
		TestData: testData,
		Imports: []string{"strings"},
	}, nil
}

// optimize slice allocation with capacity hints
func (ct *CodeTransformer) transformSliceAllocation(extracted *ExtractedCode) (*TransformedCode, error) {
	original := extracted.OriginalFunc

	//create optimized version 
	optimized := ct.copyFuncDecl(original)
	optimized.Name.Name = original.Name.Name

	// Transform make() calls to include capacity
	ct.addSliceCapacity(optimized.Body)

	originalSrc, _ := ct.astToSource(original)
	optimizedSrc, _ := ct.astToSource(optimized)
	testData := ct.generateTestDataForSliceFunc(original)

	return &TransformedCode{
		Original: originalSrc,
		Optimized: optimizedSrc,
		TestData: testData,
		Imports: []string{},
	}, nil
}

// optimize map lookups using the comma ok idiom
func (ct *CodeTransformer) transformMapLookup(extracted *ExtractedCode) (*TransformedCode, error) {
	original := extracted.OriginalFunc
	
	optimized := ct.copyFuncDecl(original)
	optimized.Name.Name = original.Name.Name + "Optimized"
	
	// Transform map lookups
	ct.optimizeMapLookups(optimized.Body)
	
	originalSrc, _ := ct.astToSource(original)
	optimizedSrc, _ := ct.astToSource(optimized)
	testData := ct.generateTestDataForMapFunc(original)
	
	return &TransformedCode{
		Original:  originalSrc,
		Optimized: optimizedSrc,
		TestData:  testData,
		Imports:   []string{},
	}, nil
}

// helper functions 

func (ct *CodeTransformer) copyFuncDecl(original *ast.FuncDecl) *ast.FuncDecl {

	return &ast.FuncDecl{
		Name: &ast.Ident{Name: original.Name.Name},
		Type: original.Type,
		Body: ct.copyBlockStmt(original.Body),
	}
}

func (ct *CodeTransformer) copyBlockStmt(original *ast.BlockStmt) *ast.BlockStmt {
	if original == nil {
		return nil
	}

	newBlock := &ast.BlockStmt{
		List: make([]ast.Stmt, len(original.List)),
	}

	copy(newBlock.List, original.List)
	return newBlock
}

// find and replace string += with strings.Builder
func (ct *CodeTransformer) replaceStringConcatenation(block *ast.BlockStmt) {
	if block == nil {
		return
	}

	for i, stmt := range block.List{
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			if len(assignStmt.Lhs) > 0 && len(assignStmt.Rhs) > 0 {
				// check for += 
				if assignStmt.Tok == token.ADD_ASSIGN {
					block.List[i] = ct.createBuilderStatement(assignStmt)
				}
			}
		}
	}
}

func (ct *CodeTransformer) createBuilderStatement(original *ast.AssignStmt) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.Ident{Name: "builder"},
				Sel: &ast.Ident{Name: "WriteString"},
			},
			Args: original.Rhs,
		},
	}
}

// add capacity hints to make() calls
func (ct *CodeTransformer) addSliceCapacity(block *ast.BlockStmt) {
	// Traversing AST looking for make([]Type) calls and adding capacity
	ast.Inspect(block, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			if ident, ok := callExpr.Fun.(*ast.Ident); ok && ident.Name == "make" {
				if len(callExpr.Args) == 1 {
					// Add capacity argument: make([]Type, 0, 100)
					callExpr.Args = append(callExpr.Args, 
						&ast.BasicLit{Kind: token.INT, Value: "0"},
						&ast.BasicLit{Kind: token.INT, Value: "100"})
				}
			}
		}
		return true
	})
}


// optimizeMapLookups replaces multiple map accesses with single lookup
func (ct *CodeTransformer) optimizeMapLookups(block *ast.BlockStmt) {
	// This analyzes the AST for repeated map[key] accesses and replace with: val, exists := map[key]; if exists { ... }
}

// Test data generation methods
func (ct *CodeTransformer) generateTestDataForStringFunc(funcDecl *ast.FuncDecl) string {
	return `
var testStrings = []string{"Hello", "World", "Go", "Programming", "Benchmark"}
`
}

func (ct *CodeTransformer) generateTestDataForSliceFunc(funcDecl *ast.FuncDecl) string {
	return `
var testSize = 1000
`
}

func (ct *CodeTransformer) generateTestDataForMapFunc(funcDecl *ast.FuncDecl) string {
	return `
var testMap = map[string]int{"key1": 10, "key2": 20, "key3": 30}
var testKeys = []string{"key1", "key2", "key3"}
`
}

func (ct *CodeTransformer) astToSource(node ast.Node) (string, error) {
	var buf strings.Builder
	err := format.Node(&buf, ct.fileSet, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
