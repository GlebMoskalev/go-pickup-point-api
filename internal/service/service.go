package service

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
)

type Auth interface {
	DummyLogin(ctx context.Context, role string) (string, error)
	Register(ctx context.Context, email, password, role string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(tokenString string) (*entity.UserClaims, error)
}

type Services struct {
	Auth Auth
}

func NewServices(repositories *repo.Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth: NewAuthService(repositories.User, cfg.Token, cfg.Salt),
	}
}
