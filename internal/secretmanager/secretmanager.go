// Package secretmanager provides interfaces and implementations for secret management.
package secretmanager

import (
	"context"
	"time"
)

// Client defines the interface for retrieving secrets from a secret management service.
type Client interface {
	// GetSecretValue retrieves the value of a secret by its identifier.
	GetSecretValue(ctx context.Context, id string) ([]byte, error)

	// GetSecretVersion retrieves the current version of a secret by its identifier.
	GetSecretVersion(ctx context.Context, id string) (string, error)
}

// Config holds configuration options for the SecretRetriever.
type Config struct {
	Frequency time.Duration
	Timeout   time.Duration
	Path      string
}

// ConfigOption is a function that modifies Config.
type ConfigOption func(*Config)

// WithFrequency sets the frequency at which secrets are checked for updates.
func WithFrequency(frequency time.Duration) ConfigOption {
	return func(config *Config) {
		config.Frequency = frequency
	}
}

// WithTimeout sets the timeout for secret retrieval operations.
func WithTimeout(timeout time.Duration) ConfigOption {
	return func(config *Config) {
		config.Timeout = timeout
	}
}

func WithPath(path string) ConfigOption {
	return func(config *Config) {
		config.Path = path
	}
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Frequency: 15 * time.Second,
		Timeout:   10 * time.Second,
		Path:      "/tmp",
	}
}

// Secret represents a secret that has been retrieved and stored.
type Secret struct {
	Identifier string
	EnvName    string
	Version    string
	Path       string
}
