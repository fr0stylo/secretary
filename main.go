package main

import (
	"context"
	"github.com/fr0stylo/secretary/providers/aws"
	"github.com/fr0stylo/secretary/runtime"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

	sm, err := aws.NewSecretsManager(ctx)
	if err != nil {
		log.Fatal(err)
	}
	sc := NewSecretRetriever(sm, runtime.WithFrequency(15*time.Second))
	if err := sc.CreateSecretsFromEnvironment(ctx, os.Environ()); err != nil {
		log.Fatal(err)
	}

	changeCh := sc.Run(ctx)
	defer sc.Stop()

	if err := runApplication(ctx, changeCh, os.Args[1:]); err != nil {
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
