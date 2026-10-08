package ws

import (
	"errors"
	"testing"
)

type failingWebSocket struct {
	WebSocket
	err error
}

func (f failingWebSocket) Broadcast(string, interface{}) error { return f.err }

func TestEmitDeviceReturnsBroadcastError(t *testing.T) {
	want := errors.New("broadcast failed")
	h := &eventHubImpl{
		server:       failingWebSocket{err: want},
		onLoadDevice: func(string) (interface{}, error) { return "fixture", nil },
	}
	if err := h.EmitDevice("fixture"); !errors.Is(err, want) {
		t.Fatalf("EmitDevice error = %v, want %v", err, want)
	}
}
