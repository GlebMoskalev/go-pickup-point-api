package repo

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/pgxdb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User interface {
	Create(ctx context.Context, user entity.User) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetById(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

type Repositories struct {
	User
}

func NewRepositories(db *pgxpool.Pool) *Repositories {
	return &Repositories{
		User: pgxdb.NewUserRepo(db),
	}
}
