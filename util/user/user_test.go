package user_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
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
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/test"
	userUtil "github.com/ushiradineth/cron-be/util/user"
	"github.com/ushiradineth/cron-be/util/validator"
)

var (
	accessToken         string
	refreshToken        string
	user1ID             string
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

var user1Auth auth.AuthenticateBodyParams = auth.AuthenticateBodyParams{
	Email:    user1.Email,
	Password: user1.Password,
}

func TestInit(t *testing.T) {
	t.Run("Initiate Dependencies", func(t *testing.T) {
		err := godotenv.Load("../../.env")
		if err != nil {
			log.Println("Failed to load env")
		}

		db = test.NewDB("../../database/migration")
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

	t.Run("Create User", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, user1, http.StatusOK, response.StatusSuccess)
	})

	t.Run("Authenticates user 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})
}

func TestGetUserHelper(t *testing.T) {
	t.Run("Get User 1", func(t *testing.T) {
		response := httptest.NewRecorder()
		user := userUtil.GetUser(response, user1.Email, db)
		assert.NotNil(t, user, "Error getting user")

		assert.Equal(t, user1.Name, user.Name)
		assert.Equal(t, user1.Email, user.Email)
	})

	t.Run("Email does not Exist", func(t *testing.T) {
		response := httptest.NewRecorder()
		user := userUtil.GetUser(response, "not_an_user@email.com", db)
		assert.Nil(t, user, "User should be empty")
	})
}

func TestDoesUserExistHelper(t *testing.T) {
	t.Run("Get User 1 with Email", func(t *testing.T) {
		exists, count, err := userUtil.DoesUserExist("", user1.Email, db)

		assert.NoError(t, err, "Error getting user")
		assert.True(t, exists, "User should exist")
		assert.Equal(t, 1, count, "Count should be 1")
	})

	t.Run("Email does not Exist", func(t *testing.T) {
		exists, count, err := userUtil.DoesUserExist("", "not_an_user@email.com", db)

		assert.NoError(t, err, "Error getting user")
		assert.False(t, exists, "User should not exist")
		assert.Equal(t, 0, count, "Count should be 0")
	})

	t.Run("Get User 1 with ID", func(t *testing.T) {
		exists, count, err := userUtil.DoesUserExist(user1ID, user1.Email, db)

		assert.NoError(t, err, "Error getting user")
		assert.True(t, exists, "User should exist")
		assert.Equal(t, 1, count, "Count should be 1")
	})

	t.Run("ID does not Exist", func(t *testing.T) {
		exists, count, err := userUtil.DoesUserExist(uuid.New().String(), "not_an_user@email.com", db)

		assert.NoError(t, err, "Error getting user")
		assert.False(t, exists, "User should not exist")
		assert.Equal(t, 0, count, "Count should be 0")
	})

	t.Run("ID is invalid", func(t *testing.T) {
		exists, count, err := userUtil.DoesUserExist("not_an_uuid", "not_an_user@email.com", db)

		assert.Error(t, err, "This action should error")
		assert.False(t, exists, "User should not exist")
		assert.Equal(t, 0, count, "Count should be 0")
	})
}

func TestGetUserFromJWTHelper(t *testing.T) {
	t.Run("Get User with JWT", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/non-existent-path", nil)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

		response := httptest.NewRecorder()

		user := userUtil.GetUserFromJWT(request, response, db)

		assert.Equal(t, response.Code, http.StatusOK)
		assert.Equal(t, user1.Name, user.Name)
		assert.Equal(t, user1.Email, user.Email)
	})

	t.Run("User does not exist", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/non-existent-path", nil)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", expiredAccessToken))

		response := httptest.NewRecorder()

		userUtil.GetUserFromJWT(request, response, db)

		assert.Equal(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("JWT is invalid", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/non-existent-path", nil)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "not_an_jwt"))

		response := httptest.NewRecorder()

		userUtil.GetUserFromJWT(request, response, db)

		assert.Equal(t, response.Code, http.StatusUnauthorized)
	})

	t.Run("Authorization header is invalid", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/non-existent-path", nil)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("invalid %v", "not_an_jwt"))

		response := httptest.NewRecorder()

		userUtil.GetUserFromJWT(request, response, db)

		assert.Equal(t, response.Code, http.StatusBadRequest)
	})
}

func TestCleanUp(t *testing.T) {
	t.Run("Delete user", func(t *testing.T) {
		test.DeleteUserHelper(userAPI, t, http.StatusOK, response.StatusSuccess, user1ID, accessToken)
	})
}
