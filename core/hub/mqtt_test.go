package hub_test

import (
	"fmt"
	"math/rand"
	"node-herder/hub"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestMqtt(t *testing.T) {

	// TODO: either connect to online test broker or the local one in the network
	var broker = "tcp://test.mosquitto.org:1883"
	var topic = "TestDevice1"
	var maxInterval = 5

	var message = "test mqtt data 1"

	cfg := hub.Config{
		Mqtt: hub.MqttConfig{
			Username: "test",
			Password: "12345",
			Broker:   ip,
			Topics: []string{
				topic,
			},
		}}

	var opts = mqtt.NewClientOptions()
	opts.AddBroker(ip)
	opts.Username = cfg.Mqtt.Username
	opts.Password = cfg.Mqtt.Password
	var client = mqtt.NewClient(opts)
	var token = client.Connect()
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}

	mqttClient := hub.NewMqttClient(cfg.Mqtt)

	mqttClient.Connect()
	mqttClient.WithMessageHandler(messagePubHandler)

	//
	clientMain(client, topic, message, maxInterval)
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("mqtt message => Topic: %s, Payload; %s\n", msg.Topic(), msg.Payload())
}

func publishFunc(client mqtt.Client, topic string, message string, maxInterval int) {
	for {
		var token = client.Publish(topic, 2, false, message)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}
		fmt.Println(message)
		time.Sleep(time.Duration(rand.Float64() * float64(maxInterval) * float64(time.Second)))
	}
}

func clientMain(client mqtt.Client, topic string, message string, maxInterval int) {
	var closeChan = make(chan struct{})

	go publishFunc(client, topic, message, maxInterval)

	<-closeChan
}
