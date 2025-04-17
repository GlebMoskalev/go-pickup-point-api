package app

import (
	"context"
	"fmt"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	v1 "github.com/GlebMoskalev/go-pickup-point-api/internal/api/v1"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/service"
	"log/slog"
	"net/http"
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

	if err := server.ListenAndServe(); err != nil {
		slog.Error("server error: %v", err)
		return
	}
}
