package devices

import (
	"encoding/json"
	"errors"
	"fmt"
)

// type BridgeInfoFeature struct {
// 	Access      int    `json:"access"`
// 	Description string `json:"description"`
// 	Name        string `json:"name"`
// 	Property    string `json:"property"`
// 	Type        string `json:"type"`
// 	ValueOff    any    `json:"value_off,omitempty"`
// 	ValueOn     any    `json:"value_on,omitempty"`
// 	ValueToggle string `json:"value_toggle,omitempty"`
// 	ValueMax    any    `json:"value_max,omitempty"`
// 	ValueMin    any    `json:"value_min,omitempty"`
// 	Values      []any  `json:"values,omitempty"`
// 	Presets     []struct {
// 		Description string `json:"description"`
// 		Name        string `json:"name"`
// 		Value       int    `json:"value"`
// 	} `json:"presets,omitempty"`
// 	Unit string `json:"unit,omitempty"`
// }

type BridgeExpose struct {
	Type        string   `json:"type"`
	Access      int      `json:"access,omitempty"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Name        string   `json:"name,omitempty"`
	Property    string   `json:"property,omitempty"`
	Values      []string `json:"values,omitempty"`
	ValueOff    any      `json:"value_off,omitempty"`
	ValueOn     any      `json:"value_on,omitempty"`
	ValueToggle string   `json:"value_toggle,omitempty"`
	ValueMax    any      `json:"value_max,omitempty"`
	ValueMin    any      `json:"value_min,omitempty"`
	Presets     []struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Value       int    `json:"value"`
	} `json:"presets,omitempty"`
	Unit     string         `json:"unit,omitempty"`
	Features []BridgeExpose `json:"features,omitempty"`
}

type BridgeInfo struct {
	DateCode   string `json:"date_code"`
	Definition struct {
		Description string         `json:"description"`
		Exposes     []BridgeExpose `json:"exposes"`
		Model       string         `json:"model"`
		Options     []struct {
			Access      int    `json:"access"`
			Description string `json:"description"`
			Name        string `json:"name"`
			Property    string `json:"property"`
			Type        string `json:"type"`
			ValueMin    int    `json:"value_min,omitempty"`
			ValueOff    bool   `json:"value_off,omitempty"`
			ValueOn     bool   `json:"value_on,omitempty"`
		} `json:"options"`
		SupportsOta bool   `json:"supports_ota"`
		Vendor      string `json:"vendor"`
	} `json:"definition"`
	Disabled           bool   `json:"disabled"`
	FriendlyName       string `json:"friendly_name"`
	IeeeAddress        string `json:"ieee_address"`
	InterviewCompleted bool   `json:"interview_completed"`
	Interviewing       bool   `json:"interviewing"`
	Manufacturer       string `json:"manufacturer"`
	ModelID            string `json:"model_id"`
	NetworkAddress     int    `json:"network_address"`
	PowerSource        string `json:"power_source"`
	SoftwareBuildID    string `json:"software_build_id"`
	Supported          bool   `json:"supported"`
	Type               string `json:"type"`
}

func (b *BridgeInfo) IsActive() bool {

	return !b.Disabled && b.Type != "Coordinator" //&& b.InterviewCompleted
}

func (b *BridgeInfo) SanitiseProperty(name string, data any) (any, error) {
	for _, e := range b.Definition.Exposes {
		for _, f := range e.Features {

			if f.Property == name {
				sanitizedData, err := f.SanitizeData(data)
				if err != nil {
					return nil, errors.Join(fmt.Errorf("failed to sanitize feature data for property %s: %s", name, err.Error()))
				}
				return sanitizedData, nil
			}
		}

		if e.Property == name {
			sanitizedData, err := e.SanitizeData(data)
			if err != nil {
				return nil, errors.Join(fmt.Errorf("failed to sanitize expose data for property %s: %s", name, err.Error()))
			}
			return sanitizedData, nil

		}
	}

	return nil, fmt.Errorf("ailed to sanitize data. Property=%s not found", name)
}

func LoadBridgeDevices(payload []byte) ([]*BridgeInfo, error) {
	var bridgeDevices []*BridgeInfo
	err := json.Unmarshal(payload, &bridgeDevices)
	if err != nil {
		return nil, err
	}

	return bridgeDevices, nil
}

func FindByFriendlyName(payload []byte, friendlyName string) (*BridgeInfo, error) {
	bridgeDevices, err := LoadBridgeDevices(payload)
	if err != nil {
		return nil, err
	}

	for _, device := range bridgeDevices {
		if device.FriendlyName == friendlyName {
			return device, nil
		}
	}
	return nil, errors.New("friendlyName not found")
}

func FindAllExposesByCategory(payload []byte, category ExposeCategory) ([]*BridgeExpose, error) {
	bridgeDevices, err := LoadBridgeDevices(payload)
	if err != nil {
		return nil, err
	}
	exposes := []*BridgeExpose{}
	for _, device := range bridgeDevices {
		for _, expose := range device.Definition.Exposes {

			if expose.Category == "" && category == MeasurementCategory {
				exposes = append(exposes, &expose)
				continue
			}

			if expose.Category == category {
				exposes = append(exposes, &expose)
				continue
			}
		}

		for _, expose := range device.Definition.Exposes {

			for _, feature := range expose.Features {
				if feature.Category == "" && category == MeasurementCategory {
					exposes = append(exposes, &feature)
					continue
				}

				if feature.Category == category {
					exposes = append(exposes, &feature)
					continue
				}
			}
		}

	}

	return exposes, nil
}

func FindAllExposesByAccessMode(payload []byte, accesMode ExposeAccessMode) ([]*BridgeInfo, error) {
	bridgeDevices, err := LoadBridgeDevices(payload)
	if err != nil {
		return nil, err
	}
	devices := []*BridgeInfo{}
	for _, device := range bridgeDevices {
		for _, e := range device.Definition.Exposes {

			if e.Access == accesMode {
				devices = append(devices, device)
				continue
			}
		}
	}
	return devices, nil
}

func FindAllExposesByDataType(payload []byte, dataType ExposeDataType) ([]*BridgeInfo, error) {
	bridgeDevices, err := LoadBridgeDevices(payload)
	if err != nil {
		return nil, err
	}
	devices := []*BridgeInfo{}
	for _, device := range bridgeDevices {
		for _, e := range device.Definition.Exposes {

			if e.Type == dataType {
				devices = append(devices, device)
				continue
			}
		}
	}
	return devices, nil
}

func FindByExposeType(payload []byte, exposeType ExposeDataType) (*BridgeInfo, error) {
	bridgeDevices, err := LoadBridgeDevices(payload)
	if err != nil {
		return nil, err
	}

	for _, device := range bridgeDevices {
		for _, e := range device.Definition.Exposes {

			if e.Type == exposeType {
				return device, nil
			}
		}
	}
	return nil, errors.New("exposeType not found")
}

func (e *BridgeExpose) SanitizeData(data any) (any, error) {

	if e.Type == "binary" {
		if value, ok := data.(bool); ok {
			if value {
				return e.ValueOn, nil
			} else {
				return e.ValueOff, nil
			}
		}
	} else if e.Type == "numeric" {
		if value, ok := data.(int); ok {

			if min, ok := e.ValueMin.(int); ok {
				if value < min {
					return nil, fmt.Errorf("value=%d smaller than Minimum %d", value, min)
				}
			}

			if max, ok := e.ValueMax.(int); ok {
				if value > max {
					return nil, fmt.Errorf("value=%d bigger than Maximum %d", value, max)
				}
			}
		}
	} else {
		return nil, fmt.Errorf("type=%s  not implemented", e.Type)
	}

	return data, nil
}
