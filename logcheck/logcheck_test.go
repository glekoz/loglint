package logcheck_test

import (
	"testing"

	"github.com/glekoz/loglint/logcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestLowercaseRule(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logcheck.NewAnalyzer(), "lowercase")
}

func TestEnglishOnlyRule(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logcheck.NewAnalyzer(), "english")
}

func TestNoSpecialSymbolsRule(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logcheck.NewAnalyzer(), "symbols")
}

func TestNoSensitiveDataRule(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logcheck.NewAnalyzer(), "sensitive")
}
