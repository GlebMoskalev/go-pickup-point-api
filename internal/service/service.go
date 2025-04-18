package service

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"time"
)

type Auth interface {
	DummyLogin(ctx context.Context, role string) (string, error)
	Register(ctx context.Context, email, password, role string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(tokenString string) (*entity.UserClaims, error)
}

type PVZ interface {
	Create(ctx context.Context, city string) (*entity.PVZ, error)
	ListWithDetails(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]entity.PVZWithDetails, error)
}

type Reception interface {
	Create(ctx context.Context, pvzID string) (*entity.Reception, error)
	CloseLastReception(ctx context.Context, pvzID string) error
}

type Product interface {
	Create(ctx context.Context, pvzID, productType string) (*entity.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID string) error
}

type Services struct {
	Auth      Auth
	PVZ       PVZ
	Reception Reception
	Product   Product
}

func NewServices(repositories *repo.Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth:      NewAuthService(repositories.User, cfg.Token, cfg.Salt),
		PVZ:       NewPVZService(repositories.PVZ),
		Reception: NewReceptionService(repositories.Reception, repositories.PVZ),
		Product:   NewProductService(repositories.Product, repositories.Reception, repositories.PVZ),
	}
}
