package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"node-herder/internal/integrations"
	"node-herder/repository"
	"testing"
)

type listenerFunc func() error

func (f listenerFunc) Listen() error { return f() }
func TestRunWithConnectionsCleansUpEveryExit(t *testing.T) {
	for _, mode := range []string{"startup_failure", "listener_failure", "context_cancelled"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			dir := t.TempDir()
			keys := repository.IntegrationKeyring{ActiveKeyID: "fixture", Keys: map[string]string{"fixture": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{5}, 32))}}
			var service *integrations.Service
			sentinel := errors.New("isolated failure")
			err := runWithConnections(ctx, func(child context.Context) *integrations.Service {
				repo, e := repository.OpenIntegrationRepository(dir, keys)
				if e != nil {
					t.Fatal(e)
				}
				files, e := repository.OpenIntegrationDefinitions(t.TempDir())
				if e != nil {
					t.Fatal(e)
				}
				service, e = integrations.NewService(child, integrations.Options{Repository: repo, Definitions: files})
				if e != nil {
					t.Fatal(e)
				}
				return service
			}, func(context.Context, *integrations.Service) (httpListener, error) {
				if mode == "startup_failure" {
					return nil, sentinel
				}
				return listenerFunc(func() error {
					if mode == "context_cancelled" {
						cancel()
						return nil
					}
					return sentinel
				}), nil
			})
			if mode == "context_cancelled" && err != nil || mode != "context_cancelled" && !errors.Is(err, sentinel) {
				t.Fatal(err)
			}
			if _, e := service.Instances(); e == nil {
				t.Fatal("feature still admits reads after exit")
			}
			reopened, e := repository.OpenIntegrationRepository(dir, keys)
			if e != nil {
				t.Fatal("database handle was not released", e)
			}
			reopened.Close()
		})
	}
}
