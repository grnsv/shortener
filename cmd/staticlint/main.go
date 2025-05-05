// Package main implements a static analysis tool based on multichecker.
//
// This tool allows you to run a configurable set of Go static analysis checks
// (analyzers) on your codebase. The set of analyzers is configured via a JSON
// file (config.json) that specifies which analyzers to enable from several
// groups: standard analyzers, staticcheck, simple, stylecheck, custom analyzers,
// and others.
//
// Mechanism of multichecker launch:
//
//  1. The tool reads its configuration from a JSON file (config.json) located
//     in the same directory as the executable. The configuration specifies which
//     analyzers to enable by name for each group.
//  2. It parses the configuration into a ConfigData struct.
//  3. For each group of analyzers (standard, staticcheck, simple, stylecheck,
//     custom, and other), it selects only those analyzers whose names are listed
//     in the configuration.
//  4. All selected analyzers are collected into a single slice.
//  5. The multichecker.Main function is called with the selected analyzers.
//     multichecker is a driver that runs all provided analyzers on the target
//     Go packages, reporting any issues found.
//
// Analyzer groups and their purposes:
//
//   - Standard analyzers (stdAnalyzers): Provided by golang.org/x/tools/go/analysis/passes.
//     These analyzers check for a wide range of common issues in Go code, such as
//     misuse of append, atomic operations, shadowed variables, printf formatting,
//     unreachable code, and more.
//
//   - Staticcheck analyzers: Provided by honnef.co/go/tools/staticcheck.
//     These analyzers detect bugs, performance issues, and code simplifications
//     that go beyond what the standard analyzers provide.
//
//   - Simple analyzers: Provided by honnef.co/go/tools/simple.
//     These analyzers suggest code simplifications and idiomatic Go patterns.
//
//   - Stylecheck analyzers: Provided by honnef.co/go/tools/stylecheck.
//     These analyzers enforce Go code style and best practices.
//
//   - Custom analyzers (mychecks): Provided by the project's own package (github.com/grnsv/shortener/pkg/mychecks).
//     These analyzers can be used to enforce project-specific rules.
//
//   - Other analyzers (otherAnalyzers): Includes third-party analyzers such as:
//     1. bodyclose: Checks that HTTP response bodies are closed.
//     2. errcheck: Checks that error return values are used.
//
// Each analyzer is referenced by its name, and only those enabled in the config
// are run. This modular approach allows you to tailor the static analysis to
// your project's needs.
//
// Usage:
//
//  1. Create a config.json file in the same directory as the built binary, listing the analyzers to enable.
//  2. Build and run the binary. The tool will analyze the code in the current
//     module according to the enabled analyzers in config.json.
package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/grnsv/shortener/pkg/mychecks"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// Config — configuration file name.
const Config = `config.json`

// ConfigData describes the structure of the configuration file.
type ConfigData struct {
	Std         []string `json:"std"`
	Staticcheck []string `json:"staticcheck"`
	Simple      []string `json:"simple"`
	Stylecheck  []string `json:"stylecheck"`
	Mychecks    []string `json:"mychecks"`
	Other       []string `json:"other"`
}

var stdAnalyzers = []*analysis.Analyzer{
	appends.Analyzer,
	asmdecl.Analyzer,
	assign.Analyzer,
	atomic.Analyzer,
	atomicalign.Analyzer,
	bools.Analyzer,
	buildssa.Analyzer,
	buildtag.Analyzer,
	cgocall.Analyzer,
	composite.Analyzer,
	copylock.Analyzer,
	ctrlflow.Analyzer,
	deepequalerrors.Analyzer,
	defers.Analyzer,
	directive.Analyzer,
	errorsas.Analyzer,
	fieldalignment.Analyzer,
	findcall.Analyzer,
	framepointer.Analyzer,
	httpmux.Analyzer,
	httpresponse.Analyzer,
	ifaceassert.Analyzer,
	inspect.Analyzer,
	loopclosure.Analyzer,
	lostcancel.Analyzer,
	nilfunc.Analyzer,
	nilness.Analyzer,
	pkgfact.Analyzer,
	printf.Analyzer,
	reflectvaluecompare.Analyzer,
	shadow.Analyzer,
	shift.Analyzer,
	sigchanyzer.Analyzer,
	slog.Analyzer,
	sortslice.Analyzer,
	stdmethods.Analyzer,
	stdversion.Analyzer,
	stringintconv.Analyzer,
	structtag.Analyzer,
	testinggoroutine.Analyzer,
	tests.Analyzer,
	timeformat.Analyzer,
	unmarshal.Analyzer,
	unreachable.Analyzer,
	unsafeptr.Analyzer,
	unusedresult.Analyzer,
	unusedwrite.Analyzer,
	usesgenerics.Analyzer,
	waitgroup.Analyzer,
}

var otherAnalyzers = []*analysis.Analyzer{
	bodyclose.Analyzer,
	errcheck.Analyzer,
}

func main() {
	appfile, err := os.Executable()
	if err != nil {
		panic(err)
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	if err != nil {
		panic(err)
	}
	var cfg ConfigData
	if err = json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	analyzers := []*analysis.Analyzer{}
	analyzers = appendAnalyzers(analyzers, stdAnalyzers, cfg.Std)
	analyzers = appendLintAnalyzers(analyzers, staticcheck.Analyzers, cfg.Staticcheck)
	analyzers = appendLintAnalyzers(analyzers, simple.Analyzers, cfg.Simple)
	analyzers = appendLintAnalyzers(analyzers, stylecheck.Analyzers, cfg.Stylecheck)
	analyzers = appendAnalyzers(analyzers, mychecks.Analyzers, cfg.Mychecks)
	analyzers = appendAnalyzers(analyzers, otherAnalyzers, cfg.Other)

	multichecker.Main(analyzers...)
}

func appendAnalyzers(result []*analysis.Analyzer, available []*analysis.Analyzer, cfg []string) []*analysis.Analyzer {
	checks := make(map[string]bool)
	for _, v := range cfg {
		checks[v] = true
	}

	for _, v := range available {
		if checks[v.Name] {
			result = append(result, v)
		}
	}

	return result
}

func appendLintAnalyzers(result []*analysis.Analyzer, available []*lint.Analyzer, cfg []string) []*analysis.Analyzer {
	checks := make(map[string]bool)
	for _, v := range cfg {
		checks[v] = true
	}

	for _, v := range available {
		if checks[v.Analyzer.Name] {
			result = append(result, v.Analyzer)
		}
	}

	return result
}
