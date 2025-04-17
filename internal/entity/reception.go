package entity

import (
	"github.com/google/uuid"
	"time"
)

const (
	StatusInProgress = "in_progress"
	StatusClose      = "close"
)

type Reception struct {
	ID       uuid.UUID `db:"id"`
	DateTime time.Time `db:"date_time"`
	PVZID    uuid.UUID `db:"pvz_id"`
	Status   string    `db:"status"`
}
