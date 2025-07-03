package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ExitMainAnalyzer - анализатор, который сообщает о прямых вызовах
// os.Exit внутри функции main() пакета main.

var ExitMainAnalyzer = &analysis.Analyzer{
	Name:     "exitmain",
	Doc:      "reports direct calls to os.Exit in main function of package main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (interface{}, error) {
		if pass.Pkg.Name() != "main" {
			return nil, nil
		}
		insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}
		insp.Preorder(nodeFilter, func(n ast.Node) {
			fn := n.(*ast.FuncDecl)
			if fn.Name.Name != "main" || fn.Body == nil {
				return
			}
			ast.Inspect(fn.Body, func(n2 ast.Node) bool {
				call, ok := n2.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				id, ok := sel.X.(*ast.Ident)
				if !ok || id.Name != "os" || sel.Sel.Name != "Exit" {
					return true
				}
				if obj, ok := pass.TypesInfo.Uses[id].(*types.PkgName); ok && obj.Imported().Path() == "os" {
					pass.Reportf(sel.Sel.Pos(), "direct call to os.Exit is not allowed in main")
				}
				return true
			})
		})
		return nil, nil
	},
}
