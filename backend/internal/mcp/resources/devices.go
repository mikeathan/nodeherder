package resources

import (
	"encoding/json"
	"node-herder/models/bridge"
	"node-herder/models/devices"
)

// DeviceContextResponse matches the format in docs/device_context.json.
type DeviceContextResponse struct {
	Version     string          `json:"version"`
	GeneratedAt string          `json:"generatedAt"`
	Devices     []DeviceContext `json:"devices"`
}

// DeviceContext represents a device in the context response.
type DeviceContext struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Exposes     []ExposeContext `json:"exposes"`
}

// ExposeContext represents an expose in the context.
type ExposeContext struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Unit         string   `json:"unit,omitempty"`
	Values       []string `json:"values,omitempty"`
	ValueOn      any      `json:"valueOn,omitempty"`
	ValueOff     any      `json:"valueOff,omitempty"`
	ValueToggle  any      `json:"valueToggle,omitempty"`
	Aggregations []string `json:"aggregations"`
}

// DeviceStore provides access to devices for resource generation.
type DeviceStore interface {
	AllDevices() ([]*devices.Device, error)
}

// DevicesResource provides the device context as an MCP resource.
type DevicesResource struct {
	store DeviceStore
}

// NewDevicesResource creates a new DevicesResource.
func NewDevicesResource(store DeviceStore) *DevicesResource {
	return &DevicesResource{store: store}
}

// GetContent returns the device context as JSON.
func (r *DevicesResource) GetContent() ([]byte, error) {
	devs, err := r.store.AllDevices()
	if err != nil {
		return nil, err
	}

	response := buildDeviceContext(devs)
	return json.Marshal(response)
}

func buildDeviceContext(devs []*devices.Device) *DeviceContextResponse {
	contexts := make([]DeviceContext, 0, len(devs))

	for _, d := range devs {
		ctx := DeviceContext{
			ID:          d.Id,
			Name:        d.FriendlyName,
			Description: d.Description,
			Exposes:     buildExposes(d.Exposes),
		}
		contexts = append(contexts, ctx)
	}

	return &DeviceContextResponse{
		Version: "1",
		Devices: contexts,
	}
}

func buildExposes(exposes map[string]*devices.Entity) []ExposeContext {
	result := make([]ExposeContext, 0, len(exposes))

	for _, e := range exposes {
		ctx := ExposeContext{
			Name: e.Name,
			Type: string(e.Type),
			Unit: e.Unit,
		}

		// Build aggregations based on type
		switch e.Type {
		case bridge.NumericDataType:
			ctx.Aggregations = []string{"last", "min", "max", "avg"}
		case bridge.BinaryDataType:
			ctx.Aggregations = []string{"last", "count"}
			if v, ok := e.Values["on"]; ok {
				ctx.ValueOn = v
			}
			if v, ok := e.Values["off"]; ok {
				ctx.ValueOff = v
			}
			if v, ok := e.Values["toggle"]; ok {
				ctx.ValueToggle = v
			}
		case bridge.EnumDataType:
			ctx.Aggregations = []string{"last", "count"}
			for _, v := range e.Values {
				if s, ok := v.(string); ok {
					ctx.Values = append(ctx.Values, s)
				}
			}
		}

		result = append(result, ctx)
	}

	return result
}
