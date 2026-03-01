package logcheck

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RulesConfig struct {
	Lowercase        *bool `yaml:"lowercase"`
	EnglishOnly      *bool `yaml:"english_only"`
	NoSpecialSymbols *bool `yaml:"no_special_symbols"`
	NoSensitiveData  *bool `yaml:"no_sensitive_data"`
}

type Config struct {
	Rules             RulesConfig         `yaml:"rules"`
	SensitiveKeywords []string            `yaml:"sensitive_keywords"`
	KeywordsWhitelist []string            `yaml:"keywords_whitelist"`
	SymbolsWhitelist  []string            `yaml:"symbols_whitelist"`
	Loggers           map[string][]string `yaml:"loggers"`
}

func boolVal(b *bool) bool {
	return b == nil || *b
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
