package pgxdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/GlebMoskalev/go-pickup-point-api/integration/helperstest"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/pgxdb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPVZRepoCreate(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	pvzRepo := pgxdb.NewPVZRepo(dbPool)

	testCases := []struct {
		name        string
		city        string
		expectError bool
	}{
		{
			name:        "Create PVZ successfully with Moscow city",
			city:        entity.CityMoscow,
			expectError: false,
		},
		{
			name:        "Create PVZ successfully with Saint Petersburg city",
			city:        entity.CitySPB,
			expectError: false,
		},
		{
			name:        "Create PVZ successfully with Kazan city",
			city:        entity.CityKazan,
			expectError: false,
		},
		{
			name:        "Create PVZ with empty city",
			city:        "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pvz, err := pvzRepo.Create(ctx, tc.city)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, pvz)
			} else {
				require.NoError(t, err)
				require.NotNil(t, pvz)
				require.Equal(t, tc.city, pvz.City)
				require.NotEqual(t, uuid.Nil, pvz.ID)
				require.False(t, pvz.RegistrationDate.IsZero())

				var exists bool
				err = dbPool.QueryRow(ctx,
					`SELECT EXISTS (SELECT id FROM pvz WHERE id = $1)`,
					pvz.ID,
				).Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists)
			}
		})
	}
}

func TestPVZRepoExists(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	pvzRepo := pgxdb.NewPVZRepo(dbPool)

	pvz, err := pvzRepo.Create(ctx, entity.CityMoscow)
	require.NoError(t, err)

	invalidPVZID := uuid.New().String()

	testCases := []struct {
		name         string
		pvzID        string
		expectExists bool
	}{
		{
			name:         "PVZ exists",
			pvzID:        pvz.ID.String(),
			expectExists: true,
		},
		{
			name:         "PVZ does not exist",
			pvzID:        invalidPVZID,
			expectExists: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists := pvzRepo.Exists(ctx, tc.pvzID)
			require.Equal(t, tc.expectExists, exists)
		})
	}
}

func TestPVZRepoListWithDetails(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	pvzRepo := pgxdb.NewPVZRepo(dbPool)
	productRepo := pgxdb.NewProductRepo(dbPool)

	pvz1, err := pvzRepo.Create(ctx, entity.CityMoscow)
	require.NoError(t, err)

	pvz2, err := pvzRepo.Create(ctx, entity.CitySPB)
	require.NoError(t, err)

	reception1ID := helperstest.CreateReception(t, ctx, dbPool, pvz1.ID)
	reception2ID := helperstest.CreateReceptionWithStatus(t, ctx, dbPool, pvz1.ID, entity.StatusClose)
	reception3ID := helperstest.CreateReception(t, ctx, dbPool, pvz2.ID)

	_, err = productRepo.Create(ctx, reception1ID.String(), entity.ProductTypeShoes)
	require.NoError(t, err)
	_, err = productRepo.Create(ctx, reception1ID.String(), entity.ProductTypeElectronics)
	require.NoError(t, err)
	_, err = productRepo.Create(ctx, reception2ID.String(), entity.ProductTypeShoes)
	require.NoError(t, err)
	_, err = productRepo.Create(ctx, reception3ID.String(), entity.ProductTypeElectronics)
	require.NoError(t, err)

	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	testCases := []struct {
		name           string
		startDate      *time.Time
		endDate        *time.Time
		page           int
		limit          int
		expectCount    int
		expectedCities []string
	}{
		{
			name:           "List all PVZs with details",
			startDate:      nil,
			endDate:        nil,
			page:           1,
			limit:          10,
			expectCount:    2,
			expectedCities: []string{entity.CityMoscow, entity.CitySPB},
		},
		{
			name:           "List PVZs with date range",
			startDate:      &pastTime,
			endDate:        &futureTime,
			page:           1,
			limit:          10,
			expectCount:    2,
			expectedCities: []string{entity.CityMoscow, entity.CitySPB},
		},
		{
			name:           "List PVZs with future start date",
			startDate:      &futureTime,
			endDate:        nil,
			page:           1,
			limit:          10,
			expectCount:    0,
			expectedCities: []string{},
		},
		{
			name:           "List PVZs with pagination - first page",
			startDate:      nil,
			endDate:        nil,
			page:           1,
			limit:          1,
			expectCount:    1,
			expectedCities: []string{entity.CityMoscow},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pvzList, err := pvzRepo.ListWithDetails(ctx, tc.startDate, tc.endDate, tc.page, tc.limit)
			require.NoError(t, err)
			require.Len(t, pvzList, tc.expectCount)

			if tc.expectCount > 0 {
				cities := make([]string, 0, len(pvzList))
				for _, pvz := range pvzList {
					cities = append(cities, pvz.PVZ.City)
				}

				for _, expectedCity := range tc.expectedCities {
					found := false
					for _, city := range cities {
						if city == expectedCity {
							found = true
							break
						}
					}
					require.True(t, found, "Expected city %s not found in results", expectedCity)
				}
			}
		})
	}
}
