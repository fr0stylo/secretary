// Package main provides the entry point for the secretary application.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/fr0stylo/secretary/internal/providers/aws"
	"github.com/fr0stylo/secretary/internal/providers/dummy"
	"github.com/fr0stylo/secretary/internal/secretmanager"
)

var (
	provider = flag.String("provider", "aws", "The secret provider to use")
)

func main() {
	flag.Parse()
	ctx, _ := context.WithCancel(context.Background())

	var client secretmanager.Client
	switch *provider {
	case "aws":
		sm, err := aws.NewSecretsManager(ctx)
		if err != nil {
			log.Fatal(err)
		}
		client = sm
	case "dummy":
		client = dummy.NewSecretManager()
	}

	sc := secretmanager.NewRetriever(client, secretmanager.WithFrequency(15*time.Second))
	if err := sc.CreateSecretsFromEnvironment(ctx, os.Environ()); err != nil {
		log.Fatal(err)
	}

	watcher := secretmanager.NewWatcher(sc)

	changeCh := watcher.Start(ctx)
	defer watcher.Stop()

	if err := runApplication(ctx, changeCh, flag.Args()); err != nil {
		log.Fatal(err)
	}
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
