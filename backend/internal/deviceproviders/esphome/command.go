package esphome

// This file validates a requested value against the discovered descriptor and
// encodes exactly one native command. Sending, retrying and confirmation belong
// to the session and the shared device pipeline.
import (
	"errors"
	"math"
	"regexp"
	"unicode/utf8"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"google.golang.org/protobuf/proto"
)

// EncodeCommand validates a requested value against a discovered writable
// descriptor and builds one native API command. It does not send or retry it.
// A button uses nil; numbers use float64, as decoded from JSON.
func EncodeCommand(d Descriptor, value any) (proto.Message, error) {
	invalid := errors.New("invalid or unsupported ESPHome command")
	if !d.Writable {
		return nil, invalid
	}
	switch d.Kind {
	case kindSwitch:
		v, ok := value.(bool)
		if !ok {
			return nil, invalid
		}
		return &api.SwitchCommandRequest{Key: d.Key, DeviceId: d.DeviceID, State: v}, nil
	case kindLight:
		v, ok := value.(bool)
		if !ok {
			return nil, invalid
		}
		return &api.LightCommandRequest{Key: d.Key, DeviceId: d.DeviceID, HasState: true, State: v}, nil
	case kindNumber:
		v, ok := value.(float64)
		if !ok || !finite(v) || !finite(d.Min) || !finite(d.Max) || !finite(d.Step) || d.Step <= 0 || v < d.Min || v > d.Max {
			return nil, invalid
		}
		steps := (v - d.Min) / d.Step
		// Float32 native metadata is approximate, so allow a small relative error.
		if math.Abs(steps-math.Round(steps)) > 1e-5*math.Max(1, math.Abs(steps)) || math.Abs(steps-math.Round(steps)) > 0.001 {
			return nil, invalid
		}
		return &api.NumberCommandRequest{Key: d.Key, DeviceId: d.DeviceID, State: float32(v)}, nil
	case kindSelect:
		v, ok := value.(string)
		if !ok {
			return nil, invalid
		}
		for _, option := range d.Options {
			if v == option {
				return &api.SelectCommandRequest{Key: d.Key, DeviceId: d.DeviceID, State: v}, nil
			}
		}
	case kindText:
		v, ok := value.(string)
		if !ok || !utf8.ValidString(v) || len(v) < int(d.TextMin) || len(v) > int(d.TextMax) || len(v) > maxTextValueBytes {
			return nil, invalid
		}
		if d.TextPattern != "" {
			pattern, err := regexp.Compile("^(?:" + d.TextPattern + ")$")
			if err != nil || !pattern.MatchString(v) {
				return nil, invalid
			}
		}
		return &api.TextCommandRequest{Key: d.Key, DeviceId: d.DeviceID, State: v}, nil
	case kindButton:
		if value == nil {
			return &api.ButtonCommandRequest{Key: d.Key, DeviceId: d.DeviceID}, nil
		}
	}
	return nil, invalid
}
