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

func NewZ2MClient(broker string, username string, password string) *Z2MClient {

	return &Z2MClient{
		broker:   broker,
		username: username,
		password: password,
		devices:  []string{},
		client:   nil,
		cliendId: "sinkhole-mikeathan",
	}
}

var baseTopic string = "zigbee2mqtt"

type Z2MClient struct {
	devices  []string
	client   mqtt.Client
	broker   string
	username string
	password string
	cliendId string
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
		topic := fmt.Sprintf("%s/%s", baseTopic, device_name)
		token = m.client.Subscribe(topic, 1, nil)

		if token.Wait() && token.Error() != nil {
			return token.Error()
		}
		fmt.Printf("mqtt topic: %s subscribed\n", topic)
	}

	return nil
}

func (m *Z2MClient) AddDevice(device_name string) {
	m.devices = append(m.devices, device_name)
}

func (m *Z2MClient) Disconnect() {
	m.client.Disconnect(100)
	fmt.Println("mqtt disconnected")
}
