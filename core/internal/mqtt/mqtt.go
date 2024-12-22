package mqtt

import (
	"fmt"
	"node-herder/utils"
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
	"bridge/response/device/interview",
	"bridge/logging",
}

// TODO
// bridge/request/device/remove - {"id":"my_bulb","force":true} -
//  - success response = "data":{"id": "my_bulb","block":false,"force":false},"status":"ok"}
//  - error response = {"id": "my_bulb","block":false,"force":false},"status":"error","error":"Failed to remove dimmer (Error: AREQ - ZDO - mgmtLeaveRsp after 10000ms)"
// bridge/request/device/configure - {"id": "deviceID"} - response = {"data":{"id": "my_remote"},"status":"ok"}.
// bridge/request/permit_join - {"value": true, "time": 20} (will allow joining for 20 seconds).
// bridge/request/restart - empty payload - response =  {"data":{},"status":"ok"}.
// /bridge/request/backup - mpty payload - response: {"data":{"zip":"WklHQkVFMk1RVFQuUk9DS1M="},"status":"ok"}

type MqttConfig struct {
	Broker     string
	Username   string
	Password   string
	ClientType string
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

func NewMqttClient(config MqttConfig) MqttClient {

	var client = &MqttService{
		broker:         config.Broker,
		username:       config.Username,
		password:       config.Password,
		client:         nil,
		clientId:       "sinkhole-z2m",
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
