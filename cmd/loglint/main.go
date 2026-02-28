package main

import (
	"github.com/glekoz/loglint/logcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(logcheck.NewAnalyzer())
}
