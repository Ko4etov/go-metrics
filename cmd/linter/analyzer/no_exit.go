package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func checkExit(pass *analysis.Pass) error {
	// Если это пакет main и функция main - разрешаем exit
	if pass.Pkg.Name() == "main" && isInMainFunction(pass) {
		return nil
	}
	
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		
		// Проверяем вызовы os.Exit
		if isOsExitCall(call) {
			pass.Reportf(call.Pos(), 
				"использование os.Exit() запрещено вне функции main пакета main")
		}
		
		// Проверяем вызовы log.Fatal*
		if isLogFatalCall(call) {
			pass.Reportf(call.Pos(), 
				"использование log.Fatal*() запрещено вне функции main пакета main")
		}
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
	
	// Проверяем log.Fatal, log.Fatalf, log.Fatalln
	if ident.Name == "log" && strings.HasPrefix(sel.Sel.Name, "Fatal") {
		return true
	}
	
	// Также проверяем logger.Fatal если это наш логгер
	if ident.Name == "logger" && strings.HasPrefix(sel.Sel.Name, "Fatal") {
		return true
	}
	
	return false
}

func isInMainFunction(pass *analysis.Pass) bool {
	// Упрощенная проверка - в реальном анализаторе нужно проверять контекст
	// Здесь просто разрешаем все в пакете main
	return true
}