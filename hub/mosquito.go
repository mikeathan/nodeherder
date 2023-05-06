package hub

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Message %s received on topic %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

type MqttClient struct {
	broker   string
	username string
	password string
	topics   []string
	client   mqtt.Client
}

func NewMqttClient(broker string, username string, password string) *MqttClient {
	return &MqttClient{
		broker:   broker,
		username: username,
		password: password,
		topics:   []string{},
		client:   nil,
	}
}

func (m *MqttClient) AddTopic(topic string) {
	m.topics = append(m.topics, topic)
}

func (m *MqttClient) Connect() {
	options := mqtt.NewClientOptions()
	options.AddBroker(m.broker)
	options.SetClientID("sinkhole-mikeathan")
	options.Username = m.username
	options.Password = m.password
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	m.client = mqtt.NewClient(options)
	token := m.client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for _, topic := range m.topics {
		token = m.client.Subscribe(topic, 1, nil)
		token.Wait()
		fmt.Printf("topic %s\n", topic)
	}
}

func (m *MqttClient) Disconnect() {
	m.client.Disconnect(100)
}
