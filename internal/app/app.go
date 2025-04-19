package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	v1 "github.com/GlebMoskalev/go-pickup-point-api/internal/api/v1"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func Run(configPath string) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		slog.Error("config error", "error", err)
		return
	}

	setLogger(cfg.Env)

	dbPool, err := setupDB(context.Background(), cfg.Database)
	if err != nil {
		return
	}
	defer dbPool.Close()

	repositories := repo.NewRepositories(dbPool)

	services := service.NewServices(repositories, cfg)
	router := v1.NewRouter(services)

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting server", "address", serverAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
		}
	case sig := <-shutdown:
		slog.Info("shutdown signal received", "siglan", sig)

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		slog.Info("gracefully shutting down server", "timeout", cfg.Server.ShutdownTimeout)

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)

			if err := server.Close(); err != nil {
				slog.Error("server closed error", "error", err)
			}
		}
	}

	slog.Info("server stopped")
}
