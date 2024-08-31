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

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/swaggo/http-swagger/v2"
	"github.com/ushiradineth/cron-be/api/event"
	"github.com/ushiradineth/cron-be/api/user"
	"github.com/ushiradineth/cron-be/database"
	_ "github.com/ushiradineth/cron-be/docs"
)

// Graceful shutdowns based on https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

//	@title			Cron
//	@version		1.0
//	@description	Backend for Cron calendar management app.

//	@contact.name	Ushira Dineth
//	@contact.url	https://ushira.com
//	@contact.email	ushiradineth@gmail.com

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

	routes := routes(db)

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

func routes(db *sqlx.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) { user.Get(w, r, db) })
	mux.HandleFunc("POST /user", func(w http.ResponseWriter, r *http.Request) { user.Post(w, r, db) })
	mux.HandleFunc("PUT /user", func(w http.ResponseWriter, r *http.Request) { user.Put(w, r, db) })
	mux.HandleFunc("DELETE /user", func(w http.ResponseWriter, r *http.Request) { user.Delete(w, r, db) })
	mux.HandleFunc("POST /user/auth", func(w http.ResponseWriter, r *http.Request) { user.Authenticate(w, r, db) })
	mux.HandleFunc("POST /user/auth/refresh", func(w http.ResponseWriter, r *http.Request) { user.RefreshToken(w, r, db) })
	mux.HandleFunc("PUT /user/auth/password", func(w http.ResponseWriter, r *http.Request) { user.PutPassword(w, r, db) })

	mux.HandleFunc("GET /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.Get(w, r, db) })
	mux.HandleFunc("POST /event", func(w http.ResponseWriter, r *http.Request) { event.Post(w, r, db) })
	mux.HandleFunc("PUT /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.Put(w, r, db) })
	mux.HandleFunc("DELETE /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.Delete(w, r, db) })
	mux.HandleFunc("GET /event/user", func(w http.ResponseWriter, r *http.Request) { event.GetUserEvents(w, r, db) })

	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}
