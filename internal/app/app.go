package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	v1 "github.com/GlebMoskalev/go-pickup-point-api/internal/api/v1"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	metricsAddr := fmt.Sprintf(":%s", cfg.Prometheus.Port)
	metricsServer := &http.Server{
		Addr:    metricsAddr,
		Handler: mux,
	}

	serverErrors := make(chan error, 2)
	go func() {
		slog.Info("starting server", "address", serverAddr)
		serverErrors <- server.ListenAndServe()
	}()

	go func() {
		slog.Info("starting prometheus metrics server", "address", metricsAddr)
		serverErrors <- metricsServer.ListenAndServe()
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

		if err := metricsServer.Shutdown(ctx); err != nil {
			slog.Error("metrics server shutdown error", "error", err)
			if err := metricsServer.Close(); err != nil {
				slog.Error("metrics server close error", "error", err)
			}
		}
	}

	slog.Info("servers stopped")
}
