package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/api/router"
	"github.com/ushiradineth/cron-be/database"
	_ "github.com/ushiradineth/cron-be/docs"
	logger "github.com/ushiradineth/cron-be/util/log"
	validator "github.com/ushiradineth/cron-be/util/validator"
)

// @title						Cron
// @version					1.0
// @description				Backend for Cron calendar management app.
// @contact.name				Ushira Dineth
// @contact.url				https://ushira.com
// @contact.email				ushiradineth@gmail.com
// @BasePath					/api/v1
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
	log := logger.New()

	err := godotenv.Load(".env")
	if err != nil {
		log.Error.Println("Failed to load env")
	}

	db := database.New(log)
	validator := validator.New()
	router := router.New(db, validator, log, os.Getenv("FRONTEND_URL"))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: router,
	}

	go func() {
		log.Info.Printf("Listening on %s\n", httpServer.Addr)
		err := httpServer.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			log.Error.Printf("Error listening and serving: %s\n", err)
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
			log.Error.Printf("error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()

	return nil
}
