package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Title     string    `db:"title" json:"title"`
	Start     time.Time `db:"start_time" json:"start_time"`
	End       time.Time `db:"end_time" json:"end_time"`
	Timezone  string    `db:"timezone" json:"timezone"`
	Repeated  string    `db:"repeated" json:"repeated"`
}
