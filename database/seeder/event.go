package seeder

import (
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func CreateEvent(db *sqlx.DB, userId uuid.UUID) {
	var title string
	err := faker.FakeData(&title)
	if err != nil {
		panic(err)
	}

	// Generate a random day within the past 30 days
	randomDay := time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour).Truncate(24 * time.Hour)

	// Define random start time within the same day with 30-minute increments
	startHour := rand.Intn(24)
	startMinute := rand.Intn(2) * 30
	startTime := randomDay.Add(time.Duration(startHour)*time.Hour + time.Duration(startMinute)*time.Minute)

	// Define random end time within the same day with 30-minute increments
	maxDuration := 16 * time.Hour
	endHour := startHour + rand.Intn(int(maxDuration.Hours()))
	endMinute := rand.Intn(2) * 30
	endTime := randomDay.Add(time.Duration(endHour)*time.Hour + time.Duration(endMinute)*time.Minute)

	_, err = db.Exec(`INSERT INTO events (id, user_id, created_at, title, start_time, end_time, timezone, repeated) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, uuid.New(), userId, time.Now(), title, startTime, endTime, "Asia/Colombo", "No")
	if err != nil {
		panic(err)
	}
}
