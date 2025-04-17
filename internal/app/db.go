package app

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

func setupDB(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		slog.Error("failed to create a config", "error", err)
		return nil, err
	}
	dbConfig.MaxConns = cfg.MaxConns
	dbConfig.MinConns = cfg.MinConns
	dbConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	dbConfig.MaxConnLifetime = cfg.MaxConnLifeTime

	connPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		slog.Error("error while creating connection to the database", "error", err)
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()

	err = connPool.Ping(pingCtx)
	if err != nil {
		slog.Error("could not ping database", "error", err)
		connPool.Close()
		return nil, err
	}

	return connPool, nil
}
