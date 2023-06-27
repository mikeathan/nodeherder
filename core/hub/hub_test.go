package hub_test

import (
	"context"
	"fmt"
	"node-herder/hub"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestXxx(t *testing.T) {

	port := 3000
	var broker = "192.168.50.179:1883"
	var topic = "device1"
	const msg1 = "{'battery':95,'humidity':60.8,'last_seen':'2023-05-31T19:05:28+01:00','linkquality':50,'temperature':22.1,'voltage':3000}"
	const msg2 = "{'battery':95,'humidity':60.8,'last_seen':'2023-05-31T19:05:28+01:00','linkquality':50,'temperature':22.1,'voltage':3000}"
	repo := hub.NewMemoryRepository()

	repo.Store(topic, msg1)
	repo.Store("device2", msg2)
	// setup ws hub
	var wsConfig = hub.WsConfig{
		OnConnected: func() interface{} {
			devices := repo.ListAllDevices()
			bytes, err := hub.ToJson(devices)
			if err != nil {
				fmt.Printf("OnConnected error: %s \n", err)
				return []byte{}
			}
			return bytes
		},
	}

	ctx, _ := context.WithCancel(context.Background())
	wsHub := hub.NewWsHub(wsConfig)

	router := hub.NewRouter()
	router.GET("/ws", hub.NewWsHandler(wsHub))

	apiServer := hub.NewHttpServer(
		port,
		hub.WithRouter(router),
		hub.WithContext(ctx),
	)

	go apiServer.Listen()

	// need test http server
	// h := hub.NewWsHandler(wsHub)
	//NewTestWsServer(t, h)

	//setup mqtt

	var messageHandler = func(client mqtt.Client, msg mqtt.Message) {
		var name = hub.SanitizeTopic(msg.Topic())
		var payload = msg.Payload()
		fmt.Printf("mqtt Message => Topic: %s, Payload: %s\n", msg.Topic(), msg.Payload())
		repo.Store(name, payload)
		wsHub.Broadcast(hub.DeviceUpdated, payload)
	}

	cfg := GetMqttConfig(broker, messageHandler, topic)
	mqttClient := hub.NewMqttClient(cfg)
	mqttClient.Connect()

	var message = "{\"battery\":100,\"humidity\":60.4,\"last_seen\":\"2023-06-27T15:33:24+01:00\",\"linkquality\":40,\"temperature\":24,\"voltage\":3000}"
	StartMqttNodeClient(cfg, message, 1)

	//SendMessage(t, conn, []byte(message))

	fmt.Println("Exited")

}
