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

	ErrInvalidCity         = errors.New("invalid city")
	ErrInvalidPVZID        = errors.New("invalid pvz id")
	ErrOpenReceptionExists = errors.New("open reception exists")
	ErrNoOpenReception     = errors.New("no open reception exists")
	ErrInvalidProductType  = errors.New("invalid product type")
	ErrNoProducts          = errors.New("no products")
)
