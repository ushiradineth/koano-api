package user_test

import (
	"net/http"
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
	"github.com/ushiradineth/cron-be/api/resource/user"
	"github.com/ushiradineth/cron-be/database"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/test"
	"github.com/ushiradineth/cron-be/util/validator"
)

var (
	accessToken         string
	refreshToken        string
	user1ID             string
	user2ID             string
	expiredAccessToken  string
	expiredRefreshToken string
	db                  *sqlx.DB
	userAPI             *user.API
	authAPI             *auth.API
)

var user1 user.PostQueryParams = user.PostQueryParams{
	Name:     faker.Name(),
	Email:    faker.Email(),
	Password: "UPlow1234!@#",
}

var user2 user.PostQueryParams = user.PostQueryParams{
	Name:     faker.Name(),
	Email:    faker.Email(),
	Password: "lowUP1234!@#",
}

func TestInit(t *testing.T) {
	t.Run("Initiate Dependencies", func(t *testing.T) {
		assert.NoError(t, godotenv.Load("../../../.env"), "Environment variables should be loaded in")

		db = database.New()
		v := validator.New()

		userAPI = user.New(db, v)
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
}

func TestCreateUserHandler(t *testing.T) {
	body := url.Values{}
	bodyStruct := user1

	body.Set("name", user1.Name)
	body.Set("email", user1.Email)
	body.Set("password", user1.Password)
	t.Run("Create User 1", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user1)
	})

	t.Run("Email already exists", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, user1)
	})

	body.Set("name", user2.Name)
	body.Set("email", user2.Email)
	body.Set("password", user2.Password)
	bodyStruct.Name = user2.Name
	bodyStruct.Email = user2.Email
	bodyStruct.Password = user2.Password
	t.Run("Create User 2", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user2)
	})

	body.Set("email", "not_an_email")
	bodyStruct.Email = "not_an_email"
	t.Run("Email is invalid", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, user1)
	})
	body.Set("email", "placeholder@email.com")
	bodyStruct.Email = "placeholder@email.com"

	body.Set("password", "notastandardpassword")
	bodyStruct.Password = "notastandardpassword"
	t.Run("Password is invalid", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, user1)
	})
}

func TestAuthenticateUserHandler(t *testing.T) {
	body := url.Values{}
	bodyStruct := user2

	body.Set("email", user2.Email)
	body.Set("password", user2.Password)
	bodyStruct.Email = user2.Email
	bodyStruct.Password = user2.Password
	t.Run("Authenticate User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	body.Set("email", user1.Email)
	body.Set("password", user1.Password)
	bodyStruct.Email = user1.Email
	bodyStruct.Password = user1.Password
	t.Run("Authenticate User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	body.Set("email", user1.Email)
	body.Set("password", user2.Password)
	bodyStruct.Email = user1.Email
	bodyStruct.Password = user2.Password
	t.Run("Wrong credentials", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusUnauthorized, response.StatusFail, &user1ID, &accessToken, &refreshToken)
	})
}

func TestGetUserHandler(t *testing.T) {
	t.Run("Get user 1", func(t *testing.T) {
		test.GetUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1, user1ID, accessToken)
	})

	t.Run("JWT does not match user ID", func(t *testing.T) {
		test.GetUserHelper(userAPI, t, http.StatusUnauthorized, response.StatusFail, user1, user2ID, accessToken)
	})

	t.Run("JWT is invalid", func(t *testing.T) {
		test.GetUserHelper(userAPI, t, http.StatusUnauthorized, response.StatusFail, user1, user1ID, expiredAccessToken)
	})

	t.Run("UUID is invalid", func(t *testing.T) {
		test.GetUserHelper(userAPI, t, http.StatusBadRequest, response.StatusFail, user1, "not_an_uuid", accessToken)
	})
}

func TestUpdateUserHandler(t *testing.T) {
	body := url.Values{}
	bodyStruct := user.PostQueryParams{
		Name:     user1.Name,
		Email:    user1.Email,
		Password: user1.Password,
	}

	body.Set("name", user2.Name)
	body.Set("email", user1.Email)
	bodyStruct.Name = user2.Name
	bodyStruct.Email = user1.Email
	t.Run("Update user name", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, bodyStruct, user1ID, accessToken)
	})

	body.Set("email", user2.Email)
	bodyStruct.Email = user2.Email
	t.Run("Email already exists", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, bodyStruct, user1ID, accessToken)
	})

	t.Run("JWT does not match user ID", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusUnauthorized, response.StatusFail, bodyStruct, user2ID, accessToken)
	})

	body.Set("email", "newemail@gmail.com")
	bodyStruct.Email = "newemail@gmail.com"
	t.Run("Update user 1 email", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, bodyStruct, user1ID, accessToken)
	})

	t.Run("JWT is invalid", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusUnauthorized, response.StatusFail, bodyStruct, user1ID, expiredAccessToken)
	})

	body.Set("email", user1.Email)
	bodyStruct.Email = user1.Email
	t.Run("Reset user 1", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, bodyStruct, user1ID, accessToken)
	})

	t.Run("UUID is invalid", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, bodyStruct, "not_an_uuid", accessToken)
	})

	body.Set("name", "")
	bodyStruct.Name = ""
	t.Run("Name is required", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, bodyStruct, user1ID, accessToken)
	})
	body.Set("name", user1.Name)
	bodyStruct.Name = user1.Name

	body.Set("email", "not_an_email")
	bodyStruct.Email = "not_an_email"
	t.Run("Email is invalid", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, bodyStruct, user1ID, accessToken)
	})
}

func TestDeleteUserHandler(t *testing.T) {
	t.Run("Delete User 1", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	t.Run("User does not exist", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusBadRequest, response.StatusFail, uuid.NewString(), accessToken)
	})

	body := url.Values{}
	body.Set("email", user2.Email)
	body.Set("password", user2.Password)
	t.Run("Authenticate User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("JWT does not match user ID", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusUnauthorized, response.StatusFail, user1ID, accessToken)
	})

	t.Run("Delete User 2", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user2ID, accessToken)
	})

	t.Run("JWT is invalid", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusUnauthorized, response.StatusFail, user1ID, expiredAccessToken)
	})

	t.Run("UUID is invalid", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusBadRequest, response.StatusFail, "not_an_uuid", accessToken)
	})
}
