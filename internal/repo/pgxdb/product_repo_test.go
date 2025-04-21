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
)

func TestProductRepoCreate(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	productRepo := pgxdb.NewProductRepo(dbPool)
	pvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	validReceptionID := helperstest.CreateReception(t, ctx, dbPool, pvzID)
	invalidReceptionID := uuid.New()

	testCases := []struct {
		name        string
		receptionID string
		productType string
		expectError bool
	}{
		{
			name:        "Create product successfully with shoes type",
			receptionID: validReceptionID.String(),
			productType: entity.ProductTypeShoes,
			expectError: false,
		},
		{
			name:        "Create product successfully with electronics type",
			receptionID: validReceptionID.String(),
			productType: entity.ProductTypeElectronics,
			expectError: false,
		},
		{
			name:        "Create product with invalid reception ID",
			receptionID: invalidReceptionID.String(),
			productType: entity.ProductTypeElectronics,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			product, err := productRepo.Create(ctx, tc.receptionID, tc.productType)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, product)
			} else {
				require.NoError(t, err)
				require.NotNil(t, product)
				require.Equal(t, tc.productType, product.Type)
				var exists bool
				err = dbPool.QueryRow(ctx,
					`SELECT EXISTS (SELECT id FROM products WHERE id = $1)`,
					product.ID,
				).Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists)
			}
		})
	}
}

func TestProductRepoDeleteLastProduct(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	productRepo := pgxdb.NewProductRepo(dbPool)
	pvzID := helperstest.CreatePVZ(t, ctx, dbPool)
	validReceptionID := helperstest.CreateReception(t, ctx, dbPool, pvzID)
	invalidReceptionID := uuid.New()

	product1, err := productRepo.Create(ctx, validReceptionID.String(), entity.ProductTypeShoes)
	require.NoError(t, err)

	product2, err := productRepo.Create(ctx, validReceptionID.String(), entity.ProductTypeElectronics)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		receptionID string
		expectError bool
		expectedErr error
	}{
		{
			name:        "Delete last product successfully",
			receptionID: validReceptionID.String(),
			expectError: false,
		},
		{
			name:        "Delete last product with invalid reception ID",
			receptionID: invalidReceptionID.String(),
			expectError: true,
			expectedErr: repoerr.ErrNoRows,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := productRepo.DeleteLastProduct(ctx, tc.receptionID)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != nil {
					require.Equal(t, tc.expectedErr, err)
				}
			} else {
				require.NoError(t, err)

				var exists bool
				err = dbPool.QueryRow(ctx,
					`SELECT EXISTS (SELECT id FROM products WHERE id = $1)`,
					product2.ID,
				).Scan(&exists)
				require.NoError(t, err)
				require.False(t, exists)

				err = dbPool.QueryRow(ctx,
					`SELECT EXISTS (SELECT id FROM products WHERE id = $1)`,
					product1.ID,
				).Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists)
			}
		})
	}
}
