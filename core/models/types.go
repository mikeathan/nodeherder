package models

import "encoding/json"

type Device struct {
	Name    string      `json:"name"`
	Payload interface{} `json:"payload"`
}

func (d Device) MarshalJSON() ([]byte, error) {
	p := d.Payload
	if v, ok := d.Payload.([]byte); ok {
		p = string(v)
	}
	return json.Marshal(&struct {
		Name    string      `json:"name"`
		Payload interface{} `json:"payload"`
	}{
		Name:    d.Name,
		Payload: p,
	})
}
