package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func checkPanic(pass *analysis.Pass) error {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	
	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		
		if isPanicCall(call) {
			pass.Reportf(call.Pos(), 
				"использование panic() запрещено, используйте возврат ошибок")
		}
	})
	
	return nil
}

func isPanicCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "panic"
}