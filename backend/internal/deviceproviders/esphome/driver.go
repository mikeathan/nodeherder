package esphome

// This file constructs ESPHome provider instances for the shared manager. It
// validates enrollment configuration without performing network I/O.
import (
	"context"
	"errors"

	"node-herder/internal/deviceproviders"
)

// Driver constructs ESPHome instances for the shared provider manager.
type Driver struct{ pipeline DevicePipeline }

// NewDriver binds ESPHome sessions to the shared device lifecycle pipeline.
func NewDriver(pipeline DevicePipeline) *Driver { return &Driver{pipeline: pipeline} }

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
	return &driverInstance{endpoint: endpoint, key: enrollment.Credential, sink: sink, supervisor: NewSupervisor()}, nil
}

type driverInstance struct {
	endpoint   deviceproviders.Endpoint
	key        string
	sink       *DeviceSink
	supervisor *Supervisor
}

func (i *driverInstance) SetIdentityValidator(validate func(context.Context, string) error) {
	i.sink.identity = validate
}

func (i *driverInstance) SetInventoryValidator(validate func(context.Context, map[string]string) error) {
	i.sink.inventory = validate
}

func (i *driverInstance) Run(ctx context.Context) error {
	return i.supervisor.Run(ctx, i.endpoint, i.key, i.sink)
}
func (i *driverInstance) Command(ctx context.Context, id string, value any) error {
	return i.supervisor.Command(ctx, id, value)
}

func (i *driverInstance) Health() deviceproviders.Health { return i.supervisor.Health() }
