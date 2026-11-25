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

// recordingResultWebSocket fails the result frame, as a frame over maxBroadcastBytes
// would, and records the follow-up frame.
type recordingResultWebSocket struct {
	WebSocket
	events []string
}

func (r *recordingResultWebSocket) Broadcast(name string, _ interface{}) error {
	r.events = append(r.events, name)
	if name == Metrics {
		return errors.New("frame exceeds the outbound limit")
	}
	return nil
}

// A rejected result still answers the pending request; LoadMetrics uses ReportError=false.
func TestExecuteReportsRejectedResult(t *testing.T) {
	server := &recordingResultWebSocket{}
	h := &eventHubImpl{server: server}
	h.execute(&eventExecutorOptions{
		Action:        func(interface{}) (interface{}, error) { return "metrics result", nil },
		SuccessEvent:  Metrics,
		ReportResult:  true,
		ReportSuccess: true,
		ReportError:   false,
	})
	if len(server.events) != 2 || server.events[0] != Metrics || server.events[1] != OperationFailed {
		t.Fatalf("broadcasts = %v, want [%s %s]", server.events, Metrics, OperationFailed)
	}
}
