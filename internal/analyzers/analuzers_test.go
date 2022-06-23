package analyzers

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ExitCheckAnalyzer, "./...")
}
