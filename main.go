package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/event"
	"github.com/ushiradineth/cron-be/user"
)

func main() {
	godotenv.Load(".env")
	db := database.Configure(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASSWORD"), os.Getenv("PG_URL"), os.Getenv("PG_DATABASE")))

	mux := http.NewServeMux()

	mux.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) { user.GetUserHandler(w, r, db) })
	mux.HandleFunc("POST /user", func(w http.ResponseWriter, r *http.Request) { user.PostUserHandler(w, r, db) })
	mux.HandleFunc("PUT /user", func(w http.ResponseWriter, r *http.Request) { user.PutUserHandler(w, r, db) })
	mux.HandleFunc("DELETE /user", func(w http.ResponseWriter, r *http.Request) { user.DeleteUserHandler(w, r, db) })
	mux.HandleFunc("GET /user/auth", func(w http.ResponseWriter, r *http.Request) { user.AuthenticateUserHandler(w, r, db) })
	mux.HandleFunc("POST /user/auth/refresh", func(w http.ResponseWriter, r *http.Request) { user.RefreshTokenHandler(w, r, db) })
	mux.HandleFunc("PUT /user/auth/password", func(w http.ResponseWriter, r *http.Request) { user.PutUserPasswordHandler(w, r, db) })

	mux.HandleFunc("GET /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.GetEventHandler(w, r, db) })
	mux.HandleFunc("POST /event", func(w http.ResponseWriter, r *http.Request) { event.PostEventHandler(w, r, db) })
	mux.HandleFunc("PUT /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.PutEventHandler(w, r, db) })
	mux.HandleFunc("DELETE /event/{event_id}", func(w http.ResponseWriter, r *http.Request) { event.DeleteEventHandler(w, r, db) })
	mux.HandleFunc("GET /event/user", func(w http.ResponseWriter, r *http.Request) { event.GetUserEventsHandler(w, r, db) })

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), mux)
}
