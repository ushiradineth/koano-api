package auth_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/util/auth"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		password := "password"

		hashedPassword, err := auth.HashPassword(password)
		assert.NoError(t, err, "HashPassword should not return an error")
		assert.NotEmpty(t, hashedPassword, "Hashed password should not be empty")
		assert.NotEqual(t, password, hashedPassword, "Hashed password should not match the original password")

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		assert.NoError(t, err, "The bcrypt hash comparison should not return an error for the correct password")
	})

	t.Run("Wrong Password", func(t *testing.T) {
		password := "password"

		hashedPassword, err := auth.HashPassword(password)
		assert.NoError(t, err, "HashPassword should not return an error")
		assert.True(t, auth.CheckPasswordHash(password, hashedPassword), "CheckPasswordHash should return true for a correct password")

		incorrectPassword := "wrongpassword"
		assert.False(t, auth.CheckPasswordHash(incorrectPassword, hashedPassword), "CheckPasswordHash should return false for an incorrect password")
	})

	t.Run("Empty Password", func(t *testing.T) {
		password := ""
		hashedPassword, err := auth.HashPassword(password)

		assert.NoError(t, err, "HashPassword should not return an error for an empty password")
		assert.True(t, auth.CheckPasswordHash(password, hashedPassword), "CheckPasswordHash should return true for an empty password")
		assert.False(t, auth.CheckPasswordHash("nonemptypassword", hashedPassword), "CheckPasswordHash should return false for a non-empty password")
	})
}

func TestNewAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	id := uuid.New()
	name := "Test User"
	email := "test@example.com"

	token, expiresIn, expiresAt, err := auth.NewAccessToken(id, name, email)
	assert.NoError(t, err, "NewAccessToken should not return an error")
	assert.NotEmpty(t, token, "NewAccessToken should return a non-empty token")
	assert.NotEmpty(t, expiresIn, "NewAccessToken should return a non-empty expiresIn")
	assert.NotEmpty(t, expiresAt, "NewAccessToken should return a non-empty expiresAt")
}

func TestNewRefreshToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	token, err := auth.NewRefreshToken()
	assert.NoError(t, err, "NewRefreshToken should not return an error")
	assert.NotEmpty(t, token, "NewRefreshToken should return a non-empty token")
}

func TestParseAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	id := uuid.New()
	name := "Test User"
	email := "test@example.com"
	token, expiresIn, expiresAt, err := auth.NewAccessToken(id, name, email)

	w := httptest.NewRecorder()
	claims := auth.ParseAccessToken(w, token)
	assert.NotNil(t, claims, "Parsed access token claims should not be nil")
	assert.Equal(t, id, claims.Id, "Parsed token ID should match")
	assert.Equal(t, name, claims.Name, "Parsed token name should match")
	assert.Equal(t, email, claims.Email, "Parsed token email should match")
	assert.Equal(t, expiresAt, claims.StandardClaims.ExpiresAt, "Parsed token expiresAt should match")
	assert.NotEmpty(t, expiresIn, "Parsed token expiresIn should not be empty")
	assert.NoError(t, err, "Parsed access token should not return an error")

	w = httptest.NewRecorder()
	claims = auth.ParseAccessToken(w, "invalidtoken")
	assert.Nil(t, claims, "Parsed claims should be nil for an invalid token")
}

func TestParseRefreshToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	w := httptest.NewRecorder()
	token, _ := auth.NewRefreshToken()

	parsedClaims := auth.ParseRefreshToken(w, token)
	assert.NotNil(t, parsedClaims, "Parsed token claims should not be nil")

	w = httptest.NewRecorder()
	parsedClaims = auth.ParseRefreshToken(w, "invalidtoken")
	assert.Nil(t, parsedClaims, "Parsed claims should be nil for an invalid token")
}

func TestGetJWT(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer testtoken")

	token, err := auth.GetJWT(req)
	assert.NoError(t, err, "GetJWT should not return an error for a valid Authorization header")
	assert.Equal(t, "testtoken", token, "Extracted token should match")

	req.Header.Del("Authorization")
	_, err = auth.GetJWT(req)
	assert.Error(t, err, "GetJWT should return an error if the Authorization header is missing")

	req.Header.Set("Authorization", "InvalidFormat")
	_, err = auth.GetJWT(req)
	assert.Error(t, err, "GetJWT should return an error for an invalid Authorization header format")
}

func TestParseExpiredAccessToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")

	w := httptest.NewRecorder()
	claims := auth.UserClaim{
		Id:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	parsedClaims := auth.ParseExpiredAccessToken(w, expiredToken)
	assert.NotNil(t, parsedClaims, "Parsed expired access token claims should not be nil")
	assert.Equal(t, claims.Id, parsedClaims.Id, "Parsed token ID should match")
	assert.Equal(t, claims.Email, parsedClaims.Email, "Parsed token email should match")

	w = httptest.NewRecorder()
	validToken, _, _, _ := auth.NewAccessToken(claims.Id, claims.Name, claims.Email)
	parsedClaims = auth.ParseExpiredAccessToken(w, validToken)
	assert.Nil(t, parsedClaims, "Parsed claims should be nil for a valid token")
}
