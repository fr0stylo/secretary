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

/*
Runtime - Provides core logic for creating secrets from user defined secrets
*/
type Runtime struct {
	client         providers.IProvider
	config         *Options
	pulledVersions []*Secret
	runCancel      context.CancelFunc
}

/*
NewRuntime - Initializes a new Runtime structure with default values and returns a pointer to it
*/
func NewRuntime(client providers.IProvider, opts ...SecretRetrieverOpts) *Runtime {
	config := DefaultOptions()
	for _, opt := range opts {
		opt(config)
	}
	return &Runtime{
		client:         client,
		config:         config,
		pulledVersions: make([]*Secret, 0),
	}
}

/*
Run - Begins runtime execution and constantly watches for secret changes on a separate go-routine until its provided
context has been cancelled
*/
func (r *Runtime) Run(ctx context.Context) chan string {
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

/*
Stop - Stop Secretary runtime execution
*/
func (r *Runtime) Stop() {
	if r.runCancel != nil {
		r.runCancel()
	}
}

/*
CreateSecretsFromEnvironment - Creates new mounted secret files for secrets declared in user defined environmental
variables
*/
func (r *Runtime) CreateSecretsFromEnvironment(ctx context.Context, envSecrets []string) error {
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

/*
CreateSecret - Creates a new file within the path defined in the provided secret for the wrapped process to use
*/
func (r *Runtime) CreateSecret(ctx context.Context, secret *Secret) error {
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
