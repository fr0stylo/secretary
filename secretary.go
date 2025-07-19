package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())
	changeCh := make(chan string)

	if err := fetchSecrets(ctx); err != nil {
		log.Fatal(err)
	}

	if err := runApplication(ctx, changeCh, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func fetchSecrets(ctx context.Context) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	sm := secretsmanager.NewFromConfig(cfg)

	env := os.Environ()
	for _, e := range env {
		if !strings.HasPrefix(e, "SECRETARY_") {
			continue
		}
		str := strings.SplitN(e, "=", 2)
		if len(str) != 2 {
			continue
		}
		secretName := strings.Trim(str[0], "SECRETARY_")
		secretPath := fmt.Sprintf("/tmp/%s", secretName)
		secretArn := str[1]
		if !strings.HasPrefix(secretArn, "arn:aws:secretsmanager:") {
			log.Printf("Skipping %s=%s, not an ARN", secretName, secretArn)
		}

		secret, err := sm.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
			SecretId: &secretArn,
		})
		if err != nil {
			return err
		}
		f, err := os.Create(secretPath)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(*secret.SecretString))
		f.Close()

		if err := os.Unsetenv(str[0]); err != nil {
			return err
		}
		return os.Setenv(secretName, secretPath)
	}

	return nil
}

func runApplication(ctx context.Context, changeCh chan string, args []string) error {
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
