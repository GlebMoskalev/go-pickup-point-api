package entity

import (
	"github.com/google/uuid"
	"time"
)

const (
	ProductTypeElectronics = "электроника"
	ProductTypeClothes     = "одежда"
	ProductTypeShoes       = "обувь"
)

type Product struct {
	ID          uuid.UUID `db:"id"`
	DateTime    time.Time `db:"date_time"`
	Type        string    `db:"type"`
	ReceptionId uuid.UUID `db:"reception_id"`
	OrderNumber int       `db:"order_number"`
}
