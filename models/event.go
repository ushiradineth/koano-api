package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at"`
	Active    bool       `db:"active" json:"active"`

	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	Title    string    `db:"title" json:"title"`
	Start    time.Time `db:"start_time" json:"start_time"`
	End      time.Time `db:"end_time" json:"end_time"`
	Timezone string    `db:"timezone" json:"timezone"`
	Repeated string    `db:"repeated" json:"repeated"`
}
