package providers

import "context"

/*
IProvider - A shared interface for all secret provider clients to implement
*/
type IProvider interface {
	GetSecretValue(ctx context.Context, name string) ([]byte, error)
	GetSecretVersion(context.Context, string) (string, error)
}
