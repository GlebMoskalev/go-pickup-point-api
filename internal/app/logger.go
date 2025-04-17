package app

import (
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/prettyslog"
	"log/slog"
	"os"
)

func setLogger(env string) {
	var handler slog.Handler
	switch env {
	case "local":
		handler = prettyslog.NewPrettySlog(slog.LevelDebug)
	case "dev":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		})
	case "prod":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})
	default:
		slog.Warn("unknown env, setting default logger", "env", env)
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	slog.SetDefault(slog.New(handler))
}
