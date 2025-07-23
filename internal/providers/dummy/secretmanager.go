package dummy

import (
	"context"
	"fmt"
	"math/rand"
)

type SecretManager struct {
	version int
}

func (s *SecretManager) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	return []byte("dummy-secret-value"), nil
}

func (s *SecretManager) GetSecretVersion(ctx context.Context, id string) (string, error) {
	if rand.Int()%5 == 0 {
		s.version++
	}
	return fmt.Sprintf("v%d", s.version), nil
}

func NewSecretManager() *SecretManager {
	return &SecretManager{
		version: 0,
	}
}
