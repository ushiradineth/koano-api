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
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/models"
	"github.com/ushiradineth/cron-be/util"
)

func assertResponse(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("response is wrong, got %q want %q", got, want)
	}
}

var user1_access_token string = ""
var user1_refresh_token string = ""
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

func TestInit(t *testing.T) {
	t.Run("Creates user", func(t *testing.T) {
		godotenv.Load("../.env")

		// Harded coded so I don't delete main DB :)
		db = database.Configure()
	})
}

func CreateUserHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, "/user", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	PostUserHandler(response, request, db)

	got := fmt.Sprint(response.Code)
	want := fmt.Sprint(want_code)

	assertResponse(t, got, want)
}

func TestPostUserHandler(t *testing.T) {
	body := url.Values{}
	body.Set("name", user1.name)
	body.Set("email", user1.email)
	body.Set("password", user1.password)

	t.Run("Creates user 1", func(t *testing.T) {
		CreateUserHelper(t, body, http.StatusOK)
	})

	t.Run("Doesnt create user as the email already exists", func(t *testing.T) {
		CreateUserHelper(t, body, http.StatusBadRequest)
	})

	body.Set("name", user2.name)
	body.Set("email", user2.email)
	body.Set("password", user2.password)

	t.Run("Creates user 2", func(t *testing.T) {
		CreateUserHelper(t, body, http.StatusOK)
	})
}

func AuthenticateUserHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPost, "/user/auth", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	AuthenticateUserHandler(response, request, db)

	assertResponse(t, fmt.Sprint(response.Code), fmt.Sprint(want_code))

	var responseBody AuthenticateUserResponse
	if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	if responseBody.User.Email != body.Get("email") {
		t.Errorf("Email in response does not match expected: got %s, want %s", responseBody.User.Email, body.Get("email"))
	}

	if responseBody.User.Password != "" {
		t.Errorf("Password in response does not match expected: got %s, want %s", responseBody.User.Password, "")
	}

	if responseBody.AccessToken == "" {
		t.Error("Access Token is missing")
	}

	user1_access_token = responseBody.AccessToken
	user1_refresh_token = responseBody.RefreshToken
}

func TestAuthenticateUserHandler(t *testing.T) {
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

		AuthenticateUserHandler(response, request, db)

		got := fmt.Sprint(response.Code)
		want := fmt.Sprint(http.StatusUnauthorized)

		assertResponse(t, got, want)
	})
}

func TestGetUserHandler(t *testing.T) {
	body := url.Values{}
	body.Set("name", user1.name)
	body.Set("email", user1.email)
	body.Set("password", user1.password)

	t.Run("Get User", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/user", nil)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", user1_access_token))
		response := httptest.NewRecorder()
		GetUserHandler(response, request, db)

		assertResponse(t, fmt.Sprint(response.Code), fmt.Sprint(http.StatusOK))

		var responseBody models.User
		if err := json.NewDecoder(response.Body).Decode(&responseBody); err != nil {
			t.Errorf("Failed to decode response body: %v", err)
		}

		if responseBody.Name != body.Get("name") {
			t.Errorf("Name in response does not match expected: got %s, want %s", responseBody.Name, body.Get("name"))
		}

		if responseBody.Email != body.Get("email") {
			t.Errorf("Email in response does not match expected: got %s, want %s", responseBody.Email, body.Get("email"))
		}

		if responseBody.Password != "" {
			t.Errorf("Password in response does not match expected: got %s, want %s", responseBody.Password, "")
		}
	})
}

func UpdateUserHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPut, "/user", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", user1_access_token))
	response := httptest.NewRecorder()

	PutUserHandler(response, request, db)

	got := fmt.Sprint(response.Code)
	want := fmt.Sprint(want_code)

	assertResponse(t, got, want)
}

func TestPutUserHandler(t *testing.T) {
	body := url.Values{}
	body.Set("name", "Edited Ushira Dineth")
	body.Set("email", user1.email)

	t.Run("Update user name", func(t *testing.T) {
		UpdateUserHelper(t, body, http.StatusOK)
	})

	body.Set("name", user1.name)
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
}

func UpdateUserPasswordHelper(t testing.TB, body url.Values, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodPut, "/user/auth/password", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", user1_access_token))
	response := httptest.NewRecorder()

	PutUserPasswordHandler(response, request, db)

	got := fmt.Sprint(response.Code)
	want := fmt.Sprint(want_code)

	assertResponse(t, got, want)
}

func TestPutUserPasswordHandler(t *testing.T) {
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

		AuthenticateUserHandler(response, request, db)

		got := fmt.Sprint(response.Code)
		want := fmt.Sprint(http.StatusUnauthorized)

		assertResponse(t, got, want)
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
}

func TestGetUserHelper(t *testing.T) {
	t.Run("Get User", func(t *testing.T) {
		user, err := util.GetUser(user1.email, db)
		if err != nil {
			t.Errorf("Error getting user: %s", err)
		}

		user1_id = user.ID.String()

		if user.Name != user1.name {
			t.Errorf("Name in response does not match expected: got %s, want %s", user.Name, user1.name)
		}

		if user.Email != user1.email {
			t.Errorf("Email in response does not match expected: got %s, want %s", user.Email, user1.email)
		}
	})

	t.Run("Fail getting non existant user", func(t *testing.T) {
		user, err := util.GetUser("iamanonexistantuser@email.com", db)
		if err == nil {
			t.Error("There should be an error when trying to get a non existant user")
		}

		if user != nil {
			t.Error("User should be empty")
		}
	})
}

func TestDoesUserExistHelper(t *testing.T) {
	t.Run("Get User with email", func(t *testing.T) {
		exists, count, err := util.DoesUserExist("", user1.email, db)
		if err != nil {
			t.Errorf("Error getting user: %s", err)
		}

		if !exists {
			t.Error("User should exists")
		}

		assertResponse(t, fmt.Sprint(count), fmt.Sprint(1))
	})

	t.Run("Fail getting non existant user with email", func(t *testing.T) {
		exists, count, err := util.DoesUserExist("", "iamanonexistantuser@email.com", db)
		if err != nil {
			t.Errorf("Error getting user: %s", err)
		}

		if exists {
			t.Error("User shouldnt exists")
		}

		assertResponse(t, fmt.Sprint(count), fmt.Sprint(0))
	})

	t.Run("Get User with id", func(t *testing.T) {
		exists, count, err := util.DoesUserExist(user1_id, user1.email, db)
		if err != nil {
			t.Errorf("Error getting user: %s", err)
		}

		if !exists {
			t.Error("User should exists")
		}

		assertResponse(t, fmt.Sprint(count), fmt.Sprint(1))
	})

	t.Run("Fail getting non existant user with id", func(t *testing.T) {
		exists, count, err := util.DoesUserExist(uuid.New().String(), "iamanonexistantuser@email.com", db)
		if err != nil {
			t.Errorf("Error getting user: %s", err)
		}

		if exists {
			t.Error("User shouldnt exists")
		}

		assertResponse(t, fmt.Sprint(count), fmt.Sprint(0))
	})
}

func TestGetUserFromJWTHelper(t *testing.T) {
	t.Run("Get User with token", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/user/non-existant-path", nil)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", user1_access_token))

		user, err := util.GetUserFromJWT(request, db)
		if err != nil {
			t.Errorf("Error getting user: %s", err)
		}

		if user.Name != user1.name {
			t.Errorf("Name in response does not match expected: got %s, want %s", user.Name, user1.name)
		}

		if user.Email != user1.email {
			t.Errorf("Email in response does not match expected: got %s, want %s", user.Email, user1.email)
		}
	})

	t.Run("Fail getting non existant user with token", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/user/non-existant-path", nil)
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "asdaund1id10dj"))

		util.GetUserFromJWT(request, db)
	})
}

func RefreshTokenHelper(t testing.TB, want_code int, body url.Values, access_token string) {
	t.Helper()
	request, _ := http.NewRequest("POST", "/user/auth/refresh", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	response := httptest.NewRecorder()

	RefreshTokenHandler(response, request, db)

	assertResponse(t, fmt.Sprint(response.Code), fmt.Sprint(want_code))
}

func TestRefreshTokenHandler(t *testing.T) {
	body := url.Values{}
	body.Set("refresh_token", user1_refresh_token)

	t.Run("Valid refresh token, Valid access token", func(t *testing.T) {
		RefreshTokenHelper(t, http.StatusBadRequest, body, user1_access_token)
	})

	t.Run("Valid refresh token, Expired access token", func(t *testing.T) {
		RefreshTokenHelper(t, http.StatusOK, body, expired_access_token)
	})

	body.Set("refresh_token", expired_refresh_token)

	t.Run("Expired refresh token, Valid access token", func(t *testing.T) {
		RefreshTokenHelper(t, http.StatusBadRequest, body, user1_access_token)
	})

	t.Run("Expired refresh token, Expired access token", func(t *testing.T) {
		RefreshTokenHelper(t, http.StatusBadRequest, body, expired_access_token)
	})
}

func DeleteUserHelper(t testing.TB, want_code int) {
	t.Helper()
	request, _ := http.NewRequest(http.MethodDelete, "/user", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", user1_access_token))
	response := httptest.NewRecorder()

	DeleteUserHandler(response, request, db)

	got := fmt.Sprint(response.Code)
	want := fmt.Sprint(want_code)

	assertResponse(t, got, want)
}

func TestDeleteUserHandler(t *testing.T) {
	t.Run("Deletes user 1", func(t *testing.T) {
		DeleteUserHelper(t, http.StatusOK)
	})

	t.Run("Doesnt delete user as user doesnt exist", func(t *testing.T) {
		DeleteUserHelper(t, http.StatusInternalServerError)
	})

	body := url.Values{}
	body.Set("email", user2.email)
	body.Set("password", user2.password)

	t.Run("Authenticates user 1", func(t *testing.T) {
		AuthenticateUserHelper(t, body, http.StatusOK)
	})

	t.Run("Deletes user", func(t *testing.T) {
		DeleteUserHelper(t, http.StatusOK)
	})
}
