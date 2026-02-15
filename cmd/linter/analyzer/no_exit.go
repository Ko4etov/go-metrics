package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func checkExit(pass *analysis.Pass) error {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.CallExpr)(nil),
	}
	
	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		
		// Проверяем os.Exit
		if isOsExitCall(call) {
			if !isInMainFunctionCall(pass, stack) {
				pass.Reportf(call.Pos(), 
					"использование os.Exit() запрещено вне функции main пакета main")
			}
			return true
		}
		
		// Проверяем log.Fatal*
		if isLogFatalCall(call) {
			if !isInMainFunctionCall(pass, stack) {
				pass.Reportf(call.Pos(), 
					"использование log.Fatal*() запрещено вне функции main пакета main")
			}
			return true
		}
		
		return true
	})
	
	return nil
}

func isOsExitCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	
	return ident.Name == "os" && sel.Sel.Name == "Exit"
}

func isLogFatalCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	
	if ident.Name == "log" && strings.HasPrefix(sel.Sel.Name, "Fatal") {
		return true
	}
	
	if ident.Name == "logger" && strings.HasPrefix(sel.Sel.Name, "Fatal") {
		return true
	}
	
	return false
}

func isInMainFunctionCall(pass *analysis.Pass, stack []ast.Node) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}
	
	for i := len(stack) - 1; i >= 0; i-- {
		node := stack[i]
		
		if fn, ok := node.(*ast.FuncDecl); ok {
			return fn.Name.Name == "main"
		}
	}
	
	return false
}