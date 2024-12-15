package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/api/resource/user"
)

func CreateUserHelper(userAPI *user.API, t testing.TB, body user.PostBodyParams, want_code int, want_status string) {
	t.Helper()

	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	userAPI.Post(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, body.Name, datamap["name"])
		assert.Equal(t, body.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
	}
}

func GetUserHelper(userAPI *user.API, t testing.TB, want_code int, want_status string, user user.PostBodyParams, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/users/{user_id}", nil)
	req.SetPathValue("user_id", userId)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Get(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, userId, datamap["id"])
		assert.Equal(t, user.Name, datamap["name"])
		assert.Equal(t, user.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
	}
}

func UpdateUserHelper(userAPI *user.API, t testing.TB, body user.PostBodyParams, want_code int, want_status string, userId string, accessToken string) {
	t.Helper()

	requestBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("Failed to marshal body: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPut, "/users/{user_id}", bytes.NewBuffer(requestBody))
	req.SetPathValue("user_id", userId)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Put(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		datamap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, userId, datamap["id"])
		assert.Equal(t, body.Name, datamap["name"])
		assert.Equal(t, body.Email, datamap["email"])
		assert.Equal(t, "redacted", datamap["password"], "password in response should be redacted")
	}
}

func DeleteUserHelper(userAPI *user.API, t testing.TB, want_code int, want_status string, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, "/users/{user_id}", nil)
	req.SetPathValue("user_id", userId)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	userAPI.Delete(res, req)

	GenericAssert(t, want_code, want_status, res)
}
