package auth_test

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
	"github.com/ushiradineth/cron-be/api/resource/event"
	"github.com/ushiradineth/cron-be/api/resource/user"
	"github.com/ushiradineth/cron-be/database"
	authUtil "github.com/ushiradineth/cron-be/util/auth"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/test"
	"github.com/ushiradineth/cron-be/util/validator"
)

var (
	accessToken            string
	refreshToken           string
	user1ID                string
	user2ID                string
	expiredAccessToken     string
	expiredRefreshToken    string
	deletedUserAccessToken string
	db                     *sqlx.DB
	userAPI                *user.API
	authAPI                *auth.API
	eventAPI               *event.API
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
		eventAPI = event.New(db, v)
		authAPI = auth.New(db, v)

		body := url.Values{}
		body.Set("name", user1.Name)
		body.Set("email", user1.Email)
		body.Set("password", user1.Password)
		t.Run("Create User 1", func(t *testing.T) {
			test.CreateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user1)
		})

		t.Run("Authenticates User 1", func(t *testing.T) {
			test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
		})

		body.Set("name", user2.Name)
		body.Set("email", user2.Email)
		body.Set("password", user2.Password)
		t.Run("Create User 2", func(t *testing.T) {
			test.CreateUserHelper(userAPI, t, body, http.StatusOK, response.StatusSuccess, user2)
		})

		userIDUUID, _ := uuid.Parse(user1ID)
		expiredClaim := authUtil.UserClaim{
			Id:    userIDUUID,
			Name:  user1.Name,
			Email: user1.Email,
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
			},
		}

		expiredAccessToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaim).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()

		expiredRefreshToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaim).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()

    expiredClaim.Email = faker.Email()
		deletedUserAccessToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaim).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()
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

  body.Set("email", "not_an_user@email.com")
  bodyStruct.Email = "not_an_user@email.com"
  t.Run("Email is not registered", func(t *testing.T) {
    test.AuthenticateUserHelper(authAPI, t, body, http.StatusBadRequest, response.StatusFail, &user1ID, &accessToken, &refreshToken)
  })

	body.Set("email", "not_an_email")
	bodyStruct.Email = "not_an_email"
	t.Run("Email is invalid", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusBadRequest, response.StatusFail, &user1ID, &accessToken, &refreshToken)
	})

	body.Set("email", user1.Email)
	bodyStruct.Email = user1.Email
	body.Set("password", "not_a_password")
	bodyStruct.Password = "not_a_password"
	t.Run("Password is invalid", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusBadRequest, response.StatusFail, &user1ID, &accessToken, &refreshToken)
	})

	body.Set("email", user1.Email)
	body.Set("password", user2.Password)
	bodyStruct.Email = user1.Email
	bodyStruct.Password = user2.Password
	t.Run("Wrong credentials", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusUnauthorized, response.StatusFail, &user1ID, &accessToken, &refreshToken)
	})
}

func TestUpdateUserPasswordHandler(t *testing.T) {
	t.Run("Update User Password", func(t *testing.T) {
		body := url.Values{}
		body.Set("email", user1.Email)
		body.Set("password", user1.Password)

		t.Run("Authenticates user 1", func(t *testing.T) {
			test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
		})

		body.Set("password", user2.Password)
		t.Run("Update user 1 password", func(t *testing.T) {
			test.UpdateUserPasswordHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, accessToken)
		})

		body.Set("password", user1.Password)
		t.Run("Reset user 1", func(t *testing.T) {
			test.UpdateUserPasswordHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, accessToken)
		})

		t.Run("JWT is invalid", func(t *testing.T) {
			test.UpdateUserPasswordHelper(authAPI, t, body, http.StatusUnauthorized, response.StatusFail, expiredAccessToken)
		})

		t.Run("JWT is expired", func(t *testing.T) {
			test.UpdateUserPasswordHelper(authAPI, t, body, http.StatusUnauthorized, response.StatusFail, expiredAccessToken)
		})

		body.Set("password", "not_a_password")
		t.Run("Password is invalid", func(t *testing.T) {
			test.UpdateUserPasswordHelper(authAPI, t, body, http.StatusBadRequest, response.StatusFail, accessToken)
		})
	})
}

func TestRefreshTokenHandler(t *testing.T) {
	t.Run("Refresh Token", func(t *testing.T) {
		body := url.Values{}

		body.Set("refresh_token", refreshToken)
		t.Run("Valid refresh token, Valid access token", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, accessToken, http.StatusBadRequest, response.StatusFail)
		})

		t.Run("Valid refresh token, Expired access token", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, expiredAccessToken, http.StatusOK, response.StatusSuccess)
		})

		t.Run("JWT user does not exist", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, deletedUserAccessToken, http.StatusBadRequest, response.StatusFail)
		})

		body.Set("refresh_token", "not_a_refresh_token")
		t.Run("Refresh token is invalid", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, accessToken, http.StatusBadRequest, response.StatusFail)
		})

		body.Set("refresh_token", expiredRefreshToken)
		t.Run("Expired refresh token, Valid access token", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, accessToken, http.StatusBadRequest, response.StatusFail)
		})

		t.Run("Expired refresh token, Expired access token", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, expiredAccessToken, http.StatusBadRequest, response.StatusFail)
		})

		t.Run("Expired refresh token, Valid access token", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, accessToken, http.StatusBadRequest, response.StatusFail)
		})

		t.Run("JWT is invalid", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, "not_an_access_token", http.StatusBadRequest, response.StatusFail)
		})

		t.Run("JWT is expired", func(t *testing.T) {
			test.RefreshTokenHelper(authAPI, t, body, expiredAccessToken, http.StatusBadRequest, response.StatusFail)
		})
	})
}

func TestCleanUp(t *testing.T) {
	body := url.Values{}

	body.Set("email", user1.Email)
	body.Set("password", user1.Password)
	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})

	t.Run("Delete User 1", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})

	body.Set("email", user2.Email)
	body.Set("password", user2.Password)
	t.Run("Authenticate User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, body, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("Delete User 2", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user2ID, accessToken)
	})
}
