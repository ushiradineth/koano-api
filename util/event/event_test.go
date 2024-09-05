package event_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/api/resource/auth"
	"github.com/ushiradineth/cron-be/api/resource/event"
	"github.com/ushiradineth/cron-be/api/resource/user"
	eventUtil "github.com/ushiradineth/cron-be/util/event"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/test"
	"github.com/ushiradineth/cron-be/util/validator"
)

var (
	accessToken         string
	refreshToken        string
	user1ID             string
	event1ID            string
	expiredAccessToken  string
	expiredRefreshToken string
	db                  *sqlx.DB
	userAPI             *user.API
	eventAPI            *event.API
	authAPI             *auth.API
)

var user1 user.PostQueryParams = user.PostQueryParams{
	Name:     faker.Name(),
	Email:    faker.Email(),
	Password: "UPlow1234!@#",
}

var event1 event.EventQueryParams = event.EventQueryParams{
	Title:     "Test",
	StartTime: "2020-01-02T15:04:05Z",
	EndTime:   "2023-01-02T14:04:05Z",
	Timezone:  "Asia/Colombo",
	Repeated:  "daily",
}

func TestInit(t *testing.T) {
	t.Run("Initiate Dependencies", func(t *testing.T) {
		assert.NoError(t, godotenv.Load("../../.env"), "Environment variables should be loaded in")

		db = test.NewDB("../../database/migration")
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

	user := url.Values{}
	user.Set("name", user1.Name)
	user.Set("email", user1.Email)
	user.Set("password", user1.Password)
	t.Run("Create User", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, user, http.StatusOK, response.StatusSuccess, user1)
	})

	t.Run("Authenticates user 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	body := url.Values{}
	body.Set("title", event1.Title)
	body.Set("start_time", event1.StartTime)
	body.Set("end_time", event1.EndTime)
	body.Set("timezone", event1.Timezone)
	body.Set("repeated", event1.Repeated)
	t.Run("Create Event", func(t *testing.T) {
		test.CreateEventHelper(eventAPI, t, body, http.StatusOK, response.StatusSuccess, event1, &event1ID, accessToken)
	})
}

func TestGetEventHelper(t *testing.T) {
	t.Run("Get Event", func(t *testing.T) {
		response := httptest.NewRecorder()
		event := eventUtil.GetEvent(response, event1ID, user1ID, db)
		assert.NotNil(t, event, "Error getting event")

		assert.Equal(t, event1.Title, event.Title)
		assert.Equal(t, event1.StartTime, time.Time(event.Start).Format(time.RFC3339))
		assert.Equal(t, event1.EndTime, time.Time(event.End).Format(time.RFC3339))
		assert.Equal(t, event1.Timezone, event.Timezone)
		assert.Equal(t, event1.Repeated, event.Repeated)
	})

	t.Run("Event ID is invalid", func(t *testing.T) {
		response := httptest.NewRecorder()
		event := eventUtil.GetEvent(response, "not_an_id", user1ID, db)
		assert.Nil(t, event, "Event should be empty")
	})

	t.Run("UUID is not a Event ID", func(t *testing.T) {
		response := httptest.NewRecorder()
		event := eventUtil.GetEvent(response, uuid.NewString(), user1ID, db)
		assert.Nil(t, event, "Event should be empty")
	})
}

func TestDoesEventExistHelper(t *testing.T) {
	t.Run("Get Event 1 with ID", func(t *testing.T) {
		event := eventUtil.DoesEventExist(event1ID, "", "", user1ID, db)
		assert.True(t, event, "Event should exist")
	})

	t.Run("Get no event if User ID is empty", func(t *testing.T) {
		event := eventUtil.DoesEventExist(event1ID, "", "", "", db)
		assert.False(t, event, "Event should not exist")
	})

	t.Run("Get Event 1 with Start and End Time", func(t *testing.T) {
		event := eventUtil.DoesEventExist("", event1.StartTime, event1.EndTime, user1ID, db)
		assert.True(t, event, "Event should exist")
	})
}

func TestCleanUp(t *testing.T) {
	t.Run("Delete user", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})
}
