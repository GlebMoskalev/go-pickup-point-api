package service

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=Auth --output=./mocks
type Auth interface {
	DummyLogin(ctx context.Context, role string) (string, error)
	Register(ctx context.Context, email, password, role string) (*entity.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(tokenString string) (*entity.UserClaims, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=PVZ --output=./mocks
type PVZ interface {
	Create(ctx context.Context, city string) (*entity.PVZ, error)
	ListWithDetails(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]entity.PVZWithDetails, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=Reception --output=./mocks
type Reception interface {
	Create(ctx context.Context, pvzID string) (*entity.Reception, error)
	CloseLastReception(ctx context.Context, pvzID string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=Product --output=./mocks
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
