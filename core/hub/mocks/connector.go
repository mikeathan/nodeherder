package mocks

import (
	"fmt"
	"node-herder/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

type NopWsServer struct {
}

func (w *NopWsServer) Broadcast(eventName string, data interface{}) error {
	fmt.Println("Empty Broadcast")
	return nil
}
func (w *NopWsServer) RegisterNewClient(conn *websocket.Conn) models.EventEmitter {
	fmt.Println("Empty RegisterNewClient")
	return nil
}

type NopMqttClient struct {
}

func (w *NopMqttClient) Connect() error {
	fmt.Println("Empty Connect")
	return nil
}

func (w *NopMqttClient) WithMessageHandler(messageHandler func(client mqtt.Client, msg mqtt.Message)) {
	fmt.Println("Empty WithMessageHandler")
}
func (w *NopMqttClient) AddTopic(topic string) {
	fmt.Println("Empty BroaAddTopicdcast")

}
func (w *NopMqttClient) Disconnect() {
	fmt.Println("Empty Disconnect")
}

type NopRepository struct {
}

func (w *NopRepository) Store(deviceName string, payload interface{}) {
	fmt.Println("Empty Store")
}

func (w *NopRepository) ListAllDevices() []*models.Device {

	fmt.Println("Empty ListAllDevices")
	return []*models.Device{}
}
