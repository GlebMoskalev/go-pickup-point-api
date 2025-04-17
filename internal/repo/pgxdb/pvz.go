package pgxdb

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type PVZRepo struct {
	db *pgxpool.Pool
}

func NewPVZRepo(db *pgxpool.Pool) *PVZRepo {
	return &PVZRepo{db: db}
}

func (r *PVZRepo) Create(ctx context.Context, city string) (*entity.PVZ, error) {
	log := slog.With("layer", "PVZRepo", "operation", "Create", "city", city)
	log.Debug("starting pvz creation")

	query := `
	INSERT INTO pvz (city)
	VALUES ($1)
	RETURNING id, registration_date
`
	var id uuid.UUID
	var date time.Time
	err := r.db.QueryRow(ctx, query, city).Scan(&id, &date)
	if err != nil {
		log.Error("failed to create pvz", "error", err)
		return nil, err
	}

	pvz := entity.PVZ{
		ID:               id,
		RegistrationDate: date,
		City:             city,
	}

	log.Info("pvz created successfully", "pvzID", id.String())
	return &pvz, nil
}

func (r *PVZRepo) Exists(ctx context.Context, pvzID string) bool {
	log := slog.With("layer", "PVZRepo", "operation", "Exists", "pvzID", pvzID)
	log.Debug("checking pvz existence")

	query := `
	SELECT EXISTS(
		SELECT 1
		FROM pvz
		WHERE id = $1
	)
`

	var exists bool
	err := r.db.QueryRow(ctx, query, pvzID).Scan(&exists)
	if err != nil {
		log.Error("failed to check pvz existence", "error", err)
		return false
	}

	log.Debug("pvz existence check completed", "exists", exists)
	return exists
}
