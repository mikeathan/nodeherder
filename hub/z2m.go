package hub

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type device struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var topic = msg.Topic()
	var payload = msg.Payload()
	fmt.Printf("DEBUG - mqtt message => Topic: %s, Payload; %s\n", topic, payload)

	// TODO: store device in map, global store
	var devicePayload = device{Name: getDeviceName(topic), Payload: json.RawMessage(msg.Payload())}
	Broadcast(DeviceUpdated, devicePayload)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("zigbee2mqtt client connected")
	// TODO: send devices from global store
	// Broadcast(ClientConnected, nil)
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

func NewZ2MClient(broker string, username string, password string) *Z2MClient {

	return &Z2MClient{
		broker:   broker,
		username: username,
		password: password,
		devices:  []string{}, // not used , remove ???
		client:   nil,
		cliendId: "sinkhole-z2m",
	}
}

var baseTopic string = "zigbee2mqtt/"

type Z2MClient struct {
	devices  []string // not used , remove ???
	client   mqtt.Client
	broker   string
	username string
	password string
	cliendId string
}

func getDeviceName(topic string) string {
	return strings.Replace(topic, baseTopic, "", -1)
}

func (m *Z2MClient) Connect() error {

	options := mqtt.NewClientOptions()
	options.AddBroker(m.broker)
	options.SetClientID(m.cliendId)
	options.Username = m.username
	options.Password = m.password

	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	m.client = mqtt.NewClient(options)
	token := m.client.Connect()

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	for _, device_name := range m.devices {
		topic := fmt.Sprintf("%s%s", baseTopic, device_name)
		token = m.client.Subscribe(topic, 1, nil)

		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
		fmt.Printf("subscribe zigbee2mqtt topic: %s\n", topic)
	}

	return nil
}

func (m *Z2MClient) AddDevice(device_name string) {
	m.devices = append(m.devices, device_name)
}

func (m *Z2MClient) Disconnect() {
	m.client.Disconnect(100)
	fmt.Println("zigbee2mqtt client disconnected")
}
