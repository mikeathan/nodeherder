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

Occupancy devices

UI:
Occupancy
luminance
Status/availability: online or offline
No last seen
Or golang filters messages and uptates the last seen

Backend:

filters messages and uptates the last seen
Handle no battery in payload
Device logic for extending functionality

Setting configuration for devices
Triggers and binding
Add api endpoint for collecting data from WiFi sensors

Circuit breaker
