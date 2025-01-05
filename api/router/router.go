package router

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"github.com/ushiradineth/cron-be/api/resource/auth"
	"github.com/ushiradineth/cron-be/api/resource/event"
	"github.com/ushiradineth/cron-be/api/resource/health"
	"github.com/ushiradineth/cron-be/api/resource/user"
	logger "github.com/ushiradineth/cron-be/util/log"
)

func New(db *sqlx.DB, validator *validator.Validate, logger *logger.Logger) http.Handler {
	router := http.NewServeMux()
	router.Handle("/", Base())

	group := "/api/v1"
	router.Handle(fmt.Sprintf("%s/", group), V1(group, db, validator, logger))

	if os.Getenv("CORS_ENABLED") == "true" {
		allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
		logger.Info.Println("CORS Enabled")
		if allowedOrigin == "" {
			logger.Warn.Println("CORS_ALLOWED_ORIGIN is not set or empty. Defaulting to no CORS.")
		} else {
			logger.Info.Printf("CORS Allowed Origin: %s", allowedOrigin)
		}

		c := cors.New(cors.Options{
			AllowedOrigins: []string{allowedOrigin},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Authorization", "Content-Type"},
		})

		return c.Handler(router)
	}

	logger.Info.Println("CORS Disabled")
	return router
}

func Base() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /health", health.Health)
	router.Handle("GET /metrics", promhttp.Handler())

	if os.Getenv("ENV") == "DEVELOPMENT" {
		router.HandleFunc("/swagger/", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	return router
}

func V1(group string, db *sqlx.DB, validator *validator.Validate, logger *logger.Logger) http.Handler {
	router := http.NewServeMux()

	userAPI := user.New(db, validator, logger)
	router.HandleFunc("GET /users/{user_id}", userAPI.Get)
	router.HandleFunc("POST /users", userAPI.Post)
	router.HandleFunc("PUT /users/{user_id}", userAPI.Put)
	router.HandleFunc("DELETE /users/{user_id}", userAPI.Delete)

	authAPI := auth.New(db, validator, logger)
	router.HandleFunc("POST /auth/login", authAPI.Authenticate)
	router.HandleFunc("POST /auth/refresh", authAPI.RefreshToken)
	router.HandleFunc("PUT /auth/reset-password", authAPI.PutPassword)

	eventAPI := event.New(db, validator, logger)
	router.HandleFunc("GET /events/{event_id}", eventAPI.Get)
	router.HandleFunc("POST /events", eventAPI.Post)
	router.HandleFunc("PUT /events/{event_id}", eventAPI.Put)
	router.HandleFunc("DELETE /events/{event_id}", eventAPI.Delete)
	router.HandleFunc("GET /events", eventAPI.GetUserEvents)

	return http.StripPrefix(group, router)
}
