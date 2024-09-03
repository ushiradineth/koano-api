package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/api/resource/auth"
	"github.com/ushiradineth/cron-be/api/resource/user"
)

func CreateUserHelper(userAPI *user.API, t testing.TB, body url.Values, want_code int, want_status string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, "/users", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	userAPI.Post(res, req)

	assert.Equal(t, want_code, res.Code)

	GenericAssert(t, want_code, want_status, res)
}

func AuthenticateUserHelper(authAPI *auth.API, t testing.TB, body url.Values, want_code int, want_status string, userId *string, accessToken *string, refreshToken *string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	authAPI.Authenticate(res, req)

	assert.Equal(t, want_code, res.Code)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if want_code == http.StatusOK {
		dataMap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		userMap, ok := dataMap["user"].(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, body.Get("email"), userMap["email"])
		assert.Equal(t, "redacted", userMap["password"], "Password in response should be redacted")

		assert.NotEmpty(t, userMap["id"], "User ID is missing")
		assert.NotEmpty(t, dataMap["access_token"], "Access Token is missing")
		assert.NotEmpty(t, dataMap["refresh_token"], "Refresh Token is missing")

		*userId, _ = userMap["id"].(string)
		*accessToken, _ = dataMap["access_token"].(string)
		*refreshToken, _ = dataMap["refresh_token"].(string)
	}
}

func GetUserHelper(userAPI *user.API, t testing.TB, want_code int, want_status string, user user.PostQueryParams, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/users/{user_id}", nil)
	req.SetPathValue("user_id", userId)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Get(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if want_code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, user.Name, datamap["name"])
		assert.Equal(t, user.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
	}
}

func UpdateUserHelper(userAPI *user.API, t testing.TB, body url.Values, want_code int, want_status string, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPut, "/users/{user_id}", strings.NewReader(body.Encode()))
	req.SetPathValue("user_id", userId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Put(res, req)

	GenericAssert(t, want_code, want_status, res)
}

func UpdateUserPasswordHelper(authAPI *auth.API, t testing.TB, body url.Values, want_code int, want_status string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPut, "/auth/reset-password", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	authAPI.PutPassword(res, req)

	GenericAssert(t, want_code, want_status, res)
}

func RefreshTokenHelper(authAPI *auth.API, t testing.TB, body url.Values, access_token string, want_code int, want_status string) {
	t.Helper()
	req, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", access_token))
	res := httptest.NewRecorder()

	authAPI.RefreshToken(res, req)

	GenericAssert(t, want_code, want_status, res)
}

func DeleteUserHelper(userAPI *user.API, t testing.TB, want_code int, want_status string, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, "/users/{user_id}", nil)
	req.SetPathValue("user_id", userId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Delete(res, req)

	GenericAssert(t, want_code, want_status, res)
}
