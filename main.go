package main

import (
	"fmt"
	"log"
	"net/http"
	"node-herder/hub"
)

func main() {

	broker := "192.168.50.179:1883"
	username := "sinkhole"
	password := "pwd"

	fileServer := http.FileServer(http.Dir("./frontend/dist"))
	http.Handle("/", fileServer)

	client := hub.NewZ2MClient(broker, username, password)
	client.AddDevice("TH1")
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
