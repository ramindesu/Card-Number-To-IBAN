package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"
)

const defaultZarinHubRequestTimeout = 10 * time.Second

// ZarinHub contains the external API connection and authentication settings.
type ZarinHub struct {
	BaseURL             string
	Username            string
	Password            string
	RequestTimeout      time.Duration
	RequestHandlingType string
}

// LoadZarinHub loads ZarinHub settings from the environment.
func LoadZarinHub() (ZarinHub, error) {
	if err := loadEnvironmentFile(); err != nil {
		return ZarinHub{}, err
	}

	requestTimeout, err := loadDuration(
		"ZARINHUB_REQUEST_TIMEOUT",
		defaultZarinHubRequestTimeout,
	)
	if err != nil {
		return ZarinHub{}, err
	}

	configuration := ZarinHub{
		BaseURL:             os.Getenv("ZARINHUB_BASE_URL"),
		Username:            os.Getenv("ZARINHUB_API_USERNAME"),
		Password:            os.Getenv("ZARINHUB_API_PASSWORD"),
		RequestTimeout:      requestTimeout,
		RequestHandlingType: valueOrDefault("ZARINHUB_REQUEST_HANDLING_TYPE", "standard"),
	}

	if err := configuration.Validate(); err != nil {
		return ZarinHub{}, err
	}

	return configuration, nil
}

// Validate checks whether the ZarinHub settings are safe and complete.
func (z ZarinHub) Validate() error {
	var validationErrors []error

	for _, required := range []struct {
		name  string
		value string
	}{
		{name: "ZARINHUB_BASE_URL", value: z.BaseURL},
		{name: "ZARINHUB_API_USERNAME", value: z.Username},
		{name: "ZARINHUB_API_PASSWORD", value: z.Password},
	} {
		if required.value == "" {
			validationErrors = append(
				validationErrors,
				fmt.Errorf("%s is required", required.name),
			)
		}
	}

	parsedURL, err := url.Parse(z.BaseURL)
	if err != nil ||
		(parsedURL.Scheme != "http" && parsedURL.Scheme != "https") ||
		parsedURL.Host == "" {
		validationErrors = append(validationErrors, errors.New("ZARINHUB_BASE_URL must be a valid URL"))
	}

	if z.RequestTimeout <= 0 {
		validationErrors = append(validationErrors, errors.New("ZARINHUB_REQUEST_TIMEOUT must be positive"))
	}

	if z.RequestHandlingType != "standard" && z.RequestHandlingType != "advanced" {
		validationErrors = append(
			validationErrors,
			errors.New("ZARINHUB_REQUEST_HANDLING_TYPE must be standard or advanced"),
		)
	}

	return errors.Join(validationErrors...)
}
