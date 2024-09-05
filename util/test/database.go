package test

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewDB(migrationPath string) *sqlx.DB {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatalf("Could not start Postgres container: %v", err)
	}

	connectionString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("Could not get connection string: %v", err)
	}

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	if err := RunMigrations(connectionString, migrationPath); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	fmt.Println("Connected to Postgres Database")
	return db
}

func RunMigrations(connectionString string, migrationPath string) error {
	// Build the migration command
	cmd := exec.Command("migrate", "-path", migrationPath, "-database", connectionString, "-verbose", "up")

	// Capture standard output and error
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("could not run migrations: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}
