package analyzers

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func GetStaticCheckAnalyzers() []*analysis.Analyzer {
	analyzers := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers)+1)
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, v.Analyzer)
	}

	return analyzers
}
