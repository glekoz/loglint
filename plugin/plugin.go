package loglint

import (
	"github.com/glekoz/loglint/logcheck"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

type Settings struct {
	Config string `json:"config"`
}

type loglintPlugin struct {
	settings Settings
}

func (p *loglintPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	a := logcheck.NewAnalyzer()
	if p.settings.Config != "" {
		if err := a.Flags.Set("config", p.settings.Config); err != nil {
			return nil, err
		}
	}
	return []*analysis.Analyzer{a}, nil
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
