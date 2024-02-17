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
	database.Configure()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /user", user.GetUserHandler)
	mux.HandleFunc("POST /user", user.PostUserHandler)
	mux.HandleFunc("PUT /user", user.PutUserHandler)
	mux.HandleFunc("DELETE /user", user.DeleteUserHandler)
	mux.HandleFunc("GET /user/auth", user.AuthenticateUserHandler)
	mux.HandleFunc("GET /user/auth/refresh", user.RefreshTokenHandler)
	mux.HandleFunc("PUT /user/auth/password", user.PutUserPasswordHandler)

	mux.HandleFunc("GET /event/{event_id}", event.GetEventHandler)
	mux.HandleFunc("POST /event", event.PostEventHandler)
	mux.HandleFunc("PUT /event/{event_id}", event.PutEventHandler)
	mux.HandleFunc("DELETE /event/{event_id}", event.DeleteEventHandler)
	mux.HandleFunc("GET /event/user", event.GetUserEventsHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), mux)
}
