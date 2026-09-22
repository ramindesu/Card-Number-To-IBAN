package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	defaultHTTPReadTimeout     = 10 * time.Second
	defaultHTTPWriteTimeout    = 15 * time.Second
	defaultHTTPIdleTimeout     = 60 * time.Second
	defaultHTTPShutdownTimeout = 10 * time.Second
)

// HTTPServer contains the public HTTP server settings.
type HTTPServer struct {
	Host            string
	Port            uint16
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// LoadHTTPServer loads and validates HTTP server settings from the environment.
func LoadHTTPServer() (HTTPServer, error) {
	if err := loadEnvironmentFile(); err != nil {
		return HTTPServer{}, err
	}

	port, err := loadPort("HTTP_PORT")
	if err != nil {
		return HTTPServer{}, err
	}

	readTimeout, err := loadDuration("HTTP_READ_TIMEOUT", defaultHTTPReadTimeout)
	if err != nil {
		return HTTPServer{}, err
	}

	writeTimeout, err := loadDuration("HTTP_WRITE_TIMEOUT", defaultHTTPWriteTimeout)
	if err != nil {
		return HTTPServer{}, err
	}

	idleTimeout, err := loadDuration("HTTP_IDLE_TIMEOUT", defaultHTTPIdleTimeout)
	if err != nil {
		return HTTPServer{}, err
	}

	shutdownTimeout, err := loadDuration(
		"SHUTDOWN_TIMEOUT",
		defaultHTTPShutdownTimeout,
	)
	if err != nil {
		return HTTPServer{}, err
	}

	configuration := HTTPServer{
		Host:            os.Getenv("HTTP_HOST"),
		Port:            port,
		ReadTimeout:     readTimeout,
		WriteTimeout:    writeTimeout,
		IdleTimeout:     idleTimeout,
		ShutdownTimeout: shutdownTimeout,
	}
	if err := configuration.Validate(); err != nil {
		return HTTPServer{}, err
	}

	return configuration, nil
}

// Validate checks whether HTTP server settings are safe and complete.
func (h HTTPServer) Validate() error {
	var validationErrors []error

	if h.Host == "" {
		validationErrors = append(validationErrors, errors.New("HTTP_HOST is required"))
	} else if net.ParseIP(h.Host) == nil && h.Host != "localhost" {
		validationErrors = append(
			validationErrors,
			errors.New("HTTP_HOST must be an IP address or localhost"),
		)
	}

	if h.Port == 0 {
		validationErrors = append(validationErrors, errors.New("HTTP_PORT must be greater than zero"))
	}

	for _, timeout := range []struct {
		name  string
		value time.Duration
	}{
		{name: "HTTP_READ_TIMEOUT", value: h.ReadTimeout},
		{name: "HTTP_WRITE_TIMEOUT", value: h.WriteTimeout},
		{name: "HTTP_IDLE_TIMEOUT", value: h.IdleTimeout},
		{name: "SHUTDOWN_TIMEOUT", value: h.ShutdownTimeout},
	} {
		if timeout.value <= 0 {
			validationErrors = append(
				validationErrors,
				fmt.Errorf("%s must be positive", timeout.name),
			)
		}
	}

	return errors.Join(validationErrors...)
}

// Address returns the host:port value accepted by net/http.
func (h HTTPServer) Address() string {
	return net.JoinHostPort(h.Host, strconv.FormatUint(uint64(h.Port), 10))
}
