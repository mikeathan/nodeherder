package esphome

// This file maps native discovery messages to bounded descriptors. Unknown
// kinds become read-only descriptors; malformed or oversized metadata is
// rejected before it can reach the shared device model.
import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"unicode/utf8"

	"node-herder/models/bridge"

	"github.com/mycontroller-org/esphome_api/pkg/api"
	"google.golang.org/protobuf/proto"
)

// Descriptor is a discovery result, not a store of live device state. The ID is
// stable within its device, while Name, Unit and Options are untrusted metadata.
type Descriptor struct {
	ID, Name, Kind, Unit, Category string
	DeviceClass                    string
	Key, DeviceID                  uint32
	Writable                       bool
	LimitedControl                 bool
	Min, Max, Step                 float64
	Options                        []string
	TextMin, TextMax               uint32
	TextPattern                    string
}

// Describe maps supported native discovery messages to bounded descriptors. It
// returns (nil, nil) for other messages, and rejects unsafe or invalid metadata.
func Describe(message proto.Message) (*Descriptor, error) {
	var d Descriptor
	if message == nil || !message.ProtoReflect().IsValid() {
		return nil, errors.New("invalid ESPHome discovery message")
	}
	switch m := message.(type) {
	case *api.ListEntitiesSensorResponse:
		d.Key, d.DeviceID, d.Name, d.Kind, d.Unit = m.Key, m.DeviceId, m.Name, kindSensor, m.UnitOfMeasurement
	case *api.ListEntitiesBinarySensorResponse:
		d.Key, d.DeviceID, d.Name, d.Kind = m.Key, m.DeviceId, m.Name, kindBinarySensor
	case *api.ListEntitiesTextSensorResponse:
		d.Key, d.DeviceID, d.Name, d.Kind = m.Key, m.DeviceId, m.Name, kindTextSensor
	case *api.ListEntitiesTextResponse:
		d.Key, d.DeviceID, d.Name, d.Kind = m.Key, m.DeviceId, m.Name, kindText
		if m.Mode == api.TextMode_TEXT_MODE_PASSWORD {
			d.Kind = kindSecretText
			break
		}
		d.Writable, d.TextMin, d.TextMax, d.TextPattern = true, m.MinLength, m.MaxLength, m.Pattern
		if d.TextMin > d.TextMax || d.TextMax > maxTextValueBytes || len(d.TextPattern) > maxTextPatternBytes || !utf8.ValidString(d.TextPattern) {
			return nil, errors.New("invalid ESPHome text constraints")
		}
		if d.TextPattern != "" {
			if _, err := regexp.Compile("^(?:" + d.TextPattern + ")$"); err != nil {
				// Preserve its state without accepting commands we cannot validate safely.
				d.Writable = false
			}
		}
	case *api.ListEntitiesSwitchResponse:
		d.Key, d.DeviceID, d.Name, d.Kind, d.Writable = m.Key, m.DeviceId, m.Name, kindSwitch, true
	case *api.ListEntitiesLightResponse:
		d.Key, d.DeviceID, d.Name, d.Kind, d.Writable = m.Key, m.DeviceId, m.Name, kindLight, true
		d.LimitedControl = m.LegacySupportsBrightness || m.LegacySupportsRgb || m.LegacySupportsWhiteValue || m.LegacySupportsColorTemperature || len(m.Effects) > 0
		for _, mode := range m.SupportedColorModes {
			if mode != api.ColorMode_COLOR_MODE_UNKNOWN && mode != api.ColorMode_COLOR_MODE_ON_OFF {
				d.LimitedControl = true
			}
		}
	case *api.ListEntitiesNumberResponse:
		d.Key, d.DeviceID, d.Name, d.Kind, d.Unit, d.Writable = m.Key, m.DeviceId, m.Name, kindNumber, m.UnitOfMeasurement, true
		d.Min, d.Max, d.Step = float64(m.MinValue), float64(m.MaxValue), float64(m.Step)
		if !finite(d.Min) || !finite(d.Max) || !finite(d.Step) || d.Min > d.Max || d.Step <= 0 {
			return nil, errors.New("invalid ESPHome number constraints")
		}
	case *api.ListEntitiesSelectResponse:
		d.Key, d.DeviceID, d.Name, d.Kind, d.Writable = m.Key, m.DeviceId, m.Name, kindSelect, true
		if len(m.Options) == 0 || len(m.Options) > maxSelectOptions {
			return nil, errors.New("invalid ESPHome select options")
		}
		d.Options = append([]string(nil), m.Options...)
	case *api.ListEntitiesButtonResponse:
		d.Key, d.DeviceID, d.Name, d.Kind, d.Writable = m.Key, m.DeviceId, m.Name, kindButton, true
	default:
		// Unknown discovery kinds remain visible but read-only. Plain protocol
		// messages lack this descriptor interface and are ignored.
		entity, ok := message.(interface {
			GetKey() uint32
			GetDeviceId() uint32
			GetName() string
		})
		if !ok {
			return nil, nil
		}
		d.Key, d.DeviceID, d.Name, d.Kind = entity.GetKey(), entity.GetDeviceId(), entity.GetName(), kindUnsupported
	}
	if classified, ok := message.(interface{ GetDeviceClass() string }); ok {
		d.DeviceClass = classified.GetDeviceClass()
		if !validText(d.DeviceClass) {
			return nil, errors.New("invalid ESPHome device class")
		}
	}
	if categorized, ok := message.(interface{ GetEntityCategory() api.EntityCategory }); ok {
		switch categorized.GetEntityCategory() {
		case api.EntityCategory_ENTITY_CATEGORY_CONFIG:
			d.Category = bridge.ConfigCategory
		case api.EntityCategory_ENTITY_CATEGORY_DIAGNOSTIC:
			d.Category = bridge.DiagnosticCategory
		}
	}
	if !validText(d.Name) || !validText(d.Unit) {
		return nil, errors.New("invalid ESPHome entity metadata")
	}
	for _, option := range d.Options {
		if !validText(option) {
			return nil, errors.New("invalid ESPHome select option")
		}
	}
	d.ID = fmt.Sprintf("%08x-%08x", d.DeviceID, d.Key)
	return &d, nil
}

func finite(n float64) bool   { return !math.IsNaN(n) && !math.IsInf(n, 0) }
func validText(s string) bool { return len(s) <= maxEntityNameBytes && utf8.ValidString(s) }
