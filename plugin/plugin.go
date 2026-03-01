package loglint

import (
	"github.com/glekoz/loglint/logcheck"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

type RulesSettings struct {
	Lowercase        *bool `json:"lowercase"`
	EnglishOnly      *bool `json:"english_only"`
	NoSpecialSymbols *bool `json:"no_special_symbols"`
	NoSensitiveData  *bool `json:"no_sensitive_data"`
}

type Settings struct {
	Rules             RulesSettings `json:"rules"`
	SensitiveKeywords []string      `json:"sensitive_keywords"`
	KeywordsWhitelist []string      `json:"keywords_whitelist"`
	SymbolsWhitelist  []string      `json:"symbols_whitelist"`
}

type loglintPlugin struct {
	settings Settings
}

func (p *loglintPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	cfg := &logcheck.Config{
		Rules: logcheck.RulesConfig{
			Lowercase:        p.settings.Rules.Lowercase,
			EnglishOnly:      p.settings.Rules.EnglishOnly,
			NoSpecialSymbols: p.settings.Rules.NoSpecialSymbols,
			NoSensitiveData:  p.settings.Rules.NoSensitiveData,
		},
		SensitiveKeywords: p.settings.SensitiveKeywords,
		KeywordsWhitelist: p.settings.KeywordsWhitelist,
		SymbolsWhitelist:  p.settings.SymbolsWhitelist,
	}
	return []*analysis.Analyzer{logcheck.NewAnalyzerWithConfig(cfg)}, nil
}

func (p *loglintPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func New(conf any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](conf)
	if err != nil {
		return nil, err
	}
	return &loglintPlugin{settings: settings}, nil
}

func init() {
	register.Plugin("loglint", New)
}
