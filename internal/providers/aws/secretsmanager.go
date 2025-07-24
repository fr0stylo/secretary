// Package aws provides AWS-specific implementations of secret management interfaces.
package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManager implements the secretmanager.Client interface for AWS Secrets Manager.
type SecretsManager struct {
	client *secretsmanager.Client
}

// GetSecretValue retrieves the value of a secret from AWS Secrets Manager.
func (a *SecretsManager) GetSecretValue(ctx context.Context, s string) ([]byte, error) {
	value, err := a.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &s,
	})
	if err != nil {
		return nil, err
	}
	return []byte(*value.SecretString), nil
}

// GetSecretVersion retrieves the current version of a secret from AWS Secrets Manager.
func (a *SecretsManager) GetSecretVersion(ctx context.Context, s string) (string, error) {
	value, err := a.client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: &s,
	})
	if err != nil {
		return "", err
	}

	for k, v := range value.VersionIdsToStages {
		for _, stage := range v {
			if stage == "AWSCURRENT" {
				return k, nil
			}
		}
	}
	return "", fmt.Errorf("no current version found")
}

// NewSecretsManager creates a new AWS Secrets Manager client.
func NewSecretsManager(ctx context.Context) (*SecretsManager, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &SecretsManager{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}
