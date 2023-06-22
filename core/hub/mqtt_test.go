package hub_test

import (
	"fmt"
	"node-herder/hub"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// public mqtt test broker
// var broker = "broker.emqx.io"
//
//	var port = 1883
//	opts := mqtt.NewClientOptions()
//	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
//	opts.SetClientID("go_mqtt_client")
//	opts.SetUsername("emqx")
//	opts.SetPassword("public")
func TestMqtt(t *testing.T) {
	t.Skip("component test. ignore")
	// TODO: either connect to online test broker or the local one in the network
	var broker = "192.168.50.179:1883"
	var topic = "device1"

	cfg := hub.Config{
		Mqtt: hub.MqttConfig{
			Username: "sinkhole",
			Password: "mqtt2023",
			Broker:   broker,
			Topics: []string{
				topic,
			},
		}}

	// hub mqtt client
	mqttClient := hub.NewMqttClient(cfg.Mqtt)
	mqttClient.WithMessageHandler(hubMessageHandler())
	mqttClient.Connect()

	startMqttNodeClient(cfg.Mqtt)
}

func hubMessageHandler() func(client mqtt.Client, msg mqtt.Message) {
	return func(client mqtt.Client, msg mqtt.Message) {
		var name = msg.Topic()
		var payload = msg.Payload()
		fmt.Printf("Hub received message => Topic: %s, Payload; %s\n", name, payload)
	}
}

func startMqttNodeClient(cfg hub.MqttConfig) {

	// some fake external device mqqtclient
	var numOfEvents = 5
	var message = "test mqtt data"

	var opts = mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.Username = cfg.Username
	opts.Password = cfg.Password
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnectionLost = connectLostHandler
	opts.OnConnect = connectHandler

	var client = mqtt.NewClient(opts)
	var token = client.Connect()
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}

	//
	var topic = fmt.Sprintf("zigbee2mqtt/%s", cfg.Topics[0])
	var closeChan = make(chan bool)

	go publishFunc(closeChan, client, topic, message, numOfEvents)

	<-closeChan
	fmt.Printf("Device Shutdown \n")
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Device Message = Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Device Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Device Connection lost: %v", err)
}

func publishFunc(closeChan chan bool, client mqtt.Client, topic string, message string, numOfEvents int) {

	for i := 1; i <= numOfEvents; i++ {
		var payload = fmt.Sprintf("%s %d", message, i)
		var token = client.Publish(topic, 2, false, payload)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}

		fmt.Printf("Device publish: %s \n", payload)

		time.Sleep(2 * time.Second)
	}

	closeChan <- true

}
