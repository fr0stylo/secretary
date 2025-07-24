package providers

import (
	"context"
	"strings"

	"github.com/fr0stylo/secretary/internal/providers/aws"
	"github.com/fr0stylo/secretary/internal/providers/dummy"
)

type Mux struct {
}

func (m Mux) GetSecretValue(ctx context.Context, id string) ([]byte, error) {
	if strings.HasPrefix(id, "arn:aws:ssm:") {
		e, err := aws.NewSSM(ctx)
		if err != nil {
			return nil, err
		}
		return e.GetSecretValue(ctx, id)
	}

	if strings.HasPrefix(id, "arn:aws:secretsmanager:") {
		e, err := aws.NewSecretsManager(ctx)
		if err != nil {
			return nil, err
		}
		return e.GetSecretValue(ctx, id)
	}

	d := dummy.NewSecretManager()
	return d.GetSecretValue(ctx, id)
}

func (m Mux) GetSecretVersion(ctx context.Context, id string) (string, error) {
	if strings.HasPrefix(id, "arn:aws:ssm:") {
		e, err := aws.NewSSM(ctx)
		if err != nil {
			return "", err
		}
		return e.GetSecretVersion(ctx, id)
	}

	if strings.HasPrefix(id, "arn:aws:secretsmanager:") {
		e, err := aws.NewSecretsManager(ctx)
		if err != nil {
			return "", err
		}
		return e.GetSecretVersion(ctx, id)
	}

	d := dummy.NewSecretManager()
	return d.GetSecretVersion(ctx, id)
}

func NewMux() *Mux {
	return &Mux{}
}
