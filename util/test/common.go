package test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/cron-be/util/response"
)

func GenericAssert(t testing.TB, want_code int, want_status string, res *httptest.ResponseRecorder) response.Response {
	assert.Equal(t, want_code, res.Code)

	var responseBody response.Response
	err := json.NewDecoder(res.Body).Decode(&responseBody)
	assert.NoError(t, err)

	assert.Equal(t, want_code, responseBody.Code)
	assert.Equal(t, want_status, responseBody.Status)

	return responseBody
}
