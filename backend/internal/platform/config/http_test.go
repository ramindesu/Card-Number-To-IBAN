package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ramindesu/Card-Number-To-IBAN/backend/internal/platform/config"
)

func TestHTTPServerValidateAndAddress(t *testing.T) {
	t.Parallel()

	configuration := config.HTTPServer{
		Host:            "127.0.0.1",
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     time.Minute,
		ShutdownTimeout: 10 * time.Second,
	}

	require.NoError(t, configuration.Validate())
	assert.Equal(t, "127.0.0.1:8080", configuration.Address())
}

func TestHTTPServerValidateRejectsUnsafeConfiguration(t *testing.T) {
	t.Parallel()

	err := (config.HTTPServer{}).Validate()

	require.Error(t, err)
	assert.ErrorContains(t, err, "HTTP_HOST")
	assert.ErrorContains(t, err, "HTTP_PORT")
	assert.ErrorContains(t, err, "HTTP_READ_TIMEOUT")
	assert.ErrorContains(t, err, "HTTP_WRITE_TIMEOUT")
	assert.ErrorContains(t, err, "HTTP_IDLE_TIMEOUT")
	assert.ErrorContains(t, err, "SHUTDOWN_TIMEOUT")
}

func TestCardProtectionRequiresLongFingerprintKey(t *testing.T) {
	t.Parallel()

	assert.Error(t, (config.CardProtection{FingerprintKey: []byte("too-short")}).Validate())
	assert.NoError(t, (config.CardProtection{
		FingerprintKey: []byte("0123456789abcdef0123456789abcdef"),
	}).Validate())
}
