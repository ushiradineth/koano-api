package mysql

import (
	"time"

	"github.com/google/uuid"
)

var userSchema = `
CREATE TABLE IF NOT EXISTS user (
    id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    name TEXT,
    email VARCHAR(255),

    PRIMARY KEY (id),
    UNIQUE (id, email(255))
);
`

var eventSchema = `
CREATE TABLE IF NOT EXISTS event (
    id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    title TEXT,
    start TIMESTAMP,
    end TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE (id),
    CONSTRAINT fk_event_user FOREIGN KEY (user_id) REFERENCES user(id)
);
`

type User struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
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
