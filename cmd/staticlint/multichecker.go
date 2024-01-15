// It initializes a slice of analyzers and runs them using the multichecker.Main function.
package main

import (
	"go/ast"

	"github.com/jingyugao/rowserrcheck/passes/rowserr"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/ast/inspector"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Main is the entry point of the application.
// It initializes a slice of analyzers and runs them using the multichecker.Main function
func main() {

	analyzers := []*analysis.Analyzer{
		// buildssa.Analyzer is an analyzer that builds SSA form of packages.
		// It provides a representation of the program that is easier to analyze.
		buildssa.Analyzer,

		// findcall.Analyzer is an analyzer that finds all function calls in the program.
		findcall.Analyzer,

		// printf.Analyzer is an analyzer that checks for incorrect usage of Printf.
		printf.Analyzer,

		// stdmethods.Analyzer is an analyzer that checks for incorrect usage of standard methods.
		stdmethods.Analyzer,

		// structtag.Analyzer is an analyzer that checks for incorrect usage of struct tags.
		structtag.Analyzer,

		// tests.Analyzer is an analyzer that checks for common mistakes in tests.
		tests.Analyzer,

		// unmarshal.Analyzer is an analyzer that checks for incorrect usage of Unmarshal.
		unmarshal.Analyzer,

		// inspect.Analyzer is an analyzer that provides a way to traverse the syntax tree.
		inspect.Analyzer,

		// asmdecl.Analyzer is an analyzer that checks for incorrect assembly declarations.
		asmdecl.Analyzer,

		// assign.Analyzer is an analyzer that checks for incorrect assignment operations.
		assign.Analyzer,

		// atomic.Analyzer is an analyzer that checks for incorrect usage of atomic operations.
		atomic.Analyzer,

		// bools.Analyzer is an analyzer that checks for incorrect boolean expressions.
		bools.Analyzer,

		// buildtag.Analyzer is an analyzer that checks for incorrect build tags.
		buildtag.Analyzer,

		// cgocall.Analyzer is an analyzer that checks for incorrect cgo calls.
		cgocall.Analyzer,

		// composite.Analyzer is an analyzer that checks for incorrect composite literals.
		composite.Analyzer,

		// copylock.Analyzer is an analyzer that checks for incorrect lock copying.
		copylock.Analyzer,

		// errorsas.Analyzer is an analyzer that checks for incorrect usage of errors.As.
		errorsas.Analyzer,

		// httpresponse.Analyzer is an analyzer that checks for incorrect HTTP responses.
		httpresponse.Analyzer,

		// loopclosure.Analyzer is an analyzer that checks for incorrect loop closures.
		loopclosure.Analyzer,

		// lostcancel.Analyzer is an analyzer that checks for lost cancelations.
		lostcancel.Analyzer,

		// nilness.Analyzer is an analyzer that checks for incorrect nil comparisons.
		nilness.Analyzer,

		// shift.Analyzer is an analyzer that checks for incorrect shift operations.
		shift.Analyzer,

		// unreachable.Analyzer is an analyzer that checks for unreachable code.
		unreachable.Analyzer,

		// unsafeptr.Analyzer is an analyzer that checks for incorrect unsafe pointer operations.
		unsafeptr.Analyzer,

		// unusedresult.Analyzer is an analyzer that checks for unused results.
		unusedresult.Analyzer,

		// bodyclose.Analyzer is an analyzer that checks for incorrect body closing.
		bodyclose.Analyzer,

		// ExitAnalyzer is a custom analyzer that checks for direct calls to os.Exit in the main function.
		ExitAnalyzer,

		// rowserr.NewAnalyzer is an analyzer that checks whether sql.Rows.Err is correctly checked.
		rowserr.NewAnalyzer(
			"github.com/jackc/pgx/v5",
		),
	}

	for _, analyzer := range staticcheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}
	for _, analyzer := range stylecheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	multichecker.Main(
		analyzers...,
	)
}

// ExitAnalyzer is a custom analyzer that checks for direct calls to os.Exit in the main function.
var ExitAnalyzer = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "Checks for direct calls to os.Exit in the main function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run is the function that gets executed when the Analyzer is run.
func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		if sel.Sel.Name == "Exit" && sel.X.(*ast.Ident).Name == "os" {
			pass.Reportf(call.Pos(), "direct call to os.Exit is forbidden")
		}
	})

	return nil, nil
}
