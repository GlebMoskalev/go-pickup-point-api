package entity

import (
	"github.com/google/uuid"
	"time"
)

const (
	RoleEmployee  = "employee"
	RoleModerator = "moderator"
)

type User struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	Role         string    `db:"role"`
	CreatedAt    time.Time `db:"created_at"`
}
