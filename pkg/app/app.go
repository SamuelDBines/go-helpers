package app

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"sync"
	"syscall"
)

type App interface {
	Name() string
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func Run(apps ...App) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	chanlen := len(apps)

	errCh := make(chan error, chanlen)

	var wg sync.WaitGroup
	wg.Add(chanlen)

	for _, app := range apps {

		go func() {
			defer wg.Done()
			log.Printf("starting app: %s", app.Name())
			if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
				return
			}
			log.Printf("app stopped: %s", app.Name())
		}()
	}

	select {
	case <-ctx.Done():
		stop()
	case err := <-errCh:
		log.Fatal("app failed: ", err)
	}

	for _, app := range apps {
		if err := app.Shutdown(ctx); err != nil {
			log.Println("shutdown error: ", app.Name(), err)
		}
	}
	wg.Wait()
	return nil
}
