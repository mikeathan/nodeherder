package hub_test

import (
	"context"
	"fmt"
	"node-herder/hub"
	"testing"
	"time"
)

func TestXxx(t *testing.T) {

	t.Skip("component test. ignore")
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

	ctx, _ := context.WithCancel(context.Background())

	eventHub := hub.NewWsHub()

	cfg := GetMqttConfig(broker, nil, topic)

	mqtt := hub.NewMqttClient(cfg)
	_, err := hub.NewController(
		hub.WithRepository(repo),
		hub.WithMqtt(mqtt),
		hub.WithEventHub(eventHub))

	if err != nil {
		fmt.Printf("Hub connector error: %s \n", err.Error())
		<-ctx.Done()
	}

	apiServer := createHttpServer(port, eventHub, ctx)

	go apiServer.Listen()

	// this to directy trigger http request
	// TEST ONLY   - comment
	//NewTestWsServer(t, h)
	// #######################
	//time.Sleep(5 * time.Minute)

	// wait before sending some random mqtt device data
	time.Sleep(5 * time.Second)
	//setup mqtt

	StartMqttNodeClient(cfg, data3, 1)

	//SendMessage(t, conn, []byte(message))

	fmt.Println("Exited")
}

func createHttpServer(port int, ws hub.EventHub, ctx context.Context) hub.HttpServer {
	router := hub.NewRouter()
	router.GET("/ws", hub.NewWsHandler(ws))
	//router.GET("/", http.FileServer(http.Dir("../../frontend/dist")))

	apiServer := hub.NewHttpServer(
		port,
		hub.WithRouter(router),
		hub.WithContext(ctx),
	)

	return apiServer

}
