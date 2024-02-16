package event

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `db:"id"`
	UserID    int       `db:"user_id"`
	Title     string    `db:"title"`
	Start     time.Time `db:"start"`
	End       time.Time `db:"end"`
	CreatedAt time.Time `db:"created_at"`
}
