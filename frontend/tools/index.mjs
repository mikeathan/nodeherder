    import path from 'path'
    import { fileURLToPath } from 'url'
    
    const __filename = fileURLToPath(import.meta.url);
    const __dirname = path.dirname(__filename);
    
    import express from 'express'
    import expressWs from 'express-ws'
    import http from 'http'
    
    // Our port
    let port = 3000;
    
    // App and server
    let app = express();
    let server = http.createServer(app).listen(port);    
    let deviceId = 1;

    const deviceData = ref([
    {
        "name": "TH1",
        "payload":
        {
        "battery": 100,
        "humidity": 69.8,
        "last_seen": "2023-05-06T19:13:37+01:00",
        "linkquality": 29,
        "temperature": 20,
        "voltage": 3000
        }
    }, {
        "name": "TH2",
        "payload":
        {
        "battery": 100,
        "humidity": 61.8,
        "last_seen": "2023-05-09T17:07:22+01:00",
        "linkquality": 32,
        "temperature": 22.2,
        "voltage": 3000
        }
    }
    ]);

    let data = [{
        "name":"device 1",
        "payload":""
    },{
        "name":"device 2",
        "payload":""
    }]

    expressWs(app, server);
   
    // Get the /ws websocket route
    app.ws('/ws', async function(ws, req) {
        console.log("connected");
        data.forEach((device)=>{
            device.payload = "connected"
            ws.send(JSON.stringify(device));
        });
       
        setInterval(function(){

            let msg = "some message id ="+deviceId++;
            data.forEach((device)=>{
                device.payload = msg
                ws.send(JSON.stringify(device));
            });
         },5000);  
            

        ws.on('message', async function(msg) {
            console.log("message received" +msg);
        });
    });


function mockTHDeviceData(device){
   device.battery = 100;
   device.humidity = 60.1;
   device.last_seen = Date.now();
   device.linkquality = 47;
   device.temperature = 19.1;
   device.voltage = 3000;
}