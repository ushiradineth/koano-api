package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/api/resource/event"
)

func CreateEventHelper(eventAPI *event.API, t testing.TB, body url.Values, want_code int, want_status string, event event.EventQueryParams, eventId *string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, "/events", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	eventAPI.Post(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		dataMap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.NotEmpty(t, dataMap["id"], "Event ID is missing")
		*eventId = dataMap["id"].(string)

		assert.Equal(t, event.Title, dataMap["title"])
		assert.Equal(t, event.StartTime, dataMap["start_time"])
		assert.Equal(t, event.EndTime, dataMap["end_time"])
		assert.Equal(t, event.Timezone, dataMap["timezone"])
		assert.Equal(t, event.Repeated, dataMap["repeated"])
	}
}

func GetEventHelper(eventAPI *event.API, t testing.TB, want_code int, want_status string, event event.EventQueryParams, eventId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/events/{event_id}", nil)
	req.SetPathValue("event_id", eventId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	eventAPI.Get(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		dataMap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.NotEmpty(t, dataMap["id"], "Event ID is missing")
		assert.Equal(t, event.Title, dataMap["title"])
		assert.Equal(t, event.StartTime, dataMap["start_time"])
		assert.Equal(t, event.EndTime, dataMap["end_time"])
		assert.Equal(t, event.Timezone, dataMap["timezone"])
		assert.Equal(t, event.Repeated, dataMap["repeated"])
	}
}

func UpdateEventHelper(eventAPI *event.API, t testing.TB, body url.Values, want_code int, want_status string, event event.EventQueryParams, eventId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPut, "/events/{event_id}", strings.NewReader(body.Encode()))
	req.SetPathValue("event_id", eventId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	eventAPI.Put(res, req)

	responseBody := GenericAssert(t, want_code, want_status, res)

	if res.Code == http.StatusOK {
		dataMap, ok := responseBody.Data.(map[string]interface{})
		assert.True(t, true, ok)

		assert.Equal(t, eventId, dataMap["id"])
		assert.Equal(t, event.Title, dataMap["title"])
		assert.Equal(t, event.StartTime, dataMap["start_time"])
		assert.Equal(t, event.EndTime, dataMap["end_time"])
		assert.Equal(t, event.Timezone, dataMap["timezone"])
		assert.Equal(t, event.Repeated, dataMap["repeated"])
	}
}

func DeleteEventHelper(eventAPI *event.API, t testing.TB, want_code int, want_status string, eventId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, "/events/{event_id}", nil)
	req.SetPathValue("event_id", eventId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	eventAPI.Delete(res, req)

	GenericAssert(t, want_code, want_status, res)
}

func GetUserEventsHelper(eventAPI *event.API, t testing.TB, body url.Values, want_code int, want_status string, userId string, accessToken string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/users/{user_id}/events", nil)
	req.URL.RawQuery = body.Encode()
	req.SetPathValue("user_id", userId)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	res := httptest.NewRecorder()

	eventAPI.GetUserEvents(res, req)

	GenericAssert(t, want_code, want_status, res)
}
