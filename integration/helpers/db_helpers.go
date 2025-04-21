package helpers

import (
	"context"
	"fmt"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func SetupPostgresContainer(t *testing.T, ctx context.Context) (*postgres.PostgresContainer, config.Database) {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("pvz_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")

	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err, "failed to get container host")

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err, "failed to get container port")

	cfg := config.Database{
		User:            "postgres",
		Password:        "postgres",
		Host:            host,
		Port:            port.Port(),
		Name:            "pvz_test",
		SslMode:         "disable",
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifeTime: 15 * time.Minute,
		MaxConnIdleTime: 30 * time.Minute,
	}

	return postgresContainer, cfg
}

func SetupDatabaseConnection(t *testing.T, ctx context.Context, dbConfig config.Database) *pgxpool.Pool {
	pgxConfig, err := pgxpool.ParseConfig(dbConfig.DSN())
	require.NoError(t, err, "failed to parse database connection string")

	pgxConfig.MaxConns = dbConfig.MaxConns
	pgxConfig.MinConns = dbConfig.MinConns
	pgxConfig.MaxConnLifetime = dbConfig.MaxConnLifeTime
	pgxConfig.MaxConnIdleTime = dbConfig.MaxConnIdleTime

	var dbPool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		dbPool, err = pgxpool.NewWithConfig(ctx, pgxConfig)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err, "failed to connect to database after multiple attempts")

	err = dbPool.Ping(ctx)
	require.NoError(t, err, "failed to ping database")

	return dbPool
}

func ApplyMigrations(t *testing.T, dbConfig config.Database) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbConfig.DSN())
	require.NoError(t, err, "failed to connect to database for migrations")
	defer pool.Close()

	_, currentFile, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))

	migrationsDir := filepath.Join(projectRoot, "migrations")

	files, err := os.ReadDir(migrationsDir)
	require.NoError(t, err, fmt.Sprintf("failed to read migrations directory: %s", migrationsDir))
	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrations = append(migrations, file.Name())
		}
	}

	sort.Strings(migrations)

	for _, migrationFile := range migrations {
		migrationPath := filepath.Join(migrationsDir, migrationFile)
		sqlBytes, err := os.ReadFile(migrationPath)
		require.NoError(t, err, fmt.Sprintf("failed to read migration file: %s", migrationPath))

		_, err = pool.Exec(ctx, string(sqlBytes))
		require.NoError(t, err, fmt.Sprintf("failed to apply migration: %s", migrationFile))

		t.Logf("Successfully applied migration: %s", migrationFile)
	}
}
