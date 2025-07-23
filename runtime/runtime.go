package runtime

import (
	"context"
	"fmt"
	"github.com/fr0stylo/secretary/providers"
	"log"
	"os"
	"slices"
	"strings"
	"time"
)

type SecretRetriever struct {
	client         providers.IProvider
	config         *Options
	pulledVersions []*Secret
	runCancel      context.CancelFunc
}

func NewSecretRetriever(client providers.IProvider, opts ...SecretRetrieverOpts) *SecretRetriever {
	config := DefaultOptions()
	for _, opt := range opts {
		opt(config)
	}
	return &SecretRetriever{
		client:         client,
		config:         config,
		pulledVersions: make([]*Secret, 0),
	}
}

func (r *SecretRetriever) Run(ctx context.Context) chan string {
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

func (r *SecretRetriever) Stop() {
	if r.runCancel != nil {
		r.runCancel()
	}
}

func (r *SecretRetriever) CreateSecretsFromEnvironment(ctx context.Context, envSecrets []string) error {
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

func (r *SecretRetriever) CreateSecret(ctx context.Context, secret *Secret) error {
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
