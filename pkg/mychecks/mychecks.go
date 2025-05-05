// Package mychecks provides custom static analysis checks for Go code.
//
// This package contains analyzers that can be used with golang.org/x/tools/go/analysis
// to enforce project-specific coding standards and restrictions.
package mychecks

import (
	"github.com/grnsv/shortener/pkg/mychecks/noosexit"
	"golang.org/x/tools/go/analysis"
)

// Analyzers is a list of all custom analyzers provided by this package.
var Analyzers = []*analysis.Analyzer{
	noosexit.NoOsExitAnalyzer,
}
