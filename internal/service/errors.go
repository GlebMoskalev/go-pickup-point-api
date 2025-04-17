package service

import "errors"

var (
	ErrInternal = errors.New("internal server error")

	ErrInvalidRole        = errors.New("invalid role")
	ErrUserExists         = errors.New("user exists")
	ErrInvalidCredentials = errors.New("invalid credentials ")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidEmail       = errors.New("invalid email")
)
