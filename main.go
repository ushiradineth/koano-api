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

	mux.HandleFunc("GET /user/{user_id}", user.GetUserHandler)
	mux.HandleFunc("POST /user", user.PostUserHandler)
	mux.HandleFunc("PUT /user/{user_id}", user.PutUserHandler)
	mux.HandleFunc("PUT /user/password/{user_id}", user.PutUserPasswordHandler)
	mux.HandleFunc("DELETE /user/{user_id}", user.DeleteUserHandler)
	mux.HandleFunc("GET /user/auth", user.AuthenticateUserHandler)

	mux.HandleFunc("GET /event/{event_id}", event.GetEventHandler)
	mux.HandleFunc("POST /event", event.PostEventHandler)
	mux.HandleFunc("PUT /event/{event_id}", event.PutEventHandler)
	mux.HandleFunc("DELETE /event/{event_id}", event.DeleteEventHandler)
	mux.HandleFunc("GET /event/user/{user_id}", event.GetUserEventsHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), mux)
}
