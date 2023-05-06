sample data : TH1
{
    "battery": 100,
    "humidity": 69.8,
    "last_seen": "2023-05-06T19:13:37+01:00",
    "linkquality": 29,
    "temperature": 20,
    "voltage": 3000
}

// libs
go get github.com/mehdihadeli/go-mediatry


Web server

Starts mqtt
   Receives message
   Send websocket message
   Send metrics rest-api request
 
Websocket listener

Metrics listener