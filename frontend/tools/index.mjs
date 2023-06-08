import path from "path";
import { fileURLToPath } from "url";
import moment from "moment";
import "moment-timezone";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

import express from "express";
import expressWs from "express-ws";
import http from "http";

// temperature
const temperatureChangeDelaySec = 5;
const temperatureMin = 10.0;
const temperatureMax = 40.0;

const humidityChangeDelaySec = 30;
const humidityMin = 30.0;
const humidityMax = 100.0;

// Our port
let port = 3000;

let counter = 0;
// App and server
let app = express();
let server = http.createServer(app).listen(port);
console.log("[" + currentTime() + "] server listening at port " + port);

let devicesConfig = [
  {
    name: "device 1",
    temperatureOffset: 0.6,
    humidityOffset: 11.3,
    delayInMs: 20000,
    humidity: humidityMin,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
  {
    name: "device 2",
    temperatureOffset: 1.1,
    humidityOffset: 9.7,
    delayInMs: 30000,
    humidity: humidityMin,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
];

expressWs(app, server);

// Get the /ws websocket route
app.ws("/ws", async function (ws, req) {
  console.log("client connected");

  devicesConfig.forEach((config) => {
    var device = buildPayload("connected", config);
    ws.send(device);
  });

  devicesConfig.forEach((config) => {
    setInterval(function () {
      counter++;
      var device = buildPayload("updated", config);
      ws.send(device);
    }, config.delayInMs);
  });

  ws.on("message", async function (msg) {
    console.log("message received" + msg);
  });
});

function buildPayload(status, settings) {
  var data = {
    name: settings.name,
    payload: mockTHDevicePayload(status, settings),
  };
  var device = JSON.stringify(data);
  return device;
}

function currentTime() {
  var isoNow = moment().tz("Europe/London");
  return isoNow.format();
}

function mockTHDevicePayload(state, settings) {
  var device = {
    state: state + "=" + counter,
    battery: 100,
    humidity: getMockHumidity(settings),
    last_seen: currentTime(),
    linkquality: 47,
    temperature: getMockTemperature(settings),
    voltage: 3000,
  };

  return device;
}

function getMockTemperature(settings) {
  var diff = moment().diff(settings.temperatureLastChanged);
  var duration = moment.duration(diff);

  if (duration.seconds() < temperatureChangeDelaySec) {
    return settings.temperature;
  }

  if (settings.temperature > temperatureMax) {
    settings.temperature = temperatureMin;
  }

  settings.temperature += settings.temperatureOffset;
  settings.temperatureLastChanged = moment();
  return settings.temperature;
}

function getMockHumidity(settings) {
  var diff = moment().diff(settings.humidityLastChanged);
  var duration = moment.duration(diff);

  if (duration.seconds() < humidityChangeDelaySec) {
    return settings.humidity;
  }

  if (settings.humidity > humidityMax) {
    settings.humidity = humidityMin;
  }

  settings.humidity += settings.humidityOffset;
  settings.humidityLastChanged = moment();
  return settings.humidity;
}
