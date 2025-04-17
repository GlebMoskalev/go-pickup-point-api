package entity

import (
	"github.com/google/uuid"
	"time"
)

const (
	CityMoscow = "Москва"
	CitySPB    = "Санкт-Петербург"
	CityKazan  = "Казань"
)

type PVZ struct {
	ID               uuid.UUID `db:"id"`
	RegistrationDate time.Time `db:"registration_date"`
	City             string    `db:"city"`
}
