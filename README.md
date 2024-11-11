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
Add api endpoint for collecting data from WiFi sensors
Circuit breaker

TEST:

curl -X POST http://192.168.50.69:4100/collect -H 'Content-Type: application/json' -d '{"label":"weather node 1","temperature":45.6,"Timestamp":"2023-03-19T19:57:28.961193655Z"}'

gas sensor
42["node_data_updated","{\"3\": {\"label\": \"gas_monitor\", \"node_id\": \"3\", \"temperature\": \"22.2 *C\", \"humidity\": \"33 %RH\", \"air_quality_score\": \"95 %\", \"PM1.0\": \"1 ug/m3 (ultrafine particles)\", \"PM2.5\": \"1 ug/m3 (combustion particles, organic compounds, metal)\", \"PM10.0\": \"2 ug/m3 (dust, pollen, mould spores)\", \"timestamp\": 1691517687.940329}}"]

enviro weather
{"nickname": "enviro_node", "uid": "e661410403554934", "timestamp": "2023-08-09T16:50:12Z", "readings": {"temperature": 27.38, "humidity": 45.48, "pressure": 1003.16, "luminance": 54.98, "wind_speed": 0, "rain": 0, "rain_per_second": 0.0, "wind_direction": 90, "voltage": 0.0}, "model": "weather"}

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

## Deployment

- add build makefile - DONE
- deploy to docker

## Backend

- add type in device.expose for http data
- add refresh functionality to ping mqtt device for when we just started server an we want to awake devices ?
- add remove /force remove/block functionality
- add configure exposes device functionality

- add auth0
- add support for https and websocket TLS

- http polling devices support
- add more ws operation responses - eg success or error
- test autiomation loading. configureAction for sanitizing numeric type data
- do we need to unsubsribe from removed/renamed topic ??
- Test new logic in RegisterBridge
- backup automations

## Features

- Add schemes - living room with grouped devices

## frontend

- add type in device.expose for http data ?
- add functionality to enable/disable a trigger
- add log window in frontend

BUGS:

- Non bridge new device
  when new nont Bridge device joins
  because it hasnt Id, we build one Id on regisration. So we cant store it in metrics straighr away.
  we would have to set it up afterwards
  Also not sure if server is restarted that we have stored that information eg device id in the store, TO be tested
-

TODO

- frontend - device settings component - DONE
- frontend - test metrics graph - need mocked data in test node server ! - DONE
- metrics results could have property from/to so we know the range for ui purposes - DONE
- Non bridge devices . eg HTTP need more investigation/testing
- error reporting - important
- metrics repo - keep for x days - DONE

frontend - tabs - load tab on click -(leave for now)
frontend - add app settings in main page - fix layout
frontend - add navigation for pages - use vuetify and redesign layout
frontend - handle timerange enum colours

toggle for live data ? later

once we send the request
store response in metrics store ? needs thinking if we need that

# Logging

add log rotation - DONE
send mqqt message to enable log type from bridge to be emmited for zigbee2mqtt event logs
add remote log enable in UI
add download file log in UI

Backend TODO
automation schedule
if automation has schedule then it will be enabled/disabled accordingly during the set hours/days

device bridge utilize diagnostics category
api limiter
cache with expiration

METRICS backend TODO

- returns lis of period for ui to choose from - NO
- consider sampling data if too large data set ?
- index entries = bolt.Bucket.CreateIndex

Check for disalbed items in bridge - see if we can add them if online

and prolem below ?

[INFO]: 2024-09-15 19:50:31 - Start metrics cleanup
[ERROR]: 2024-09-15 19:50:31 - Error during metrics cleanup: parsing time "brightness2024-08-25T18:04:56." as "2006-01-02T15:04:05.000000000Z": cannot parse "brightness2024-08-25T18:04:56." as "2006"
[INFO]: 2024-09-15 19:50:31 - End metrics cleanup
[INFO]: 2024-09-15 19:50:31 - Start metrics cleanup
[ERROR]: 2024-09-15 19:50:31 - Error during metrics cleanup: parsing time "brightness2024-08-25T18:04:56." as "2006-01-02T15:04:05.000000000Z": cannot parse "brightness2024-08-25T18:04:56." as "2006"
[INFO]: 2024-09-15 19:50:31 - End metrics cleanup



npm install @vuepic/vue-datepicker