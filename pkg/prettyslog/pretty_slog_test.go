package prettyslog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPrettySlog_Enabled(t *testing.T) {
	handler := NewPrettySlog(slog.LevelInfo)
	testCases := []struct {
		level    slog.Level
		expected bool
	}{
		{slog.LevelDebug, false},
		{slog.LevelInfo, true},
		{slog.LevelWarn, true},
		{slog.LevelError, true},
	}

	for _, tc := range testCases {
		t.Run(tc.level.String(), func(t *testing.T) {
			if got := handler.Enabled(context.Background(), tc.level); got != tc.expected {
				t.Errorf("Enabled(%v) = %v, want %v", tc.level, got, tc.expected)
			}
		})
	}
}

func TestPrettySlog_WithAttrs(t *testing.T) {
	originalHandler := NewPrettySlog(slog.LevelInfo)

	attrs := []slog.Attr{
		slog.String("key1", "value1"),
		slog.Int("key2", 42),
	}

	newHandler := originalHandler.WithAttrs(attrs)
	prettyHandler, ok := newHandler.(*PrettySlog)
	if !ok {
		t.Fatalf("WithAttrs did not return a *PrettySlog")
	}

	if len(prettyHandler.attrs) != len(attrs) {
		t.Errorf("WithAttrs() added %d attrs, want %d", len(prettyHandler.attrs), len(attrs))
	}

	if len(originalHandler.attrs) != 0 {
		t.Errorf("Original handler was modified, has %d attrs", len(originalHandler.attrs))
	}
}

func TestPrettySlog_WithGroup(t *testing.T) {
	handler := NewPrettySlog(slog.LevelInfo)
	groupHandler := handler.WithGroup("test_group")

	if handler != groupHandler {
		t.Errorf("WithGroup() returned different handler instance")
	}
}

func TestPrettySlog_Handle(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
	}()

	handler := NewPrettySlog(slog.LevelInfo)

	record := slog.Record{
		Time:    time.Date(2025, 4, 15, 12, 34, 56, 789000000, time.UTC),
		Level:   slog.LevelInfo,
		Message: "Test message",
	}
	record.AddAttrs(slog.String("test_key", "test_value"))

	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedParts := []string{
		"12:34:56.789",
		"INFO",
		"Test message",
		"test_key=test_value",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output does not contain '%s'", part)
		}
	}
}

func TestPrettySlog_HandleWithAttrs(t *testing.T) {
	// Перенаправляем stdout во временный буфер
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout // Восстанавливаем stdout
	}()

	handler := NewPrettySlog(slog.LevelInfo)

	// Добавляем пре-аттрибуты к обработчику
	handler = handler.WithAttrs([]slog.Attr{slog.String("pre_attr", "pre_value")}).(*PrettySlog)

	// Создаем запись лога
	record := slog.Record{
		Time:    time.Date(2023, 10, 15, 12, 34, 56, 789000000, time.UTC),
		Level:   slog.LevelInfo,
		Message: "Test message with attrs",
	}
	record.AddAttrs(slog.String("record_attr", "record_value"))

	// Обрабатываем запись
	err := handler.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle() returned error: %v", err)
	}

	// Закрываем pipe writer и читаем вывод
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Проверяем, что вывод содержит и пре-аттрибуты, и аттрибуты из записи
	if !strings.Contains(output, "pre_attr=pre_value") {
		t.Errorf("Output does not contain pre-attribute")
	}
	if !strings.Contains(output, "record_attr=record_value") {
		t.Errorf("Output does not contain record attribute")
	}
}

func TestPrettySlog_DifferentLevels(t *testing.T) {
	// Проверяем работу с разными уровнями логирования
	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		t.Run(level.String(), func(t *testing.T) {
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			defer func() {
				os.Stdout = oldStdout
			}()

			handler := NewPrettySlog(slog.LevelDebug)

			record := slog.Record{
				Time:    time.Now(),
				Level:   level,
				Message: "Test " + level.String() + " message",
			}

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("Handle() returned error: %v", err)
			}

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, strings.ToUpper(level.String())) {
				t.Errorf("Output does not contain level name '%s'", strings.ToUpper(level.String()))
			}
		})
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		color    string
		text     string
		expected string
	}{
		{red, "ERROR", "\033[31mERROR\033[0m"},
		{cyan, "INFO", "\033[36mINFO\033[0m"},
		{gray, "DEBUG", "\033[90mDEBUG\033[0m"},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			result := colorize(test.color, test.text)
			if result != test.expected {
				t.Errorf("colorize(%q, %q) = %q, want %q", test.color, test.text, result, test.expected)
			}
		})
	}
}
