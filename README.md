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


Frontend
settings page
      "unit_system" : "metric",
home
sensor_data

Web server

Starts mqtt
   Receives message
   Send websocket message
   Send metrics rest-api request
 
Websocket listener

Metrics listener

TODO:
pass options to mqttclient and then create
mediator 
websockets

