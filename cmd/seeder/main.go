package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/database/seeder"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Failed to load env")
	}

	db := database.New()

	for i := 0; i < 100; i++ {
		userId := seeder.CreateUser(db)

		for i := 0; i < 10; i++ {
			seeder.CreateEvent(db, userId)
		}
	}

	return nil
}
