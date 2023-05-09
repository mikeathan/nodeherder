package main

import (
	"fmt"
	"log"
	"net/http"
	"node-herder/hub"
)

func main() {

	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fileServer)

	options := hub.NewMqttOptions("192.168.50.179", "username", "password")
	client := hub.NewMqttClient(options)
	client.AddTopic("TH1")

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
