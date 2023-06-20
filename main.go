package main

import (
	"fmt"
	"log"
	"net/http"
	"node-herder/hub"
)

func main() {

	cfg := hub.HubConfig{WSPath: "/ws",
		MqttConfig: hub.MqttConfig{
			Username: "sinkhole",
			Password: "mqtt2023",
			Broker:   "192.168.50.179:1883",
			Topics: []string{
				"TH1",
			},
		}}

	hub.Init(cfg)

	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fileServer)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
