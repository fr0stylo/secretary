// Package secretmanager provides interfaces and implementations for secret management.
package secretmanager

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"
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

// Run starts the secret monitoring process and returns a channel that will receive
// notifications when secrets change.
func (r *Retriever) Run(ctx context.Context) chan string {
	t := time.NewTicker(r.config.Frequency)
	changeCh := make(chan string)
	ctx, cancel := context.WithCancel(ctx)
	r.runCancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				found := false
				for _, secret := range r.pulledVersions {
					v, err := r.client.GetSecretVersion(ctx, secret.Identifier)
					if err != nil {
						log.Printf("Error retrieving secret version: %s", err)
						continue
					}
					if v == secret.Version {
						continue
					}
					log.Printf("Secret %s changed, recreating", secret.Identifier)
					found = true
					if err := r.CreateSecret(ctx, secret); err != nil {
						log.Printf("Error creating secret: %s", err)
						continue
					}
				}
				if found {
					changeCh <- time.Now().String()
				}
			}
		}
	}()
	return changeCh
}

// Stop stops the secret monitoring process.
func (r *Retriever) Stop() {
	if r.runCancel != nil {
		r.runCancel()
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
		secretPath := fmt.Sprintf("/tmp/%s", secretName)
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

// CreateSecret creates a secret file and sets an environment variable pointing to it.
func (r *Retriever) CreateSecret(ctx context.Context, secret *Secret) error {
	version, err := r.client.GetSecretVersion(ctx, secret.Identifier)
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

	retrievedSecret, err := r.client.GetSecretValue(ctx, secret.Identifier)
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
