package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/api/event"
	"github.com/ushiradineth/cron-be/api/user"
	"github.com/ushiradineth/cron-be/database"
)

// Graceful shutdowns based on https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	godotenv.Load(".env")
	db := database.Configure()

	routes := routes(db)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: routes,
	}

	go func() {
		log.Printf("Listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
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

	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) { user.GetUserHandler(w, r, db) })
	mux.HandleFunc("POST /user", func(w http.ResponseWriter, r *http.Request) { user.PostUserHandler(w, r, db) })
	mux.HandleFunc("PUT /user", func(w http.ResponseWriter, r *http.Request) { user.PutUserHandler(w, r, db) })
	mux.HandleFunc("DELETE /user", func(w http.ResponseWriter, r *http.Request) { user.DeleteUserHandler(w, r, db) })
	mux.HandleFunc("POST /user/auth", func(w http.ResponseWriter, r *http.Request) { user.AuthenticateUserHandler(w, r, db) })
	mux.HandleFunc("POST /user/auth/refresh", func(w http.ResponseWriter, r *http.Request) { user.RefreshTokenHandler(w, r, db) })
	mux.HandleFunc("PUT /user/auth/password", func(w http.ResponseWriter, r *http.Request) { user.PutUserPasswordHandler(w, r, db) })

	mux.HandleFunc("GET /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.GetEventHandler(w, r, db) })
	mux.HandleFunc("POST /event", func(w http.ResponseWriter, r *http.Request) { event.PostEventHandler(w, r, db) })
	mux.HandleFunc("PUT /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.PutEventHandler(w, r, db) })
	mux.HandleFunc("DELETE /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.DeleteEventHandler(w, r, db) })
	mux.HandleFunc("GET /event/user", func(w http.ResponseWriter, r *http.Request) { event.GetUserEventsHandler(w, r, db) })

	return mux
}
