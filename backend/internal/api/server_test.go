package api

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestApiServerReturnsOnListenFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := NewHttpServer(-1, WithContext(ctx))
	done := make(chan struct{})
	go func() {
		if err := server.Listen(); err == nil {
			t.Error("listener failure was hidden")
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("failed listener retained background waiters")
	}
}

func TestApiServerJoinsShutdownHooksOnListenFailure(t *testing.T) {
	server := NewHttpServer(-1)
	want := errors.New("shutdown failure")
	called := false
	server.AddShutdownHook(func(ctx context.Context) error {
		called = true
		if _, ok := ctx.Deadline(); !ok {
			t.Error("shutdown context has no deadline")
		}
		return want
	})
	if err := server.Listen(); !called || !errors.Is(err, want) {
		t.Fatalf("called=%v error=%v", called, err)
	}
}
