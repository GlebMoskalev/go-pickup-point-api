package pgxdb

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type ReceptionRepo struct {
	db *pgxpool.Pool
}

func NewReceptionRepo(db *pgxpool.Pool) *ReceptionRepo {
	return &ReceptionRepo{db: db}
}

func (r *ReceptionRepo) Create(ctx context.Context, pvzID string) (*entity.Reception, error) {
	log := slog.With("layer", "ReceptionRepo", "operation", "Create", "pvzID", pvzID)
	log.Debug("starting reception creation")

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
	INSERT INTO receptions (pvz_id, status)
	VALUES ($1, 'in_progress')
	RETURNING id, date_time, pvz_id
`
	var id, pvzUUID uuid.UUID
	var dateTime time.Time
	err = tx.QueryRow(ctx, query, pvzID).Scan(&id, &dateTime, &pvzUUID)
	if err != nil {
		log.Error("failed to create reception", "error", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return nil, err
	}

	reception := &entity.Reception{
		ID:       id,
		DateTime: dateTime,
		PVZID:    pvzUUID,
		Status:   entity.StatusInProgress,
	}

	log.Info("reception created successfully", "receptionID", id.String())
	return reception, nil
}

func (r *ReceptionRepo) HasOpenReception(ctx context.Context, pvzID string) (bool, error) {
	log := slog.With("layer", "ReceptionRepo", "operation", "HasOpenReception", "pvzID", pvzID)
	log.Debug("checking for open reception")

	query := `
    SELECT EXISTS (
        SELECT 1
        FROM receptions
        WHERE pvz_id = $1 AND status = 'in_progress'
    )
    `
	var exists bool
	err := r.db.QueryRow(ctx, query, pvzID).Scan(&exists)
	if err != nil {
		log.Error("failed to check open reception", "error", err)
		return false, err
	}

	log.Debug("open reception check completed", "exists", exists)
	return exists, nil
}

func (r *ReceptionRepo) GetLastOpenReception(ctx context.Context, pvzID string) (*entity.Reception, error) {
	log := slog.With("layer", "ReceptionRepo", "operation", "GetLastOpenReception", "pvzID", pvzID)
	log.Debug("retrieving last open reception")

	query := `
	SELECT id, pvz_id, status, date_time
	FROM receptions
	WHERE pvz_id = $1 AND status = 'in_progress'
	ORDER BY date_time DESC
	LIMIT 1
`

	var reception entity.Reception
	err := r.db.QueryRow(ctx, query, pvzID).Scan(&reception.ID, &reception.PVZID, &reception.Status, &reception.DateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("not found open reception")
			return nil, repoerr.ErrNoRows
		}
		log.Error("failed to get last open reception", "error", err)
		return nil, err
	}
	log.Info("last open reception retrieved", "receptionID", reception.ID.String())
	return &reception, nil
}

func (r *ReceptionRepo) Close(ctx context.Context, receptionID string) error {
	log := slog.With("layer", "ReceptionRepo", "operation", "Close", "receptionID", receptionID)
	log.Debug("starting reception closure")

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error("failed to begin transaction", "error", err)
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				log.Error("failed to rollback transaction", "error", rollbackErr)
			}
		}
	}()

	query := `
	UPDATE receptions
	SET status = 'close'
	WHERE id = $1 AND status = 'in_progress'
	RETURNING id
`
	var id uuid.UUID
	err = tx.QueryRow(ctx, query, receptionID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("not found in_progress reception")
			return repoerr.ErrNoRows
		}
		log.Error("failed to close reception", "error", err)
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error("failed to commit transaction", "error", err)
		return err
	}

	log.Info("reception closed successfully", "receptionID", id.String())
	return nil
}
