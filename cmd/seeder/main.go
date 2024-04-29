package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/database/seeder"
)

func main() {
	godotenv.Load(".env")
	db := database.Configure()

	for i := 0; i < 100; i++ {
		userId := seeder.CreateUser(db)

		for i := 0; i < 10; i++ {
			seeder.CreateEvent(db, userId)
		}
	}

	os.Exit(0)
}
