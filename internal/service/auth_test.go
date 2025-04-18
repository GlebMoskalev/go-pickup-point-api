package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/mocks"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/privacy"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestAuthService_generateJWT(t *testing.T) {
	testCases := []struct {
		name          string
		userID        uuid.UUID
		role          string
		cfgToken      config.Token
		expectedError error
		expectedToken bool
	}{
		{
			name:          "successful generation",
			userID:        uuid.New(),
			role:          entity.RoleEmployee,
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedError: nil,
			expectedToken: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewAuthService(nil, tc.cfgToken, "salt")
			token, err := service.generateJWT(tc.userID, tc.role)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, err := jwt.ParseWithClaims(token, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(tc.cfgToken.SignKey), nil
				})
				assert.NoError(t, err)
				claims, ok := parsedToken.Claims.(*entity.UserClaims)
				assert.True(t, ok)
				assert.True(t, parsedToken.Valid)
				assert.Equal(t, tc.userID, claims.UserID)
				assert.Equal(t, tc.role, claims.Role)
				assert.Equal(t, tc.userID.String(), claims.Subject)
				assert.WithinDuration(t, time.Now().Add(tc.cfgToken.TTL), claims.ExpiresAt.Time, time.Second)
				assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, time.Second)
			}
		})
	}
}

func TestAuthService_DummyLogin(t *testing.T) {
	testCases := []struct {
		name          string
		role          string
		cfgToken      config.Token
		expectedError error
		expectedToken bool
	}{
		{
			name:          "successful employee login",
			role:          entity.RoleEmployee,
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedError: nil,
			expectedToken: true,
		},
		{
			name:          "successful moderator login",
			role:          entity.RoleModerator,
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedError: nil,
			expectedToken: true,
		},
		{
			name:          "invalid role",
			role:          "admin",
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedError: ErrInvalidRole,
			expectedToken: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewAuthService(nil, tc.cfgToken, "salt")
			ctx := context.Background()

			token, err := service.DummyLogin(ctx, tc.role)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, err := jwt.ParseWithClaims(token, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(tc.cfgToken.SignKey), nil
				})
				assert.NoError(t, err)
				claims, ok := parsedToken.Claims.(*entity.UserClaims)
				assert.True(t, ok)
				assert.True(t, parsedToken.Valid)
				assert.Equal(t, tc.role, claims.Role)
				_, err = uuid.Parse(claims.UserID.String())
				assert.NoError(t, err, "UserID should be a valid UUID")
			}
		})
	}
}

func TestAuthService_Register(t *testing.T) {
	testCases := []struct {
		name          string
		email         string
		password      string
		role          string
		prepareRepo   func(repo *mocks.User)
		expectedUser  *entity.User
		expectedError error
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			password: "password123",
			role:     entity.RoleEmployee,
			prepareRepo: func(repo *mocks.User) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, repoerr.ErrNotFound)
				userID := uuid.New()
				repo.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).
					Return(&entity.User{
						ID:           userID,
						Email:        "test@example.com",
						PasswordHash: privacy.HashPassword("password123", "salt"),
						Role:         entity.RoleEmployee,
						CreatedAt:    time.Now(),
					}, nil)
			},
			expectedUser: &entity.User{
				ID:           uuid.UUID{},
				Email:        "test@example.com",
				PasswordHash: "password",
				Role:         entity.RoleEmployee,
				CreatedAt:    time.Time{},
			},
			expectedError: nil,
		},
		{
			name:          "invalid role",
			email:         "test@example.com",
			password:      "password123",
			role:          "admin",
			prepareRepo:   func(repo *mocks.User) {},
			expectedUser:  nil,
			expectedError: ErrInvalidRole,
		},
		{
			name:          "invalid email",
			email:         "invalid-email",
			password:      "password123",
			role:          entity.RoleModerator,
			prepareRepo:   func(repo *mocks.User) {},
			expectedUser:  nil,
			expectedError: ErrInvalidEmail,
		},
		{
			name:     "user already exists",
			email:    "test@example.com",
			password: "password123",
			role:     entity.RoleEmployee,
			prepareRepo: func(repo *mocks.User) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(&entity.User{Email: "test@example.com"}, nil)
			},
			expectedUser:  nil,
			expectedError: ErrUserExists,
		},
		{
			name:     "user repo error on check",
			email:    "test@example.com",
			password: "password123",
			role:     entity.RoleEmployee,
			prepareRepo: func(repo *mocks.User) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: ErrInternal,
		},
		{
			name:     "user repo error on create",
			email:    "test@example.com",
			password: "password123",
			role:     entity.RoleEmployee,
			prepareRepo: func(repo *mocks.User) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, repoerr.ErrNotFound)
				repo.On("Create", mock.Anything, mock.AnythingOfType("entity.User")).
					Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: ErrInternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewUser(t)
			tc.prepareRepo(userRepo)
			service := NewAuthService(userRepo, config.Token{SignKey: "secret", TTL: time.Hour}, "salt")
			ctx := context.Background()

			user, err := service.Register(ctx, tc.email, tc.password, tc.role)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				_, err := uuid.Parse(user.ID.String())
				assert.NoError(t, err, "User ID should be a valid UUID")
				assert.Equal(t, tc.expectedUser.Email, user.Email)
				assert.Equal(t, tc.expectedUser.Role, user.Role)
				assert.True(t, privacy.VerifyPassword(tc.password, user.PasswordHash, "salt"))
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	testCases := []struct {
		name          string
		email         string
		password      string
		prepareRepo   func(repo *mocks.User)
		cfgToken      config.Token
		expectedToken bool
		expectedError error
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			prepareRepo: func(repo *mocks.User) {
				userID := uuid.New()
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(&entity.User{
						ID:           userID,
						Email:        "test@example.com",
						PasswordHash: privacy.HashPassword("password123", "salt"),
						Role:         entity.RoleEmployee,
					}, nil)
			},
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedToken: true,
			expectedError: nil,
		},
		{
			name:     "user not found",
			email:    "test@example.com",
			password: "password123",
			prepareRepo: func(repo *mocks.User) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, repoerr.ErrNotFound)
			},
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedToken: false,
			expectedError: ErrInvalidCredentials,
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			prepareRepo: func(repo *mocks.User) {
				userID := uuid.New()
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(&entity.User{
						ID:           userID,
						Email:        "test@example.com",
						PasswordHash: privacy.HashPassword("password123", "salt"),
						Role:         entity.RoleEmployee,
					}, nil)
			},
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedToken: false,
			expectedError: ErrInvalidCredentials,
		},
		{
			name:     "repo error",
			email:    "test@example.com",
			password: "password123",
			prepareRepo: func(repo *mocks.User) {
				repo.On("GetByEmail", mock.Anything, "test@example.com").
					Return(nil, errors.New("database error"))
			},
			cfgToken:      config.Token{SignKey: "secret", TTL: time.Hour},
			expectedToken: false,
			expectedError: ErrInternal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := mocks.NewUser(t)
			tc.prepareRepo(userRepo)
			service := NewAuthService(userRepo, tc.cfgToken, "salt")
			ctx := context.Background()

			token, err := service.Login(ctx, tc.email, tc.password)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, err := jwt.ParseWithClaims(token, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(tc.cfgToken.SignKey), nil
				})
				assert.NoError(t, err)
				claims, ok := parsedToken.Claims.(*entity.UserClaims)
				assert.True(t, ok)
				assert.True(t, parsedToken.Valid)
				_, err = uuid.Parse(claims.UserID.String())
				assert.NoError(t, err, "UserID should be a valid UUID")
				assert.Equal(t, entity.RoleEmployee, claims.Role)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	userID := uuid.New()
	validToken := func(signKey string, ttl time.Duration, role string) string {
		claims := entity.UserClaims{
			UserID: userID,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Subject:   userID.String(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(signKey))
		return tokenString
	}
	expiredToken := func(signKey string, role string) string {
		claims := entity.UserClaims{
			UserID: userID,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Subject:   userID.String(),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(signKey))
		return tokenString
	}

	testCases := []struct {
		name           string
		tokenString    string
		cfgToken       config.Token
		expectedClaims *entity.UserClaims
		expectedError  error
	}{
		{
			name:        "valid token",
			tokenString: validToken("secret", time.Hour, entity.RoleEmployee),
			cfgToken:    config.Token{SignKey: "secret", TTL: time.Hour},
			expectedClaims: &entity.UserClaims{
				UserID: userID,
				Role:   entity.RoleEmployee,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: userID.String(),
				},
			},
			expectedError: nil,
		},
		{
			name:           "expired token",
			tokenString:    expiredToken("secret", entity.RoleEmployee),
			cfgToken:       config.Token{SignKey: "secret", TTL: time.Hour},
			expectedClaims: nil,
			expectedError:  ErrTokenExpired,
		},
		{
			name:           "invalid signature",
			tokenString:    validToken("wrongkey", time.Hour, entity.RoleEmployee),
			cfgToken:       config.Token{SignKey: "secret", TTL: time.Hour},
			expectedClaims: nil,
			expectedError:  ErrInvalidToken,
		},
		{
			name:           "invalid token format",
			tokenString:    "invalid.token.format",
			cfgToken:       config.Token{SignKey: "secret", TTL: time.Hour},
			expectedClaims: nil,
			expectedError:  ErrInvalidToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewAuthService(nil, tc.cfgToken, "salt")

			claims, err := service.ValidateToken(tc.tokenString)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, tc.expectedClaims.UserID, claims.UserID)
				assert.Equal(t, tc.expectedClaims.Role, claims.Role)
				assert.Equal(t, tc.expectedClaims.Subject, claims.Subject)
				assert.WithinDuration(t, time.Now().Add(tc.cfgToken.TTL), claims.ExpiresAt.Time, time.Second)
				assert.WithinDuration(t, time.Now(), claims.IssuedAt.Time, time.Second)
			}
		})
	}
}
