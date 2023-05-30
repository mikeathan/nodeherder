import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

import express from 'express'
import WebSocket from 'ws'
import http from 'http'


let port = 3000;

const app = express();
const server = http.createServer(app);
const websocketServer = new WebSocket.Server({ server });

websocketServer.on("connection", (webSocketClient) => {
    console.log("Websocket connection opened");
    webSocketClient.send(JSON.stringify({ "message" : "hello" }));
    // webSocketClient.send("The time is: ");
    // setInterval(() => {
    //     let time = new Date();
    //     webSocketClient.send("The time is: " + time.toTimeString());
    // }, 1000);
  });
  
  //start the web server
  server.listen(port, () => {
    console.log("Websocket server started on port 3000");
  });
