package app

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestSetLogger_ProdLevel(t *testing.T) {
	var buf bytes.Buffer
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	setLogger("prod")

	slog.Debug("this is debug")
	slog.Info("this is info")
	slog.Warn("this is warn")
	slog.Error("this is error")
	w.Close()
	os.Stdout = originalStdout
	buf.ReadFrom(r)
	logOutput := buf.String()

	if strings.Contains(logOutput, "this is debug") {
		t.Errorf("unexpected debug log in prod")
	}
	if strings.Contains(logOutput, "this is info") {
		t.Errorf("unexpected info log in prod")
	}

	if !strings.Contains(logOutput, "this is warn") {
		t.Errorf("expected warn log missing")
	}
	if !strings.Contains(logOutput, "this is error") {
		t.Errorf("expected error log missing")
	}
}
