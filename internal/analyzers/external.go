package analyzers

import (
	"golang.org/x/tools/go/analysis"

	"github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gostaticanalysis/nilerr"
)

func GetExternalAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{analyzer.Analyzer, nilerr.Analyzer}
}
