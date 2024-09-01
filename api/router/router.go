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
	mux.HandleFunc("GET /user", userAPI.Get)
	mux.HandleFunc("POST /user", userAPI.Post)
	mux.HandleFunc("PUT /user", userAPI.Put)
	mux.HandleFunc("DELETE /user", userAPI.Delete)

	authAPI := auth.New(db, v)
	mux.HandleFunc("POST /auth/login", authAPI.Authenticate)
	mux.HandleFunc("POST /auth/refresh", authAPI.RefreshToken)
	mux.HandleFunc("PUT /auth/reset-password", authAPI.PutPassword)

	eventAPI := event.New(db, v)
	mux.HandleFunc("GET /event/{event_id}", eventAPI.Get)
	mux.HandleFunc("POST /event", eventAPI.Post)
	mux.HandleFunc("PUT /event/{event_id}", eventAPI.Put)
	mux.HandleFunc("DELETE /event/{event_id}", eventAPI.Delete)
	mux.HandleFunc("GET /event/user", eventAPI.GetUserEvents)

	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}
