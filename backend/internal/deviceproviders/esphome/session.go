package esphome

// This file owns one native connection attempt: handshake, discovery,
// subscription, ordered state delivery and command sending. Reconnection,
// health and persistence belong to supervisor.go and manager.go.
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"node-herder/internal/deviceproviders"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"github.com/mycontroller-org/esphome_api/pkg/client"
	types "github.com/mycontroller-org/esphome_api/pkg/types"
	"google.golang.org/protobuf/proto"
)

// Sink receives inventory and ordered reports for one session. Implementations
// must return on context cancellation and must not retain or mutate inventory.
type Sink interface {
	Inventory(context.Context, []Descriptor) error
	Ready(context.Context) error
	Observe(context.Context, Observation) error
}

type nativeClient interface {
	Hello() (*types.HelloResponse, error)
	DeviceInfo() (*types.DeviceInfo, error)
	ListEntities() error
	SubscribeStates() error
	Ping() error
	CloseContext(context.Context) error
	Send(proto.Message) error
	Done() <-chan struct{}
}

type dialSession func(context.Context, string, string, func(proto.Message)) (nativeClient, error)

// Session performs one connection attempt. Its caller owns reconnection, health,
// command routing and cancellation, and supplies the sink for confirmed reports.
type Session struct {
	dial           dialSession
	started        atomic.Bool
	ready          atomic.Bool
	done           chan struct{}
	commands       chan sessionCommand
	keepaliveTicks <-chan time.Time
}

type sessionCommand struct {
	ctx    context.Context
	id     string
	value  any
	result chan error
}

// NewSession creates a session using the patched native API client. A session
// does not connect until Run is called.
func NewSession() *Session {
	return newSession(func(ctx context.Context, address, key string, callback func(proto.Message)) (nativeClient, error) {
		return client.DialContext(ctx, "NodeHerder", address, key, sessionTimeout, callback)
	})
}

func newSession(dial dialSession) *Session {
	return &Session{dial: dial, done: make(chan struct{}), commands: make(chan sessionCommand, commandQueueCapacity)}
}

// Command submits one command to a ready session. Cancellation after a send
// starts leaves its outcome unknown; the caller must not automatically retry it.
func (s *Session) Command(ctx context.Context, entityID string, value any) error {
	if ctx == nil {
		return errors.New("command context is required")
	}
	if !s.ready.Load() {
		return errors.New("ESPHome device is offline")
	}
	request := sessionCommand{ctx: ctx, id: entityID, value: value, result: make(chan error, 1)}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		return errors.New("ESPHome device is offline")
	case s.commands <- request:
	default:
		return errors.New("ESPHome command queue is full")
	}
	select {
	case err := <-request.result:
		return err
	case <-ctx.Done():
		return fmt.Errorf("ESPHome command outcome may be unknown: %w", ctx.Err())
	case <-s.done:
		return errors.New("ESPHome connection closed; command outcome may be unknown")
	}
}

// Run connects once, discovers entities, subscribes and forwards confirmed state
// until cancellation, disconnect or failure. It does not retry or expose secrets.
func (s *Session) Run(ctx context.Context, endpoint deviceproviders.Endpoint, key string, sink Sink) error {
	if !s.started.CompareAndSwap(false, true) {
		return errors.New("ESPHome session already started")
	}
	defer close(s.done)
	if ctx == nil || sink == nil || endpoint.NetworkAddress() == "" {
		return errors.New("invalid ESPHome session configuration")
	}
	if err := ValidateEncryptionKey(key); err != nil {
		return err
	}
	incoming := make(chan proto.Message, ingressCapacity)
	overflow := make(chan struct{})
	var overflowOnce sync.Once
	callback := func(msg proto.Message) {
		select {
		case incoming <- msg:
		default:
			overflowOnce.Do(func() { close(overflow) })
		}
	}
	c, err := s.dial(ctx, endpoint.NetworkAddress(), key, callback)
	if err != nil {
		return fmt.Errorf("ESPHome connection failed: %w", err)
	}
	watchDone := make(chan struct{})
	watchExited := make(chan struct{})
	go func() {
		defer close(watchExited)
		select {
		case <-ctx.Done():
			closeCtx, cancel := context.WithTimeout(context.Background(), closeTimeout)
			defer cancel()
			_ = c.CloseContext(closeCtx)
		case <-watchDone:
		}
	}()
	defer func() {
		close(watchDone)
		closeCtx, cancel := context.WithTimeout(context.Background(), closeTimeout)
		defer cancel()
		_ = c.CloseContext(closeCtx)
		<-watchExited
	}()
	hello, err := c.Hello()
	if err != nil {
		return fmt.Errorf("ESPHome hello failed: %w", err)
	}
	if hello == nil || hello.ApiVersionMajor != client.ClientAPIVersionMajor {
		return errors.New("unsupported ESPHome API major version")
	}
	// Older native API firmware requires a legacy authentication request, even
	// when no password is configured. Newer firmware ignores this message.
	if err := c.Send(&api.AuthenticationRequest{}); err != nil { //nolint:staticcheck // required by older firmware
		return fmt.Errorf("ESPHome authentication setup failed: %w", err)
	}
	info, err := c.DeviceInfo()
	if err != nil {
		return fmt.Errorf("ESPHome device information failed: %w", err)
	}
	if identifier, ok := sink.(interface {
		Identify(context.Context, *types.DeviceInfo) error
	}); ok {
		if err := identifier.Identify(ctx, info); err != nil {
			return fmt.Errorf("ESPHome device identity failed: %w", err)
		}
	}
	if err := c.ListEntities(); err != nil {
		return fmt.Errorf("ESPHome discovery failed: %w", err)
	}
	timer := time.NewTimer(sessionTimeout)
	defer timer.Stop()
	descriptors := make(map[string]Descriptor)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.Done():
			return errors.New("ESPHome connection closed during discovery")
		case <-overflow:
			return errors.New("ESPHome ingress overflow; resynchronization required")
		case <-timer.C:
			return errors.New("ESPHome discovery timed out")
		case msg := <-incoming:
			if _, ok := msg.(*api.ListEntitiesDoneResponse); ok {
				return s.runSubscribed(ctx, c, sink, descriptors, incoming, overflow)
			}
			d, err := Describe(msg)
			if err != nil {
				return err
			}
			if d == nil {
				continue
			}
			if _, exists := descriptors[d.ID]; exists || len(descriptors) >= maxEntities {
				return errors.New("ESPHome discovery contains duplicate or too many entities")
			}
			descriptors[d.ID] = *d
		}
	}
}

func (s *Session) runSubscribed(ctx context.Context, c nativeClient, sink Sink, descriptors map[string]Descriptor, incoming <-chan proto.Message, overflow <-chan struct{}) error {
	select {
	case <-overflow:
		return errors.New("ESPHome ingress overflow; resynchronization required")
	default:
	}
	inventory := make([]Descriptor, 0, len(descriptors))
	for _, d := range descriptors {
		inventory = append(inventory, d)
	}
	if err := sink.Inventory(ctx, inventory); err != nil {
		return fmt.Errorf("ESPHome inventory delivery failed: %w", err)
	}
	if err := c.SubscribeStates(); err != nil {
		return fmt.Errorf("ESPHome subscription failed: %w", err)
	}
	select {
	case <-overflow:
		return errors.New("ESPHome ingress overflow; resynchronization required")
	default:
	}
	if err := sink.Ready(ctx); err != nil {
		return fmt.Errorf("ESPHome readiness delivery failed: %w", err)
	}
	s.ready.Store(true)
	defer s.ready.Store(false)
	commandStop := make(chan struct{})
	commandDone := make(chan struct{})
	commandFailed := make(chan struct{}, 1)
	go func() {
		defer close(commandDone)
		s.runCommands(ctx, c, descriptors, overflow, commandStop, commandFailed)
	}()
	defer func() { close(commandStop); <-commandDone }()
	keepalive := time.NewTicker(keepaliveInterval)
	defer keepalive.Stop()
	ticks := s.keepaliveTicks
	if ticks == nil {
		ticks = keepalive.C
	}
	pingCtx, cancelPing := context.WithCancel(ctx)
	pingDone := make(chan struct{})
	pingFailed := make(chan struct{}, 1)
	// Waiting for a ping reply must not stall observations or automation inputs.
	go func() {
		defer close(pingDone)
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-ticks:
				if c.Ping() != nil {
					pingFailed <- struct{}{}
					return
				}
			}
		}
	}()
	defer func() {
		cancelPing()
		// Closing interrupts a pending ping before joining the owned worker.
		closeCtx, cancel := context.WithTimeout(context.Background(), closeTimeout)
		defer cancel()
		_ = c.CloseContext(closeCtx)
		<-pingDone
	}()
	for {
		// An overflow invalidates all queued observations, even when the queue remains readable.
		select {
		case <-overflow:
			return errors.New("ESPHome ingress overflow; resynchronization required")
		default:
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.Done():
			return errors.New("ESPHome connection closed")
		case <-overflow:
			return errors.New("ESPHome ingress overflow; resynchronization required")
		case <-pingFailed:
			return errors.New("ESPHome keepalive failed")
		case <-commandFailed:
			return errors.New("ESPHome command transport failed")
		case msg := <-incoming:
			id, ok := stateID(msg)
			if !ok {
				continue
			}
			d, ok := descriptors[id]
			if !ok {
				return errors.New("ESPHome state for undiscovered entity")
			}
			observation, err := DecodeState(d, msg)
			if err != nil {
				return err
			}
			if err := sink.Observe(ctx, observation); err != nil {
				return fmt.Errorf("ESPHome state delivery failed: %w", err)
			}
		}
	}
}

// Commands use a separate worker because state delivery can synchronously trigger
// an automation on the device lane; waiting for that lane in the reader and
// dispatching commands on the reader would deadlock the automation.
func (s *Session) runCommands(ctx context.Context, c nativeClient, descriptors map[string]Descriptor, overflow, stop <-chan struct{}, failed chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.Done():
			return
		case <-overflow:
			return
		case <-stop:
			return
		case request := <-s.commands:
			select {
			case <-ctx.Done():
				request.result <- ctx.Err()
				return
			case <-c.Done():
				request.result <- errors.New("ESPHome device is offline")
				return
			case <-overflow:
				request.result <- errors.New("ESPHome device is desynchronized")
				return
			case <-stop:
				request.result <- errors.New("ESPHome device is offline")
				return
			default:
			}
			if err := request.ctx.Err(); err != nil {
				request.result <- err
				continue
			}
			d, ok := descriptors[request.id]
			if !ok {
				request.result <- errors.New("ESPHome entity not found")
				continue
			}
			message, err := EncodeCommand(d, request.value)
			if err != nil {
				request.result <- err
				continue
			}
			err = c.Send(message)
			request.result <- err
			if err != nil {
				failed <- struct{}{}
				return
			}
		}
	}
}

func stateID(msg proto.Message) (string, bool) {
	var key, device uint32
	switch m := msg.(type) {
	case *api.SensorStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.BinarySensorStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.TextSensorStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.TextStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.SwitchStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.LightStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.NumberStateResponse:
		key, device = m.Key, m.DeviceId
	case *api.SelectStateResponse:
		key, device = m.Key, m.DeviceId
	default:
		return "", false
	}
	return fmt.Sprintf("%08x-%08x", device, key), true
}
