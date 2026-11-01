package esphome

// This file translates a native state message into a confirmed observation. It
// validates the message against the discovered descriptor and never mutates
// shared state or treats a submitted command as confirmation.
import (
	"errors"
	"unicode/utf8"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"google.golang.org/protobuf/proto"
)

// Observation represents a single confirmed report. Value is nil when the device
// reports a missing value or a numeric value that cannot be represented in JSON.
type Observation struct {
	EntityID  string
	Available bool
	Value     any
}

// DecodeState validates a state message against its discovered descriptor; it
// does not mutate state or interpret a command as confirmation.
func DecodeState(d Descriptor, message proto.Message) (Observation, error) {
	if message == nil || !message.ProtoReflect().IsValid() {
		return Observation{}, errors.New("invalid ESPHome state message")
	}
	var key, device uint32
	var kind string
	var value any
	available := true
	switch m := message.(type) {
	case *api.SensorStateResponse:
		key, device, kind, value, available = m.Key, m.DeviceId, kindSensor, float64(m.State), !m.MissingState && finite(float64(m.State))
	case *api.BinarySensorStateResponse:
		key, device, kind, value, available = m.Key, m.DeviceId, kindBinarySensor, m.State, !m.MissingState
	case *api.TextSensorStateResponse:
		key, device, kind, value, available = m.Key, m.DeviceId, kindTextSensor, m.State, !m.MissingState
	case *api.TextStateResponse:
		key, device, kind, value, available = m.Key, m.DeviceId, kindText, m.State, !m.MissingState
		if d.Kind == kindSecretText {
			kind, value, available = kindSecretText, nil, false
		}
	case *api.SwitchStateResponse:
		key, device, kind, value = m.Key, m.DeviceId, kindSwitch, m.State
	case *api.LightStateResponse:
		key, device, kind, value = m.Key, m.DeviceId, kindLight, m.State
	case *api.NumberStateResponse:
		key, device, kind, value, available = m.Key, m.DeviceId, kindNumber, float64(m.State), !m.MissingState && finite(float64(m.State))
	case *api.SelectStateResponse:
		key, device, kind, value, available = m.Key, m.DeviceId, kindSelect, m.State, !m.MissingState
	default:
		return Observation{}, errors.New("unsupported ESPHome state message")
	}
	if d.Key != key || d.DeviceID != device || d.Kind != kind {
		return Observation{}, errors.New("ESPHome state does not match discovery")
	}
	if s, ok := value.(string); available && ok {
		limit := maxStateTextBytes
		if kind == kindText {
			limit = maxTextValueBytes
		}
		if len(s) > limit || !utf8.ValidString(s) {
			return Observation{}, errors.New("invalid ESPHome state text")
		}
	}
	if !available {
		value = nil
	}
	return Observation{EntityID: d.ID, Available: available, Value: value}, nil
}
