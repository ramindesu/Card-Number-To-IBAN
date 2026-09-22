package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const (
	defaultCardQueueCapacity           = 100
	defaultCardQueueWorkers            = 4
	defaultZarinHubRateLimitRPS        = 5
	defaultZarinHubRateLimitBurst      = 5
	defaultZarinHubMaxAttempts         = 3
	defaultZarinHubRetryBaseDelay      = 250 * time.Millisecond
	maximumConfiguredZarinHubAttempts  = 5
	maximumConfiguredRequestsPerSecond = 1000
)

// CardQueue controls upstream concurrency, rate limiting, and retry policy.
type CardQueue struct {
	Capacity          int
	Workers           int
	RequestsPerSecond int
	Burst             int
	MaxAttempts       int
	RetryBaseDelay    time.Duration
}

// LoadCardQueue loads queue and ZarinHub rate policies from the environment.
func LoadCardQueue() (CardQueue, error) {
	if err := loadEnvironmentFile(); err != nil {
		return CardQueue{}, err
	}

	capacity, err := loadPositiveInteger("CARD_QUEUE_CAPACITY", defaultCardQueueCapacity)
	if err != nil {
		return CardQueue{}, err
	}
	workers, err := loadPositiveInteger("CARD_QUEUE_WORKERS", defaultCardQueueWorkers)
	if err != nil {
		return CardQueue{}, err
	}
	requestsPerSecond, err := loadPositiveInteger(
		"ZARINHUB_RATE_LIMIT_RPS",
		defaultZarinHubRateLimitRPS,
	)
	if err != nil {
		return CardQueue{}, err
	}
	burst, err := loadPositiveInteger(
		"ZARINHUB_RATE_LIMIT_BURST",
		defaultZarinHubRateLimitBurst,
	)
	if err != nil {
		return CardQueue{}, err
	}
	maxAttempts, err := loadPositiveInteger(
		"ZARINHUB_MAX_ATTEMPTS",
		defaultZarinHubMaxAttempts,
	)
	if err != nil {
		return CardQueue{}, err
	}
	retryBaseDelay, err := loadDuration(
		"ZARINHUB_RETRY_BASE_DELAY",
		defaultZarinHubRetryBaseDelay,
	)
	if err != nil {
		return CardQueue{}, err
	}

	configuration := CardQueue{
		Capacity:          capacity,
		Workers:           workers,
		RequestsPerSecond: requestsPerSecond,
		Burst:             burst,
		MaxAttempts:       maxAttempts,
		RetryBaseDelay:    retryBaseDelay,
	}
	if err := configuration.Validate(); err != nil {
		return CardQueue{}, err
	}

	return configuration, nil
}

// Validate bounds resource usage and retry amplification.
func (c CardQueue) Validate() error {
	var validationErrors []error

	for _, value := range []struct {
		name  string
		value int
	}{
		{name: "CARD_QUEUE_CAPACITY", value: c.Capacity},
		{name: "CARD_QUEUE_WORKERS", value: c.Workers},
		{name: "ZARINHUB_RATE_LIMIT_RPS", value: c.RequestsPerSecond},
		{name: "ZARINHUB_RATE_LIMIT_BURST", value: c.Burst},
		{name: "ZARINHUB_MAX_ATTEMPTS", value: c.MaxAttempts},
	} {
		if value.value <= 0 {
			validationErrors = append(
				validationErrors,
				fmt.Errorf("%s must be positive", value.name),
			)
		}
	}

	if c.MaxAttempts > maximumConfiguredZarinHubAttempts {
		validationErrors = append(
			validationErrors,
			errors.New("ZARINHUB_MAX_ATTEMPTS must not exceed 5"),
		)
	}
	if c.RequestsPerSecond > maximumConfiguredRequestsPerSecond {
		validationErrors = append(
			validationErrors,
			errors.New("ZARINHUB_RATE_LIMIT_RPS must not exceed 1000"),
		)
	}
	if c.RetryBaseDelay <= 0 {
		validationErrors = append(
			validationErrors,
			errors.New("ZARINHUB_RETRY_BASE_DELAY must be positive"),
		)
	}

	return errors.Join(validationErrors...)
}

func loadPositiveInteger(name string, fallback int) (int, error) {
	value := valueOrDefault(name, strconv.Itoa(fallback))
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	if parsedValue <= 0 {
		return 0, fmt.Errorf("%s must be positive", name)
	}

	return parsedValue, nil
}
