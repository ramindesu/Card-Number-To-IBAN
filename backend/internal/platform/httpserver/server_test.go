package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/config"
)

func TestNewConfiguresHTTPServer(t *testing.T) {
	t.Parallel()

	configuration := validHTTPConfiguration()
	server, err := New(configuration, http.HandlerFunc(Health))

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:8080", server.httpServer.Addr)
	assert.Equal(t, configuration.ReadTimeout, server.httpServer.ReadHeaderTimeout)
	assert.Equal(t, configuration.ReadTimeout, server.httpServer.ReadTimeout)
	assert.Equal(t, configuration.WriteTimeout, server.httpServer.WriteTimeout)
	assert.Equal(t, configuration.IdleTimeout, server.httpServer.IdleTimeout)
	assert.Equal(t, configuration.ShutdownTimeout, server.shutdownTimeout)
}

func TestNewRejectsMissingHandler(t *testing.T) {
	t.Parallel()

	_, err := New(validHTTPConfiguration(), nil)

	assert.EqualError(t, err, "HTTP handler is required")
}

func TestHealth(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	Health(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"ok"}`, recorder.Body.String())
}

func validHTTPConfiguration() config.HTTPServer {
	return config.HTTPServer{
		Host:            "127.0.0.1",
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     time.Minute,
		ShutdownTimeout: 10 * time.Second,
	}
}
