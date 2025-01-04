package user_test

import (
	"database/sql"
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
	logger "github.com/ushiradineth/cron-be/util/log"
	"github.com/ushiradineth/cron-be/util/response"
	"github.com/ushiradineth/cron-be/util/test"
	userUtil "github.com/ushiradineth/cron-be/util/user"
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

var user1Auth auth.AuthenticateBodyParams = auth.AuthenticateBodyParams{
	Email:    user1.Email,
	Password: user1.Password,
}

var user2 user.PostBodyParams = user.PostBodyParams{
	Name:     faker.Name(),
	Email:    faker.Email(),
	Password: "UPlow1234!@#",
}

var user2Auth auth.AuthenticateBodyParams = auth.AuthenticateBodyParams{
	Email:    user2.Email,
	Password: user2.Password,
}

func TestInit(t *testing.T) {
	t.Run("Initiate Dependencies", func(t *testing.T) {
		err := godotenv.Load("../../.env")
		if err != nil {
			log.Println("Failed to load env")
		}

		db = test.NewDB("../../database/migration")
		v := validator.New()
		l := logger.New()

		userAPI = user.New(db, v, l)
		authAPI = auth.New(db, v, l)

		expiredAccessToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1234567890", "iat": time.Now().Unix(), "exp": time.Now().Add(-1 * time.Hour).Unix()}).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()

		expiredRefreshToken = func() string {
			token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1234567890", "iat": time.Now().Unix(), "exp": time.Now().Add(-1 * time.Hour).Unix()}).SignedString([]byte(os.Getenv("JWT_SECRET")))
			return token
		}()
	})

	t.Run("Create User 2", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, user2, http.StatusOK, response.StatusSuccess)
	})

	t.Run("Authenticates User 2", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user2Auth, http.StatusOK, response.StatusSuccess, &user2ID, &accessToken, &refreshToken)
	})

	t.Run("Create User 1", func(t *testing.T) {
		test.CreateUserHelper(userAPI, t, user1, http.StatusOK, response.StatusSuccess)
	})

	t.Run("Authenticates User 1", func(t *testing.T) {
		test.AuthenticateUserHelper(authAPI, t, user1Auth, http.StatusOK, response.StatusSuccess, &user1ID, &accessToken, &refreshToken)
	})
}

func TestGetUserByEmailHelper(t *testing.T) {
	t.Run("Get User 1", func(t *testing.T) {
		user, err := userUtil.GetUserByEmail(user1.Email, db)
		assert.NotNil(t, user, "Error getting user")
		assert.NoError(t, err, "Error getting user")

		assert.Equal(t, user1.Name, user.Name)
		assert.Equal(t, user1.Email, user.Email)
	})

	t.Run("Email does not Exist", func(t *testing.T) {
		user, err := userUtil.GetUserByEmail("not_an_user@email.com", db)
		assert.Nil(t, user, "User should be empty")
		assert.Equal(t, err, sql.ErrNoRows, "Error should be sql.ErrNoRows")
	})
}

func TestGetUserByIdHelper(t *testing.T) {
	t.Run("Get User 1", func(t *testing.T) {
		user, err := userUtil.GetUserById(user1ID, db)
		assert.NotNil(t, user, "Error getting user")
		assert.NoError(t, err, "Error getting user")

		assert.Equal(t, user1.Name, user.Name)
		assert.Equal(t, user1.Email, user.Email)
	})

	t.Run("ID does not Exist", func(t *testing.T) {
		user, err := userUtil.GetUserById(uuid.New().String(), db)
		assert.Nil(t, user, "User should be empty")
		assert.Equal(t, err, sql.ErrNoRows, "Error should be sql.ErrNoRows")
	})

	t.Run("ID is invalid", func(t *testing.T) {
		user, err := userUtil.GetUserById("not_an_uuid", db)
		assert.Nil(t, user, "User should be empty")
		assert.Error(t, err, "This action should error")
	})
}

func TestIsEmailInUseHelper(t *testing.T) {
	t.Run("Email in use by a different user", func(t *testing.T) {
		exists, err := userUtil.IsEmailInUse(user2.Email, user1ID, db)

		assert.NoError(t, err, "Error getting user")
		assert.True(t, exists, "Email should be in use by a different user")
	})

	t.Run("Email in use by the same user", func(t *testing.T) {
		exists, err := userUtil.IsEmailInUse(user1.Email, user1ID, db)

		assert.NoError(t, err, "Error getting user")
		assert.False(t, exists, "Email should be in use by the same user")
	})

	t.Run("Email is not in use", func(t *testing.T) {
		exists, err := userUtil.IsEmailInUse("not_in_use@email.com", user1ID, db)

		assert.NoError(t, err, "Error getting user")
		assert.False(t, exists, "Email should not be in use")
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
