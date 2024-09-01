package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/api/router"
	"github.com/ushiradineth/cron-be/database"
	_ "github.com/ushiradineth/cron-be/docs"
)

// @title						Cron
// @version					1.0
// @description				Backend for Cron calendar management app.
// @contact.name				Ushira Dineth
// @contact.url				https://ushira.com
// @contact.email				ushiradineth@gmail.com
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	err := godotenv.Load(".env")
	if err != nil {
		return errors.New("Failed to load env")
	}

	db := database.Configure()
	v := validator.New()

	routes := router.New(db, v)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: routes,
	}

	go func() {
		log.Printf("Listening on %s\n", httpServer.Addr)
		err := httpServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()

	return nil
}
