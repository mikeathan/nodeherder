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
    
    expressWs(app, server);
    
    // Get the /ws websocket route
    app.ws('/ws', async function(ws, req) {
        console.log("connected");
        ws.send(JSON.stringify({ "message" : "connected" }));

        ws.on('message', async function(msg) {
            console.log(msg);
            ws.send(JSON.stringify({ "message" : "received" }));
        });
    });