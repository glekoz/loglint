package english

import "log/slog"

func examples() {
	slog.Info("запуск сервера")             // want `log message contains non-English character`
	slog.Error("ошибка подключения к базе") // want `log message contains non-English character`
	slog.Info("starting server")
	slog.Error("failed to connect to database")
}
