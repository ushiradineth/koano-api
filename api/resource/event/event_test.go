package event_test

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/ushiradineth/cron-be/api/resource/auth"
	"github.com/ushiradineth/cron-be/api/resource/event"
	"github.com/ushiradineth/cron-be/api/resource/user"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/test"
	"github.com/ushiradineth/cron-be/util/validator"
)

var (
	accessToken         string
	refreshToken        string
	user1ID             string
	user2ID             string
	eventId             string
	expiredAccessToken  string
	expiredRefreshToken string
	db                  *sqlx.DB
	userAPI             *user.API
	authAPI             *auth.API
	eventAPI            *event.API
)

var user1 user.PostBodyParams = user.PostBodyParams{
	Name:     faker.Name(),
	Email:    faker.Email(),
	Password: "UPlow1234!@#",
}

var user2 user.PostBodyParams = user.PostBodyParams{
	Name:     faker.Name(),
	Email:    faker.Email(),
	Password: "lowUP1234!@#",
}

var user1Auth auth.AuthenticateBodyParams = auth.AuthenticateBodyParams{
	Email:    user1.Email,
	Password: user1.Password,
}

var user2Auth auth.AuthenticateBodyParams = auth.AuthenticateBodyParams{
	Email:    user2.Email,
	Password: user2.Password,
}

var event1 event.EventBodyParams = event.EventBodyParams{
	Title:     "Test",
	StartTime: "2020-01-02T15:04:05Z",
	EndTime:   "2023-01-02T14:04:05Z",
	Timezone:  "Asia/Colombo",
	Repeated:  "daily",
}

func TestInit(t *testing.T) {
	t.Run("Initiate Dependencies", func(t *testing.T) {
		err := godotenv.Load("../../../.env")
		if err != nil {
			log.Println("Failed to load env")
		}

		db = test.NewDB("../../../database/migration")
		v := validator.New()

		userAPI = user.New(db, v)
		eventAPI = event.New(db, v)
		authAPI = auth.New(db, v)

		expiredAccessToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1234567890", "iat": time.Now().Unix(), "exp": time.Now().Add(-1 * time.Hour).Unix()}).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()

		expiredRefreshToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1234567890", "iat": time.Now().Unix(), "exp": time.Now().Add(-1 * time.Hour).Unix()}).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()
	})

	t.Run("Create User 1", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, user1, http.StatusOK, response.StatusSuccess)
	})

	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	t.Run("Create User 2", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, user2, http.StatusOK, response.StatusSuccess)
	})
}

func TestCreateEventHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, event1, http.StatusOK, response.StatusSuccess, &eventId, accessToken)
	})

	t.Run("Event Already Exists", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, event1, http.StatusBadRequest, response.StatusFail, &eventId, accessToken)
	})

	t.Run("JWT is Invalid", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, event1, http.StatusUnauthorized, response.StatusFail, &eventId, expiredAccessToken)
	})

	body := event.EventBodyParams{
		Title:     event1.Title,
		StartTime: "not_datetime_with_timezone",
		EndTime:   event1.EndTime,
		Timezone:  event1.Timezone,
		Repeated:  event1.Repeated,
	}
	t.Run("Start time is invalid", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, &eventId, accessToken)
	})
	body.StartTime = event1.StartTime

	body.EndTime = "not_datetime_with_timezone"
	t.Run("End time is invalid", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, &eventId, accessToken)
	})

	body.StartTime = event1.EndTime
	body.EndTime = event1.StartTime
	t.Run("Start time occurs after End time", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, &eventId, accessToken)
	})
	body.StartTime = event1.StartTime
	body.EndTime = event1.EndTime

	body.Timezone = "not_a_timezone"
	t.Run("Timezone is invalid", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, &eventId, accessToken)
	})
	body.Timezone = event1.Timezone

	body.Repeated = "not_a_repeated_value"
	t.Run("Repeated is invalid", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, &eventId, accessToken)
	})
}

func TestGetEventHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		test.GetEventHelper(eventAPI, t, http.StatusOK, response.StatusSuccess, event1, eventId, accessToken)
	})

	t.Run("Event ID is invalid", func(t *testing.T) {
		test.GetEventHelper(eventAPI, t, http.StatusBadRequest, response.StatusFail, event1, "not_an_id", accessToken)
	})

	t.Run("JWT is Invalid", func(t *testing.T) {
		test.GetEventHelper(eventAPI, t, http.StatusUnauthorized, response.StatusFail, event1, eventId, expiredAccessToken)
	})

	t.Run("Authenticates User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("JWT does not match user ID", func(t *testing.T) {
		test.GetEventHelper(eventAPI, t, http.StatusBadRequest, response.StatusFail, event1, eventId, accessToken)
	})
}

func TestUpdateEventHandler(t *testing.T) {
	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	t.Run("Success", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, event1, http.StatusOK, response.StatusSuccess, eventId, accessToken)
	})

	t.Run("Event ID is invalid", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, event1, http.StatusBadRequest, response.StatusFail, "not_an_id", accessToken)
	})

	t.Run("JWT is Invalid", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, event1, http.StatusUnauthorized, response.StatusFail, eventId, expiredAccessToken)
	})

	t.Run("Authenticates User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("JWT does not match user ID", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, event1, http.StatusBadRequest, response.StatusFail, eventId, accessToken)
	})

	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	body := event.EventBodyParams{
		Title:     event1.Title,
		StartTime: event1.StartTime,
		EndTime:   "not_datetime_with_timezone",
		Timezone:  event1.Timezone,
		Repeated:  event1.Repeated,
	}
	t.Run("Start time is invalid", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, eventId, accessToken)
	})
	body.StartTime = event1.StartTime

	body.EndTime = "not_datetime_with_timezone"
	t.Run("End time is invalid", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, eventId, accessToken)
	})

	body.StartTime = event1.EndTime
	body.EndTime = event1.StartTime
	t.Run("Start time occurs after End time", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, eventId, accessToken)
	})
	body.StartTime = event1.StartTime
	body.EndTime = event1.EndTime

	body.Timezone = "not_a_timezone"
	t.Run("Timezone is invalid", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, eventId, accessToken)
	})
	body.Timezone = event1.Timezone

	body.Repeated = "not_a_repeated_value"
	t.Run("Repeated is invalid", func(t *testing.T) {
		test.UpdateEventHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, eventId, accessToken)
	})
}

func TestGetUserEventsHandler(t *testing.T) {
	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	body := event.GetUserEventsQueryParams{
		StartDay: "not_datetime_with_timezone",
		EndDay:   "2006-01-02",
	}
	t.Run("End time is invalid", func(t *testing.T) {
		test.GetUserEventsHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, user1ID, accessToken)
	})
	body.StartDay = "2001-01-02"

	body.EndDay = "not_datetime_with_timezone"
	t.Run("End time is invalid", func(t *testing.T) {
		test.GetUserEventsHelper(eventAPI, t, body, http.StatusBadRequest, response.StatusFail, user1ID, accessToken)
	})

	body.StartDay = "2001-01-02"
	body.EndDay = "2006-01-02"
	t.Run("Success", func(t *testing.T) {
		test.GetUserEventsHelper(eventAPI, t, body, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	t.Run("Authenticates User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("JWT is Invalid", func(t *testing.T) {
		test.GetUserEventsHelper(eventAPI, t, body, http.StatusUnauthorized, response.StatusFail, user1ID, expiredAccessToken)
	})
}

func TestDeleteEventHandler(t *testing.T) {
	t.Run("Authenticate User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	t.Run("Success", func(t *testing.T) {
		test.DeleteEventHelper(eventAPI, t, http.StatusOK, response.StatusSuccess, eventId, accessToken)
	})

	t.Run("Event does not exist", func(t *testing.T) {
		test.DeleteEventHelper(eventAPI, t, http.StatusBadRequest, response.StatusFail, uuid.NewString(), accessToken)
	})

	t.Run("JWT is invalid", func(t *testing.T) {
		test.DeleteEventHelper(eventAPI, t, http.StatusUnauthorized, response.StatusFail, eventId, expiredAccessToken)
	})

	t.Run("UUID is invalid", func(t *testing.T) {
		test.DeleteEventHelper(eventAPI, t, http.StatusBadRequest, response.StatusFail, "not_an_uuid", accessToken)
	})
}

func TestCleanUp(t *testing.T) {
	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	t.Run("Delete User 1", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	t.Run("Authenticate User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("Delete User 2", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user2ID, accessToken)
	})
}
