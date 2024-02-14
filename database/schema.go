package database

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
}

type Event struct {
	ID        uuid.UUID `db:"id"`
	UserID    int       `db:"user_id"`
	Title     string    `db:"title"`
	Start     time.Time `db:"start"`
	End       time.Time `db:"end"`
	CreatedAt time.Time `db:"created_at"`
}
