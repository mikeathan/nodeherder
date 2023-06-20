package hub

import (
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// type MqttClient interface{
// 	Connect() error
// 	AddDevice(deviceName string)
// 	Disconnect()
// }

const baseTopic string = "zigbee2mqtt/"

type device struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type MqttConfig struct {
	Broker   string
	Username string
	Password string
	Topics   []string
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var topic = msg.Topic()
	var payload = msg.Payload()
	fmt.Printf("DEBUG - mqtt message => Topic: %s, Payload; %s\n", topic, payload)

	// TODO: store device in map, global store
	//var devicePayload = device{Name: getDeviceName(topic), Payload: json.RawMessage(msg.Payload())}
	//Broadcast(DeviceUpdated, devicePayload)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("zigbee2mqtt client connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

func NewZ2MClient(config MqttConfig) *Z2MClient {

	var client = &Z2MClient{
		broker:   config.Broker,
		topics:   []string{},
		username: config.Username,
		password: config.Password,
		client:   nil,
		cliendId: "sinkhole-z2m",
	}
	for _, topic := range config.Topics {
		client.AddTopic(topic)
	}

	return client
}

type Z2MClient struct {
	topics   []string
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

	for _, device_name := range m.topics {
		topic := fmt.Sprintf("%s%s", baseTopic, device_name)
		token = m.client.Subscribe(topic, 1, nil)

		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
		fmt.Printf("subscribe zigbee2mqtt topic: %s\n", topic)
	}

	return nil
}

func (m *Z2MClient) AddTopic(topic string) {
	m.topics = append(m.topics, topic)
}

func (m *Z2MClient) Disconnect() {
	m.client.Disconnect(100)
	fmt.Println("zigbee2mqtt client disconnected")
}
