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
    console.log("server listening at port "+ port);
    let deviceId = 1;


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

            let msg = "some message id ="+deviceId++;
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


    // TODO:
    // match time to format: 2023-05-31T19:02:28+01:00
function mockTHDevicePayload(){
    var device={
        state: "",
        battery : 100,
        humidity : 60.1,
        last_seen :  new Date().toLocaleString(),
        linkquality : 47,
        temperature : 19.1,
        voltage : 3000
    };
    //var jsonText = JSON.stringify(device)
    return device;
}