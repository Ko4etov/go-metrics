package main

import (
	"github.com/Ko4etov/go-metrics/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.NewAnalyzer())
}