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

const luminance_luxMin = 10;
const luminance_luxMax = 600;
// Our port
let port = 3000;

// App and server
let app = express();
let server = http.createServer(app).listen(port);
console.log("[" + currentTime() + "] server listening at port " + port);

let devicesConfig = [
  {
    name: "device 1",
    method: "mqtt",
    type: "TH",
    temperatureOffset: 0.6,
    humidityOffset: 11.3,
    delayInMs: 20000,
    humidity: humidityMin + 6,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
  {
    name: "device 2",
    method: "mqtt",
    type: "presence",
    presence: true,
    luminance_lux_offset: 12,
    delayInMs: 30000,
    luminance_lux: luminance_luxMin,
    luminance_luxLastChanged: moment(),
    presenceLastChanged: moment(),
  },
  {
    name: "device 3",
    method: "http",
    type: "TH",
    temperatureOffset: 0.1,
    humidityOffset: 15.9,
    delayInMs: 40000,
    humidity: humidityMin + 11,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
];

expressWs(app, server);

// Get the /ws websocket route
app.ws("/ws", async function (ws, req) {
  console.log("client connected");

  var devices = onConnectBuildPayload(devicesConfig);
  var r = JSON.stringify({ type: "connected", payload: devices });
  ws.send(r);

  // simulate random websocket messages
  devicesConfig.forEach((config) => {
    setInterval(function () {
      var device = buildPayload("newdata", config);

      var d = JSON.stringify({ type: "deviceUpdated", payload: device });
      ws.send(d);
    }, config.delayInMs);
  });

  ws.on("message", async function (msg) {
    console.log("message received" + msg);
  });
});

function onConnectBuildPayload(devicesConfig) {
  var devices = [];
  devicesConfig.forEach((config) => {
    var device = buildPayload("connected", config);
    devices.push(device);
  });

  return devices;
}

function buildPayload(status, settings) {
  var payload;
  if (settings.method == "mqtt") {
    if (settings.type == "TH") {
      payload = mockMqttTHDevicePayload(status, settings);
    } else if (settings.type == "presence") {
      payload = mockMqttPresenceDevicePayload(status, settings);
    }
  } else if (settings.method == "http") {
    if (settings.type == "TH") {
      payload = mockHttpTHDevicePayload(status, settings);
    }
  }

  var data = {
    id: settings.name,
    payload: payload,
  };
  return data;
}

function currentTime() {
  var isoNow = moment().tz("Europe/London");
  return isoNow.format();
}
function mockHttpTHDevicePayload(state, settings) {
  var device = {
    id: settings.name,
    type: "http",
    sensors: {
      humidity: getMockHumidity(settings),
      temperature: getMockTemperature(settings),
      pressure: 68,
    },
    stats: {
      availability: "online",
      last_seen: currentTime(),
    },
  };

  return device;
}

function mockMqttTHDevicePayload(state, settings) {
  var device = {
    id: settings.name,
    type: "mqtt",
    sensors: {
      humidity: getMockHumidity(settings),
      temperature: getMockTemperature(settings),
    },
    stats: {
      availability: "online",
      last_seen: currentTime(),
      linkquality: 47,
      battery: 98,
    },
  };

  return device;
}

function mockMqttPresenceDevicePayload(state, settings) {
  var device = {
    id: settings.name,
    type: "mqtt",
    sensors: {
      presence: false,
      illuminance_lux: 103,
    },
    stats: {
      availability: "online",
      last_seen: currentTime(),
      linkquality: 67,
    },
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
