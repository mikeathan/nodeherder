package hub_test

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func TestMqtt(t *testing.T) {
	var ip = flag.String("ip", "tcp://127.0.0.1:1883", "MQTT broker address")
	//var server = flag.Bool("server", false, "Run in server mode")
	var debug = flag.Bool("debug", false, "Run in debug mode")
	var maxInterval = flag.Int("maxinterval", 5, "Max interval for sending words")

	var message = "test mqtt data 1"

	var opts = mqtt.NewClientOptions()
	opts.AddBroker(*ip)

	var client = mqtt.NewClient(opts)
	var token = client.Connect()
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}

	clientMain(client, message, *maxInterval, *debug)
}

func publishFunc(client mqtt.Client, word string, maxInterval int, debug bool) {
	var topic = fmt.Sprintf("topic_%s", word)
	for {
		var token = client.Publish(topic, 2, false, word)
		token.Wait()
		if token.Error() != nil {
			panic(token.Error())
		}
		if debug {
			fmt.Println(word)
		}
		time.Sleep(time.Duration(rand.Float64() * float64(maxInterval) * float64(time.Second)))
	}
}

func clientMain(client mqtt.Client, message string, maxInterval int, debug bool) {
	var closeChan = make(chan struct{})

	for _, w := range strings.Split(message, " ") {
		go publishFunc(client, w, maxInterval, debug)
	}

	<-closeChan
}
