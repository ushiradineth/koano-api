package database

import (
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ushiradineth/cron-be/event"
	"github.com/ushiradineth/cron-be/user"
)

func Configure() {
	dataSource := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_URL"), os.Getenv("PG_DATABASE"))

	DB, err := sqlx.Connect("postgres", dataSource)
	if err != nil {
		log.Fatalln(err)
	}

	pingErr := DB.Ping()

	if pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected to Postgres Database")

	user.DB = DB
	event.DB = DB
}
