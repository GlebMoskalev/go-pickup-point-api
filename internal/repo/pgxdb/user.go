package pgxdb

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/privacy"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user entity.User) (*entity.User, error) {
	log := slog.With("layer", "UserRepo", "operation", "Create", "email", privacy.MaskEmail(user.Email))
	log.Debug("starting user creation")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return nil, err
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Error("failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	query := `
	INSERT INTO users 
	    (email, password_hash, role) 
	VALUES ($1, $2, $3)
	RETURNING id
`
	var id uuid.UUID
	err = tx.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role).Scan(&id)
	if err != nil {
		var pgxError *pgconn.PgError
		if errors.As(err, &pgxError) {
			if pgxError.Code == "23505" {
				log.Warn("duplicate entry for user")
				return nil, repoerr.ErrDuplicateEntry
			}
		}
		log.Error("failed to create user", "error", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return nil, err
	}

	user.ID = id

	log.Info("user created successfully", "userID", id.String())
	return &user, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	log := slog.With("layer", "UserRepo", "operation", "GetByEmail", "email", privacy.MaskEmail(email))
	log.Debug("starting get user by email")

	query := `
	SELECT id, password_hash, role
	FROM users
	WHERE email = $1
`
	row := r.db.QueryRow(ctx, query, email)

	var user entity.User
	if err := row.Scan(&user.ID, &user.PasswordHash, &user.Role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("not found user")
			return nil, repoerr.ErrNotFound
		}
		log.Error("failed to get user by email", "error", err)
		return nil, err
	}
	user.Email = email

	log.Info("successfully get user by email", "userID", user.ID.String())
	return &user, nil
}

func (r *UserRepo) GetById(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	log := slog.With("layer", "UserRepo", "operation", "GetById", "userID", id.String())
	log.Debug("starting get user by id")

	query := `
	SELECT email, password_hash, role
	FROM users
	WHERE id = $1
`
	row := r.db.QueryRow(ctx, query, id)

	var user entity.User
	if err := row.Scan(&user.Email, &user.PasswordHash, &user.Role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Warn("not found user")
			return nil, repoerr.ErrNotFound
		}
		log.Error("failed to get user by id", "error", err)
		return nil, err
	}
	user.ID = id

	log.Info("successfully get user by id", "email", privacy.MaskEmail(user.Email))
	return &user, nil
}
