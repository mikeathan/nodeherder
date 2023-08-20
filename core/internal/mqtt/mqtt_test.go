package mqtt_test

import (
	"fmt"
	"node-herder/transport/mqtt"
	"testing"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

func GetMqttConfig(broker string, topics ...string) mqtt.MqttConfig {
	return mqtt.MqttConfig{
		Username: "sinkhole",
		Password: "mqtt2023",
		Broker:   broker,
		Topics:   topics,
	}
}

func TestMqttClientReceivesMessage(t *testing.T) {
	var broker = "192.168.50.179:1883"
	var topic = "device1"
	var message = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"
	var messageHandler = func(id string, payload []byte) {
		// TODO: test topic
		if string(payload) != message {
			t.Errorf("Payload mismatch - want %s, got %s", message, string(payload))
		}
	}

	cfg := GetMqttConfig(broker, topic)
	mqttClient := mqtt.NewMqttClient(cfg)
	mqttClient.OnMessageHandler(messageHandler)
	mqttClient.Connect()
	StartMqttNodeClient(cfg, message, 2)
}

func StartMqttNodeClient(cfg mqtt.MqttConfig, message string, nEvents int) {

	// some fake external device mqqtclient
	var opts = mqttlib.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.Username = cfg.Username
	opts.Password = cfg.Password
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnectionLost = connectLostHandler
	opts.OnConnect = connectHandler

	var client = mqttlib.NewClient(opts)
	var token = client.Connect()
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}

	//
	var topic = fmt.Sprintf("zigbee2mqtt/%s", cfg.Topics[0])
	var closeChan = make(chan bool)

	go publishFunc(closeChan, client, topic, message, nEvents)

	<-closeChan
	fmt.Printf("Device Shutdown \n")
}

var messagePubHandler mqttlib.MessageHandler = func(client mqttlib.Client, msg mqttlib.Message) {
	fmt.Printf("Device Message = Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
}

var connectHandler mqttlib.OnConnectHandler = func(client mqttlib.Client) {
	fmt.Println("Device Connected")
}

var connectLostHandler mqttlib.ConnectionLostHandler = func(client mqttlib.Client, err error) {
	fmt.Printf("Device Connection lost: %v", err)
}

func publishFunc(closeChan chan bool, client mqttlib.Client, topic string, message string, numOfEvents int) {

	for i := 1; i <= numOfEvents; i++ {
		var token = client.Publish(topic, 2, false, message)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}

		fmt.Printf("Device publish: %s \n", message)

		time.Sleep(2 * time.Second)
	}

	closeChan <- true

}
