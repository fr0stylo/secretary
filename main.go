package main

import (
	"context"
	"github.com/fr0stylo/secretary/providers/aws"
	"github.com/fr0stylo/secretary/runtime"
	"log"
	"os"
	"time"
)

func main() {
	if len(os.Args) == 1 {
		log.Printf("Missing program to execute. Provide this in the first argument")
		os.Exit(1)
	}

	ctx, _ := context.WithCancel(context.Background())

	sm, err := aws.NewSecretsManager(ctx)
	if err != nil {
		log.Fatal(err)
	}
	sc := runtime.NewRuntime(sm, runtime.WithFrequency(15*time.Second))
	if err := sc.CreateSecretsFromEnvironment(ctx, os.Environ()); err != nil {
		log.Fatal(err)
	}

	changeCh := sc.WatchChanges(ctx)
	defer sc.StopWatchChanges()

	if err := sc.ExecuteProgram(ctx, changeCh, os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
