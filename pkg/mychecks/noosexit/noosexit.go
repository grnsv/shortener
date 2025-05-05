// Package noosexit provides an analyzer that disallows direct calls to os.Exit
// inside the main function of the main package.
package noosexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// NoOsExitAnalyzer is an analyzer that disallows direct calls to os.Exit
// inside the main function of the main package.
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "disallow direct os.Exit call in main function of main package",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (any, error) {
	// Only check the main package
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		// Skip generated files
		if len(file.Comments) > 0 {
			firstComment := file.Comments[0].Text()
			if strings.Contains(firstComment, "Code generated") && strings.Contains(firstComment, "DO NOT EDIT") {
				return nil, nil
			}
		}

		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Look for the main function
			if fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if pkgIdent, ok := selector.X.(*ast.Ident); ok {
					if pkgIdent.Name == "os" && selector.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "direct call to os.Exit is forbidden in main function")
					}
				}

				return true
			})

			return false
		})
	}

	return nil, nil
}
