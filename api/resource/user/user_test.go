package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util"
)

var access_token string = ""
var refresh_token string = ""
var user1_id string = ""
var expired_access_token string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6ImIxYzJhYjRiLTNiODEtNDAyOS1hZjU5LTdjNzhkODcyZDU1MSIsIm5hbWUiOiJVc2hpcmEgRGluZXRoIiwiZW1haWwiOiJ1c2hpcmFkaW5ldGhAZ21haWwuY29tIiwiZXhwIjoxNzA4MzY3NDcxLCJpYXQiOjE3MDgzNjc0NzF9.Zr54CJPw_7s_L-h2yVSEUtHzRi4uVII8CJ6SsJp4I8E"
var expired_refresh_token string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDgzNjc1NDIsImlhdCI6MTcwODM2NzU0Mn0.LSBFevNUnXbFZ5a0yXMbR5tmA6j3GssDFCkkH262Jag"

type UserType struct {
	name     string
	email    string
	password string
}

var user1 UserType = UserType{
	name:     "Ushira Dineth",
	email:    "ushiradineth@gmail.com",
	password: "Ushira1234!",
}

var user2 UserType = UserType{
	name:     "Not Ushira Dineth",
	email:    "notushiradineth@gmail.com",
	password: "Ushira!1234",
}

var db *sqlx.DB

func TestInitDB(t *testing.T) {
	t.Run("Initiate DB Connection", func(t *testing.T) {
		assert.NoError(t, godotenv.Load("../../.env"), "Environment variables should be loaded in")
		db = database.Configure()
	})
}

func TestUserHandlers(t *testing.T) {
	t.Run("Create User", func(t *testing.T) {
		body := url.Values{}
		body.Set("name", user1.name)
		body.Set("email", user1.email)
		body.Set("password", user1.password)

		t.Run("Success", func(t *testing.T) {
			CreateUserHelper(t, body, http.StatusOK)
		})

		t.Run("Email Already Exists", func(t *testing.T) {
			CreateUserHelper(t, body, http.StatusBadRequest)
		})

		body.Set("name", user2.name)
		body.Set("email", user2.email)
		body.Set("password", user2.password)

		t.Run("User 2", func(t *testing.T) {
			CreateUserHelper(t, body, http.StatusOK)
		})
	})

	t.Run("Authenticate User", func(t *testing.T) {
		body := url.Values{}
		body.Set("email", user1.email)
		body.Set("password", user1.password)

		t.Run("Authenticates user 1", func(t *testing.T) {
			AuthenticateUserHelper(t, body, http.StatusOK)
		})

		body.Set("email", user1.email)
		body.Set("password", user2.password)

		t.Run("Fail authentication with wrong credentials", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodPost, "/user/auth", strings.NewReader(body.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()

			Authenticate(response, request, db)

			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})
	})

	t.Run("Get User", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/user", nil)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
			response := httptest.NewRecorder()
			Get(response, request, db)

			var responseBody models.User
			err := json.NewDecoder(response.Body).Decode(&responseBody)
			assert.NoError(t, err)

			assert.Equal(t, user1.name, responseBody.Name)
			assert.Equal(t, user1.email, responseBody.Email)
			assert.Equal(t, "redacted", responseBody.Password, "Password in response should be redacted")

			assert.Equal(t, http.StatusOK, response.Code)
		})

		t.Run("Access token is invalid", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/user", nil)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "invalid token"))
			response := httptest.NewRecorder()
			Get(response, request, db)

			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})
	})

	t.Run("Update User", func(t *testing.T) {
		body := url.Values{}
		body.Set("name", user2.name)
		body.Set("email", user1.email)

		t.Run("Update user name", func(t *testing.T) {
			UpdateUserHelper(t, body, http.StatusOK)
		})

		body.Set("email", user2.email)

		t.Run("Fail to update user 1 as email already exists", func(t *testing.T) {
			UpdateUserHelper(t, body, http.StatusBadRequest)
		})

		body.Set("email", "newemail@gmail.com")
		t.Run("Update user 1 email", func(t *testing.T) {
			UpdateUserHelper(t, body, http.StatusOK)
		})

		body.Set("email", user1.email)
		t.Run("Reset user 1", func(t *testing.T) {
			UpdateUserHelper(t, body, http.StatusOK)
		})
	})

	t.Run("Update User Password", func(t *testing.T) {
		body := url.Values{}
		body.Set("email", user1.email)
		body.Set("password", user1.password)

		t.Run("Authenticates user 1", func(t *testing.T) {
			AuthenticateUserHelper(t, body, http.StatusOK)
		})

		body.Set("email", user1.email)
		body.Set("password", user2.password)

		t.Run("Authentication fails due to wrong password", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodPost, "/user/auth", strings.NewReader(body.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			response := httptest.NewRecorder()

			Authenticate(response, request, db)

			assert.Equal(t, http.StatusUnauthorized, response.Code)
		})

		body.Del("email")
		body.Set("password", user2.password)

		t.Run("Update user 1 password", func(t *testing.T) {
			UpdateUserPasswordHelper(t, body, http.StatusOK)
		})

		body.Set("email", user1.email)
		body.Set("password", user2.password)

		t.Run("Authenticates user 2", func(t *testing.T) {
			AuthenticateUserHelper(t, body, http.StatusOK)
		})
	})

	t.Run("Refresh Token", func(t *testing.T) {
		body := url.Values{}
		body.Set("refresh_token", refresh_token)

		t.Run("Valid refresh token, Valid access token", func(t *testing.T) {
			RefreshTokenHelper(t, http.StatusBadRequest, body, access_token)
		})

		t.Run("Valid refresh token, Expired access token", func(t *testing.T) {
			RefreshTokenHelper(t, http.StatusOK, body, expired_access_token)
		})

		body.Set("refresh_token", expired_refresh_token)

		t.Run("Expired refresh token, Valid access token", func(t *testing.T) {
			RefreshTokenHelper(t, http.StatusBadRequest, body, access_token)
		})

		t.Run("Expired refresh token, Expired access token", func(t *testing.T) {
			RefreshTokenHelper(t, http.StatusBadRequest, body, expired_access_token)
		})
	})

	t.Run("Delete User", func(t *testing.T) {
		t.Run("Deletes user 1", func(t *testing.T) {
			DeleteUserHelper(t, http.StatusOK)
		})

		t.Run("Doesnt delete user as user doesnt exist", func(t *testing.T) {
			DeleteUserHelper(t, http.StatusBadRequest)
		})

		body := url.Values{}
		body.Set("email", user2.email)
		body.Set("password", user2.password)

		t.Run("Authenticates user 2", func(t *testing.T) {
			AuthenticateUserHelper(t, body, http.StatusOK)
		})

		t.Run("Deletes user", func(t *testing.T) {
			DeleteUserHelper(t, http.StatusOK)
		})
	})
}

func TestUserHelpers(t *testing.T) {
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

	t.Run("GetUser", func(t *testing.T) {
		t.Run("Get User", func(t *testing.T) {
			user, err := util.GetUser(user1.email, db)
			assert.NoError(t, err, "Error getting user")

			user1_id = user.ID.String()

			assert.Equal(t, user1.name, user.Name)
			assert.Equal(t, user1.email, user.Email)
		})

		t.Run("Email does not Exist", func(t *testing.T) {
			user, err := util.GetUser("iamanonexistantuser@email.com", db)

			assert.Error(t, err, "There should be an error when trying to get a non existent user")
			assert.Nil(t, user, "User should be empty")
		})
	})

	t.Run("DoesUserExist Helper", func(t *testing.T) {
		t.Run("Get User with email", func(t *testing.T) {
			exists, count, err := util.DoesUserExist("", user1.email, db)

			assert.NoError(t, err, "Error getting user")
			assert.True(t, exists, "User should exist")
			assert.Equal(t, 1, count, "Count should be 1")
		})

		t.Run("Email does not Exist", func(t *testing.T) {
			exists, count, err := util.DoesUserExist("", "iamanonexistantuser@email.com", db)

			assert.NoError(t, err, "Error getting user")
			assert.False(t, exists, "User should not exist")
			assert.Equal(t, 0, count, "Count should be 0")
		})

		t.Run("Get User with id", func(t *testing.T) {
			exists, count, err := util.DoesUserExist(user1_id, user1.email, db)

			assert.NoError(t, err, "Error getting user")
			assert.True(t, exists, "User should exist")
			assert.Equal(t, 1, count, "Count should be 1")
		})

		t.Run("ID does not Exist", func(t *testing.T) {
			exists, count, err := util.DoesUserExist(uuid.New().String(), "iamanonexistantuser@email.com", db)

			assert.NoError(t, err, "Error getting user")
			assert.False(t, exists, "User should not exist")
			assert.Equal(t, 0, count, "Count should be 0")
		})
	})

	t.Run("GetUserFromJWT Helper", func(t *testing.T) {
		t.Run("Get User with JWT", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/non-existent-path", nil)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))

			user, code, err := util.GetUserFromJWT(request, db)
			assert.NoError(t, err, "Error getting user")

			assert.Equal(t, code, http.StatusOK)
			assert.Equal(t, user1.name, user.Name)
			assert.Equal(t, user1.email, user.Email)
		})

		t.Run("Fail getting non existent user with JWT", func(t *testing.T) {
			request, _ := http.NewRequest(http.MethodGet, "/non-existent-path", nil)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "asdaund1id10dj"))

			_, code, err := util.GetUserFromJWT(request, db)

			assert.Equal(t, code, http.StatusUnauthorized)
			assert.Error(t, err, "This action should fail as this JWT is not owned by a valid user")
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

	Post(response, request, db)

	assert.Equal(t, want_code, response.Code)
}

func AuthenticateUserHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, "/user/auth", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	Authenticate(response, request, db)

	var responseBody AuthenticateResponse
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	assert.NoError(t, err)

	assert.Equal(t, body.Get("email"), responseBody.User.Email)
	assert.Equal(t, "redacted", responseBody.User.Password, "Password in response should be redacted")
	assert.NotEmpty(t, responseBody.AccessToken, "Access Token is missing")
	assert.NotEmpty(t, responseBody.RefreshToken, "Refresh Token is missing")

	assert.Equal(t, want_code, response.Code)

	access_token = responseBody.AccessToken
	refresh_token = responseBody.RefreshToken
}

func UpdateUserHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPut, "/user", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	Put(response, request, db)

	assert.Equal(t, want_code, response.Code)
}

func UpdateUserPasswordHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPut, "/user/auth/password", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	PutPassword(response, request, db)

	assert.Equal(t, want_code, response.Code)
}

func RefreshTokenHelper(t testing.TB, want_code int, body url.Values, access_token string) {
	t.Helper()
	request, _ := http.NewRequest("POST", "/user/auth/refresh", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	RefreshToken(response, request, db)

	assert.Equal(t, want_code, response.Code)
}

func DeleteUserHelper(t testing.TB, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodDelete, "/user", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	Delete(response, request, db)

	assert.Equal(t, want_code, response.Code)
}
