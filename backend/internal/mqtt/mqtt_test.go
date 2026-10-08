package mqtt_test

import (
	"fmt"
	"node-herder/internal/mqtt"
	"node-herder/mocks"
	"slices"
	"testing"
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

func newFakeMqttService(t *testing.T) (mqtt.MqttClient, *mocks.FakePahoClient) {
	t.Helper()
	var paho *mocks.FakePahoClient
	client := mqtt.NewMqttClient(mqtt.WithClientFactory(func(options *mqttlib.ClientOptions) mqttlib.Client {
		paho = mocks.NewFakePahoClient(options)
		return paho
	}))
	if err := client.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { paho.Disconnect(250) })
	// OnConnect subscribes the six bridge topics on its own goroutine.
	for i := 0; i < 6; i++ {
		select {
		case <-paho.Subscriptions():
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for bridge subscription %d", i+1)
		}
	}
	return client, paho
}

func TestRemoveTopicUnsubscribesTheSubscribedTopic(t *testing.T) {
	client, paho := newFakeMqttService(t)
	if err := client.AddTopic("fixture"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(paho.Subscribed(), "zigbee2mqtt/fixture") {
		t.Fatalf("subscriptions %v lack zigbee2mqtt/fixture", paho.Subscribed())
	}
	if err := client.RemoveTopic("fixture"); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(paho.Subscribed(), "zigbee2mqtt/fixture") {
		t.Fatalf("zigbee2mqtt/fixture still subscribed after RemoveTopic: %v", paho.Subscribed())
	}
}

func TestInboundTopicStripsOnlyTheBasePrefix(t *testing.T) {
	client, paho := newFakeMqttService(t)
	received := make(chan string, 1)
	client.OnMessageHandler(func(topic string, _ []byte) { received <- topic })

	paho.Deliver("zigbee2mqtt/garage/zigbee2mqtt/sensor", []byte(`{}`))
	select {
	case got := <-received:
		if got != "garage/zigbee2mqtt/sensor" {
			t.Fatalf("handler topic = %q, want %q", got, "garage/zigbee2mqtt/sensor")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the message")
	}
}

func TestRemoveTopicLeavesOtherClientsBridgeTopics(t *testing.T) {
	client, _ := newFakeMqttService(t)
	if err := client.RemoveTopic("bridge/devices"); err != nil {
		t.Fatal(err)
	}
	_, other := newFakeMqttService(t)
	if !slices.Contains(other.Subscribed(), "zigbee2mqtt/bridge/devices") {
		t.Fatalf("a new client's bridge subscriptions %v lack zigbee2mqtt/bridge/devices", other.Subscribed())
	}
}
