package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicUseAnalyzer = &analysis.Analyzer{
	Name: "nopanicnoexit",
	Doc:  "check for panic use and exit/log.Fatal use outside main function",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {

	isPanicCall := func(call *ast.CallExpr) {
		if funIdent, ok := call.Fun.(*ast.Ident); ok {
			if isPanicFunction(funIdent) {
				pass.Reportf(call.Pos(), "panic use error")
			}
		}
	}

	isExitCall := func(call *ast.CallExpr) {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if isOSExitFunction(sel) {
				pass.Reportf(sel.Pos(), "os.Exit use error")
			}
		}
	}

	isLogFatalCall := func(call *ast.CallExpr) {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if isLogFatalFunction(sel) {
				pass.Reportf(sel.Pos(), "log.Fatal use error")
			}
		}
	}

	for _, file := range pass.Files {
		//packageName := file.Name.Name

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				isPanicCall(x)
				isExitCall(x)
				isLogFatalCall(x)
			}
			return true
		})
	}
	return nil, nil
}

func isPanicFunction(funIdent *ast.Ident) bool {
	return funIdent.Name == "panic"
}

func isOSExitFunction(sel *ast.SelectorExpr) bool {
	if x, ok := sel.X.(*ast.Ident); ok && x.Name == "os" && sel.Sel.Name == "Exit" {
		return true
	}
	return false
}

func isLogFatalFunction(sel *ast.SelectorExpr) bool {
	if x, ok := sel.X.(*ast.Ident); ok && x.Name == "log" && sel.Sel.Name == "Fatal" {
		return true
	}
	return false
}
