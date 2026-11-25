package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

const DefaultSessionTTL = 24 * time.Hour

type Config struct {
	Username   string
	Password   string
	SessionTTL time.Duration
}

func ConfigFromEnv(prefix string) Config {
	prefix = strings.TrimSpace(prefix)
	config := Config{
		Username: strings.TrimSpace(os.Getenv(prefix + "_AUTH_USERNAME")),
		Password: os.Getenv(prefix + "_AUTH_PASSWORD"),
	}
	rawTTL := strings.TrimSpace(os.Getenv(prefix + "_AUTH_SESSION_TTL"))
	if rawTTL == "" {
		return config
	}

	ttl, err := time.ParseDuration(rawTTL)
	if err != nil || ttl <= 0 {
		slog.Warn(
			"ignore invalid mock auth session TTL",
			"environment", prefix+"_AUTH_SESSION_TTL",
			"value", rawTTL,
		)
		return config
	}
	config.SessionTTL = ttl
	return config
}

func (c Config) withDefaults() Config {
	c.Username = strings.TrimSpace(c.Username)
	if c.SessionTTL <= 0 {
		c.SessionTTL = DefaultSessionTTL
	}
	return c
}

func (c Config) validate() error {
	if c.Username == "" {
		return errors.New("mock auth username is required")
	}
	if c.Password == "" {
		return errors.New("mock auth password is required")
	}
	if len(c.Username) > maxUsernameBytes {
		return fmt.Errorf(
			"mock auth username must be at most %d bytes",
			maxUsernameBytes,
		)
	}
	if len(c.Password) < minPasswordBytes || len(c.Password) > maxPasswordBytes {
		return fmt.Errorf(
			"mock auth password must be between %d and %d bytes",
			minPasswordBytes,
			maxPasswordBytes,
		)
	}
	return nil
}
