package event

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/api/user"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/models"
)

type UserType struct {
	name     string
	email    string
	password string
}

type EventType struct {
	title      string
	start_time string
	end_time   string
	timezone   string
	repeated   string
}

var user1 UserType = UserType{
	name:     "Ushira Dineth",
	email:    "ushiradineth@gmail.com",
	password: "Ushira1234!",
}

var event1 EventType = EventType{
	title:      "Event 1",
	start_time: "2001-11-30T10:00:00Z",
	end_time:   "2001-11-30T10:15:00Z",
	timezone:   "Asia/Colombo",
	repeated:   "No",
}

var access_token string = ""
var refresh_token string = ""
var event_id uuid.UUID

var db *sqlx.DB

func TestInitDB(t *testing.T) {
	t.Run("Initiate DB Connection", func(t *testing.T) {
		assert.NoError(t, godotenv.Load("../../.env"), "Environment variables should be loaded in")
		db = database.Configure()
	})
}

func TestEventHandlers(t *testing.T) {
	t.Run("Create User", func(t *testing.T) {
		body := url.Values{}
		body.Set("name", user1.name)
		body.Set("email", user1.email)
		body.Set("password", user1.password)

		CreateUserHelper(t, body, http.StatusOK)

		t.Run("Authenticates user", func(t *testing.T) {
			AuthenticateUserHelper(t, body, http.StatusOK)
		})
	})

	t.Run("Create Event", func(t *testing.T) {
		body := url.Values{}
		body.Set("title", event1.title)
		body.Set("start_time", event1.start_time)
		body.Set("end_time", event1.end_time)
		body.Set("timezone", event1.timezone)
		body.Set("repeated", event1.repeated)

		t.Run("Success", func(t *testing.T) {
			CreateEventHelper(t, body, http.StatusOK)
		})

		t.Run("Event Already Exists", func(t *testing.T) {
			CreateEventHelper(t, body, http.StatusBadRequest)
		})
	})

	t.Run("Get Event", func(t *testing.T) {
		t.Run("Fail with invalid Event ID", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/event/{event_id}", nil)
			request.SetPathValue("event_id", "INVALID_ID")
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
			response := httptest.NewRecorder()

			Get(response, request, db)

			assert.Equal(t, http.StatusInternalServerError, response.Code)
		})

		t.Run("Success", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/event/{event_id}", nil)
			request.SetPathValue("event_id", event_id.String())
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
			response := httptest.NewRecorder()

			Get(response, request, db)

			var responseBody models.Event

			err := json.NewDecoder(response.Body).Decode(&responseBody)
			assert.NoError(t, err)

			parsed_start, err := time.Parse(time.RFC3339, event1.start_time)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				return
			}

			parsed_end, err := time.Parse(time.RFC3339, event1.end_time)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				return
			}

			assert.Equal(t, event1.title, responseBody.Title)
			assert.Equal(t, parsed_start, responseBody.Start)
			assert.Equal(t, parsed_end, responseBody.End)
			assert.Equal(t, event1.timezone, responseBody.Timezone)
			assert.Equal(t, event1.repeated, responseBody.Repeated)
			assert.Equal(t, http.StatusOK, response.Code)
		})
	})

	t.Run("Delete User", func(t *testing.T) {
		t.Run("Deletes user", func(t *testing.T) {
			DeleteUserHelper(t, http.StatusOK)
		})
	})
}

func TestEventHelpers(t *testing.T) {
	t.Run("Create User", func(t *testing.T) {
		body := url.Values{}
		body.Set("name", user1.name)
		body.Set("email", user1.email)
		body.Set("password", user1.password)

		CreateUserHelper(t, body, http.StatusOK)

		t.Run("Authenticates user 1", func(t *testing.T) {
			AuthenticateUserHelper(t, body, http.StatusOK)
		})
	})

	t.Run("Delete User", func(t *testing.T) {
		t.Run("Deletes user", func(t *testing.T) {
			DeleteUserHelper(t, http.StatusOK)
		})
	})
}

func CreateUserHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, "/user", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	user.Post(response, request, db)

	assert.Equal(t, http.StatusOK, response.Code)
}

func AuthenticateUserHelper(t testing.TB, body url.Values, wantCode int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, "/user/auth", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	user.Authenticate(response, request, db)

	var responseBody user.AuthenticateResponse
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	assert.NoError(t, err)

	assert.Equal(t, body.Get("email"), responseBody.User.Email)
	assert.Equal(t, "", responseBody.User.Password)
	assert.NotEmpty(t, responseBody.AccessToken)
	assert.NotEmpty(t, responseBody.RefreshToken)

	assert.Equal(t, http.StatusOK, response.Code)

	access_token = responseBody.AccessToken
	refresh_token = responseBody.RefreshToken
}

func DeleteUserHelper(t testing.TB, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodDelete, "/user", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	user.Delete(response, request, db)

	assert.Equal(t, http.StatusOK, response.Code)
}

func CreateEventHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, "/event", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	Post(response, request, db)

	assert.Equal(t, want_code, response.Code)

	if response.Code == http.StatusOK {
		var responseBody models.Event

		err := json.NewDecoder(response.Body).Decode(&responseBody)
		assert.NoError(t, err)

		event_id = responseBody.ID
	}
}
