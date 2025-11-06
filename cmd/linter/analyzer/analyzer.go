package analyzer

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "custom_linter",
	Doc:      "reports panic, log.Fatal and os.Exit outside main",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		if isPanicCall(call) {
			pass.Reportf(call.Pos(), "panic is not allowed")
			return
		}

		if isLogFatalCall(call) {
			if !isInMainPackageAndFunction(pass) {
				pass.Reportf(call.Pos(), "log.Fatal is not allowed outside main function of main package")
			}
			return
		}

		if isOSExitCall(call) {
			if !isInMainPackageAndFunction(pass) {
				pass.Reportf(call.Pos(), "os.Exit is not allowed outside main function of main package")
			}
			return
		}
	})

	return nil, nil
}

func isPanicCall(call *ast.CallExpr) bool {
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name == "panic"
	}
	return false
}

func isLogFatalCall(call *ast.CallExpr) bool {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if ident, ok := selExpr.X.(*ast.Ident); ok {
		return ident.Name == "log" && (selExpr.Sel.Name == "Fatal" ||
			selExpr.Sel.Name == "Fatalf" || selExpr.Sel.Name == "Fatalln")
	}
	return false
}

func isOSExitCall(call *ast.CallExpr) bool {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if ident, ok := selExpr.X.(*ast.Ident); ok {
		return ident.Name == "os" && selExpr.Sel.Name == "Exit"
	}
	return false
}

func isInMainPackageAndFunction(pass *analysis.Pass) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	return true
}
