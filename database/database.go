package database

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func New() *sqlx.DB {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_URL"), os.Getenv("PG_DATABASE"), os.Getenv("PG_SSLMODE"))

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	fmt.Println("Connected to Postgres Database")
	return db
}
