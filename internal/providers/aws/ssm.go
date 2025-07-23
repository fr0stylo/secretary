package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Ssm struct {
	client *ssm.Client
}

func (s Ssm) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	p, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: &id,
	})
	if err != nil {
		return nil, err
	}

	return []byte(*p.Parameter.Value), nil
}

func (s Ssm) GetSecretVersion(ctx context.Context, id string) (string, error) {
	p, err := s.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name: &id,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", p.Parameter.Version), nil
}

// NewSSM creates a new SSM client.
func NewSSM(ctx context.Context) (*Ssm, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &Ssm{
		client: ssm.NewFromConfig(cfg),
	}, nil
}
