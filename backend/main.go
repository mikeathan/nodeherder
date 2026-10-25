package main

import (
	"context"
	"errors"
	"fmt"
	"node-herder/internal/integrations"
	"time"

	hub "node-herder/internal"
	"node-herder/store"
	"node-herder/utils"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	args := readArgs()
	utils.InitFileLogger()
	utils.SetLogLevel(args.logLevel)
	if err := run(args.port); err != nil {
		utils.LogErrorf("startup or shutdown failed: %v", err)
		os.Exit(1)
	}
}

func run(port int) error {
	if err := utils.LoadEnviromentConfig(); err != nil {
		return fmt.Errorf("load environment: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runServer(ctx, port)
}

type httpListener interface{ Listen() error }

func runServer(ctx context.Context, port int) error {
	return runWithConnections(ctx, integrations.Bootstrap, func(ctx context.Context, connections *integrations.Service) (httpListener, error) {
		appStore, err := store.Create(ctx)
		if err != nil {
			return nil, fmt.Errorf("create store: %w", err)
		}
		h, err := hub.Register(port, appStore, ctx, hub.WithConnections(connections))
		if err != nil {
			return nil, fmt.Errorf("register server: %w", err)
		}
		return h, nil
	})
}

// Keep feature cleanup outside store/hub construction so every failure and listener exit joins it.
func runWithConnections(ctx context.Context, bootstrap func(context.Context) *integrations.Service, start func(context.Context, *integrations.Service) (httpListener, error)) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	connections := bootstrap(ctx)
	defer func() {
		cancel()
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer closeCancel()
		if closeErr := connections.Close(closeCtx); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	listener, err := start(ctx, connections)
	if err != nil {
		return err
	}
	return listener.Listen()
}
