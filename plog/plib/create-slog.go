/*
© 2025–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package plib

import (
	"log/slog"

	"github.com/haraldrudell/parl"
)

func CreateSlog(log parl.PrintfFunc) (logger *slog.Logger) {
	var writer = NewWriter(log)
	// Debug(msg string, args ...any)
	// DebugContext(ctx context.Context, msg string, args ...any)
	// Enabled(ctx context.Context, level slog.Level) bool
	// Error(msg string, args ...any)
	// ErrorContext(ctx context.Context, msg string, args ...any)
	// Handler() slog.Handler
	// Info(msg string, args ...any)
	// InfoContext(ctx context.Context, msg string, args ...any)
	// Log(ctx context.Context, level slog.Level, msg string, args ...any)
	// LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
	// Warn(msg string, args ...any)
	// WarnContext(ctx context.Context, msg string, args ...any)
	// With(args ...any) *slog.Logger
	// WithGroup(name string) *slog.Logger
	var handler = slog.NewTextHandler(writer, nil)
	logger = slog.New(handler)

	return
}
