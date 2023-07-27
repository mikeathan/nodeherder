package mqttlistener_test

import (
	"fmt"
	"node-herder/transport/mqttlistener"
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

func GetMqttConfig(broker string, messageHandler func(client mqtt.Client, msg mqtt.Message), topics ...string) mqttlistener.MqttConfig {
	return mqttlistener.MqttConfig{
		Username:       "sinkhole",
		Password:       "mqtt2023",
		Broker:         broker,
		Topics:         topics,
		MessageHandler: messageHandler,
	}
}

func TestMqttClientReceivesMessage(t *testing.T) {
	var broker = "192.168.50.179:1883"
	var topic = "device1"
	var message = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"
	var messageHandler = func(client mqtt.Client, msg mqtt.Message) {
		// TODO: test topic
		if string(msg.Payload()) != message {
			t.Errorf("Payload mismatch - want %s, got %s", message, string(msg.Payload()))
		}
	}

	cfg := GetMqttConfig(broker, messageHandler, topic)
	mqttClient := mqttlistener.NewMqttClient(cfg)
	mqttClient.Connect()
	StartMqttNodeClient(cfg, message, 2)
}

func StartMqttNodeClient(cfg mqttlistener.MqttConfig, message string, nEvents int) {

	// some fake external device mqqtclient
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

	go publishFunc(closeChan, client, topic, message, nEvents)

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
