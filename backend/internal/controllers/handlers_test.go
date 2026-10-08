package controllers

import (
	"errors"
	"node-herder/internal/services"
	"node-herder/internal/ws"
	"node-herder/mocks"
	"reflect"
	"strings"
	"testing"
)

type failureMQTT struct {
	*mocks.MockMqttClient
	calls *[]string
	err   error
}

func (m *failureMQTT) RemoveTopic(topic string) error {
	*m.calls = append(*m.calls, "unsubscribe:"+topic)
	return m.err
}

func TestBridgeResponsesRejectMalformedDataBeforeEffects(t *testing.T) {
	for _, tc := range []struct{ topic, payload string }{
		{"rename", `{"status":"ok","data":{"from":"old"}}`},
		{"rename", `{"status":"ok","data":{"from":"old","to":42}}`},
		{"rename", `{"status":"ok","data":{"from":"","to":"new"}}`},
		{"rename", `{"status":"ok","data":{"from":"old","to":""}}`},
		{"rename", `{"status":"ok","data":null}`},
		{"remove", `{"status":"ok","data":{}}`},
		{"remove", `{"status":"ok","data":{"id":42}}`},
		{"interview", `{"status":"ok","data":{"id":""}}`},
		{"interview", `null`},
		{"interview", `{"data":{"id":"device"}}`},
		{"permit_join", `{"status":"unknown"}`},
		{"permit_join", `{"status":"ok","transaction":42}`},
	} {
		t.Run(tc.topic+"/"+tc.payload, func(t *testing.T) {
			var calls []string
			w := &failureEventHub{calls: &calls}
			m := &failureMQTT{calls: &calls}
			r := services.NewHubRegisterService(&failureStore{calls: &calls}, w, 0)
			var h handler
			switch tc.topic {
			case "rename":
				h = newBridgeDeviceRenameResponseHandler(w, m)
			case "remove":
				h = newBridgeDeviceRemoveResponseHandler(r, w, m)
			case "interview":
				h = newBridgeDeviceInterviewResponseHandler(w, m)
			case "permit_join":
				h = newBridgePermitJoinResponseHandler(w, m)
			}
			// Recover only to make the pre-fix panic a useful regression failure.
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("malformed response panicked: %v; effects=%v", p, calls)
				}
			}()
			topic := "bridge/response/device/" + tc.topic
			if tc.topic == "permit_join" {
				topic = "bridge/response/permit_join"
			}
			if err := h.ProcessPayload(topic, "mqtt", []byte(tc.payload)); err == nil {
				t.Error("malformed response returned no error")
			}
			if len(calls) != 0 {
				t.Fatalf("malformed response caused effects: %v", calls)
			}
		})
	}
}

func TestDeviceHandlerReturnsInvalidJSON(t *testing.T) {
	for _, payload := range []string{"", "{", "null", "[]", `"text"`} {
		t.Run(payload, func(t *testing.T) {
			h := newDeviceHandler(nil) // invalid input must never reach processing
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("invalid payload reached processor: %v", p)
				}
			}()
			if err := h.ProcessPayload("fixture", "mqtt", []byte(payload)); err == nil {
				t.Error("invalid device JSON returned no error")
			} else if payload == "{" {
				if !strings.Contains(err.Error(), "device payload") {
					t.Errorf("error lacks stage: %v", err)
				}
			}
		})
	}
}

func TestBridgeResponseEffectsAndFailures(t *testing.T) {
	emitErr, removeErr := errors.New("emit failed"), errors.New("remove failed")
	t.Run("rename success and emit error", func(t *testing.T) {
		var calls []string
		h := newBridgeDeviceRenameResponseHandler(&failureEventHub{calls: &calls, err: emitErr}, &failureMQTT{calls: &calls})
		err := h.ProcessPayload("bridge/response/device/rename", "mqtt", []byte(`{"status":"ok","data":{"from":"old","to":"new"}}`))
		if !errors.Is(err, emitErr) || !reflect.DeepEqual(calls, []string{"unsubscribe:old", "emit:new"}) {
			t.Fatalf("error/effects = %v/%v", err, calls)
		}
	})
	t.Run("remove store failure still reports to browser", func(t *testing.T) {
		var calls []string
		w := &failureEventHub{calls: &calls, err: emitErr}
		r := services.NewHubRegisterService(&failureStore{calls: &calls, err: removeErr}, w, 0)
		h := newBridgeDeviceRemoveResponseHandler(r, w, nil)
		err := h.ProcessPayload("bridge/response/device/remove", "mqtt", []byte(`{"status":"ok","data":{"id":"fixture"}}`))
		if !errors.Is(err, removeErr) || !errors.Is(err, emitErr) || !reflect.DeepEqual(calls, []string{"remove", "broadcast"}) {
			t.Fatalf("error/effects = %v/%v", err, calls)
		}
	})
	t.Run("interview broadcast failure", func(t *testing.T) {
		var calls []string
		h := newBridgeDeviceInterviewResponseHandler(&failureEventHub{calls: &calls, err: emitErr}, nil)
		if err := h.ProcessPayload("bridge/response/device/interview", "mqtt", []byte(`{"status":"ok","data":{"id":"fixture"}}`)); !errors.Is(err, emitErr) {
			t.Fatalf("broadcast error lost: %v", err)
		}
	})
}

func TestBridgeResponseNotificationCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, payload, event, message string
	}{
		{"remove", `{"status":"ok","data":{"id":"fixture","extra":true},"extra":true}`, ws.OperationSuccess, "Device fixture removed"},
		{"interview", `{"status":"ok","data":{"id":"fixture"}}`, ws.OperationSuccess, "Device fixture interview successful"},
		{"remove", `{"status":"error","error":"bridge rejected","data":[1,2]}`, ws.OperationFailed, "bridge rejected"},
		{"interview", `{"status":"error","error":"bridge rejected"}`, ws.OperationFailed, "bridge rejected"},
		{"rename", `{"status":"error","error":"bridge rejected"}`, ws.OperationFailed, "error"},
		{"permit_join", `{"status":"error","error":"bridge rejected"}`, ws.OperationFailed, "bridge rejected"},
	} {
		t.Run(tc.name+"/"+tc.event, func(t *testing.T) {
			var calls []string
			w := &failureEventHub{calls: &calls}
			r := services.NewHubRegisterService(&failureStore{calls: &calls}, w, 0)
			var h handler
			topic := "bridge/response/device/" + tc.name
			switch tc.name {
			case "remove":
				h = newBridgeDeviceRemoveResponseHandler(r, w, nil)
			case "interview":
				h = newBridgeDeviceInterviewResponseHandler(w, nil)
			case "rename":
				h = newBridgeDeviceRenameResponseHandler(w, nil)
			case "permit_join":
				h = newBridgePermitJoinResponseHandler(w, nil)
				topic = "bridge/response/permit_join"
			}
			if err := h.ProcessPayload(topic, "mqtt", []byte(tc.payload)); err != nil {
				t.Fatal(err)
			}
			if w.lastEvent != tc.event || w.lastPayload != tc.message {
				t.Fatalf("notification = %q/%v, want %q/%q", w.lastEvent, w.lastPayload, tc.event, tc.message)
			}
		})
	}
}

func TestBridgeLoggingNullAndBroadcastError(t *testing.T) {
	var calls []string
	want := errors.New("broadcast failed")
	h := newBridgeLoggingHandler(&failureEventHub{calls: &calls, err: want})
	if err := h.ProcessPayload("bridge/logging", "mqtt", []byte("null")); err != nil || len(calls) != 0 {
		t.Fatalf("null logging response = %v, calls %v", err, calls)
	}
	if err := h.ProcessPayload("bridge/logging", "mqtt", []byte(`{"level":"error","message":"fixture error"}`)); !errors.Is(err, want) {
		t.Fatalf("broadcast error lost: %v", err)
	}
}

var _ ws.EventHub = (*failureEventHub)(nil)
