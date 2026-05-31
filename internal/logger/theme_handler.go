package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/SamyRai/juleson/internal/presentation/views/theme"
)

// ThemeHandler formats slog Records using the CLI lipgloss theme
type ThemeHandler struct {
	out   io.Writer
	opts  *slog.HandlerOptions
	group string
	attrs []slog.Attr
}

// NewThemeHandler creates a new handler that writes themed output
func NewThemeHandler(out io.Writer, opts *slog.HandlerOptions) *ThemeHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &ThemeHandler{
		out:  out,
		opts: opts,
	}
}

// Enabled checks if the given level is enabled
func (h *ThemeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

// Handle formats the log record using the theme package
func (h *ThemeHandler) Handle(ctx context.Context, r slog.Record) error {
	var builder strings.Builder

	// Write message using theme
	switch {
	case r.Level == LevelSuccess:
		builder.WriteString(theme.SuccessStyle.Render("✅ " + r.Message))
	case r.Level >= slog.LevelError:
		builder.WriteString(theme.ErrorStyle.Render("❌ " + r.Message))
	case r.Level >= slog.LevelWarn:
		builder.WriteString(theme.WarnStyle.Render("⚠️  " + r.Message))
	case r.Level >= slog.LevelInfo:
		// For Info, we treat it as a Step
		builder.WriteString(theme.InfoStyle.Render("• " + r.Message))
	default:
		// Debug
		builder.WriteString(theme.MutedStyle.Render("🔍 " + r.Message))
	}

	// Format attributes
	var attrs []string

	// Add handler attributes
	for _, a := range h.attrs {
		attrs = append(attrs, formatAttr(a))
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, formatAttr(a))
		return true
	})

	if len(attrs) > 0 {
		builder.WriteString(theme.MutedStyle.Render("  [" + strings.Join(attrs, " ") + "]"))
	}

	builder.WriteString("\n")
	_, err := fmt.Fprint(h.out, builder.String())
	return err
}

func formatAttr(a slog.Attr) string {
	val := a.Value.String()
	return fmt.Sprintf("%s=%s", a.Key, val)
}

// WithAttrs returns a new handler with the given attributes appended
func (h *ThemeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = append(h2.attrs, attrs...)
	return &h2
}

// WithGroup returns a new handler with the given group appended
func (h *ThemeHandler) WithGroup(name string) slog.Handler {
	h2 := *h
	if h2.group != "" {
		h2.group += "." + name
	} else {
		h2.group = name
	}
	return &h2
}
