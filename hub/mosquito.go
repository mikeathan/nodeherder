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

type MqttOptions struct {
	broker   string
	username string
	password string
}

func NewMqttOptions(broker string, username string, password string) *MqttOptions {
	return &MqttOptions{
		broker:   broker,
		username: username,
		password: password,
	}
}

type MqttClient struct {
	options *MqttOptions
	topics  []string
	client  mqtt.Client
}

func (o *MqttOptions) options() *mqtt.ClientOptions {
	options := mqtt.NewClientOptions()
	options.AddBroker(o.broker)
	options.SetClientID("sinkhole-mikeathan")
	options.Username = o.username
	options.Password = o.password
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler
	return options
}

func NewMqttClient(options *MqttOptions) *MqttClient {
	return &MqttClient{
		options: options,
		topics:  []string{},
		client:  nil,
	}
}

func (m *MqttClient) AddTopic(topic string) {

	m.topics = append(m.topics, topic)
}

func (m *MqttClient) Connect() {
	options := m.options.options()
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
