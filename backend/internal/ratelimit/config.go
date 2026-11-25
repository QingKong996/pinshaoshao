package ratelimit

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	defaultWindow               = time.Minute
	defaultAPITotalLimit        = 3000
	defaultUnauthenticatedLimit = 60
	defaultAPILimit             = 60
	defaultSubjectLimit         = 300
)

// Config controls all rate limit layers.
type Config struct {
	Window               time.Duration
	APITotalLimit        int
	UnauthenticatedLimit int
	APILimit             int
	SubjectLimit         int
}

// ConfigFromEnv loads rate limit configuration from environment variables.
func ConfigFromEnv() (Config, error) {
	config := Config{
		Window:               defaultWindow,
		APITotalLimit:        defaultAPITotalLimit,
		UnauthenticatedLimit: defaultUnauthenticatedLimit,
		APILimit:             defaultAPILimit,
		SubjectLimit:         defaultSubjectLimit,
	}

	if value := os.Getenv("RATE_LIMIT_WINDOW"); value != "" {
		window, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("parse RATE_LIMIT_WINDOW: %w", err)
		}
		config.Window = window
	}

	var err error
	if config.APITotalLimit, err = positiveIntEnv("RATE_LIMIT_API_TOTAL_LIMIT", defaultAPITotalLimit); err != nil {
		return Config{}, err
	}
	if config.UnauthenticatedLimit, err = positiveIntEnv(
		"RATE_LIMIT_UNAUTHENTICATED_LIMIT",
		defaultUnauthenticatedLimit,
	); err != nil {
		return Config{}, err
	}
	if config.APILimit, err = positiveIntEnv("RATE_LIMIT_API_LIMIT", defaultAPILimit); err != nil {
		return Config{}, err
	}
	if config.SubjectLimit, err = positiveIntEnv("RATE_LIMIT_SUBJECT_LIMIT", defaultSubjectLimit); err != nil {
		return Config{}, err
	}
	if err := config.validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) validate() error {
	if c.Window < time.Millisecond {
		return fmt.Errorf("RATE_LIMIT_WINDOW must be at least 1ms")
	}
	if c.APITotalLimit <= 0 {
		return fmt.Errorf("RATE_LIMIT_API_TOTAL_LIMIT must be positive")
	}
	if c.UnauthenticatedLimit <= 0 {
		return fmt.Errorf("RATE_LIMIT_UNAUTHENTICATED_LIMIT must be positive")
	}
	if c.APILimit <= 0 {
		return fmt.Errorf("RATE_LIMIT_API_LIMIT must be positive")
	}
	if c.SubjectLimit <= 0 {
		return fmt.Errorf("RATE_LIMIT_SUBJECT_LIMIT must be positive")
	}
	return nil
}

func positiveIntEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}
