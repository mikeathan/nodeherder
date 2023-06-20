package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"node-herder/hub"
	"os"
)

// port := readPort()

// ctx, cancelCtx := context.WithCancel(context.Background())

// fmt.Println("Starting up server.")
// c := make(chan os.Signal)
// signal.Notify(c, os.Interrupt, syscall.SIGTERM)

// go func() {
// 	defer close(c)
// 	<-c
// 	fmt.Println("SIGTERM signal notified")
// 	cancelCtx()
// }()

// repo := node.NewMemoryRepository()

// router := hub.NewRouter()

// // collectorService := api.NewCollectorService(repo)
// // router.PUT("/collector", api.NewDataHandler(collectorService))
// // router.POST("/collector", api.NewFindDeviceEventsHandler(collectorService))
// // router.GET("/collector/history", api.NewHistoryHandler(collectorService))
// // router.GET("/collector/devices", api.NewDevicesHandler(collectorService))
// // router.GET("/collector/ping", api.NewHealthCheckHandler())

// //router.Use(api.LoggingMiddleware(logger))

// // apiServer := api.NewServer(
// // 	port,
// // 	api.WithRouter(router),
// // 	api.WithLogger(logger),
// // 	api.WithContext(ctx),
// // )

// // apiServer.Listen()
// fmt.Println("Exited")

func readPort() int {

	port := flag.Int("port", 4100, "port number")
	flag.Parse()
	if *port <= 0 {

		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return *port
}
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
