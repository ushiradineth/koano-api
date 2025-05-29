package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ushiradineth/koano-api/api/resource/health"
)

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	res := httptest.NewRecorder()

	health.Health(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "healthy", res.Body.String())
}
