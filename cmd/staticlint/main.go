package main

import (
	"github.com/itd27m01/go-metrics-service/internal/analyzers"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		GetAnalyzers()...,
	)
}

func GetAnalyzers() []*analysis.Analyzer {
	var analyzersSlice []*analysis.Analyzer
	analyzersSlice = append(analyzersSlice, analyzers.GetAnalysisAnalyzers()...)
	analyzersSlice = append(analyzersSlice, analyzers.GetStaticCheckAnalyzers()...)
	analyzersSlice = append(analyzersSlice, analyzers.GetExternalAnalyzers()...)
	analyzersSlice = append(analyzersSlice, analyzers.ExitCheckAnalyzer)

	return analyzersSlice
}
