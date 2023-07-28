package mqtt

import (
	"fmt"
	"strings"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient interface {
	Connect() error
	AddTopic(topic string)
	Disconnect()
	OnMessageHandler(handler func(client mqttlib.Client, msg mqttlib.Message))
}

const baseTopic string = "zigbee2mqtt/"

type MqttConfig struct {
	Broker         string
	Username       string
	Password       string
	Topics         []string
	MessageHandler func(client mqttlib.Client, msg mqttlib.Message)
}

var _messagePubHandler mqttlib.MessageHandler = func(client mqttlib.Client, msg mqttlib.Message) {
	fmt.Printf("mqtt Message => Topic: %s, Payload; %s\n", msg.Topic(), msg.Payload())
}

var _connectHandler mqttlib.OnConnectHandler = func(client mqttlib.Client) {
	fmt.Println("mqtt Client connected")
}

var _connectionLostHandler mqttlib.ConnectionLostHandler = func(client mqttlib.Client, err error) {
	fmt.Printf("mqtt Connection Lost: %s\n", err.Error())
}

func NewMqttClient(config MqttConfig) MqttClient {

	var client = &MqttService{
		broker:         config.Broker,
		topics:         []string{},
		username:       config.Username,
		password:       config.Password,
		client:         nil,
		cliendId:       "sinkhole-z2m",
		messageHandler: config.MessageHandler,
	}

	for _, topic := range config.Topics {
		client.AddTopic(topic)
	}
	// temporary
	if client.messageHandler == nil {
		client.messageHandler = _messagePubHandler
	}
	return client
}

type MqttService struct {
	topics         []string
	client         mqttlib.Client
	broker         string
	username       string
	password       string
	cliendId       string
	messageHandler func(client mqttlib.Client, msg mqttlib.Message)
}

func SanitizeTopic(topic string) string {
	return strings.Replace(topic, baseTopic, "", -1)
}

func (m *MqttService) OnMessageHandler(handler func(client mqttlib.Client, msg mqttlib.Message)) {
	m.messageHandler = handler
}

func (m *MqttService) Connect() error {

	options := mqttlib.NewClientOptions()
	options.AddBroker(m.broker)
	options.SetClientID(m.cliendId)
	options.Username = m.username
	options.Password = m.password

	options.SetDefaultPublishHandler(m.messageHandler)
	options.OnConnect = _connectHandler
	options.OnConnectionLost = _connectionLostHandler

	m.client = mqttlib.NewClient(options)
	token := m.client.Connect()

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	for _, device_name := range m.topics {
		// todo: pass full topic dont build them here
		topic := fmt.Sprintf("%s%s", baseTopic, device_name)
		token = m.client.Subscribe(topic, 1, nil)

		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
		fmt.Printf("subscribe zigbee2mqtt topic: %s\n", topic)
	}

	return nil
}

func (m *MqttService) AddTopic(topic string) {
	m.topics = append(m.topics, topic)
	fmt.Printf("Topic %s added\n", topic)
}

func (m *MqttService) Disconnect() {
	m.client.Disconnect(100)
	fmt.Println("zigbee2mqtt client disconnected")
}
