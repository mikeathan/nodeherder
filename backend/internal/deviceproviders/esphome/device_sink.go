package esphome

// This file is the adapter between a native session and the shared device
// pipeline. It converts descriptors into the existing entity model, binds the
// physical identity before publishing inventory and forwards confirmed reports.
import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"node-herder/models/bridge"
	"node-herder/models/devices"

	"github.com/google/uuid"
	types "github.com/mycontroller-org/esphome_api/pkg/types"
)

// DevicePipeline is the shared device lifecycle boundary. Its methods serialize
// registration, observations and offline transitions for the same device.
type DevicePipeline interface {
	RegisterProviderDevice(context.Context, *devices.Device) error
	UpdateProviderDevice(context.Context, string, map[string]interface{}) error
	ProviderDeviceOffline(context.Context, string) error
	RemoveProviderDevice(context.Context, string) error
}

// DeviceSink adapts one ESPHome device's session to the existing device pipeline.
// Its methods are called sequentially by one supervisor, not concurrently.
type DeviceSink struct {
	pipeline      DevicePipeline
	id, name      string
	pending       *devices.Device
	registered    bool
	automaticName bool
	identity      func(context.Context, string) error
	inventory     func(context.Context, map[string]string) error
	metadata      map[string]interface{}
}

// Identify binds the enrollment to the device before publishing its inventory.
func (s *DeviceSink) Identify(ctx context.Context, info *types.DeviceInfo) error {
	if info == nil || s.identity == nil {
		return errors.New("ESPHome device identity is unavailable")
	}
	mac, err := net.ParseMAC(info.MacAddress)
	if err != nil || len(mac) != 6 || mac[0]&1 != 0 {
		return errors.New("invalid ESPHome device identity")
	}
	zero := true
	for _, b := range mac {
		zero = zero && b == 0
	}
	if zero {
		return errors.New("invalid ESPHome device identity")
	}
	if err := s.identity(ctx, mac.String()); err != nil {
		return err
	}
	s.metadata = make(map[string]interface{})
	for key, value := range map[string]string{"manufacturer": info.Manufacturer, "model": info.Model, "firmware_version": info.EsphomeVersion, "project_name": info.ProjectName, "project_version": info.ProjectVersion} {
		if value != "" && validText(value) && !strings.ContainsRune(value, '\x00') {
			s.metadata[key] = value
		}
	}
	if s.automaticName {
		for _, name := range []string{info.FriendlyName, info.Name} {
			name = strings.TrimSpace(name)
			if name != "" && validText(name) && !strings.ContainsRune(name, '\x00') {
				s.name = name
				break
			}
		}
	}
	return nil
}

// NewDeviceSink binds an enrollment UUID and display name to the shared pipeline.
// The caller owns the sink and must not call it from multiple sessions concurrently.
func NewDeviceSink(pipeline DevicePipeline, id, name string) (*DeviceSink, error) {
	parsed, err := uuid.Parse(id)
	if pipeline == nil || err != nil || parsed.String() != id || name == "" || !validText(name) {
		return nil, errors.New("invalid provider device identity")
	}
	return &DeviceSink{pipeline: pipeline, id: id, name: name}, nil
}

// Inventory prepares a new device generation without publishing it until Ready.
func (s *DeviceSink) Inventory(ctx context.Context, descriptors []Descriptor) error {
	if len(descriptors) > maxEntities {
		return errors.New("too many ESPHome entities")
	}
	if s.inventory != nil {
		kinds := make(map[string]string, len(descriptors))
		for _, descriptor := range descriptors {
			kinds[descriptor.ID] = descriptor.Kind
		}
		if err := s.inventory(ctx, kinds); err != nil {
			return err
		}
	}
	d := devices.NewDevice(s.id)
	d.Provider, d.ConnectionType, d.FriendlyName = ProviderName, ProviderName, s.name
	for key, value := range s.metadata {
		d.Properties[key] = value
	}
	for _, descriptor := range descriptors {
		entity, err := deviceEntity(descriptor)
		if err != nil {
			return err
		}
		if _, exists := d.Exposes[entity.Name]; exists {
			return errors.New("duplicate ESPHome entity")
		}
		d.Exposes[entity.Name] = entity
	}
	s.pending = d
	return nil
}

// Ready commits the prepared inventory after subscription succeeds.
func (s *DeviceSink) Ready(ctx context.Context) error {
	if s.pending == nil {
		return errors.New("ESPHome inventory unavailable")
	}
	if err := s.pipeline.RegisterProviderDevice(ctx, s.pending); err != nil {
		return err
	}
	s.registered = true
	s.pending = nil
	return nil
}

// Observe forwards one confirmed report; unavailable observations become nil.
func (s *DeviceSink) Observe(ctx context.Context, observation Observation) error {
	if !s.registered {
		return errors.New("ESPHome device is not registered")
	}
	value := observation.Value
	if !observation.Available {
		value = nil
	}
	return s.pipeline.UpdateProviderDevice(ctx, s.id, map[string]interface{}{observation.EntityID: value})
}

// Status propagates disconnection without claiming stale values are live.
func (s *DeviceSink) Status(ctx context.Context, status ConnectionStatus) error {
	if status == Offline && s.registered {
		return s.pipeline.ProviderDeviceOffline(ctx, s.id)
	}
	return nil
}

func deviceEntity(d Descriptor) (*devices.Entity, error) {
	if d.ID != fmt.Sprintf("%08x-%08x", d.DeviceID, d.Key) || !validText(d.Name) || !validText(d.Unit) {
		return nil, errors.New("invalid ESPHome descriptor")
	}
	e := devices.NewEntity(d.ID)
	e.Description, e.Unit = d.Name, d.Unit
	e.Attributes[attributeLabel] = d.Name
	e.Attributes[attributeEntityKind] = d.Kind
	if d.ObjectID != "" {
		e.Attributes[attributeObjectID] = d.ObjectID
	}
	if d.DeviceClass != "" {
		e.Attributes[attributeDeviceClass] = d.DeviceClass
	}
	if d.LimitedControl {
		e.Attributes[attributeControlUnavailable] = "Only on/off control is supported for this light."
	}
	if d.Kind == kindUnsupported {
		e.Attributes[attributeControlUnavailable] = "This entity type is not supported yet."
	}
	e.Category = d.Category
	if e.Category == "" {
		if d.Writable && d.Kind != kindSwitch && d.Kind != kindLight {
			e.Category = bridge.ConfigCategory
		} else {
			e.Category = bridge.MeasurementCategory
		}
	}
	e.AccessMode = bridge.ReadAccessMode
	switch d.Kind {
	case kindSensor:
		e.Type = bridge.NumericDataType
	case kindBinarySensor, kindSwitch, kindLight:
		e.Type = bridge.BinaryDataType
		e.Values[bridge.ValueKeyOn], e.Values[bridge.ValueKeyOff] = true, false
	case kindTextSensor, kindSecretText, kindUnsupported:
		e.Type = bridge.TextDataType
	case kindText:
		e.Type = bridge.TextDataType
		e.Attributes[attributeMinLength], e.Attributes[attributeMaxLength] = d.TextMin, d.TextMax
		if !d.Writable {
			e.Attributes[attributeControlUnavailable] = "This device's text validation is not supported; the value is read-only."
		}
	case kindNumber:
		if !finite(d.Min) || !finite(d.Max) || !finite(d.Step) || d.Min > d.Max || d.Step <= 0 {
			return nil, errors.New("invalid ESPHome number constraints")
		}
		e.Type = bridge.NumericDataType
		e.Attributes[attributeMin], e.Attributes[attributeMax], e.Attributes[attributeStep] = d.Min, d.Max, d.Step
	case kindSelect:
		if len(d.Options) == 0 || len(d.Options) > maxSelectOptions {
			return nil, errors.New("invalid ESPHome options")
		}
		e.Type = bridge.EnumDataType
		for i, option := range d.Options {
			if !validText(option) {
				return nil, errors.New("invalid ESPHome option")
			}
			e.Values[strconv.Itoa(i)] = option
		}
	case kindButton:
		e.Type = bridge.ButtonDataType
	default:
		return nil, errors.New("unsupported ESPHome descriptor")
	}
	if d.Writable {
		e.AccessMode = bridge.ReadWriteAccessMode
	}
	if d.Kind == kindButton {
		e.AccessMode = bridge.WriteAccessMode
	}
	return e, nil
}
