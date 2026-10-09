package mocks

import (
	"slices"
	"strings"
	"sync"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

// FakePahoClient is an in-memory mqttlib.Client for driving mqtt.MqttService
// without a broker. It reproduces paho v1.5 dispatch (router.go):
//   - OrderMatters false: every inbound message runs the handler on its own goroutine.
//   - OrderMatters true: handlers run one at a time on a single router goroutine.
//     Subscription acknowledgements are completed on that same goroutine, so a
//     handler that blocks on a subscribe deadlocks, as it would with paho.
//
// OnConnect runs on its own goroutine, as in paho. Outbound publishes are
// passed to the OnPublish hook, which plays the broker/device side.
type FakePahoClient struct {
	options *mqttlib.ClientOptions

	mu         sync.Mutex
	connected  bool
	router     chan func()
	routerDone chan struct{}
	onPublish  func(topic string, payload []byte)
	subscribed []string
	subscribe  chan string
}

// NewFakePahoClient returns a ClientFactory-compatible constructor result.
func NewFakePahoClient(options *mqttlib.ClientOptions) *FakePahoClient {
	return &FakePahoClient{options: options, subscribe: make(chan string, 1024)}
}

// OnPublish sets the broker/device simulation for outbound publishes.
func (f *FakePahoClient) OnPublish(hook func(topic string, payload []byte)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onPublish = hook
}

// Subscriptions reports each topic once its subscribe has been acknowledged.
func (f *FakePahoClient) Subscriptions() <-chan string { return f.subscribe }

// Deliver injects an inbound message using the configured dispatch mode.
func (f *FakePahoClient) Deliver(topic string, payload []byte) {
	msg := &fakeMessage{topic: topic, payload: append([]byte(nil), payload...)}
	handler := f.options.DefaultPublishHandler
	if handler == nil {
		return
	}
	if !f.options.Order {
		go handler(f, msg)
		return
	}
	f.routerChan() <- func() { handler(f, msg) }
}

func (f *FakePahoClient) routerChan() chan func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.router
}

func (f *FakePahoClient) IsConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connected
}

func (f *FakePahoClient) IsConnectionOpen() bool { return f.IsConnected() }

func (f *FakePahoClient) Connect() mqttlib.Token {
	f.mu.Lock()
	f.connected = true
	if f.options.Order && f.router == nil {
		f.router = make(chan func(), 4096)
		f.routerDone = make(chan struct{})
		go f.route(f.router, f.routerDone)
	}
	f.mu.Unlock()
	if f.options.OnConnect != nil {
		go f.options.OnConnect(f)
	}
	return completedToken()
}

func (f *FakePahoClient) route(router <-chan func(), done chan<- struct{}) {
	defer close(done)
	for fn := range router {
		fn()
	}
}

// Disconnect stops the router, waiting up to quiesce milliseconds for queued
// work to drain (a deadlocked handler must not hang the caller).
func (f *FakePahoClient) Disconnect(quiesce uint) {
	f.mu.Lock()
	router, done := f.router, f.routerDone
	f.router, f.routerDone = nil, nil
	f.connected = false
	f.mu.Unlock()
	if router != nil {
		close(router)
		select {
		case <-done:
		case <-time.After(time.Duration(quiesce) * time.Millisecond):
		}
	}
}

func (f *FakePahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) mqttlib.Token {
	f.mu.Lock()
	hook := f.onPublish
	f.mu.Unlock()
	if hook != nil {
		var data []byte
		switch v := payload.(type) {
		case []byte:
			data = append([]byte(nil), v...)
		case string:
			data = []byte(v)
		}
		hook(topic, data)
	}
	return completedToken()
}

func (f *FakePahoClient) Subscribe(topic string, qos byte, callback mqttlib.MessageHandler) mqttlib.Token {
	token := newFakeToken()
	ack := func() {
		f.mu.Lock()
		f.subscribed = append(f.subscribed, topic)
		f.mu.Unlock()
		f.subscribe <- strings.TrimPrefix(topic, "zigbee2mqtt/")
		token.complete()
	}
	if f.options.Order {
		f.routerChan() <- ack
	} else {
		ack()
	}
	return token
}

func (f *FakePahoClient) SubscribeMultiple(filters map[string]byte, callback mqttlib.MessageHandler) mqttlib.Token {
	return completedToken()
}

// Unsubscribe removes exactly the given topic filters, as a broker would; a
// filter that was never subscribed is ignored.
func (f *FakePahoClient) Unsubscribe(topics ...string) mqttlib.Token {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, topic := range topics {
		f.subscribed = slices.DeleteFunc(f.subscribed, func(s string) bool { return s == topic })
	}
	return completedToken()
}

// Subscribed returns the topic filters currently subscribed, in subscription order.
func (f *FakePahoClient) Subscribed() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.subscribed)
}

func (f *FakePahoClient) AddRoute(topic string, callback mqttlib.MessageHandler) {}

func (f *FakePahoClient) OptionsReader() mqttlib.ClientOptionsReader {
	return mqttlib.ClientOptionsReader{}
}

type fakeMessage struct {
	topic   string
	payload []byte
}

func (m *fakeMessage) Duplicate() bool   { return false }
func (m *fakeMessage) Qos() byte         { return 0 }
func (m *fakeMessage) Retained() bool    { return false }
func (m *fakeMessage) Topic() string     { return m.topic }
func (m *fakeMessage) MessageID() uint16 { return 0 }
func (m *fakeMessage) Payload() []byte   { return m.payload }
func (m *fakeMessage) Ack()              {}

type fakeToken struct {
	done chan struct{}
	once sync.Once
}

func newFakeToken() *fakeToken { return &fakeToken{done: make(chan struct{})} }

func completedToken() *fakeToken {
	t := newFakeToken()
	t.complete()
	return t
}

func (t *fakeToken) complete() { t.once.Do(func() { close(t.done) }) }

func (t *fakeToken) Wait() bool {
	<-t.done
	return true
}

func (t *fakeToken) WaitTimeout(d time.Duration) bool {
	select {
	case <-t.done:
		return true
	case <-time.After(d):
		return false
	}
}

func (t *fakeToken) Done() <-chan struct{} { return t.done }
func (t *fakeToken) Error() error          { return nil }
