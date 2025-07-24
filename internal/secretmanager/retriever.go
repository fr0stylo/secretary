// Package secretmanager provides interfaces and implementations for secret management.
package secretmanager

import (
	"context"
	"log"
	"os"
	"path"
	"slices"
	"strings"
)

// Retriever manages the retrieval and monitoring of secrets.
type Retriever struct {
	client         Client
	config         *Config
	pulledVersions []*Secret
	runCancel      context.CancelFunc
}

// NewRetriever creates a new SecretRetriever with the given client and options.
func NewRetriever(client Client, opts ...ConfigOption) *Retriever {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return &Retriever{
		client:         client,
		config:         config,
		pulledVersions: make([]*Secret, 0),
	}
}

// CreateSecretsFromEnvironment creates secrets from environment variables with the SECRETARY_ prefix.
func (r *Retriever) CreateSecretsFromEnvironment(ctx context.Context, envSecrets []string) error {
	for _, envSecret := range envSecrets {
		if !strings.HasPrefix(envSecret, "SECRETARY_") {
			continue
		}
		str := strings.SplitN(envSecret, "=", 2)
		if len(str) != 2 {
			log.Printf("invalid secret name: %s", envSecret)
			continue
		}
		secretName := strings.TrimPrefix(str[0], "SECRETARY_")
		secretPath := path.Join(r.config.Path, secretName)
		secretIdentifier := str[1]

		s := &Secret{
			Identifier: secretIdentifier,
			EnvName:    secretName,
			Version:    "",
			Path:       secretPath,
		}
		if err := r.CreateSecret(ctx, s); err != nil {
			return err
		}
		if err := os.Unsetenv(str[0]); err != nil {
			return err
		}
	}
	return nil
}

// Clean removes all secret files and unsets related environment variables.
// This should be called when the application is shutting down to ensure secrets are not left on disk.
func (r *Retriever) Clean() error {
	for _, secret := range r.pulledVersions {
		if err := os.Remove(secret.Path); err != nil {
			log.Printf("error removing secret file %s: %v", secret.Path, err)
		}
		if err := os.Unsetenv(secret.EnvName); err != nil {
			log.Printf("error unsetting environment variable %s: %v", secret.EnvName, err)
		}
	}
	return nil
}

// CreateSecret creates a secret file and sets an environment variable pointing to it.
func (r *Retriever) CreateSecret(ctx context.Context, secret *Secret) error {
	tctx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()
	version, err := r.client.GetSecretVersion(tctx, secret.Identifier)
	if err != nil {
		return err
	}
	secret.Version = version
	if !slices.ContainsFunc(r.pulledVersions, func(s *Secret) bool {
		return s.Identifier == secret.Identifier
	}) {
		r.pulledVersions = append(r.pulledVersions, secret)
	}
	log.Printf(
		"Creating secret %s (version %s) at %s",
		secret.Identifier,
		secret.Version,
		secret.Path,
	)

	retrievedSecret, err := r.client.GetSecretValue(tctx, secret.Identifier)
	if err != nil {
		return err
	}
	f, err := os.Create(secret.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(retrievedSecret)
	if err != nil {
		return err
	}

	return os.Setenv(secret.EnvName, secret.Path)
}
