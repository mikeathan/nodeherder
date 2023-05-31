package hub_test

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Topic: zigbee2mqtt/TH1, Payload; {"battery":100,"humidity":59.8,"last_seen":"2023-05-31T19:02:28+01:00","linkquality":51,"temperature":18.4,"voltage":3000}
type device struct {
	Name    string           `json:"name"`
	Payload *json.RawMessage `json:"payload"`
}

func TestXxx(t *testing.T) {
	var topic = "zigbee2mqtt/TH1"
	var payload = []byte("{'battery':100,'humidity':59.8,'last_seen':'2023-05-31T19:02:28+01:00','linkquality':51,'temperature':18.4,'voltage':3000}")
	payload1 := (*json.RawMessage)(&payload)
	var device = device{Name: topic, Payload: payload1}
	bytes, err := json.Marshal(device)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}
