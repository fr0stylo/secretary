// Package dummy provides a mock implementation of secret management interfaces for testing purposes.
package dummy

import (
	"context"
	"fmt"
	"math/rand"
)

// SecretManager implements the secretmanager.Client interface with dummy values for testing.
type SecretManager struct {
	version int
}

// GetSecretValue returns a dummy secret value regardless of the provided ID.
// This is useful for testing without requiring actual secret storage.
func (s *SecretManager) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	return []byte("dummy-secret-value"), nil
}

// GetSecretVersion returns a version string and occasionally increments the version.
// It randomly increments the version (20% chance) to simulate version changes for testing.
func (s *SecretManager) GetSecretVersion(ctx context.Context, id string) (string, error) {
	if rand.Int()%5 == 0 {
		s.version++
	}
	return fmt.Sprintf("v%d", s.version), nil
}

// NewSecretManager creates a new dummy secret manager for testing purposes.
// It initializes with version 0 and returns a pointer to SecretManager.
func NewSecretManager() *SecretManager {
	return &SecretManager{
		version: 0,
	}
}
