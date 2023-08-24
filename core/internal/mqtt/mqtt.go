package mqtt

import (
	"errors"
	"fmt"
	"strings"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient interface {
	Connect() error
	AddTopic(topic string) error
	Disconnect()
	OnMessageHandler(handler func(string, []byte))
	Publish(friendlyName string, payload interface{})
}

const baseTopic string = "zigbee2mqtt/"

var bridgeTopics = []string{
	"bridge/devices",
	"bridge/logging",
}

type MqttConfig struct {
	Broker   string
	Username string
	Password string
}

func (m *MqttService) onConnectedHandler() func(client mqttlib.Client) {
	return func(client mqttlib.Client) {
		fmt.Println("mqtt Client connected")
		m.subscribeTopics()
	}
}

func (m *MqttService) connectionLostHandler() func(client mqttlib.Client, err error) {
	return func(client mqttlib.Client, err error) {
		fmt.Printf("mqtt Connection Lost: %s\n", err.Error())
	}
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
		username:       config.Username,
		password:       config.Password,
		client:         nil,
		cliendId:       "sinkhole-z2m",
		messageHandler: func(s string, b []byte) {},
		topics:         bridgeTopics,
	}

	return client
}

type MqttService struct {
	client         mqttlib.Client
	broker         string
	username       string
	password       string
	cliendId       string
	messageHandler func(string, []byte)
	topics         []string
}

func sanitizeTopic(topic string) string {
	return strings.Replace(topic, baseTopic, "", -1)
}

func (m *MqttService) OnMessageHandler(handler func(string, []byte)) {
	m.messageHandler = handler
}

func (m *MqttService) Publish(friendlyName string, payload interface{}) {
	topic := fmt.Sprintf("%s%s", baseTopic, friendlyName)

	//fmt.Printf("Publish: %s \n", topic)
	m.client.Publish(topic, 0, false, payload)
}

func (m *MqttService) Connect() error {

	options := mqttlib.NewClientOptions()
	options.AddBroker(m.broker)
	options.SetClientID(m.cliendId)
	options.Username = m.username
	options.Password = m.password
	options.AutoReconnect = true

	options.SetDefaultPublishHandler(m.messagePubHandler())
	options.OnConnect = m.onConnectedHandler()
	options.OnConnectionLost = m.connectionLostHandler()
	options.SetReconnectingHandler(func(c mqttlib.Client, options *mqttlib.ClientOptions) {
		fmt.Println("...... mqtt reconnecting ......")
	})
	m.client = mqttlib.NewClient(options)
	token := m.client.Connect()

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MqttService) subscribeTopics() {
	// fix
	//  not currently connected and ResumeSubs not set
	for _, topic := range m.topics {
		fmt.Printf("Subscribe topic: %s\n", topic)
		err := m.subscribe(topic)
		if err != nil {
			fmt.Printf("%s failed: %s \n", topic, err.Error())
		}
	}
}

func (m *MqttService) subscribe(topic string) error {

	fullTopic := fmt.Sprintf("%s%s", baseTopic, topic)
	token := m.client.Subscribe(fullTopic, 1, nil)

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (m *MqttService) AddTopic(topic string) error {

	for _, t := range m.topics {
		if t == topic {
			return errors.New("topic is subscribed")
		}
	}
	fmt.Printf("Add topic: %s\n", topic)
	err := m.subscribe(topic)
	if err != nil {
		fmt.Printf("%s failed: %s \n", topic, err.Error())
	}
	m.topics = append(m.topics, topic)
	return nil
}

func (m *MqttService) Disconnect() {
	m.client.Disconnect(100)
	fmt.Println("zigbee2mqtt client disconnected")
}
