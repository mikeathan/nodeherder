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

advanced:

Setting configuration for devices
Triggers and binding
Add api endpoint for collecting data from WiFi sensors
Circuit breaker

TEST:

curl -X POST http://192.168.50.69:4100/collect -H 'Content-Type: application/json' -d '{"label":"weather node 1","temperature":45.6,"Timestamp":"2023-03-19T19:57:28.961193655Z"}'

TODO:

- worker pool that notify consumers when task added ??
- on startup get mqtt device state from z2m and build the connected devices
  register to mqtt state topic

hive bulb
"color_temp" flaot
"color_mode" string

"brightness" flaot
"state" string

gas sensor
42["node_data_updated","{\"3\": {\"label\": \"gas_monitor\", \"node_id\": \"3\", \"temperature\": \"22.2 *C\", \"humidity\": \"33 %RH\", \"air_quality_score\": \"95 %\", \"PM1.0\": \"1 ug/m3 (ultrafine particles)\", \"PM2.5\": \"1 ug/m3 (combustion particles, organic compounds, metal)\", \"PM10.0\": \"2 ug/m3 (dust, pollen, mould spores)\", \"timestamp\": 1691517687.940329}}"]

enviro weather
{"nickname": "enviro_node", "uid": "e661410403554934", "timestamp": "2023-08-09T16:50:12Z", "readings": {"temperature": 27.38, "humidity": 45.48, "pressure": 1003.16, "luminance": 54.98, "wind_speed": 0, "rain": 0, "rain_per_second": 0.0, "wind_direction": 90, "voltage": 0.0}, "model": "weather"}

TODO:
Automation

can only do timer/device and webrequests
Devices can be done only once we can send mqqt messages o devices that allow setting
Trigger
Timer
Device

Condition
Timer
value
Sensor
value

Action
Device
sensor set something
Web request
post payload

    icons

    <i class="fa-solid fa-wind"></i>
    https://icons8.com/icons/set/particulate-matter
    https://icons8.com/icons/set/air-quality

{"type":"deviceAdded","payload":{"id":"0xa4c13894070052fc","name":"Human presence","connection_type":"mqtt","power_source":"mains","exposes":{"illuminance_lux":{"name":"illuminance_lux","description":"Measured illuminance in lux","unit":"lx","data":3,"properties":{}},"presence":{"name":"presence","description":"Indicates whether the device detected presence","data":false,"properties":{}}},"properties":{"availability":"online","last_seen":"2023-10-10T05:55:29+01:00","linkquality":58}}}

{"type":"deviceAdded","payload":{"id":"0x00124b00146c31cd","name":"Motion sensor 1","connection_type":"mqtt","power_source":"battery","exposes":{"occupancy":{"name":"occupancy","description":"Indicates whether the device detected occupancy","data":false,"properties":{}},"temperature":{"name":"temperature","description":"Measured temperature value","unit":"°C","data":23.75,"properties":{}}},"properties":{"availability":"online","battery":7,"last_seen":"2023-10-08T06:24:24+01:00","linkquality":29}}}

{"type":"deviceAdded","payload":{"id":"0x00124b0029207763","name":"TH01","connection_type":"mqtt","power_source":"battery","exposes":{"humidity":{"name":"humidity","description":"Measured relative humidity","unit":"%","data":83.91,"properties":{}},"temperature":{"name":"temperature","description":"Measured temperature value","unit":"°C","data":19.87,"properties":{}}},"properties":{"availability":"online","battery":100,"last_seen":"2023-10-08T06:25:37+01:00","linkquality":32}}}

{"type":"deviceUpdated","payload":{"id":"0x00124b0029207763","last_seen":"2023-10-11T06:26:18+01:00","data":{"temperature":20.02}}}
{"type":"deviceUpdated","payload":{"id":"0x00124b0029207763","last_seen":"2023-10-11T06:27:28+01:00","data":{"humidity":83}}}

TODO:

fix golang server routing on refresh
add more ws operation responses - eg success or error
test autiomation loading. configureAction for sanitizing numeric type data
do we need to unsubsribe from removed/renamed topic ??
Test new logic in RegisterBridge
backup automations

###### frontend

allow brigtness in actions. {"brightness":197}
action doenst contain brightness when running form golang

add support for hue switch rotation
add log window in frontend
device card - add ways to control device if its feature
fix sizing for mobile
