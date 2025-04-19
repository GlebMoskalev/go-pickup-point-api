package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"time"
)

type (
	Config struct {
		Env        string     `env-required:"true", yaml:"env"`
		Server     Server     `yaml:"server"`
		Database   Database   `yaml:"database"`
		Token      Token      `yaml:"token"`
		Salt       string     `env-required:"true" env:"SALT"`
		Prometheus Prometheus `yaml:"prometheus"`
	}
	Server struct {
		Host            string        `env-required:"true" env:"HOST"`
		Port            string        `env-required:"true" env:"PORT"`
		ShutdownTimeout time.Duration `env-required:"true" yaml:"shutdown_timeout"`
	}
	Database struct {
		User            string        `env-required:"true" env:"DB_USER"`
		Password        string        `env-required:"true" env:"DB_PASSWORD"`
		Host            string        `env-required:"true" env:"DB_HOST"`
		Port            string        `env-required:"true" env:"DB_PORT"`
		Name            string        `env-required:"true" env:"DB_NAME"`
		SslMode         string        `env-required:"true" env:"SSL_MODE"`
		MaxConns        int32         `env-required:"true" yaml:"max_conns"`
		MinConns        int32         `env-required:"true" yaml:"min_conns"`
		MaxConnLifeTime time.Duration `env-required:"true" yaml:"max_conn_life_time"`
		MaxConnIdleTime time.Duration `env-required:"true" yaml:"max_conn_idle_time"`
	}
	Token struct {
		SignKey string        `env-required:"true" env:"JWT_SIGN_KEY"`
		TTL     time.Duration `env-required:"true" yaml:"ttl"`
	}

	Prometheus struct {
		Port string `env-required:"true" yaml:"port"`
		Path string `env-required:"true" yaml:"path"`
	}
)

func (db Database) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SslMode,
	)
}

func NewConfig(configPath string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	cfg := &Config{}
	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, err
}
