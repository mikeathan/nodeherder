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

{"nickname": "enviro_node", "uid": "e661410403554934", "timestamp": "2023-08-01T16:30:04Z", "readings": {"id": "enviro_node", "temperature": 27.03, "humidity": 43.63, "pressure": 985.46, "luminance": 4.65, "wind_speed": 0, "rain": 0, "rain_per_second": 0.0, "wind_direction": 90, "voltage": 0.0}, "model": "weather"}

TEST:

curl -X POST http://192.168.50.183:4100/collect -H 'Content-Type: application/json' -d '{"nodeid":"node1","temperature":45.6,"Timestamp":"2023-03-19T19:57:28.961193655Z"}'

TODO:

- worker pool that notify consumers when task added ??
- on startup get mqtt device state from z2m and build the connected devices
  register to mqtt state topic


hive bulb
"color_temp"  flaot
"color_mode" string

"brightness" flaot
"state" string


gas sensor
42["node_data_updated","{\"3\": {\"label\": \"gas_monitor\", \"node_id\": \"3\", \"temperature\": \"22.2 *C\", \"humidity\": \"33 %RH\", \"air_quality_score\": \"95 %\", \"PM1.0\": \"1 ug/m3 (ultrafine particles)\", \"PM2.5\": \"1 ug/m3 (combustion particles, organic compounds, metal)\", \"PM10.0\": \"2 ug/m3 (dust, pollen, mould spores)\", \"timestamp\": 1691517687.940329}}"]

enviro weather
{"nickname": "enviro_node", "uid": "e661410403554934", "timestamp": "2023-08-09T16:50:12Z", "readings": {"temperature": 27.38, "humidity": 45.48, "pressure": 1003.16, "luminance": 54.98, "wind_speed": 0, "rain": 0, "rain_per_second": 0.0, "wind_direction": 90, "voltage": 0.0}, "model": "weather"}
