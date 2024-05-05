package main

import (
	"context"
	"flag"
	"fmt"
	hub "node-herder/internal"
	"node-herder/internal/mqtt"
	repository "node-herder/repository/devices"
	"node-herder/utils"
	"os"
	"os/signal"
	"syscall"
)

type cmdArgs struct {
	port      int
	buildType string
	logLevel  string
}

func readArgs() *cmdArgs {

	port := flag.Int("port", 4100, "port number")
	buildType := flag.String("buildType", "", "client build type")
	logLevel := flag.String("logLevel", "info", "client build type")

	flag.Parse()
	if *port <= 0 {

		fmt.Print("Invalid port number")
		os.Exit(-1)
	}

	return &cmdArgs{port: *port, buildType: *buildType, logLevel: *logLevel}
}

func main() {

	args := readArgs()

	utils.InitFileLogger()
	utils.SetLogLevel(args.logLevel)

	ctx, cancelCtx := context.WithCancel(context.Background())

	utils.LogInfo("starting up server")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer close(c)
		<-c
		utils.LogWarn("system termination signal received.")
		cancelCtx()
	}()

	repo := repository.NewMemoryDeviceRepo()
	mqttConfig := mqtt.MqttConfig{
		Username:   "sinkhole",
		Password:   "mqtt2023",
		Broker:     "192.168.50.179:1883",
		ClientType: args.buildType,
	}

	h := hub.Register(args.port, repo, mqttConfig, ctx)
	h.Listen()
	utils.LogInfo("exit")
}

// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"time"

// 	"github.com/nakabonne/tstorage"
// )

// type SensorData struct {
// 	DeviceID string  `json:"device_id"`
// 	Sensor   string  `json:"sensor"`
// 	Value    float64 `json:"value"`
// }

// func main() {
// 	storage, err := tstorage.NewStorage(tstorage.WithTimestampPrecision(tstorage.Seconds))
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer storage.Close()

// 	devices := []struct {
// 		id   string
// 		data chan SensorData
// 	}{
// 		{id: "device1", data: make(chan SensorData)},
// 		{id: "device2", data: make(chan SensorData)},
// 	}

// 	for _, device := range devices {
// 		data := SensorData{
// 			DeviceID: device.id,
// 			Sensor:   "temperature",
// 			Value:    generateRandomFloat(20.0, 30.0),
// 		}
// 		err = insert(storage, data)
// 		fmt.Println("device ", device.id, " stored")

// 		if err != nil {
// 			fmt.Println("Error inserting data for device1:", err)
// 		}
// 	}
// 	time.Sleep(time.Second * 2)

// 	for _, device := range devices {

// 		err = query(storage, device.id, "temperature", time.Now().Add(time.Minute*1), time.Now().Add(-time.Minute*1))
// 		if err != nil {
// 			fmt.Println("Error quering data for device1:", err)
// 		}
// 	}

// }

// func insert(storage tstorage.Storage, data SensorData) error {
// 	labels := []tstorage.Label{
// 		{Name: "sensor", Value: data.Sensor},
// 	}
// 	return storage.InsertRows([]tstorage.Row{
// 		{
// 			Metric:    data.DeviceID,
// 			Labels:    labels,
// 			DataPoint: tstorage.DataPoint{Timestamp: time.Now().Unix(), Value: data.Value},
// 		}})

// }

// func query(storage tstorage.Storage, device string, sensor string, start time.Time, end time.Time) error {
// 	labels := []tstorage.Label{
// 		{Name: "sensor", Value: sensor},
// 	}
// 	points, err := storage.Select(device, labels, end.Unix(), start.Unix())
// 	if err != nil {
// 		return err
// 	}

// 	for _, p := range points {
// 		fmt.Printf("timestamp: %v, value: %v\n", p.Timestamp, p.Value)
// 	}

// 	return nil
// }
// func generateRandomFloat(min, max float64) float64 {
// 	return min + (max-min)*rand.Float64()
// }
