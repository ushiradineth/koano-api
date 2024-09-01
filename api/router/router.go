package router

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/swaggo/http-swagger/v2"
	"github.com/ushiradineth/cron-be/api/resource/auth"
	"github.com/ushiradineth/cron-be/api/resource/event"
	"github.com/ushiradineth/cron-be/api/resource/health"
	"github.com/ushiradineth/cron-be/api/resource/user"
)

func New(db *sqlx.DB, v *validator.Validate) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", health.Health)

	userAPI := user.New(db, v)
	mux.HandleFunc("GET /users/{user_id}", userAPI.Get)
	mux.HandleFunc("POST /users", userAPI.Post)
	mux.HandleFunc("PUT /users/{user_id}", userAPI.Put)
	mux.HandleFunc("DELETE /users/{user_id}", userAPI.Delete)

	authAPI := auth.New(db, v)
	mux.HandleFunc("POST /auth/login", authAPI.Authenticate)
	mux.HandleFunc("POST /auth/refresh", authAPI.RefreshToken)
	mux.HandleFunc("PUT /auth/reset-password", authAPI.PutPassword)

	eventAPI := event.New(db, v)
	mux.HandleFunc("GET /events/{event_id}", eventAPI.Get)
	mux.HandleFunc("POST /events", eventAPI.Post)
	mux.HandleFunc("PUT /events/{event_id}", eventAPI.Put)
	mux.HandleFunc("DELETE /events/{event_id}", eventAPI.Delete)
	mux.HandleFunc("GET /users/{user_id}/events", eventAPI.GetUserEvents)

	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}
