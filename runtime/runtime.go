package runtime

import (
	"context"
	"fmt"
	"github.com/fr0stylo/secretary/providers"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"
)

/*
Runtime - Provides core logic for creating secrets from user defined secrets
*/
type Runtime struct {
	Client         providers.IProvider
	Config         *Options
	PulledVersions []*Secret
	RunCancel      context.CancelFunc
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
		Client:         client,
		Config:         config,
		PulledVersions: make([]*Secret, 0),
	}
}

/*
WatchChanges - Starts a go-routine that constantly watches for changes to the user provided secret, and if changes are
found then they are created
*/
func (r *Runtime) WatchChanges(ctx context.Context) chan string {
	t := time.NewTicker(r.Config.Frequency)
	changeCh := make(chan string)
	ctx, cancel := context.WithCancel(ctx)
	r.RunCancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				found := false
				for _, secret := range r.PulledVersions {
					v, err := r.Client.GetSecretVersion(ctx, secret.Identifier)
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
StopWatchChanges - Stops the go-routine from watching for new secret changes
*/
func (r *Runtime) StopWatchChanges() {
	if r.RunCancel != nil {
		r.RunCancel()
	}
}

/*
CreateSecret - Creates a new file within the path defined in the provided secret for the wrapped process to use
*/
func (r *Runtime) CreateSecret(ctx context.Context, secret *Secret) error {
	version, err := r.Client.GetSecretVersion(ctx, secret.Identifier)
	if err != nil {
		return err
	}
	secret.Version = version
	if !slices.ContainsFunc(r.PulledVersions, func(s *Secret) bool {
		return s.Identifier == secret.Identifier
	}) {
		r.PulledVersions = append(r.PulledVersions, secret)
	}
	log.Printf(
		"Creating secret %s (version %s) at %s",
		secret.Identifier,
		secret.Version,
		secret.Path,
	)

	retrievedSecret, err := r.Client.GetSecretValue(ctx, secret.Identifier)
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
ExecuteProgram - Begins execution of the user provided program and passes signals from secretary to the wrapped
process
*/
func (r *Runtime) ExecuteProgram(ctx context.Context, changeCh chan string, args []string) error {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return err
	}

	complete := make(chan error)
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		complete <- cmd.Wait()
	}()

	for {
		select {
		case change := <-changeCh:
			log.Printf("Change detected: %s, Sending SIGHUP to %d", change, cmd.Process.Pid)
			if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
				return err
			}
		case <-signalCh:
			log.Printf("Received signal, sending SIGKILL to %d", cmd.Process.Pid)
			if err := cmd.Process.Signal(syscall.SIGKILL); err != nil {
				return err
			}
			return nil
		case err := <-complete:
			return err
		}
	}
}
