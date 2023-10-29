import path from "path";
import { fileURLToPath } from "url";
import moment from "moment";
import "moment-timezone";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

import express from "express";
import expressWs from "express-ws";
import http from "http";
import { createRequire } from "module";
const devicesFullPath = "../../docs/devices.json";

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
//const fs = require("fs");

var automationMap = new Map([
  [
    "0xa4c13894070052fc",
    {
      id: "0xa4c13894070052fc",
      friendlyName: "Human presence",
      description: "Attic light test automation",
      enabled: true,
      triggers: [
        {
          name: "presence",
          conditions: [{ name: "presence", value: false, equality: "=" }],
          action: {
            id: "0x70ac08fffefafeca",
            friendlyname: "Attic light",
            type: "light",
            property: "state",
            data: "OFF",
            delay: 300000000000,
          },
        },
        {
          name: "presence",
          conditions: [
            { name: "presence", value: true, equality: "=" },
            { name: "lux", value: 30, equality: "<=" },
          ],
          action: {
            id: "0x70ac08fffefafeca",
            friendlyname: "Attic light",
            type: "light",
            property: "state",
            data: "ON",
          },
        },
      ],
    },
  ],
  [
    "0x00124b0029207763",
    {
      id: "0x00124b0029207763",
      friendlyName: "TH1",
      description: "Temperature and humidity test automation",
      enabled: false,
      triggers: [
        {
          name: "presence",
          conditions: [{ name: "presence", value: false, equality: "=" }],
          action: {
            id: "0x70ac08fffefafeca",
            friendlyname: "Attic light",
            type: "light",
            property: "state",
            data: "OFF",
            delay: 300000000000,
          },
        },
        {
          name: "presence",
          conditions: [
            { name: "presence", value: true, equality: "=" },
            { name: "lux", value: 30, equality: "<=" },
          ],
          action: {
            id: "0x70ac08fffefafeca",
            friendlyname: "Attic light",
            type: "light",
            property: "state",
            data: "ON",
          },
        },
      ],
    },
  ],
]);

let automation1Trigger =
  '{"type":"automations","payload":[{"id":"0xa4c13894070052fc","friendlyName":"Human presence","description":"Attic light test automation","enabled":true,"triggers":[{"name":"presence","conditions":[{"name":"presence","value":false,"equality":"="}],"action":{"id":"0x70ac08fffefafeca","friendlyname":"Attic light","type":"light","property":"state","data":"OFF","delay":300000000000}},{"name":"presence","conditions":[{"name":"presence","value":true,"equality":"="},{"name":"lux","value":30,"equality":"<="}],"action":{"id":"0x70ac08fffefafeca","friendlyname":"Attic light","type":"light","property":"state","data":"ON"}}]}]}';

expressWs(app, server);

var devicesPayload = loadDevices();

// Get the /ws websocket route
app.ws("/ws", async function (ws, req) {
  console.log("client connected");

  settings.forEach((s) => {
    setInterval(function () {
      var updatePayload = buildDeviceUpdatedPayload(s);

      console.log("debug ", updatePayload);
      var d = JSON.stringify({
        type: "deviceUpdated",
        payload: updatePayload,
      });

      ws.send(d);
    }, s.delayInMs);
  });

  ws.on("message", async function (msg) {
    console.log("message received" + msg);

    const obj = JSON.parse(msg);
    switch (obj.type) {
      case "loadAutomations":
        sendMessage(ws, "automations", getAutomations());
        break;

      case "loadDevices":
        //var payload = buildNewDevicesPayload();
        sendMessage(ws, "devices", devicesPayload);
        break;

      case "saveAutomation":
        var automation = obj.payload;
        automationMap.set(automation.id, automation);
        sendOperationSuccess(ws);
        break;

      case "deleteAutomation":
        if (!automationMap.has(obj.payload.id)) {
          sendOperationFailed(
            "Delete failed. Automation id " + obj.payload.id + " not found"
          );
          return;
        }

        automationMap.delete(obj.payload.id);
        sendMessage(ws, "automations", getAutomations());

        break;

      case "deleteAutomationTrigger":
        var aId = obj.payload.automationId;
        var tId = obj.payload.triggerId;

        if (!automationMap.has(aId)) {
          sendOperationFailed(
            "Delete trigger. automation id " + aId + " not found"
          );
          return;
        }

        var automation = automationMap.get(aId);
        if (tId >= automation.triggers.length) {
          sendOperationFailed("trigger index" + tId + " out of bounds.");
          return;
        }

        automation.triggers.splice(tId, 1);
        sendMessage(ws, "automationUpdated", automation);
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

function getAutomations() {
  const items = [];
  automationMap.forEach((values, k) => {
    items.push(values);
  });

  return items;
}

function sendMessage(ws, event, payload) {
  var msg = JSON.stringify({
    type: event,
    payload: payload,
  });

  ws.send(msg);
}

function sendOperationSuccess(ws) {
  var msg = JSON.stringify({
    type: "operationSuccess",
    payload: {},
  });

  ws.send(msg);
}

function sendOperationFailed(ws, message) {
  var msg = JSON.stringify({
    type: "operationFailed",
    payload: message,
  });

  ws.send(msg);
}

function saveAutomation(ws, automation) {
  automation1Trigger = automation;
}

function buildDeviceUpdatedPayload(s) {
  var func = updateDeviceMap[s.id];
  try {
    var payload = func(s);
    return payload;
  } catch (error) {
    console.log("id:", s.id + "error:" + error);
  }
  return undefined;
}

function currentTime() {
  var isoNow = moment().tz("Europe/London");
  return isoNow.format();
}

let settings = [
  {
    id: "0xa4c13894070052fc",
    friendlyName: "Human presence",
    availability: "offline",
    method: "mqtt",
    luminance_lux_offset: 12,
    delayInMs: 35000,
    luminance_lux: luminance_luxMin,
    luminance_luxLastChanged: moment(),
    presenceLastChanged: moment(),
  },
  // {
  //   id: "0x00124b00146c31cd",
  //   friendlyName: "Motion sensor 1",
  //   availability: "offline",
  //   method: "mqtt",
  //   temperatureOffset: 1.2,
  //   delayInMs: 20000,
  //   temperature: temperatureMin,
  //   temperatureLastChanged: moment(),
  // },
  // {
  //   id: "92fe86b7",
  //   friendlyName: "weather node 1",
  //   availability: "offline",
  //   method: "http",
  //   delayInMs: 120000,
  // },
  // {
  //   id: "0x00124b0029207763",
  //   friendlyName: "TH01",
  //   availability: "offline",
  //   method: "mqtt",
  //   temperatureOffset: 0.6,
  //   humidityOffset: 11.3,
  //   delayInMs: 5000,
  //   humidity: humidityMin + 6,
  //   temperature: temperatureMin,
  //   temperatureLastChanged: moment(),
  //   humidityLastChanged: moment(),
  // },
];

// let newDeviceMap = {};
// newDeviceMap["92fe86b7"] = mockAddWeatherNode1v2;
// newDeviceMap["0x00124b0029207763"] = mockAddTH01v2;
// newDeviceMap["0xa4c13894070052fc"] = mockAddHumanPresencev2;
// newDeviceMap["0x00124b00146c31cd"] = mockAddMotionSensorv2;

let updateDeviceMap = {};
updateDeviceMap["92fe86b7"] = mockUpdateWeatherNode1v2;
updateDeviceMap["0x00124b0029207763"] = mockUpdateTH01v2;
updateDeviceMap["0xa4c13894070052fc"] = mockUpdateHumanPresencev2;
updateDeviceMap["0x00124b00146c31cd"] = mockUpdateMotionSensorv2;

function mockUpdateHumanPresencev2(settings) {
  var device = {
    id: "0xa4c13894070052fc",
    last_seen: currentTime(),
    data: {
      illuminance_lux: 9,
      presence: true,
    },
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.data.availability = availability;
  }
  return device;
}

function mockUpdateMotionSensorv2(settings) {
  var device = {
    id: "0x00124b00146c31cd",
    last_seen: currentTime(),
    data: {
      occupancy: true,
      temperature: getMockTemperature(settings),
    },
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.data.availability = availability;
  }

  return device;
}

function mockUpdateWeatherNode1v2(settings) {
  var device = {
    id: "92fe86b7",
    last_seen: currentTime(),
    data: {
      temperature: getMockTemperature(settings),
    },
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.data.availability = availability;
  }
  return device;
}

function mockUpdateTH01v2(settings) {
  var device = {
    id: "0x00124b0029207763",
    last_seen: currentTime(),
    data: {
      temperature: getMockTemperature(settings),
      humidity: getMockHumidity(settings),
    },
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.data.availability = availability;
  }
  return device;
}

function setDeviceOnline(settings) {
  if (settings.availability == "offline") {
    settings.availability = "online";
    return settings.availability;
  }

  return undefined;
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

function loadDevices() {
  const require = createRequire(import.meta.url);
  var data = require(devicesFullPath);
  return data.payload;
}
