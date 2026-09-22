package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	defaultDatabaseSSLMode        = "disable"
	defaultDatabaseConnectTimeout = 5 * time.Second
)

// Database contains the PostgreSQL connection configuration.
type Database struct {
	Name           string
	User           string
	Password       string
	Host           string
	Port           uint16
	SSLMode        string
	ConnectTimeout time.Duration
}

// LoadDatabase reads the local .env file when present and loads PostgreSQL
// configuration from environment variables.
func LoadDatabase() (Database, error) {
	if err := loadEnvironmentFile(); err != nil {
		return Database{}, err
	}

	port, err := loadPort("DATABASE_PORT")
	if err != nil {
		return Database{}, err
	}

	connectTimeout, err := loadDuration(
		"DATABASE_CONNECT_TIMEOUT",
		defaultDatabaseConnectTimeout,
	)
	if err != nil {
		return Database{}, err
	}

	database := Database{
		Name:           os.Getenv("DATABASE_NAME"),
		User:           os.Getenv("DATABASE_USER"),
		Password:       os.Getenv("DATABASE_PASSWORD"),
		Host:           os.Getenv("DATABASE_HOST"),
		Port:           port,
		SSLMode:        valueOrDefault("DATABASE_SSL_MODE", defaultDatabaseSSLMode),
		ConnectTimeout: connectTimeout,
	}

	if err := database.Validate(); err != nil {
		return Database{}, err
	}

	return database, nil
}

// Validate checks whether all required PostgreSQL settings are present.
func (d Database) Validate() error {
	var validationErrors []error

	for name, value := range map[string]string{
		"DATABASE_NAME":     d.Name,
		"DATABASE_USER":     d.User,
		"DATABASE_PASSWORD": d.Password,
		"DATABASE_HOST":     d.Host,
	} {
		if value == "" {
			validationErrors = append(validationErrors, fmt.Errorf("%s is required", name))
		}
	}

	if d.Port == 0 {
		validationErrors = append(validationErrors, errors.New("DATABASE_PORT must be greater than zero"))
	}

	if d.ConnectTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("DATABASE_CONNECT_TIMEOUT must be positive"))
	}

	return errors.Join(validationErrors...)
}

// DSN returns a PostgreSQL connection string without logging it.
func (d Database) DSN() string {
	query := url.Values{}
	query.Set("sslmode", d.SSLMode)

	databaseURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     net.JoinHostPort(d.Host, strconv.FormatUint(uint64(d.Port), 10)),
		Path:     d.Name,
		RawQuery: query.Encode(),
	}

	return databaseURL.String()
}

func loadPort(name string) (uint16, error) {
	value := os.Getenv(name)
	if value == "" {
		return 0, fmt.Errorf("%s is required", name)
	}

	port, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}

	return uint16(port), nil
}

func loadDuration(name string, fallback time.Duration) (time.Duration, error) {
	value := valueOrDefault(name, fallback.String())

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}

	return duration, nil
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}

	return fallback
}
