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

func TestUserRepoCreate(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	userRepo := pgxdb.NewUserRepo(dbPool)

	testCases := []struct {
		name        string
		user        entity.User
		expectError bool
		expectedErr error
	}{
		{
			name: "Create user successfully",
			user: entity.User{
				Email: "test@example.com",
				Role:  "employee",
			},
			expectError: false,
		},
		{
			name: "Create duplicate user",
			user: entity.User{
				Email: "duplicate@example.com",
				Role:  "moderator",
			},
			expectError: false,
		},
		{
			name: "Attempt to create duplicate user",
			user: entity.User{
				Email: "duplicate@example.com",
				Role:  "employee",
			},
			expectError: true,
			expectedErr: repoerr.ErrDuplicateEntry,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userRepo.Create(ctx, tc.user)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != nil {
					require.ErrorIs(t, err, tc.expectedErr)
				}
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.NotEqual(t, uuid.Nil, user.ID)
				require.Equal(t, tc.user.Email, user.Email)
				require.Equal(t, tc.user.PasswordHash, user.PasswordHash)
				require.Equal(t, tc.user.Role, user.Role)

				var exists bool
				err = dbPool.QueryRow(ctx,
					`SELECT EXISTS (SELECT id FROM users WHERE email = $1)`,
					tc.user.Email,
				).Scan(&exists)
				require.NoError(t, err)
				require.True(t, exists)
			}
		})
	}
}

func TestUserRepoGetByEmail(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	userRepo := pgxdb.NewUserRepo(dbPool)

	testUser := entity.User{
		Email: "get-by-email@example.com",
		Role:  "employee",
	}
	createdUser, err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)
	require.NotNil(t, createdUser)

	testCases := []struct {
		name        string
		email       string
		expectError bool
		expectedErr error
	}{
		{
			name:        "Get existing user by email",
			email:       "get-by-email@example.com",
			expectError: false,
		},
		{
			name:        "Get non-existent user by email",
			email:       "nonexistent@example.com",
			expectError: true,
			expectedErr: repoerr.ErrNotFound,
		},
		{
			name:        "Get with empty email",
			email:       "",
			expectError: true,
			expectedErr: repoerr.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userRepo.GetByEmail(ctx, tc.email)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != nil {
					require.ErrorIs(t, err, tc.expectedErr)
				}
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tc.email, user.Email)
				require.Equal(t, testUser.PasswordHash, user.PasswordHash)
				require.Equal(t, testUser.Role, user.Role)
				require.NotEqual(t, uuid.Nil, user.ID)
			}
		})
	}
}

func TestUserRepoGetById(t *testing.T) {
	ctx := context.Background()

	postgresContainer, dbCfg := helperstest.SetupPostgresContainer(t, ctx)
	defer postgresContainer.Terminate(ctx)

	dbPool := helperstest.SetupDatabaseConnection(t, ctx, dbCfg)
	defer dbPool.Close()

	helperstest.ApplyMigrations(t, dbCfg)

	userRepo := pgxdb.NewUserRepo(dbPool)

	testUser := entity.User{
		Email: "get-by-id@example.com",
		Role:  "moderator",
	}
	createdUser, err := userRepo.Create(ctx, testUser)
	require.NoError(t, err)
	require.NotNil(t, createdUser)

	testCases := []struct {
		name        string
		id          uuid.UUID
		expectError bool
		expectedErr error
	}{
		{
			name:        "Get existing user by ID",
			id:          createdUser.ID,
			expectError: false,
		},
		{
			name:        "Get non-existent user by ID",
			id:          uuid.New(),
			expectError: true,
			expectedErr: repoerr.ErrNotFound,
		},
		{
			name:        "Get with nil UUID",
			id:          uuid.Nil,
			expectError: true,
			expectedErr: repoerr.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := userRepo.GetById(ctx, tc.id)

			if tc.expectError {
				require.Error(t, err)
				if tc.expectedErr != nil {
					require.ErrorIs(t, err, tc.expectedErr)
				}
				require.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				require.Equal(t, tc.id, user.ID)
				require.Equal(t, testUser.Email, user.Email)
				require.Equal(t, testUser.PasswordHash, user.PasswordHash)
				require.Equal(t, testUser.Role, user.Role)
			}
		})
	}
}
