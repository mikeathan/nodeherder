package mqtt_test

import (
	"fmt"
	"node-herder/internal/mqtt"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
)

func StartMqttNodeClient(cfg mqtt.MqttConfig, topic string, message string, nEvents int) {

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
	var t = fmt.Sprintf("zigbee2mqtt/%s", topic)
	var closeChan = make(chan bool)

	go publishFunc(closeChan, client, t, message, nEvents)

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
