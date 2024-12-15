package user_test

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
	expiredAccessToken  string
	expiredRefreshToken string
	db                  *sqlx.DB
	userAPI             *user.API
	authAPI             *auth.API
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

func TestInit(t *testing.T) {
	t.Run("Initiate Dependencies", func(t *testing.T) {
		err := godotenv.Load("../../../.env")
		if err != nil {
			log.Println("Failed to load env")
		}

		db = test.NewDB("../../../database/migration")
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
	body := user.PostBodyParams{
		Name:     user1.Name,
		Email:    user1.Email,
		Password: user1.Password,
	}

	t.Run("Create User 1", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess)
	})

	t.Run("Email already exists", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail)
	})

	body = user.PostBodyParams{
		Name:     user2.Name,
		Email:    user2.Email,
		Password: user2.Password,
	}
	t.Run("Create User 2", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess)
	})

	body.Email = "not_an_email"
	t.Run("Email is invalid", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail)
	})
	body.Email = "placeholder@email.com"

	body.Password = "notastandardpassword"
	t.Run("Password is invalid", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail)
	})
}

func TestGetUserHandler(t *testing.T) {
	t.Run("Authenticate User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("Authenticate User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

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
	body := user.PostBodyParams{
		Name:     user1.Name,
		Email:    user1.Email,
		Password: user1.Password,
	}

	body.Name = user2.Name
	t.Run("Update user name", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	body.Email = user2.Email
	t.Run("Email already exists", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, user1ID, accessToken)
	})

	t.Run("JWT does not match user ID", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusUnauthorized, response.StatusFail, user2ID, accessToken)
	})

	body.Email = "newemail@gmail.com"
	t.Run("Update user 1 email", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	t.Run("JWT is invalid", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusUnauthorized, response.StatusFail, user1ID, expiredAccessToken)
	})

	body.Email = user1.Email
	t.Run("Reset user 1", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	t.Run("UUID is invalid", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, "not_an_uuid", accessToken)
	})

	body.Name = ""
	t.Run("Name is required", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, user1ID, accessToken)
	})
	body.Name = user1.Name

	body.Email = "not_an_email"
	t.Run("Email is invalid", func(t *testing.T) {
		test.UpdateUserHelper(userAPI, t, body, http.StatusBadRequest, response.StatusFail, user1ID, accessToken)
	})
}

func TestDeleteUserHandler(t *testing.T) {
	t.Run("Delete User 1", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	t.Run("User does not exist", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusBadRequest, response.StatusFail, uuid.NewString(), accessToken)
	})

	t.Run("Authenticate User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
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
