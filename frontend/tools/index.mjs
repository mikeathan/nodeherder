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
let port = 3000;
let pingTimer = 0;
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
    delayInMs: 52000,
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
    delayInMs: 35000,
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
    delayInMs: 65000,
    humidity: humidityMin + 11,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
];

let automation1Trigger =
  '{"type":"automations","payload":[{"id":"1d57516b","name":"Human presence","description":"Attic light test automation","enabled":true,"triggers":[{"name":"presence","conditions":[{"name":"presence","value":false,"equality":"="}],"action":{"friendlyname":"Attic light","type":"light","property":"state","data":false,"delay":300000000000}},{"name":"presence","conditions":[{"name":"presence","value":true,"equality":"="},{"name":"lux","value":30,"equality":"<="}],"action":{"friendlyname":"Attic light","type":"light","property":"state","data":true}}]}]}';

let devicesResponsePayload =
  '{"type":"devices","payload":[{"id":"device 1","conn":"mqtt","power_source":"battery","sensors":{"humidity":41.3,"temperature":18.400000000000002},"stats":{"availability":"online","last_seen":"2023-09-28T11:04:12+01:00","linkquality":47,"battery":98}},{"id":"device 2","conn":"mqtt","power_source":"mains","sensors":{"presence":false,"illuminance_lux":103},"stats":{"availability":"online","last_seen":"2023-09-28T11:04:12+01:00","linkquality":67}},{"id":"device 3","conn":"http","power_source":"","sensors":{"humidity":41,"temperature":10,"pressure":68},"stats":{"availability":"offline","last_seen":"2023-09-28T11:04:12+01:00"}}]}';
expressWs(app, server);

// Get the /ws websocket route
app.ws("/ws", async function (ws, req) {
  console.log("client connected");

  // simulate device updated messages
  devicesConfig.forEach((config) => {
    setInterval(function () {
      var device = buildPayload("newdata", config);
      if (device.stats.availability === "online") {
        var d = JSON.stringify({ type: "deviceUpdated", payload: device });
        ws.send(d);
      }
    }, config.delayInMs);
  });

  ws.on("message", async function (msg) {
    console.log("message received" + msg);

    const obj = JSON.parse(msg);
    switch (obj.type) {
      case "loadAutomations":
        var msg = onLoadAutomationBuildResponse();
        ws.send(msg);
        break;

      case "loadDevices":
        var msg = onLoadDevicesBuildResponse();
        ws.send(msg);
        break;
      case "saveAutomation":
        console.log(obj.payload);
        break;
      case "pong":
        break;
      default:
        console.log("ws unhandled type: ", msg);
    }
  });

  ws.on("error", function (error) {
    console.log("Cannot start server" + error);
  });

  ws.on("close", function (code, message) {
    console.log("Disconnection: " + code + ", " + message);
    clearInterval(pingTimer);
  });

  try {
    pingTimer = setInterval(() => {
      var msg = JSON.stringify({ type: "ping", payload: "" });
      ws.send(msg);
    }, 30000);
  } catch (err) {
    console.log("ping error ", err);
  }
});

function onLoadAutomationBuildResponse() {
  return automation1Trigger;
}

function saveAutomation(automation) {
  automation1Trigger = automation;
}

function onLoadDevicesBuildResponse() {
  return devicesResponsePayload;
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

  return payload;
}

function currentTime() {
  var isoNow = moment().tz("Europe/London");
  return isoNow.format();
}

function mockHttpTHDevicePayload(state, settings) {
  var device = {
    id: settings.name,
    conn: "http",
    power_source: "", // unknown
    sensors: {
      humidity: settings.humidity,
      temperature: settings.temperature,
      pressure: 68,
    },
    stats: {
      availability: "offline",
      last_seen: currentTime(),
    },
  };

  return device;
}

function mockMqttTHDevicePayload(state, settings) {
  var device = {
    id: settings.name,
    conn: "mqtt",
    power_source: "battery",
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
    conn: "mqtt",
    power_source: "mains",
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
  if (settings.availability == "offline") {
    return settings.temperature;
  }

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
  if (settings.availability == "offline") {
    return settings.humidity;
  }

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
