package repo

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/pgxdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=User --output=./mocks
type User interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetById(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=PVZ --output=./mocks
type PVZ interface {
	Create(ctx context.Context, city string) (*entity.PVZ, error)
	Exists(ctx context.Context, pvzID string) bool
	ListWithDetails(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]entity.PVZWithDetails, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=Reception --output=./mocks
type Reception interface {
	Create(ctx context.Context, pvzID string) (*entity.Reception, error)
	HasOpenReception(ctx context.Context, pvzID string) (bool, error)
	GetLastOpenReception(ctx context.Context, pvzID string) (*entity.Reception, error)
	Close(ctx context.Context, receptionID string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=Product --output=./mocks
type Product interface {
	Create(ctx context.Context, receptionID, productType string) (*entity.Product, error)
	DeleteLastProduct(ctx context.Context, receptionID string) error
}

type Repositories struct {
	User
	PVZ
	Reception
	Product
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		User:      pgxdb.NewUserRepo(db),
		PVZ:       pgxdb.NewPVZRepo(db),
		Reception: pgxdb.NewReceptionRepo(db),
		Product:   pgxdb.NewProductRepo(db),
	}
}
