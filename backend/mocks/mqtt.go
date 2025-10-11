package mocks

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MockToggleMqttClient struct {
	handler   func(string, []byte)
	responses map[string]interface{}
	states    map[string]string // track binary device states
	mu        sync.Mutex
}

func NewMockAdvanceMqttClient() *MockToggleMqttClient {
	return &MockToggleMqttClient{
		responses: make(map[string]interface{}),
		states:    make(map[string]string),
	}
}

func (m *MockToggleMqttClient) AddResponse(topic string, resp interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[topic] = resp
}

func (m *MockToggleMqttClient) Publish(topic string, payload interface{}) {
	fmt.Println("Mock Publish")
	if strings.HasSuffix(topic, "/set") {
		m.handleCommand(strings.TrimSuffix(topic, "/set"), payload)
	} else {
		m.deliver(topic, payload)
	}
}


func (m *MockToggleMqttClient) handleCommand(baseTopic string, payload interface{}) {
	resp := m.resolveResponse(baseTopic, payload)

	go func() {
		time.Sleep(50 * time.Millisecond)
		m.deliver(baseTopic, resp) // device publishes new state
		m.deliver(baseTopic, resp) // optional broker echo
	}()
}

func (m *MockToggleMqttClient) deliver(topic string, payload interface{}) {
	data, _ := json.Marshal(payload)
	if m.handler != nil {
		m.handler(topic, data)
	}
}

func (m *MockToggleMqttClient) resolveResponse(baseTopic string, payload interface{}) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// pre-defined canned response wins
	if resp, ok := m.responses[baseTopic]; ok {
		return resp
	}

	// otherwise handle toggle logic
	var cmd map[string]interface{}
	_ = toJSON(payload, &cmd)
	if state, ok := cmd["state"].(string); ok && strings.EqualFold(state, "TOGGLE") {
		if m.states[baseTopic] == "ON" {
			m.states[baseTopic] = "OFF"
		} else {
			m.states[baseTopic] = "ON"
		}
		return map[string]any{"state": m.states[baseTopic]}
	}

	// fallback: echo back what was sent
	return cmd
}

func toJSON(src interface{}, dst any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func (m *MockToggleMqttClient) Connect() error {
	fmt.Println("Mock Connect")
	return nil
}

func (m *MockToggleMqttClient) WithMessageHandler(messageHandler func(client mqtt.Client, msg mqtt.Message)) {
	fmt.Println("Mock WithMessageHandler")
}

func (m *MockToggleMqttClient) AddTopic(topic string) error {
	fmt.Println("Mock ConfigureTopic")
	return nil
}

func (m *MockToggleMqttClient) RemoveTopic(topic string) error {
	fmt.Println("Mock RemoveTopic")
	return nil
}

func (m *MockToggleMqttClient) Disconnect() {
	fmt.Println("Mock Disconnect")
}

func (m *MockToggleMqttClient) OnMessageHandler(handler func(id string, payload []byte)) {
	m.handler = handler
}
