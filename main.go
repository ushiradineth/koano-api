package main

import (
	"fmt"
	"net/http"
	"user"

	"github.com/joho/godotenv"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /event/{event_id}/", event.GetEventHandler)
	mux.HandleFunc("POST /event/", event.PostEventHandler)
	mux.HandleFunc("PUT /event/{event_id}/", event.UpdateEventHandler)
	mux.HandleFunc("GET /event/user/{user_id}/", event.GetUserEventsHandler)

	mux.HandleFunc("GET /user/{user_id}/", user.GetUserHandler)
	mux.HandleFunc("POST /user/", user.CreateUserHandler)
	mux.HandleFunc("PUT /user/{user_id}/", user.UpdateUserHandler)
	mux.HandleFunc("PUT /user/{user_id}/password/", user.UpdateUserPasswordHandler)

	http.ListenAndServe("localhost:5000", mux)
}
