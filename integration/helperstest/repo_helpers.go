package helperstest

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"testing"
)

func CreatePVZ(t *testing.T, ctx context.Context, dbPool *pgxpool.Pool) uuid.UUID {
	var pvzID uuid.UUID
	err := dbPool.QueryRow(ctx,
		`INSERT INTO pvz (city) VALUES ($1) RETURNING id`, "Москва").Scan(&pvzID)

	require.NoError(t, err)
	require.NotNil(t, pvzID)
	return pvzID
}

func CreateReception(t *testing.T, ctx context.Context, dbPool *pgxpool.Pool, pvzID uuid.UUID) uuid.UUID {
	var receptionID uuid.UUID
	err := dbPool.QueryRow(ctx,
		`INSERT INTO receptions (pvz_id, status) VALUES ($1, $2) RETURNING id`,
		pvzID.String(), entity.StatusInProgress,
	).Scan(&receptionID)

	require.NoError(t, err)
	require.NotNil(t, receptionID)
	return receptionID
}

func CreateAndCloseReception(t *testing.T, ctx context.Context, dbPool *pgxpool.Pool, pvzID uuid.UUID) uuid.UUID {
	var receptionID uuid.UUID
	err := dbPool.QueryRow(ctx,
		`INSERT INTO receptions (pvz_id, status) VALUES ($1, $2) RETURNING id`,
		pvzID.String(), entity.StatusInProgress,
	).Scan(&receptionID)
	require.NoError(t, err)

	_, err = dbPool.Exec(ctx,
		`UPDATE receptions SET status = 'close' WHERE id = $1`,
		receptionID)
	require.NoError(t, err)

	return receptionID
}

func CreateReceptionWithStatus(t *testing.T, ctx context.Context, dbPool *pgxpool.Pool, pvzID uuid.UUID, status string) uuid.UUID {
	var receptionID uuid.UUID
	err := dbPool.QueryRow(ctx,
		`INSERT INTO receptions (pvz_id, status) VALUES ($1, $2) RETURNING id`,
		pvzID, status,
	).Scan(&receptionID)
	require.NoError(t, err)
	return receptionID
}
