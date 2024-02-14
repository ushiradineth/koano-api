package main

import (
	"db"
	"event"
	"fmt"
	"net/http"
	"os"
	"user"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	db.Configure()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /event/{event_id}/", event.GetEventHandler)
	mux.HandleFunc("POST /event/", event.PostEventHandler)
	mux.HandleFunc("PUT /event/{event_id}/", event.UpdateEventHandler)
	mux.HandleFunc("GET /event/user/{user_id}/", event.GetUserEventsHandler)

	mux.HandleFunc("GET /user/{user_id}/", user.GetUserHandler)
	mux.HandleFunc("POST /user/", user.CreateUserHandler)
	mux.HandleFunc("PUT /user/{user_id}/", user.UpdateUserHandler)
	mux.HandleFunc("PUT /user/{user_id}/password/", user.UpdateUserPasswordHandler)

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), mux)
}
