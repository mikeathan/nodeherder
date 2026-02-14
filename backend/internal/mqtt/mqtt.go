package mqtt

import (
	"fmt"
	"node-herder/utils"
	"os"
	"strings"
	"sync"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

type MqttClient interface {
	Connect() error
	AddTopic(topic string) error
	RemoveTopic(topic string) error
	Disconnect()
	OnMessageHandler(handler func(string, []byte))
	Publish(friendlyName string, payload interface{})
}

const baseTopic string = "zigbee2mqtt/"

var bridgeTopics = []string{
	"bridge/devices",
	"bridge/response/device/rename",
	"bridge/response/device/remove",
	"bridge/response/device/interview",
	"bridge/response/permit_join",
	"bridge/logging",
}

type MqttConfig struct {
	Broker     string
	Username   string
	Password   string
	ClientType string
	ClientId   string
}

func (m *MqttService) onConnectedHandler() func(client mqttlib.Client) {
	return func(client mqttlib.Client) {
		//defer m.mu.Unlock()
		//m.mu.Lock()

		utils.LogInfof("mqtt Client connected")
		m.subscribeTopics()
	}
}

func (m *MqttService) connectionLostHandler() func(client mqttlib.Client, err error) {
	return func(client mqttlib.Client, err error) {
		utils.LogDebugf("mqtt Connection Lost: %s", err.Error())
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

func WithDefaultMqttConfig() func(c *MqttConfig) {
	return func(c *MqttConfig) {
		broker, err := utils.GetMQTTBrokerURL()
		if err != nil {
			utils.LogErrorf("error getting MQTT broker URL: %v", err.Error())
			broker = "tcp://localhost:1883"
		}

		username, password := utils.GetMQTTBrokerCredentials()
		if username == "" || password == "" {
			utils.LogError("MQTT broker username or password is not set")
		}

		clientId := utils.GetMQTTClientID()
		if clientId == "" {
			utils.LogError("MQTT_CLIENT_ID is not set")
		}
		c.Password = password
		c.Username = username
		c.Broker = broker

		// In development mode, append -dev to client ID to avoid conflicts with other running instances
		if strings.ToLower(os.Getenv("APP_ENV")) == "development" {
			clientId = fmt.Sprintf("%s-dev", clientId)
			utils.LogInfof("Using Development MQTT Client ID: %s", clientId)
		}
		c.ClientId = clientId
	}
}

func NewMqttClient(opts ...func(*MqttConfig)) MqttClient {

	var config = &MqttConfig{}
	for _, opt := range opts {
		opt(config)
	}

	var client = &MqttService{
		broker:         config.Broker,
		username:       config.Username,
		password:       config.Password,
		client:         nil,
		clientId:       config.ClientId,
		messageHandler: func(s string, b []byte) {},
		topics:         bridgeTopics,
		mu:             sync.Mutex{},
	}

	if config.ClientType != "" {
		client.clientId += "-" + config.ClientType
	}

	return client
}

type MqttService struct {
	client         mqttlib.Client
	broker         string
	username       string
	password       string
	clientId       string
	messageHandler func(string, []byte)
	topics         []string
	mu             sync.Mutex
}

func sanitizeTopic(topic string) string {
	return strings.Replace(topic, baseTopic, "", -1)
}

func (m *MqttService) OnMessageHandler(handler func(string, []byte)) {
	m.messageHandler = handler
}

func (m *MqttService) Publish(friendlyName string, payload interface{}) {
	topic := fmt.Sprintf("%s%s", baseTopic, friendlyName)

	utils.LogDebugf("Publish: %s", topic)
	m.client.Publish(topic, 0, false, payload)
}

func (m *MqttService) Connect() error {

	options := mqttlib.NewClientOptions()
	options.AddBroker(m.broker)
	options.SetClientID(m.clientId)
	options.Username = m.username
	options.Password = m.password
	options.SetOrderMatters(false)       // Allow out of order messages (use this option unless in order delivery is essential)
	options.ConnectTimeout = time.Second // Minimal delays on connect
	options.WriteTimeout = time.Second   // Minimal delays on writes
	options.KeepAlive = 10               // Keepalive every 10 seconds so we quickly detect network outages
	options.PingTimeout = time.Second    // local broker so response should be quick

	options.ConnectRetry = true
	options.AutoReconnect = true

	options.SetDefaultPublishHandler(m.messagePubHandler())
	options.OnConnect = m.onConnectedHandler()
	options.OnConnectionLost = m.connectionLostHandler()
	options.SetReconnectingHandler(func(c mqttlib.Client, options *mqttlib.ClientOptions) {
		utils.LogDebug("...... mqtt reconnecting ......")
	})

	m.client = mqttlib.NewClient(options)
	token := m.client.Connect()

	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m *MqttService) subscribeTopics() {

	for _, topic := range m.topics {
		utils.LogInfof("Subscribe topic: %s", topic)
		err := m.subscribe(topic)
		if err != nil {
			utils.LogErrorf("%s failed: %s", topic, err.Error())
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

func (m *MqttService) RemoveTopic(topic string) error {

	for idx, t := range m.topics {
		if t == topic {

			if token := m.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
				return token.Error()
			}

			newtopics := append(m.topics[:idx], m.topics[idx+1:]...)
			m.topics = newtopics
			utils.LogInfof("Remove topic: %s", t)

			return nil
		}
	}

	return fmt.Errorf("topic %s not found", topic)
}

func (m *MqttService) AddTopic(topic string) error {
	//todo: remove unused topics if got renamed
	for _, t := range m.topics {
		if t == topic {
			//utils.LogDebugf("skipping topic %s is already subscribed", t)
			return nil
		}
	}
	utils.LogInfof("Add topic: %s", topic)
	err := m.subscribe(topic)
	if err != nil {
		utils.LogErrorf("%s failed: %s", topic, err.Error())
		return fmt.Errorf("%s failed: %s", topic, err.Error())
	}

	m.topics = append(m.topics, topic)
	return nil
}

func (m *MqttService) Disconnect() {
	m.client.Disconnect(100)
	utils.LogInfo("zigbee2mqtt client disconnected")
}
