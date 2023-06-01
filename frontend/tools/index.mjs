    import path from 'path'
    import { fileURLToPath } from 'url'
    import moment from 'moment'
    import 'moment-timezone'

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
    console.log("["+currentTime() + "] server listening at port "+ port );

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

            var p = mockTHDevicePayload();
            p.state = "connected";
            console.log(p);
            device.payload = p
            ws.send(JSON.stringify(device));
        });
       
        setInterval(function(){

            data.forEach((device)=>{
                var p = mockTHDevicePayload();
                 p.state = "updated";
                 console.log(p);
                 device.payload = p
                ws.send(JSON.stringify(device));
            });
         },5000);  
            

        ws.on('message', async function(msg) {
            console.log("message received" +msg);
        });
    });

function currentTime(){
    var isoNow = moment().tz('Europe/London')
    return isoNow.format();
}

function mockTHDevicePayload(){
    var device={
        state: "",
        battery : 100,
        humidity : 60.1,
        last_seen :  currentTime(),
        linkquality : 47,
        temperature : 19.1,
        voltage : 3000
    };

    return device;
}

