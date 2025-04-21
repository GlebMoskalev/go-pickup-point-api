package helperstest

import (
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"time"
)

func CreateTestConfig(dbConfig config.Database) *config.Config {
	return &config.Config{
		Env: "test",
		Server: config.Server{
			Host:            "localhost",
			Port:            "8080",
			ShutdownTimeout: 5 * time.Second,
		},
		Database: dbConfig,
		Token: config.Token{
			SignKey: "test_secret_key",
			TTL:     24 * time.Hour,
		},
		Salt: "test_salt",
		Prometheus: config.Prometheus{
			Port: "9090",
			Path: "/metrics",
		},
	}
}
