// Package aws provides AWS-specific implementations of secret management interfaces.
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// Ssm implements the secretmanager.Client interface for AWS Systems Manager Parameter Store.
type Ssm struct {
	client *ssm.Client
}

// GetSecretValue retrieves the value of a secret from AWS Systems Manager Parameter Store.
// It takes a context and a parameter ID, and returns the parameter value as a byte slice.
func (s Ssm) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	p, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: &id,
	})
	if err != nil {
		return nil, err
	}

	return []byte(*p.Parameter.Value), nil
}

// GetSecretVersion retrieves the current version of a parameter from AWS Systems Manager Parameter Store.
// It takes a context and a parameter ID, and returns the parameter version as a string.
func (s Ssm) GetSecretVersion(ctx context.Context, id string) (string, error) {
	p, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: &id,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", p.Parameter.Version), nil
}

// NewSSM creates a new AWS Systems Manager Parameter Store client.
// NewSSM creates a new AWS Systems Manager Parameter Store client using the default AWS configuration.
// It returns a pointer to an Ssm instance or an error if configuration loading fails.
func NewSSM(ctx context.Context) (*Ssm, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &Ssm{
		client: ssm.NewFromConfig(cfg),
	}, nil
}
