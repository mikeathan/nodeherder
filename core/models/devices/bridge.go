package devices

import (
	"encoding/json"
	"errors"
	"fmt"
)

type BridgeInfo struct {
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
				Values      []any  `json:"values,omitempty"`
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

func (b *BridgeInfo) IsActive() bool {

	return !b.Disabled && b.Type != "Coordinator" && b.InterviewCompleted
}

// DEBUG - delete
func LoadDevices(payload []byte) error {
	var bridgeDevices []*BridgeInfo
	err := json.Unmarshal(payload, &bridgeDevices)
	if err != nil {
		return err
	}
	for _, bd := range bridgeDevices {
		fmt.Println("friendlename:", bd.FriendlyName)
		fmt.Println("id", bd.IeeeAddress)
		fmt.Println("disabled:", bd.Disabled)
		fmt.Println("Type:", bd.Type)

		for _, expose := range bd.Definition.Exposes {
			fmt.Println("Property:", expose.Property)
			for _, feature := range expose.Features {
				fmt.Println("Feature Property:", feature.Property)
				fmt.Println("ValueMax:", feature.ValueMax)
				fmt.Println("ValueMin:", feature.ValueMin)
				fmt.Println("ValueOff:", feature.ValueOff)
				fmt.Println("ValueOn:", feature.ValueOn)
			}
		}

		fmt.Println("------------------------------")
	}

	return nil
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

func FindByExposeType(payload []byte, exposeType string) (*BridgeInfo, error) {
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

type BridgeFeature struct {
	Id         string                     `json:"id"`
	Properties map[string]*BridgeProperty `json:"properties"`
}

type BridgeProperty struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Attributes map[string]any `json:"attributes"`
}

func NewBridgeFeature(id string) *BridgeFeature {
	return &BridgeFeature{Id: id, Properties: make(map[string]*BridgeProperty)}
}

func (f *BridgeFeature) Add(property *BridgeProperty) {
	f.Properties[property.Name] = property
}

func NewBridgeProperty(name string, propertyType string) *BridgeProperty {
	return &BridgeProperty{Name: name, Type: propertyType, Attributes: make(map[string]any)}
}
