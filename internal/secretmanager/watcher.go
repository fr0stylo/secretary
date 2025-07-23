package secretmanager

import (
	"context"
	"log"
	"time"
)

type Watcher struct {
	r      *Retriever
	cancel context.CancelFunc
}

func NewWatcher(retriever *Retriever) *Watcher {
	return &Watcher{r: retriever}
}

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

func (w *Watcher) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}
