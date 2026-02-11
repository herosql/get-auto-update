package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var defaultLogger *slog.Logger

func Init(infoPath, errorPath string) (func(), error) {
	infoFile, err := os.OpenFile(infoPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open info log: %w", err)
	}

	errorFile, err := os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		infoFile.Close()
		return nil, fmt.Errorf("failed to open error log: %w", err)
	}

	infoHandler := slog.NewJSONHandler(infoFile, &slog.HandlerOptions{Level: slog.LevelInfo})
	errorHandler := slog.NewJSONHandler(errorFile, &slog.HandlerOptions{Level: slog.LevelError})

	combinedHandler := &multiHandler{
		handlers: []slog.Handler{infoHandler, errorHandler},
	}

	defaultLogger = slog.New(combinedHandler)
	slog.SetDefault(defaultLogger)

	cleanup := func() {
		infoFile.Close()
		errorFile.Close()
	}

	return cleanup, nil
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
	printConsole("➜", msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
	printConsole("❌", msg, args...)
}

func printConsole(icon, msg string, args ...any) {
	if strings.Contains(msg, "%") && len(args) > 0 {
		formatted := fmt.Sprintf(msg, args...)
		if icon == "❌" {
			fmt.Fprintf(os.Stderr, "%s  Error: %s\n", icon, formatted)
		} else {
			fmt.Printf("%s  %s\n", icon, formatted)
		}
	} else {
		if icon == "❌" {
			fmt.Fprintf(os.Stderr, "%s  Error: %s %v\n", icon, msg, args)
		} else {
			if len(args) == 0 {
				fmt.Printf("%s  %s \n", icon, msg)
			} else {
				fmt.Printf("%s  %s %v\n", icon, msg, args)
			}
		}
	}
}

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.Handlers() {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) Handlers() []slog.Handler {
	return m.handlers
}
