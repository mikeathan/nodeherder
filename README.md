sample data : TH1
{
"battery": 100,
"humidity": 69.8,
"last_seen": "2023-05-06T19:13:37+01:00",
"linkquality": 29,
"temperature": 20,
"voltage": 3000
}
{
"name": "TH1",
"payload"
{
"battery": 100,
"humidity": 69.8,
"last_seen": "2023-05-06T19:13:37+01:00",
"linkquality": 29,
"temperature": 20,
"voltage": 3000
}
}
{
"name": "TH2",
"payload"
{
"battery": 100,
"humidity": 61.8,
"last_seen": "2023-05-09T17:07:22+01:00",
"linkquality": 32,
"temperature": 22.2,
"voltage": 3000
}
}
// libs
go get github.com/mehdihadeli/go-mediatry

--Frontend
settings page
location_name
latitude
longitude
elevation
unit_system
time_zone
home
sensor_data

--Web server
Starts mqtt
Receives message
Send websocket message
Send metrics rest-api request

Websocket listener

Metrics listener

TODO:
mqtt message listener changed
websocket emit device changed message
frontend listens to message and loads it

advanced:
websockets:
emit message device change
all devices or just the updated ?

{
"detection_delay": 0.5,
"fading_time": 5,
"illuminance_lux": 168,
"last_seen": "2023-07-01T11:20:43+01:00",
"linkquality": 32,
"maximum_range": 6,
"minimum_range": 0.6,
"presence": true,
"radar_sensitivity": 5,
"target_distance": 3.39
}

Setting configuration for devices
Triggers and binding
Add api endpoint for collecting data from WiFi sensors
Circuit breaker

http
conn="http"
power_source NULL or [VALUE]

mqtt
conn="mqtt"
power_source battery or mains
data:
{"type":"connected","payload":[{"id":"device 1","conn":"mqtt","power_source":"battery","sensors":{"humidity":92.49999999999999,"temperature":19.000000000000004},"stats":{"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":47,"battery":98}},{"id":"device 2","conn":"mqtt","power_source":"mains","sensors":{"presence":false,"illuminance_lux":103},"stats":{"availability":"online","last_seen":"2023-07-20T19:48:35+01:00","linkquality":67}},{"id":"device 3","conn":"http","power_source":"","sensors":{"humidity":41,"temperature":10,"pressure":68},"stats":{"availability":"offline","last_seen":"2023-07-20T19:48:35+01:00"}}]}

Mqtt payload handling :

IDEAS:

1. WorkerPool returns results channel and controller handles them
2. Processor takes controller? and does the repo.store and ws.send by calling controller function
   -problem: dependecies between Processor and controller
3. Separate controller:
   - ws is resposible for its event handling
   - mqtt for triggering processor
   - Processor still needs repo and eventHub
   - dependencies:
     - ws -> repo
     - mqtt -> Processor
     - Processor -> ws and repo
4. Replace controller with Processor or Processor becomes the controller

ws = EventHub(repo)
Processor= Processor(EventHub, Repo)
mqtt = Exposes handler for someone to use Processor

mqtt.onNewEvent{
Processor.Enqueue(id, payload)
}

ws.onConnected{
return all_devices()
}

Processor.ProcessPayload(id, payload){
if payload == new || updated{

        repo.store(payload)
        ws.send(payload)
    }

}
{"nickname": "enviro_node", "uid": "e661410403554934", "timestamp": "2023-08-01T16:30:04Z", "readings": {"id": "enviro_node", "temperature": 27.03, "humidity": 43.63, "pressure": 985.46, "luminance": 4.65, "wind_speed": 0, "rain": 0, "rain_per_second": 0.0, "wind_direction": 90, "voltage": 0.0}, "model": "weather"}
