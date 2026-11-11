package esphome

// This file constructs ESPHome provider instances for the shared manager. It
// validates enrollment configuration without performing network I/O.
import (
	"context"
	"errors"

	"node-herder/internal/deviceproviders"
	model "node-herder/models/observations"
)

// Driver constructs ESPHome instances for the shared provider manager.
type Driver struct {
	pipeline DevicePipeline
	observer Observer
}

// NewDriver binds ESPHome sessions to the shared device lifecycle pipeline and, when
// non-nil, to an observation consumer. A nil observer keeps the driver state-only, which is
// the behaviour before observations existed.
func NewDriver(pipeline DevicePipeline) *Driver { return &Driver{pipeline: pipeline} }

// WithObservationConsumer enables Bluetooth advertisement observation for every instance this
// driver creates. It is a functional method so the existing constructor signature is preserved.
func (d *Driver) WithObservationConsumer(observer Observer) *Driver {
	d.observer = observer
	return d
}

// RemoveDevice removes a stopped ESPHome device from the shared registry.
func (d *Driver) RemoveDevice(ctx context.Context, id string) error {
	if d.pipeline == nil {
		return errors.New("device pipeline is unavailable")
	}
	return d.pipeline.RemoveProviderDevice(ctx, id)
}

// SetDeviceOffline marks a stopped ESPHome device unavailable during reconfiguration.
func (d *Driver) SetDeviceOffline(ctx context.Context, id string) error {
	if d.pipeline == nil {
		return errors.New("device pipeline is unavailable")
	}
	return d.pipeline.ProviderDeviceOffline(ctx, id)
}

// NewInstance validates configuration without connecting to the device.
func (d *Driver) NewInstance(enrollment deviceproviders.Enrollment) (deviceproviders.Instance, error) {
	if d.pipeline == nil || enrollment.Provider != ProviderName {
		return nil, errors.New("invalid ESPHome enrollment")
	}
	endpoint, err := deviceproviders.NewEndpoint(enrollment.Address, enrollment.Port)
	if err != nil {
		return nil, err
	}
	if err := ValidateEncryptionKey(enrollment.Credential); err != nil {
		return nil, err
	}
	name := enrollment.Name
	if name == "" {
		name = "ESPHome " + enrollment.ID
	}
	// An empty configured name permits the reported friendly name to replace this
	// safe generated label after the first DeviceInfo response.
	sink, err := NewDeviceSink(d.pipeline, enrollment.ID, name)
	if err != nil {
		return nil, err
	}
	sink.automaticName = enrollment.Name == ""
	// sessionSink is what the session sees. Observation is opt-in per driver, and wrapping is
	// what makes the capability structural: without an observer the session's interface
	// assertion fails and no advertisement subscription is sent.
	var sessionSink SupervisorSink = sink
	if d.observer != nil {
		sessionSink = &observingSink{
			DeviceSink: sink,
			observer:   d.observer,
			sourceID:   enrollment.ID,
			sourceName: name,
		}
		// The registered device advertises the capability so the browser can gate its UI
		// before any advertisement arrives.
		sink.capabilities = []string{model.CapabilityWirelessObserve}
	}
	return &driverInstance{endpoint: endpoint, key: enrollment.Credential, sink: sink, sessionSink: sessionSink, supervisor: NewSupervisor()}, nil
}

type driverInstance struct {
	endpoint deviceproviders.Endpoint
	key      string
	// sink owns the identity and inventory validators.
	sink *DeviceSink
	// sessionSink is the same sink, optionally wrapped to add observation.
	sessionSink SupervisorSink
	supervisor  *Supervisor
}

func (i *driverInstance) SetIdentityValidator(validate func(context.Context, string) error) {
	i.sink.identity = validate
}

func (i *driverInstance) SetInventoryValidator(validate func(context.Context, map[string]string) error) {
	i.sink.inventory = validate
}

func (i *driverInstance) Run(ctx context.Context) error {
	return i.supervisor.Run(ctx, i.endpoint, i.key, i.sessionSink)
}
func (i *driverInstance) Command(ctx context.Context, id string, value any) error {
	return i.supervisor.Command(ctx, id, value)
}

func (i *driverInstance) Health() deviceproviders.Health { return i.supervisor.Health() }
