package esphome

// This file owns reconnection, health classification and command routing for one
// configured device. It allows one active session attempt at a time and never
// replays a command whose transmission outcome is unknown.
import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"sync"
	"sync/atomic"
	"time"

	types "github.com/mycontroller-org/esphome_api/pkg/types"
	"node-herder/internal/deviceproviders"
)

// ConnectionStatus describes a provider session's externally visible lifecycle.
type ConnectionStatus string

const (
	// Connecting means an attempt is establishing a session and inventory.
	Connecting ConnectionStatus = deviceproviders.StatusConnecting
	// Online means discovery and state subscription completed successfully.
	Online ConnectionStatus = deviceproviders.StatusOnline
	// Offline means no session is available and any last values are stale.
	Offline ConnectionStatus = deviceproviders.StatusOffline
)

// SupervisorSink receives ordered session updates and lifecycle status. It must
// return promptly when its context is canceled and must not expose credentials.
type SupervisorSink interface {
	Sink
	Status(context.Context, ConnectionStatus) error
}

type sessionAttempt interface {
	Run(context.Context, deviceproviders.Endpoint, string, Sink) error
	Command(context.Context, string, any) error
}

// Supervisor owns at most one session attempt for a configured device. The
// caller owns Run's context and must wait for Run to return during shutdown.
type Supervisor struct {
	started    atomic.Bool
	mu         sync.RWMutex
	current    sessionAttempt
	status     ConnectionStatus
	diagnostic string
	reconnects uint64
	newAttempt func() sessionAttempt
	wait       func(context.Context, time.Duration) error
	now        func() time.Time
	jitter     func(time.Duration) time.Duration
}

// NewSupervisor creates a supervisor with bounded backoff and jitter. It does
// not start a connection until Run is called.
func NewSupervisor() *Supervisor {
	return &Supervisor{
		status:     Offline,
		newAttempt: func() sessionAttempt { return NewSession() },
		wait:       waitBackoff,
		now:        time.Now,
		jitter: func(base time.Duration) time.Duration {
			jittered := base + time.Duration(rand.Int64N(int64(base/5)+1))
			return min(jittered, maxReconnectDelay)
		},
	}
}

// Run reconnects until canceled, delivering status and fresh inventory for each
// attempt. It never retries commands; a supervisor can be run only once.
func (s *Supervisor) Run(ctx context.Context, endpoint deviceproviders.Endpoint, key string, sink SupervisorSink) error {
	if ctx == nil || sink == nil || endpoint.NetworkAddress() == "" {
		return errors.New("invalid ESPHome supervisor configuration")
	}
	if err := ValidateEncryptionKey(key); err != nil {
		return err
	}
	if !s.started.CompareAndSwap(false, true) {
		return errors.New("ESPHome supervisor already started")
	}
	delay := minReconnectDelay
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.setHealth(Connecting, "")
		if err := sink.Status(ctx, Connecting); err != nil {
			return err
		}
		attempt := s.newAttempt()
		s.mu.Lock()
		s.current = attempt
		s.mu.Unlock()
		wrapped := &supervisedSink{SupervisorSink: sink, now: s.now, supervisor: s}
		err := attempt.Run(ctx, endpoint, key, wrapped)
		s.mu.Lock()
		s.current = nil
		s.mu.Unlock()
		if ctx.Err() != nil {
			s.setHealth(Offline, "")
			return ctx.Err()
		}
		s.setHealth(Offline, connectionDiagnostic(err))
		if statusErr := sink.Status(ctx, Offline); statusErr != nil {
			return statusErr
		}
		if err == nil {
			return errors.New("ESPHome session ended unexpectedly")
		}
		if !wrapped.readyAt.IsZero() && s.now().Sub(wrapped.readyAt) >= stableConnectionAge {
			delay = minReconnectDelay
		}
		if err := s.wait(ctx, s.jitter(delay)); err != nil {
			return err
		}
		s.mu.Lock()
		s.reconnects++
		s.mu.Unlock()
		delay = min(delay*2, maxReconnectDelay)
	}
}

func connectionDiagnostic(err error) string {
	if errors.Is(err, deviceproviders.ErrPhysicalIdentityConflict) || errors.Is(err, deviceproviders.ErrPhysicalIdentityChanged) {
		return "identity_conflict"
	}
	if errors.Is(err, deviceproviders.ErrEntityIdentityChanged) {
		return "entity_identity_conflict"
	}
	var network *net.OpError
	if errors.As(err, &network) && network.Op == "dial" {
		return "network_unreachable"
	}
	return "connection_failed"
}

// Health returns a safe connection status and a classified diagnostic code.
func (s *Supervisor) Health() deviceproviders.Health {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return deviceproviders.Health{Status: string(s.status), Diagnostic: s.diagnostic, Reconnects: s.reconnects}
}

func (s *Supervisor) setHealth(status ConnectionStatus, diagnostic string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status, s.diagnostic = status, diagnostic
}

// Command submits a command to the currently active session without retrying it.
// A disconnect or caller cancellation may leave its outcome unknown.
func (s *Supervisor) Command(ctx context.Context, entityID string, value any) error {
	s.mu.RLock()
	current := s.current
	s.mu.RUnlock()
	if current == nil {
		return errors.New("ESPHome device is offline")
	}
	return current.Command(ctx, entityID, value)
}

type supervisedSink struct {
	SupervisorSink
	now        func() time.Time
	readyAt    time.Time
	supervisor *Supervisor
}

func (s *supervisedSink) Identify(ctx context.Context, info *types.DeviceInfo) error {
	if identifier, ok := s.SupervisorSink.(interface {
		Identify(context.Context, *types.DeviceInfo) error
	}); ok {
		return identifier.Identify(ctx, info)
	}
	return nil
}

func (s *supervisedSink) Ready(ctx context.Context) error {
	if err := s.SupervisorSink.Ready(ctx); err != nil {
		return err
	}
	if err := s.SupervisorSink.Status(ctx, Online); err != nil {
		return err
	}
	s.supervisor.setHealth(Online, "")
	s.readyAt = s.now()
	return nil
}

func waitBackoff(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
