package symbols

import "log/slog"

func examples() {
	slog.Info("server started!")         // want `log message contains non-alphanumeric symbol`
	slog.Error("connection failed!!!")   // want `log message contains non-alphanumeric symbol`
	slog.Warn("something went wrong...") // want `log message contains non-alphanumeric symbol`
	slog.Info("server started")
	slog.Error("connection failed")
	slog.Warn("something went wrong")
}
