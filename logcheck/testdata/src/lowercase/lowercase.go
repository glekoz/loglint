package lowercase

import "log/slog"

func examples() {
	slog.Info("Starting server on port 8080")   // want `log message starts with an uppercase letter`
	slog.Error("Failed to connect to database") // want `log message starts with an uppercase letter`
	slog.Info("starting server on port 8080")
	slog.Error("failed to connect to database")
}
