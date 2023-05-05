package transport

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

func test() {
	var broker = "192.168.50.179:1883"
	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID("sinkhole-mikeathan")
	options.Username = "usr"
	options.Password = "pwd"
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "zigbee2mqtt/TH1"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("topic %s\n", topic)
	//payload := `{"temperature": ""}`
	//token = client.Publish(topic, 0, false, payload)
	//token.Wait()

	client.Disconnect(100)
}
