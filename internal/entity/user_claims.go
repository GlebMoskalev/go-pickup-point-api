package entity

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	UserID uuid.UUID `json:"id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}
