package hub_test

import (
	"context"
	"fmt"
	"node-herder/hub"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestXxx(t *testing.T) {

	port := 3000
	var broker = "192.168.50.179:1883"
	var topic = "device1"
	var topic2 = "device2"
	const data1 = `{"battery":100, "humidity":61.8, "last_seen":"2023-05-31T19:05:28+01:00", "linkquality":26,"temperature":29.5,"voltage":3000}`
	const data2 = `{"battery":100, "humidity":54.2, "last_seen":"2023-07-31T19:05:28+01:00", "linkquality":34,"temperature":17.1,"voltage":2999}`
	const data3 = `{"battery":98, "humidity":71.2, "last_seen":"2023-07-31T19:05:28+01:00", "linkquality":36.1,"temperature":17.1,"voltage":2999}`
	repo := hub.NewMemoryRepository()

	repo.StoreJson(topic, []byte(data1))
	repo.StoreJson(topic2, []byte(data2))
	// setup ws hub
	var wsConfig = hub.WsConfig{
		OnConnected: func() interface{} {
			return repo.ListAllDevices()
		},
	}

	ctx, _ := context.WithCancel(context.Background())
	wsHub := hub.NewWsHub(wsConfig)

	h := hub.NewWsHandler(wsHub)
	router := hub.NewRouter()
	router.GET("/ws", h)

	apiServer := hub.NewHttpServer(
		port,
		hub.WithRouter(router),
		hub.WithContext(ctx),
	)

	go apiServer.Listen()

	// this to directy trigger http request
	// TEST ONLY   - comment
	//NewTestWsServer(t, h)
	// #######################
	//time.Sleep(5 * time.Minute)

	// wait before sending some random mqtt device data
	time.Sleep(5 * time.Second)
	//setup mqtt
	var messageHandler = func(client mqtt.Client, msg mqtt.Message) {
		var name = hub.SanitizeTopic(msg.Topic())
		var payload = msg.Payload()
		fmt.Printf("mqtt Message => Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
		repo.StoreJson(name, payload)
		device, _ := repo.FindDevice(name)
		wsHub.Broadcast(hub.DeviceUpdated, device)
	}

	cfg := GetMqttConfig(broker, messageHandler, topic)
	mqttClient := hub.NewMqttClient(cfg)
	mqttClient.Connect()

	StartMqttNodeClient(cfg, data3, 1)

	//SendMessage(t, conn, []byte(message))

	fmt.Println("Exited")

}
