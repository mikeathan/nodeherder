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
	OnMessageHandler(handler func(string, []byte))
	Publish(friendlyName string, payload interface{})
}

const baseTopic string = "zigbee2mqtt/"

type MqttConfig struct {
	Broker   string
	Username string
	Password string
	Topics   []string
}

var _connectHandler mqttlib.OnConnectHandler = func(client mqttlib.Client) {
	fmt.Println("mqtt Client connected")
}

var _connectionLostHandler mqttlib.ConnectionLostHandler = func(client mqttlib.Client, err error) {
	fmt.Printf("mqtt Connection Lost: %s\n", err.Error())
}

func (m *MqttService) messagePubHandler() func(client mqttlib.Client, msg mqttlib.Message) {
	return func(client mqttlib.Client, msg mqttlib.Message) {
		//fmt.Printf("mqtt Message => Topic: %s, Payload; %s\n", msg.Topic(), msg.Payload())

		var name = sanitizeTopic(msg.Topic())
		var payload = msg.Payload()
		m.messageHandler(name, payload)
	}
}

func NewMqttClient(config MqttConfig) MqttClient {

	var client = &MqttService{
		broker:         config.Broker,
		topics:         []string{},
		username:       config.Username,
		password:       config.Password,
		client:         nil,
		cliendId:       "sinkhole-z2m",
		messageHandler: func(s string, b []byte) {},
	}

	for _, topic := range config.Topics {
		client.AddTopic(topic)
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
	messageHandler func(string, []byte)
}

func sanitizeTopic(topic string) string {
	return strings.Replace(topic, baseTopic, "", -1)
}

func (m *MqttService) OnMessageHandler(handler func(string, []byte)) {
	m.messageHandler = handler
}

func (m *MqttService) Publish(friendlyName string, payload interface{}) {
	topic := fmt.Sprintf("%s%s", baseTopic, friendlyName)

	fmt.Printf("publishing to %s \n", topic)
	m.client.Publish(topic, 0, false, payload)
}

func (m *MqttService) Connect() error {

	options := mqttlib.NewClientOptions()
	options.AddBroker(m.broker)
	options.SetClientID(m.cliendId)
	options.Username = m.username
	options.Password = m.password

	options.SetDefaultPublishHandler(m.messagePubHandler())
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
