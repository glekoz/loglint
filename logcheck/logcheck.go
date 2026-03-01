package logcheck

import (
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type customAnalyzer struct {
	pkgs              map[string][]string
	sensitiveKeywords []string
	kwWhitelist       []string
	symbolsWhitelist  []rune

	checkLowercase        bool
	checkEnglishOnly      bool
	checkNoSpecialSymbols bool
	checkNoSensitiveData  bool

	configPath   string
	inlineConfig *Config
	once         sync.Once
	configErr    error
}

func NewAnalyzerWithConfig(cfg *Config) *analysis.Analyzer {
	ca := newDefaultAnalyzer()
	ca.inlineConfig = cfg
	return &analysis.Analyzer{
		Name: "loglint",
		Doc:  "loglint checks for proper logging usage",
		Run:  ca.run,
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
	}
}

func NewAnalyzer() *analysis.Analyzer {
	// ниже представлен базовый конфиг, который может быть переопределён через флаг -config
	ca := newDefaultAnalyzer()
	a := &analysis.Analyzer{
		Name: "loglint",
		Doc:  "loglint checks for proper logging usage",
		Run:  ca.run,
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
	}
	a.Flags.StringVar(&ca.configPath, "config", "", "path to loglint config file in YAML format")
	return a
}

func newDefaultAnalyzer() *customAnalyzer {
	return &customAnalyzer{
		pkgs: map[string][]string{
			"log/slog": {
				"Debug", "DebugContext", "Error", "ErrorContext", "Info", "InfoContext", "Warn", "WarnContext",
			},
			"go.uber.org/zap": {
				"Debug", "DPanic", "Error", "Fatal", "Info", "Panic", "Warn",
			},
		},
		sensitiveKeywords:     []string{"password", "secret", "token", "key", "credential", "auth", "login", "pass", "pwd"},
		kwWhitelist:           make([]string, 0),
		checkLowercase:        true,
		checkEnglishOnly:      true,
		checkNoSpecialSymbols: true,
		checkNoSensitiveData:  true,
	}
}

func (ca *customAnalyzer) applyConfigFrom(path string) error {
	cfg, err := loadConfig(path)
	if err != nil {
		return err
	}
	ca.applyConfigStruct(cfg)
	return nil
}

func (ca *customAnalyzer) applyConfigStruct(cfg *Config) {
	ca.checkLowercase = boolVal(cfg.Rules.Lowercase)
	ca.checkEnglishOnly = boolVal(cfg.Rules.EnglishOnly)
	ca.checkNoSpecialSymbols = boolVal(cfg.Rules.NoSpecialSymbols)
	ca.checkNoSensitiveData = boolVal(cfg.Rules.NoSensitiveData)

	if len(cfg.SensitiveKeywords) > 0 {
		ca.sensitiveKeywords = cfg.SensitiveKeywords
	}
	if len(cfg.KeywordsWhitelist) > 0 {
		ca.kwWhitelist = cfg.KeywordsWhitelist
	}
	if len(cfg.SymbolsWhitelist) > 0 {
		ca.symbolsWhitelist = ca.symbolsWhitelist[:0]
		for _, s := range cfg.SymbolsWhitelist {
			for _, r := range s {
				ca.symbolsWhitelist = append(ca.symbolsWhitelist, r)
				break
			}
		}
	}
	if len(cfg.Loggers) > 0 {
		ca.pkgs = cfg.Loggers
	}
}

func (ca *customAnalyzer) run(pass *analysis.Pass) (any, error) {
	ca.once.Do(func() {
		if ca.inlineConfig != nil {
			ca.applyConfigStruct(ca.inlineConfig)
			return
		}
		path := ca.configPath
		if path == "" {
			path = findConfig(pass)
		}
		if path != "" {
			if err := ca.applyConfigFrom(path); err != nil && ca.configPath != "" {
				if fallback := findConfig(pass); fallback != "" {
					_ = ca.applyConfigFrom(fallback)
				}
			}
		}
	})
	if ca.configErr != nil {
		return nil, ca.configErr
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		obj := pass.TypesInfo.ObjectOf(fun.Sel)
		if obj == nil {
			return
		}

		if !ca.isValidLogger(obj) {
			return
		}

		// так как первым аргументом может быть контекст или конкатенация, то
		// необходимо также передавать адрес переменной isFirst
		var isFirst = true
		for _, arg := range call.Args {
			ca.checkArgument(pass, arg, &isFirst)
		}

	})

	return nil, nil
}

func findConfig(pass *analysis.Pass) string {
	dirs := []string{}
	if len(pass.Files) > 0 {
		dirs = append(dirs, filepath.Dir(pass.Fset.Position(pass.Files[0].Pos()).Filename))
	}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, cwd)
	}

	for _, start := range dirs {
		dir := start
		for {
			for _, name := range []string{".loglint.yml", ".loglint.yaml"} {
				if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
					return filepath.Join(dir, name)
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

func (ca *customAnalyzer) isValidLogger(o types.Object) bool {
	pkg := o.Pkg()
	if pkg == nil {
		return false
	}

	methods, exists := ca.pkgs[pkg.Path()]
	if !exists {
		return false
	}

	return slices.Contains(methods, o.Name())
}

func (ca *customAnalyzer) checkArgument(pass *analysis.Pass, expr ast.Expr, isFirst *bool) {
	switch v := expr.(type) {
	case *ast.BasicLit:
		if v.Kind == token.STRING {
			ca.checkStringLiteral(pass, v, isFirst)
		}

	case *ast.BinaryExpr:
		if v.Op == token.ADD {
			ca.checkArgument(pass, v.X, isFirst)
			ca.checkArgument(pass, v.Y, isFirst)
		}

	case *ast.Ident:
		ca.checkIdentifier(pass, v)

	case *ast.CallExpr:
		ca.checkCall(pass, v, isFirst)
	}
}

func (ca *customAnalyzer) checkStringLiteral(pass *analysis.Pass, lit *ast.BasicLit, isFirst *bool) {
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return
	}
	if len(value) == 0 {
		pass.Reportf(
			lit.Pos(),
			"log message is empty",
		)
		return
	}

	badsymbols := []int{}
	for i, r := range value {
		if unicode.Is(unicode.Latin, r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			continue
		}

		if slices.Contains(ca.symbolsWhitelist, r) {
			continue
		}

		if unicode.IsLetter(r) {
			if ca.checkEnglishOnly {
				pass.Reportf(
					lit.Pos()+token.Pos(i),
					"log message contains non-English character: %q", string(r),
				)
			}
			return
		}
		badsymbols = append(badsymbols, i)
	}

	if len(badsymbols) > 0 && ca.checkNoSpecialSymbols {
		newline := make([]byte, 0, len(value)-len(badsymbols))
		for i := range value {
			if slices.Contains(badsymbols, i) {
				continue
			}
			newline = append(newline, value[i])
		}
		pass.Report(analysis.Diagnostic{
			Pos:     lit.Pos(),
			End:     lit.End(),
			Message: "log message contains non-alphanumeric symbol",
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "remove non-alphanumeric symbols",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     lit.Pos(),
							End:     lit.End(),
							NewText: []byte(strconv.Quote(string(newline))),
						},
					},
				},
			},
		})
	}

	if *isFirst {
		if ca.checkLowercase && unicode.IsUpper(rune(value[0])) {
			pass.Report(analysis.Diagnostic{
				Pos:     lit.Pos(),
				End:     lit.End(),
				Message: "log message starts with an uppercase letter",
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "convert first letter to lowercase",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     lit.Pos(),
								End:     lit.End(),
								NewText: []byte(strconv.Quote(strings.ToLower(value))),
							},
						},
					},
				},
			})
		}
		*isFirst = false
	}

	if ca.checkNoSensitiveData {
		for word := range strings.FieldsSeq(strings.ToLower(value)) {
			for _, kw := range ca.sensitiveKeywords {
				if strings.Contains(word, kw) && !slices.Contains(ca.kwWhitelist, word) {
					pass.Reportf(lit.Pos(), "log message contains potentially sensitive data: %s", word)
				}
			}
		}
	}
}

func (ca *customAnalyzer) checkIdentifier(pass *analysis.Pass, ident *ast.Ident) {
	if !ca.checkNoSensitiveData {
		return
	}
	name := strings.ToLower(ident.Name)

	for _, kw := range ca.sensitiveKeywords {
		if strings.Contains(name, kw) && !slices.Contains(ca.kwWhitelist, name) {
			pass.Report(analysis.Diagnostic{
				Pos:     ident.Pos(),
				End:     ident.End(),
				Message: "log message contains potentially sensitive variable: " + ident.Name,
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "remove sensitive variable from log message",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     ident.Pos(),
								End:     ident.End(),
								NewText: []byte(strconv.Quote("credentials removed")),
							},
						},
					},
				},
			})
			break
		}
	}
}

func (ca *customAnalyzer) checkCall(pass *analysis.Pass, call *ast.CallExpr, isFirst *bool) {
	for _, arg := range call.Args {
		ca.checkArgument(pass, arg, isFirst)
	}
}
