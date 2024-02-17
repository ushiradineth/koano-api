package event

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	Title     string    `db:"title"`
	Start     time.Time `db:"start"`
	End       time.Time `db:"end"`
	Timezone  string    `db:"tz"`
	Repeated  string    `db:"repeated"`
}
