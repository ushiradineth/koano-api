package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/api/resource/user"
)

func CreateUserHelper(userAPI *user.API, t testing.TB, body url.Values, want_code int, want_status string, user user.PostQueryParams) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, "/users", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	userAPI.Post(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, user.Name, datamap["name"])
		assert.Equal(t, user.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
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

	if res.Code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, user.Name, datamap["name"])
		assert.Equal(t, user.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
		assert.Equal(t, userId, datamap["id"])
	}
}

func UpdateUserHelper(userAPI *user.API, t testing.TB, body url.Values, want_code int, want_status string, user user.PostQueryParams, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPut, "/users/{user_id}", strings.NewReader(body.Encode()))
	req.SetPathValue("user_id", userId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Put(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, user.Name, datamap["name"])
		assert.Equal(t, user.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
		assert.Equal(t, userId, datamap["id"])
	}
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
