package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	logger "github.com/ushiradineth/cron-be/util/log"
)

func New(log *logger.Logger) *sqlx.DB {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=%s",
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD"),
		os.Getenv("PG_URL"),
		os.Getenv("PG_DATABASE"),
		os.Getenv("PG_SSLMODE"),
	)

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Error.Fatalf("Error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Error.Fatalf("Error pinging database: %v", err)
	}

	log.Info.Println("Connected to Postgres Database")
	return db
}
