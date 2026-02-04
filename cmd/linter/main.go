package main

import (
	"github.com/galogen13/yandex-go-metrics/cmd/linter/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.PanicUseAnalyzer)
}
