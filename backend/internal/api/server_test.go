package api

import (
	"context"
	"testing"
	"time"
)

func TestApiServerReturnsOnListenFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := NewHttpServer(-1, WithContext(ctx))
	done := make(chan struct{})
	go func() { server.Listen(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("failed listener retained background waiters")
	}
}
