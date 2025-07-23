package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManager struct {
	client *secretsmanager.Client
}

func (a *SecretsManager) GetSecretValue(ctx context.Context, s string) ([]byte, error) {
	value, err := a.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &s,
	})
	if err != nil {
		return nil, err
	}
	return []byte(*value.SecretString), nil
}

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

func NewSecretsManager(ctx context.Context) (*SecretsManager, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &SecretsManager{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}
