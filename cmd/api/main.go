package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/api/event"
	"github.com/ushiradineth/cron-be/api/user"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	godotenv.Load(".env")
	db := database.Configure()

	return routes(db)
}

func routes(db *sqlx.DB) error {
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

	err := http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), mux)

	return err
}
