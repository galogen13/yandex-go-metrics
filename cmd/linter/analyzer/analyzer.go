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

type codeContext struct {
	packageName string
	funcName    string
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
		cc := codeContext{
			packageName: file.Name.Name,
			funcName:    "",
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.CallExpr:
				isPanicCall(x)
				if !isInMainPackage(cc) || !isInMainFunc(cc) {
					isExitCall(x)
					isLogFatalCall(x)
				}
			case *ast.FuncDecl:
				cc.funcName = x.Name.Name
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

func isInMainFunc(cc codeContext) bool {
	return cc.funcName == "main"
}

func isInMainPackage(cc codeContext) bool {
	return cc.packageName == "main"
}
