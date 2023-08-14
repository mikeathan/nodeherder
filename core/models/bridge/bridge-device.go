package bridge

import (
	"encoding/json"
	"errors"
)

type BridgeDevice struct {
	DateCode   string `json:"date_code"`
	Definition struct {
		Description string `json:"description"`
		Exposes     []struct {
			Features []struct {
				Access      int    `json:"access"`
				Description string `json:"description"`
				Name        string `json:"name"`
				Property    string `json:"property"`
				Type        string `json:"type"`
				ValueOff    string `json:"value_off,omitempty"`
				ValueOn     string `json:"value_on,omitempty"`
				ValueToggle string `json:"value_toggle,omitempty"`
				ValueMax    any    `json:"value_max,omitempty"`
				ValueMin    any    `json:"value_min,omitempty"`
				Presets     []struct {
					Description string `json:"description"`
					Name        string `json:"name"`
					Value       int    `json:"value"`
				} `json:"presets,omitempty"`
				Unit string `json:"unit,omitempty"`
			} `json:"features,omitempty"`
			Type        string   `json:"type"`
			Access      int      `json:"access,omitempty"`
			Description string   `json:"description,omitempty"`
			Name        string   `json:"name,omitempty"`
			Property    string   `json:"property,omitempty"`
			Values      []string `json:"values,omitempty"`
			Unit        string   `json:"unit,omitempty"`
			ValueMax    any      `json:"value_max,omitempty"`
			ValueMin    any      `json:"value_min,omitempty"`
		} `json:"exposes"`
		Model   string `json:"model"`
		Options []struct {
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

func Parse(payload []byte) ([]*BridgeDevice, error) {
	var bridgeDevices []*BridgeDevice
	err := json.Unmarshal(payload, &bridgeDevices)
	if err != nil {
		return nil, err
	}

	return bridgeDevices, nil
}

func FindByFriendlyName(payload []byte, friendlyName string) (*BridgeDevice, error) {
	bridgeDevices, err := Parse(payload)
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

func FindByExposeType(payload []byte, exposeType string) (*BridgeDevice, error) {
	bridgeDevices, err := Parse(payload)
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
