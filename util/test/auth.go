package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/koano-api/api/resource/auth"
)

func AuthenticateUserHelper(authAPI *auth.API, t testing.TB, body auth.AuthenticateBodyParams, want_code int, want_status string, userId *string, accessToken *string, refreshToken *string) {
	t.Helper()

	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	authAPI.Authenticate(res, req)

	assert.Equal(t, want_code, res.Code)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		dataMap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		userMap, ok := dataMap["user"].(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, body.Email, userMap["email"])
		assert.Equal(t, "redacted", userMap["password"], "Password in response should be redacted")

		assert.NotEmpty(t, userMap["id"], "User ID is missing")
		assert.NotEmpty(t, dataMap["access_token"], "Access Token is missing")
		assert.Equal(t, "Bearer", dataMap["token_type"], "Token Type should be 'Bearer'")
		assert.NotEmpty(t, dataMap["expires_in"], "Expires In is missing")
		assert.NotEmpty(t, dataMap["expires_at"], "Expires At is missing")
		assert.NotEmpty(t, dataMap["refresh_token"], "Refresh Token is missing")

		*userId, _ = userMap["id"].(string)
		*accessToken, _ = dataMap["access_token"].(string)
		*refreshToken, _ = dataMap["refresh_token"].(string)
	}
}

func UpdateUserPasswordHelper(authAPI *auth.API, t testing.TB, body auth.PutPasswordBodyParams, want_code int, want_status string, accessToken string) {
	t.Helper()

	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPut, "/auth/reset-password", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	authAPI.PutPassword(res, req)

	GenericAssert(t, want_code, want_status, res)
}

func RefreshTokenHelper(authAPI *auth.API, t testing.TB, body auth.RefreshTokenBodyParams, access_token string, want_code int, want_status string) {
	t.Helper()

	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	res := httptest.NewRecorder()

	authAPI.RefreshToken(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		dataMap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.NotEmpty(t, dataMap["access_token"], "Access Token is missing")
		assert.Equal(t, "Bearer", dataMap["token_type"], "Token Type should be 'Bearer'")
		assert.NotEmpty(t, dataMap["expires_in"], "Expires In is missing")
		assert.NotEmpty(t, dataMap["expires_at"], "Expires At is missing")
		assert.NotEmpty(t, dataMap["refresh_token"], "Refresh Token is missing")
	}
}
