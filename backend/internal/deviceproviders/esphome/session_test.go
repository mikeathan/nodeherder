package esphome

import (
	"bufio"
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"github.com/mycontroller-org/esphome_api/pkg/client"
	"github.com/mycontroller-org/esphome_api/pkg/connection"
	types "github.com/mycontroller-org/esphome_api/pkg/types"
	"google.golang.org/protobuf/proto"
	"node-herder/internal/deviceproviders"
)

type fakeSessionClient struct {
	callback     func(proto.Message)
	done         chan struct{}
	once         sync.Once
	overflow     bool
	incompatible bool
	sent         chan proto.Message
}

func TestSessionWithLocalNativeAPIPeer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	peerResult := make(chan error, 1)
	go func() {
		peer, err := listener.Accept()
		if err != nil {
			peerResult <- err
			return
		}
		defer peer.Close()
		wire, err := connection.NewPlaintextConnection(peer, time.Second)
		if err != nil {
			peerResult <- err
			return
		}
		reader := bufio.NewReader(peer)
		read := func() (proto.Message, error) { return wire.Read(reader) }
		message, err := read()
		if err != nil {
			peerResult <- err
			return
		}
		if _, ok := message.(*api.HelloRequest); !ok {
			peerResult <- errors.New("expected hello")
			return
		}
		if err = wire.Write(&api.HelloResponse{ApiVersionMajor: 1, ApiVersionMinor: 9}); err != nil {
			peerResult <- err
			return
		}
		message, err = read()
		if err != nil {
			peerResult <- err
			return
		}
		if _, ok := message.(*api.AuthenticationRequest); !ok {
			peerResult <- errors.New("expected authentication")
			return
		} //nolint:staticcheck // old firmware fixture
		if err = wire.Write(&api.AuthenticationResponse{}); err != nil {
			peerResult <- err
			return
		} //nolint:staticcheck // old firmware fixture
		message, err = read()
		if err != nil {
			peerResult <- err
			return
		}
		if _, ok := message.(*api.DeviceInfoRequest); !ok {
			peerResult <- errors.New("expected device information request")
			return
		}
		if err = wire.Write(&api.DeviceInfoResponse{MacAddress: "02:00:00:00:00:01"}); err != nil {
			peerResult <- err
			return
		}
		message, err = read()
		if err != nil {
			peerResult <- err
			return
		}
		if _, ok := message.(*api.ListEntitiesRequest); !ok {
			peerResult <- errors.New("expected discovery")
			return
		}
		if err = wire.Write(&api.ListEntitiesSensorResponse{Key: 7, Name: "Temperature"}); err != nil {
			peerResult <- err
			return
		}
		if err = wire.Write(&api.ListEntitiesSwitchResponse{Key: 8, Name: "Switch"}); err != nil {
			peerResult <- err
			return
		}
		if err = wire.Write(&api.ListEntitiesDoneResponse{}); err != nil {
			peerResult <- err
			return
		}
		message, err = read()
		if err != nil {
			peerResult <- err
			return
		}
		if _, ok := message.(*api.SubscribeStatesRequest); !ok {
			peerResult <- errors.New("expected subscription")
			return
		}
		if err = wire.Write(&api.SensorStateResponse{Key: 7, State: 22}); err != nil {
			peerResult <- err
			return
		}
		message, err = read()
		if err != nil {
			peerResult <- err
			return
		}
		if command, ok := message.(*api.SwitchCommandRequest); !ok || command.Key != 8 || !command.State {
			peerResult <- errors.New("expected switch command")
			return
		}
		message, err = read()
		if err != nil {
			peerResult <- err
			return
		}
		if _, ok := message.(*api.DisconnectRequest); !ok {
			peerResult <- errors.New("expected disconnect")
			return
		}
		peerResult <- wire.Write(&api.DisconnectResponse{})
	}()
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	session := newSession(func(ctx context.Context, _, key string, callback func(proto.Message)) (nativeClient, error) {
		return client.DialContext(ctx, "test", listener.Addr().String(), key, time.Second, callback)
	})
	sink := &fakeSessionSink{inventory: make(chan []Descriptor, 1), observations: make(chan Observation, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- session.Run(ctx, endpoint, "", sink) }()
	<-sink.inventory
	if observation := <-sink.observations; observation.Value != float64(22) {
		t.Fatalf("observation: %+v", observation)
	}
	if err := session.Command(ctx, "00000000-00000008", true); err != nil {
		t.Fatal(err)
	}
	cancel()
	<-result
	if err := <-peerResult; err != nil {
		t.Fatal(err)
	}
}

func (f *fakeSessionClient) Hello() (*types.HelloResponse, error) {
	if f.incompatible {
		return &types.HelloResponse{ApiVersionMajor: client.ClientAPIVersionMajor + 1}, nil
	}
	return &types.HelloResponse{ApiVersionMajor: client.ClientAPIVersionMajor}, nil
}
func (f *fakeSessionClient) DeviceInfo() (*types.DeviceInfo, error) {
	return &types.DeviceInfo{MacAddress: "02:00:00:00:00:01"}, nil
}
func (f *fakeSessionClient) ListEntities() error {
	if f.overflow {
		for i := 0; i < ingressCapacity+1; i++ {
			f.callback(&api.PingRequest{})
		}
		return nil
	}
	f.callback(&api.ListEntitiesSensorResponse{Key: 7, Name: "temperature"})
	f.callback(&api.ListEntitiesSwitchResponse{Key: 8, Name: "switch"})
	f.callback(&api.ListEntitiesDoneResponse{})
	return nil
}
func (f *fakeSessionClient) SubscribeStates() error {
	f.callback(&api.SensorStateResponse{Key: 7, State: 21})
	return nil
}
func (f *fakeSessionClient) CloseContext(context.Context) error {
	f.once.Do(func() { close(f.done) })
	return nil
}
func (f *fakeSessionClient) Done() <-chan struct{} { return f.done }
func (f *fakeSessionClient) Send(message proto.Message) error {
	if _, ok := message.(*api.AuthenticationRequest); !ok {
		f.sent <- message
	} //nolint:staticcheck // legacy fixture
	return nil
}
func (f *fakeSessionClient) Ping() error { return nil }

type fakeSessionSink struct {
	inventory    chan []Descriptor
	observations chan Observation
}

type commandOnObservation struct {
	session *Session
	result  chan error
}

func (s *commandOnObservation) Inventory(context.Context, []Descriptor) error { return nil }
func (s *commandOnObservation) Ready(context.Context) error                   { return nil }
func (s *commandOnObservation) Observe(ctx context.Context, _ Observation) error {
	err := s.session.Command(ctx, "00000000-00000008", true)
	s.result <- err
	return err
}

func TestSessionCommandDuringSynchronousObservation(t *testing.T) {
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	fake := &fakeSessionClient{done: make(chan struct{}), sent: make(chan proto.Message, 1)}
	session := newSession(func(_ context.Context, _, _ string, callback func(proto.Message)) (nativeClient, error) {
		fake.callback = callback
		return fake, nil
	})
	sink := &commandOnObservation{session: session, result: make(chan error, 1)}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- session.Run(ctx, endpoint, "", sink) }()
	if err := <-sink.result; err != nil {
		t.Fatal(err)
	}
	if _, ok := (<-fake.sent).(*api.SwitchCommandRequest); !ok {
		t.Fatal("switch command not sent")
	}
	cancel()
	<-done
}

func (s *fakeSessionSink) Inventory(_ context.Context, d []Descriptor) error {
	s.inventory <- d
	return nil
}
func (s *fakeSessionSink) Ready(context.Context) error { return nil }
func (s *fakeSessionSink) Observe(_ context.Context, o Observation) error {
	s.observations <- o
	return nil
}

func TestSessionDiscoversAndForwardsState(t *testing.T) {
	endpoint, err := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeSessionClient{done: make(chan struct{}), sent: make(chan proto.Message, 1)}
	session := newSession(func(_ context.Context, _, _ string, callback func(proto.Message)) (nativeClient, error) {
		fake.callback = callback
		return fake, nil
	})
	sink := &fakeSessionSink{inventory: make(chan []Descriptor, 1), observations: make(chan Observation, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- session.Run(ctx, endpoint, "", sink) }()
	inventory := <-sink.inventory
	if len(inventory) != 2 {
		t.Fatalf("inventory: %+v", inventory)
	}
	observation := <-sink.observations
	if !observation.Available || observation.Value != float64(21) {
		t.Fatalf("observation: %+v", observation)
	}
	if err := session.Command(context.Background(), "00000000-00000008", true); err != nil {
		t.Fatal(err)
	}
	message := <-fake.sent
	if command, ok := message.(*api.SwitchCommandRequest); !ok || !command.State || command.Key != 8 {
		t.Fatalf("command: %v", message)
	}
	if err := session.Command(context.Background(), "00000000-00000007", true); err == nil {
		t.Fatal("sensor command accepted")
	}
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) && err.Error() != "ESPHome connection closed" {
		t.Fatalf("got %v", err)
	}
	<-fake.done
}

func TestSessionReportsIngressOverflow(t *testing.T) {
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	fake := &fakeSessionClient{done: make(chan struct{}), overflow: true}
	session := newSession(func(_ context.Context, _, _ string, callback func(proto.Message)) (nativeClient, error) {
		fake.callback = callback
		return fake, nil
	})
	sink := &fakeSessionSink{inventory: make(chan []Descriptor, 1), observations: make(chan Observation, 1)}
	if err := session.Run(context.Background(), endpoint, "", sink); err == nil || err.Error() != "ESPHome ingress overflow; resynchronization required" {
		t.Fatalf("got %v", err)
	}
}

func TestSessionRejectsIncompatibleMajorVersion(t *testing.T) {
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	fake := &fakeSessionClient{done: make(chan struct{}), incompatible: true}
	session := newSession(func(context.Context, string, string, func(proto.Message)) (nativeClient, error) { return fake, nil })
	err := session.Run(context.Background(), endpoint, "", &fakeSessionSink{})
	if err == nil || err.Error() != "unsupported ESPHome API major version" {
		t.Fatalf("got %v", err)
	}
}

// A ping can wait for a reply while the reader continues delivering reports.
type delayedPingClient struct {
	*fakeSessionClient
	pingStarted chan struct{}
}

func (c *delayedPingClient) Ping() error {
	close(c.pingStarted)
	<-c.done
	return errors.New("closed")
}
func TestSessionDelayedPingDoesNotBlockReports(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fake := &delayedPingClient{fakeSessionClient: &fakeSessionClient{done: make(chan struct{}), sent: make(chan proto.Message, 8)}, pingStarted: make(chan struct{})}
	session := newSession(func(_ context.Context, _, _ string, callback func(proto.Message)) (nativeClient, error) {
		fake.callback = callback
		return fake, nil
	})
	ticks := make(chan time.Time, 1)
	session.keepaliveTicks = ticks
	sink := &fakeSessionSink{inventory: make(chan []Descriptor, 1), observations: make(chan Observation, 2)}
	endpoint, _ := deviceproviders.NewEndpoint("192.168.1.2", 6053)
	result := make(chan error, 1)
	go func() { result <- session.Run(ctx, endpoint, "", sink) }()
	select {
	case <-sink.observations:
	case <-time.After(time.Second):
		t.Fatal("initial report missing")
	}
	ticks <- time.Now()
	select {
	case <-fake.pingStarted:
	case <-time.After(time.Second):
		t.Fatal("ping did not start")
	}
	fake.callback(&api.SensorStateResponse{Key: 7, State: 42})
	select {
	case report := <-sink.observations:
		if report.Value != float64(42) {
			t.Fatal(report)
		}
	case <-time.After(time.Second):
		t.Fatal("ping blocked report")
	}
	cancel()
	select {
	case <-result:
	case <-time.After(time.Second):
		t.Fatal("pending ping prevented shutdown")
	}
}
