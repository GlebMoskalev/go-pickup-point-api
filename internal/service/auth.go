package service

import (
	"context"
	"errors"
	"github.com/GlebMoskalev/go-pickup-point-api/config"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo"
	"github.com/GlebMoskalev/go-pickup-point-api/internal/repo/repoerr"
	"github.com/GlebMoskalev/go-pickup-point-api/pkg/privacy"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log/slog"
	"regexp"
	"time"
)

type AuthService struct {
	userRepo repo.User
	cfgToken config.Token
	salt     string
}

func NewAuthService(userRepo repo.User, cfgToken config.Token, salt string) *AuthService {
	return &AuthService{userRepo: userRepo, cfgToken: cfgToken, salt: salt}
}

func (s *AuthService) generateJWT(userID uuid.UUID, role string) (string, error) {
	log := slog.With("layer", "AuthService", "operation", "generateJWT", "userID", userID.String())
	log.Debug("starting JWT generation")

	claims := entity.UserClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfgToken.TTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfgToken.SignKey))
	if err != nil {
		log.Error("failed to generate JWT", "error", err)
		return "", err
	}
	log.Info("JWT generated successfully")
	return tokenString, nil
}

func (s *AuthService) DummyLogin(ctx context.Context, role string) (string, error) {
	log := slog.With("layer", "AuthService", "operation", "DummyLogin", "role", role)
	log.Debug("starting dummy login")

	if role != entity.RoleEmployee && role != entity.RoleModerator {
		log.Warn("invalid role provided")
		return "", ErrInvalidRole
	}
	dummyID := uuid.New()
	token, err := s.generateJWT(dummyID, role)
	if err != nil {
		log.Error("failed to generate JWT for dummy login", "error", err)
		return "", ErrInternal
	}

	log.Info("dummy login successful", "dummyUserID", dummyID.String())
	return token, nil
}

func (s *AuthService) Register(ctx context.Context, email, password, role string) (*entity.User, error) {
	log := slog.With("layer", "AuthService", "operation", "Register", "email", privacy.MaskEmail(email))
	log.Debug("starting user registration")

	if role != entity.RoleEmployee && role != entity.RoleModerator {
		log.Warn("invalid role provided")
		return nil, ErrInvalidRole
	}

	var emailRegexp = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegexp.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		log.Warn("user already exists")
		return nil, ErrUserExists
	}
	if !errors.Is(err, repoerr.ErrNotFound) {
		log.Error("failed to check user existence", "error", err)
		return nil, ErrInternal
	}

	passwordHash := privacy.HashPassword(password, s.salt)

	user := entity.User{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		log.Error("failed to create user", "error", err)
		return nil, ErrInternal
	}

	log.Info("user registered successfully", "userID", createdUser.ID.String())
	return createdUser, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	log := slog.With("layer", "AuthService", "operation", "Login", "email", privacy.MaskEmail(email))
	log.Debug("starting user login")

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repoerr.ErrNotFound) {
			log.Warn("invalid credentials: user not found")
			return "", ErrInvalidCredentials
		}
		log.Error("failed to get user", "error", err)
		return "", ErrInternal
	}
	if !privacy.VerifyPassword(password, user.PasswordHash, s.salt) {
		log.Warn("invalid credentials: password mismatch")
		return "", ErrInvalidCredentials
	}

	token, err := s.generateJWT(user.ID, user.Role)
	if err != nil {
		log.Error("failed to generate JWT for login", "error", err)
		return "", ErrInternal
	}

	log.Info("user login successful", "userID", user.ID.String())
	return token, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*entity.UserClaims, error) {
	log := slog.With("layer", "AuthService", "operation", "ValidateToken")
	log.Debug("starting token validation")

	token, err := jwt.ParseWithClaims(tokenString, &entity.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfgToken.SignKey), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Warn("token expired")
			return nil, ErrTokenExpired
		}

		log.Error("failed to parse token", "error", err)
		return nil, ErrInternal
	}
	if claims, ok := token.Claims.(*entity.UserClaims); ok && token.Valid {
		log.Info("token validated successfully", "userID", claims.UserID.String())
		return claims, nil
	}

	log.Warn("invalid token")
	return nil, ErrInvalidToken
}
