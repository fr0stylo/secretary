// Package secretmanager provides interfaces and implementations for secret management.
package secretmanager

import (
	"context"
	"log"
	"time"
)

// Watcher monitors secrets for changes and triggers updates when they change.
// It periodically checks the version of each secret and recreates it if the version has changed.
type Watcher struct {
	r      *Retriever
	cancel context.CancelFunc
}

// NewWatcher creates a new Watcher with the given Retriever.
// The Watcher will use the Retriever to check for secret changes and update them.
func NewWatcher(retriever *Retriever) *Watcher {
	return &Watcher{r: retriever}
}

// Start begins watching for secret changes at the frequency specified in the Retriever's config.
// It returns a channel that will receive a timestamp string whenever a secret changes.
// The context can be used to stop the watcher, or the Stop method can be called.
func (w *Watcher) Start(ctx context.Context) chan string {
	t := time.NewTicker(w.r.config.Frequency)
	changeCh := make(chan string)
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				found := false
				for _, secret := range w.r.pulledVersions {
					v, err := w.r.client.GetSecretVersion(ctx, secret.Identifier)
					if err != nil {
						log.Printf("Error retrieving secret version: %s", err)
						continue
					}
					if v == secret.Version {
						continue
					}
					log.Printf("Secret %s changed, recreating", secret.Identifier)
					found = true
					if err := w.r.CreateSecret(ctx, secret); err != nil {
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

// Stop halts the watcher's goroutine.
// This should be called when the watcher is no longer needed to prevent resource leaks.
func (w *Watcher) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}
