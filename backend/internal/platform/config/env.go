package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func loadEnvironmentFile() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	return nil
}
