package providers

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/fr0stylo/secretary/internal/providers/aws"
	"github.com/fr0stylo/secretary/internal/providers/dummy"
	"github.com/fr0stylo/secretary/internal/secretmanager"
)

type Mux struct {
	providers map[string]secretmanager.Client
}

func (m *Mux) withCache(provider string, retriever func() (secretmanager.Client, error)) (secretmanager.Client, error) {
	p, ok := m.providers[provider]
	var err error
	if !ok {
		p, err = retriever()
		if err != nil {
			return nil, err
		}
		m.providers[provider] = p
	}

	return p, err
}

func (m *Mux) resolveProvider(id string) (secretmanager.Client, error) {
	// If string does not look like AWS, for now fallback to dummy
	if !strings.HasPrefix(id, "arn:aws") {
		return dummy.NewSecretManager(), nil
	}

	// AWS Support
	resource, err := arn.Parse(id)
	if err != nil {
		return nil, err
	}
	switch resource.Service {
	case "ssm":
		return m.withCache("ssm", func() (secretmanager.Client, error) {
			return aws.NewSSM(context.Background())
		})
	case "secretsmanager":
		return m.withCache("secretsmanager", func() (secretmanager.Client, error) {
			return aws.NewSecretsManager(context.Background())
		})
	}

	return nil, errors.New("unknown provider")
}

func (m *Mux) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	provider, err := m.resolveProvider(id)
	if err != nil {
		return nil, err
	}
	return provider.GetSecretValue(ctx, id)
}

func (m *Mux) GetSecretVersion(ctx context.Context, id string) (string, error) {
	provider, err := m.resolveProvider(id)
	if err != nil {
		return "", err
	}
	return provider.GetSecretVersion(ctx, id)
}

func NewMux() *Mux {
	return &Mux{
		providers: map[string]secretmanager.Client{},
	}
}
