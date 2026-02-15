package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analyzer := NewAnalyzer()

	analysistest.Run(t, testdata, analyzer, "testpkg")
	analysistest.Run(t, testdata, analyzer, "main")
}