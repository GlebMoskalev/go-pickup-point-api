package pgxdb_test

import (
	"context"
	"github.com/GlebMoskalev/go-pickup-point-api/integration/helperstest"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/pgxdb"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestReceptionRepoCreate(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	receptionRepo := pgxdb.NewReceptionRepo(dbPool)

	validPvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	invalidPvzID := uuid.New()

	testCases := []struct {
		name        string
		pvzID       string
		expectError bool
	}{
		{
			name:        "Create reception successfully",
			pvzID:       validPvzID.String(),
			expectError: false,
		},
		{
			name:        "Create reception with invalid PVZ ID",
			pvzID:       invalidPvzID.String(),
			expectError: true,
		},
		{
			name:        "Create reception with empty PVZ ID",
			pvzID:       "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reception, err := receptionRepo.Create(ctx, tc.pvzID)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, reception)
			} else {
				require.NoError(t, err)
				require.NotNil(t, reception)
				require.NotEqual(t, uuid.Nil, reception.ID)
				require.Equal(t, entity.StatusInProgress, reception.Status)

				var exists bool
				err = dbPool.QueryRow(ctx,
					`SELECT EXISTS (SELECT id FROM receptions WHERE id = $1)`,
					reception.ID,
				).Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists)
			}
		})
	}
}

func TestReceptionRepoHasOpenReception(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	receptionRepo := pgxdb.NewReceptionRepo(dbPool)

	pvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	emptyPvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	invalidPvzID := uuid.New()

	reception, err := receptionRepo.Create(ctx, pvzID.String())
	require.NoError(t, err)
	require.NotNil(t, reception)

	testCases := []struct {
		name          string
		pvzID         string
		expectError   bool
		expectedValue bool
	}{
		{
			name:          "PVZ with open reception",
			pvzID:         pvzID.String(),
			expectError:   false,
			expectedValue: true,
		},
		{
			name:          "PVZ without open reception",
			pvzID:         emptyPvzID.String(),
			expectError:   false,
			expectedValue: false,
		},
		{
			name:          "Invalid PVZ ID",
			pvzID:         invalidPvzID.String(),
			expectError:   false,
			expectedValue: false,
		},
		{
			name:          "Empty PVZ ID",
			pvzID:         "",
			expectError:   true,
			expectedValue: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasOpen, err := receptionRepo.HasOpenReception(ctx, tc.pvzID)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedValue, hasOpen)
			}
		})
	}
}

func TestReceptionRepoGetLastOpenReception(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	receptionRepo := pgxdb.NewReceptionRepo(dbPool)

	pvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	emptyPvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	invalidPvzID := uuid.New()

	firstReception, err := receptionRepo.Create(ctx, pvzID.String())
	require.NoError(t, err)
	require.NotNil(t, firstReception)

	time.Sleep(10 * time.Millisecond)

	secondReception, err := receptionRepo.Create(ctx, pvzID.String())
	require.NoError(t, err)
	require.NotNil(t, secondReception)

	testCases := []struct {
		name         string
		pvzID        string
		expectError  bool
		expectedErr  error
		checkDetails bool
	}{
		{
			name:         "Get last open reception - should return latest",
			pvzID:        pvzID.String(),
			expectError:  false,
			checkDetails: true,
		},
		{
			name:        "PVZ without any receptions",
			pvzID:       emptyPvzID.String(),
			expectError: true,
			expectedErr: repoerr.ErrNoRows,
		},
		{
			name:        "Invalid PVZ ID",
			pvzID:       invalidPvzID.String(),
			expectError: true,
			expectedErr: repoerr.ErrNoRows,
		},
		{
			name:        "Empty PVZ ID",
			pvzID:       "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reception, err := receptionRepo.GetLastOpenReception(ctx, tc.pvzID)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != nil {
					require.ErrorIs(t, err, tc.expectedErr)
				}
				require.Nil(t, reception)
			} else {
				require.NoError(t, err)
				require.NotNil(t, reception)

				if tc.checkDetails {
					require.Equal(t, secondReception.ID, reception.ID)
					require.Equal(t, entity.StatusInProgress, reception.Status)
				}
			}
		})
	}
}

func TestReceptionRepoClose(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	receptionRepo := pgxdb.NewReceptionRepo(dbPool)

	pvzID := helperstest.CreatePVZ(t, ctx, dbPool)

	openReception, err := receptionRepo.Create(ctx, pvzID.String())
	require.NoError(t, err)
	require.NotNil(t, openReception)

	closedReceptionID := helperstest.CreateAndCloseReception(t, ctx, dbPool, pvzID)

	invalidReceptionID := uuid.New()

	testCases := []struct {
		name        string
		receptionID string
		expectError bool
		expectedErr error
	}{
		{
			name:        "Close reception successfully",
			receptionID: openReception.ID.String(),
			expectError: false,
		},
		{
			name:        "Reception already closed",
			receptionID: closedReceptionID.String(),
			expectError: true,
			expectedErr: repoerr.ErrNoRows,
		},
		{
			name:        "Invalid reception ID",
			receptionID: invalidReceptionID.String(),
			expectError: true,
			expectedErr: repoerr.ErrNoRows,
		},
		{
			name:        "Empty reception ID",
			receptionID: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := receptionRepo.Close(ctx, tc.receptionID)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != nil {
					require.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				require.NoError(t, err)

				var status string
				err = dbPool.QueryRow(ctx,
					`SELECT status FROM receptions WHERE id = $1`,
					tc.receptionID,
				).Scan(&status)
				require.NoError(t, err)
				require.Equal(t, "close", status)
			}
		})
	}
}
