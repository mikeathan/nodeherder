package esphome

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"node-herder/internal/deviceproviders"
)

type supervisorAttempt struct {
	fail    bool
	entered chan struct{}
}

func (a *supervisorAttempt) Run(ctx context.Context, _ deviceproviders.Endpoint, _ string, sink Sink) error {
	if err := sink.Ready(ctx); err != nil {
		return err
	}
	close(a.entered)
	if a.fail {
		return errors.New("disconnected")
	}
	<-ctx.Done()
	return ctx.Err()
}
func (a *supervisorAttempt) Command(context.Context, string, any) error { return nil }

type supervisorTestSink struct {
	mu       sync.Mutex
	statuses []ConnectionStatus
}

func (s *supervisorTestSink) Inventory(context.Context, []Descriptor) error { return nil }
func (s *supervisorTestSink) Ready(context.Context) error                   { return nil }
func (s *supervisorTestSink) Observe(context.Context, Observation) error    { return nil }
func (s *supervisorTestSink) Status(_ context.Context, status ConnectionStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses = append(s.statuses, status)
	return nil
}

func TestSupervisorReconnectsAndStops(t *testing.T) {
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	first := &supervisorAttempt{fail: true, entered: make(chan struct{})}
	second := &supervisorAttempt{entered: make(chan struct{})}
	attempts := []sessionAttempt{first, second}
	delays := make(chan time.Duration, 1)
	s := NewSupervisor()
	s.newAttempt = func() sessionAttempt { next := attempts[0]; attempts = attempts[1:]; return next }
	s.jitter = func(d time.Duration) time.Duration { return d }
	s.wait = func(_ context.Context, d time.Duration) error { delays <- d; return nil }
	sink := &supervisorTestSink{}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- s.Run(ctx, endpoint, "", sink) }()
	if delay := <-delays; delay != time.Second {
		t.Fatalf("delay %v", delay)
	}
	<-second.entered
	if health := s.Health(); health.Reconnects != 1 {
		t.Fatalf("reconnection count: %+v", health)
	}
	if err := s.Command(context.Background(), "entity", true); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	if err := s.Command(context.Background(), "entity", true); err == nil {
		t.Fatal("command accepted after shutdown")
	}
	sink.mu.Lock()
	defer sink.mu.Unlock()
	want := []ConnectionStatus{Connecting, Online, Offline, Connecting, Online}
	if len(sink.statuses) != len(want) {
		t.Fatalf("statuses %v", sink.statuses)
	}
	for i := range want {
		if sink.statuses[i] != want[i] {
			t.Fatalf("statuses %v", sink.statuses)
		}
	}
}

func TestConnectionDiagnostic(t *testing.T) {
	if got := connectionDiagnostic(fmt.Errorf("connect: %w", &net.OpError{Op: "dial", Err: errors.New("unreachable")})); got != "network_unreachable" {
		t.Fatalf("diagnostic: %s", got)
	}
	if got := connectionDiagnostic(fmt.Errorf("inventory: %w", deviceproviders.ErrEntityIdentityChanged)); got != "entity_identity_conflict" {
		t.Fatalf("diagnostic: %s", got)
	}
}

func TestWaitBackoffCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitBackoff(ctx, time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestSupervisorBackoffIsBounded(t *testing.T) {
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	s := NewSupervisor()
	s.newAttempt = func() sessionAttempt { return &supervisorAttempt{fail: true, entered: make(chan struct{})} }
	s.jitter = func(d time.Duration) time.Duration { return d }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var delays []time.Duration
	s.wait = func(ctx context.Context, d time.Duration) error {
		delays = append(delays, d)
		if len(delays) == 8 {
			cancel()
			return ctx.Err()
		}
		return nil
	}
	if err := s.Run(ctx, endpoint, "", &supervisorTestSink{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 32 * time.Second, 60 * time.Second, 60 * time.Second}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delays %v", delays)
		}
	}
	for i := 0; i < 100; i++ {
		for _, base := range []time.Duration{time.Second, 60 * time.Second} {
			value := NewSupervisor().jitter(base)
			if value < base || value > 60*time.Second {
				t.Fatalf("jitter %v for %v", value, base)
			}
		}
	}
}
