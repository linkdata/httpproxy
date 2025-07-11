package httpproxy

// Logger matches the log/slog.Logger interface.
type Logger interface {
	Debug(msg string, keyValuePairs ...any)
	Info(msg string, keyValuePairs ...any)
	Warn(msg string, keyValuePairs ...any)
	Error(msg string, keyValuePairs ...any)
}
