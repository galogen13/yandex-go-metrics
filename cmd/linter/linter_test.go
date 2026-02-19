package main

import (
	"testing"

	"github.com/galogen13/yandex-go-metrics/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestPanicUseAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.PanicUseAnalyzer, "./...")
}
