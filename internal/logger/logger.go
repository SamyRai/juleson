package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"
	"regexp"
	"strings"
)

// LevelSuccess is a custom log level for success messages (higher than Info, lower than Warn)
const LevelSuccess = slog.Level(2)

// Config holds logger configuration
type Config struct {
	Debug      bool
	FormatJSON bool
	Output     io.Writer
}

// New creates a new context-aware, environment-adaptive logger
func New(cfg Config) *slog.Logger {
	var handler slog.Handler
	out := cfg.Output
	if out == nil {
		out = os.Stderr
	}

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: redactSecrets,
	}

	// Determine if we should use TTY theme
	isTerminal := false
	if f, ok := out.(*os.File); ok {
		isTerminal = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}

	if cfg.FormatJSON {
		handler = slog.NewJSONHandler(out, opts)
	} else if isTerminal {
		handler = NewThemeHandler(out, opts)
	} else {
		handler = slog.NewTextHandler(out, opts)
	}

	return slog.New(handler)
}

// Success is a helper to log a success message since slog doesn't have it built-in
func Success(l *slog.Logger, msg string, args ...any) {
	l.Log(context.Background(), LevelSuccess, msg, args...)
}

// SetupGlobal configures the global slog logger with default dev settings
func SetupGlobal(debug bool) {
	l := New(Config{
		Debug:      debug,
		FormatJSON: false,
		Output:     os.Stdout,
	})
	slog.SetDefault(l)
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(gh[pousr]_[a-zA-Z0-9]{36})`),
	regexp.MustCompile(`(?i)(jules_[a-zA-Z0-9]+)`),
}

func redactSecrets(groups []string, a slog.Attr) slog.Attr {
	// Redact specific keys
	lowerKey := strings.ToLower(a.Key)
	if strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "api_key") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "authorization") {
		if a.Value.Kind() == slog.KindString && a.Value.String() != "" {
			return slog.String(a.Key, "[REDACTED]")
		}
	}

	// Redact known patterns in any string value
	if a.Value.Kind() == slog.KindString {
		val := a.Value.String()
		for _, pattern := range secretPatterns {
			if pattern.MatchString(val) {
				val = pattern.ReplaceAllString(val, "[REDACTED]")
			}
		}
		if val != a.Value.String() {
			return slog.String(a.Key, val)
		}
	}
	return a
}
