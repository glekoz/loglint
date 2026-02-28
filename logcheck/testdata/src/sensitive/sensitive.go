package sensitive

import "log/slog"

var (
	userPassword = "secret123"
	apiKey       = "mykey456"
	accessToken  = "tok789"
)

func variableExamples() {
	slog.Info("user " + userPassword)     // want `log message contains potentially sensitive variable: userPassword`
	slog.Debug("request " + apiKey)       // want `log message contains potentially sensitive variable: apiKey`
	slog.Info("validated " + accessToken) // want `log message contains potentially sensitive variable: accessToken`
	slog.Info("request completed")
	slog.Debug("operation succeeded")
}
