package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	Title     string    `db:"title"`
	Start     time.Time `db:"start_time"`
	End       time.Time `db:"end_time"`
	Timezone  string    `db:"timezone"`
	Repeated  string    `db:"repeated"`
}
