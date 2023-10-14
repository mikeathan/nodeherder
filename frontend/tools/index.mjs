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

let automation1Trigger =
  '{"type":"automations","payload":[{"id":"0xa4c13894070052fc","friendlyName":"Human presence","description":"Attic light test automation","enabled":true,"triggers":[{"name":"presence","conditions":[{"name":"presence","value":false,"equality":"="}],"action":{"id":"0x70ac08fffefafeca","friendlyname":"Attic light","type":"light","property":"state","data":"OFF","delay":300000000000}},{"name":"presence","conditions":[{"name":"presence","value":true,"equality":"="},{"name":"lux","value":30,"equality":"<="}],"action":{"id":"0x70ac08fffefafeca","friendlyname":"Attic light","type":"light","property":"state","data":"ON"}}]}]}';

let featuresMsg =
  '{"type":"bridgeFeatures","payload":[{"id":"0x70ac08fffefafeca","properties":{"brightness":{"name":"brightness","type":"numeric","attributes":{"max":254,"min":0}},"color_temp":{"name":"color_temp","type":"numeric","attributes":{"max":500,"min":150}},"color_temp_startup":{"name":"color_temp_startup","type":"numeric","attributes":{"max":500,"min":150}},"state":{"name":"state","type":"binary","attributes":{"off":"OFF","on":"ON","toggle":"TOGGLE"}}}},{"id":"0x001788010b99ea7f","properties":{"brightness":{"name":"brightness","type":"numeric","attributes":{"max":254,"min":0}},"color":{"name":"color","type":"composite","attributes":{}},"color_temp":{"name":"color_temp","type":"numeric","attributes":{"max":500,"min":153}},"color_temp_startup":{"name":"color_temp_startup","type":"numeric","attributes":{"max":500,"min":153}},"state":{"name":"state","type":"binary","attributes":{"off":"OFF","on":"ON","toggle":"TOGGLE"}}}}]}';
expressWs(app, server);

// Get the /ws websocket route
app.ws("/ws", async function (ws, req) {
  console.log("client connected");

  // bridge features
  // ws.send(featuresMsg);

  settings.forEach((s) => {
    setInterval(function () {
      var updatePayload = buildDeviceUpdatedPayload(s);
      if (updatePayload.data.availability != undefined) {
        //if (s.availability === "offline") {
        // probaly check payload as well
        // return;
        // }
        var d = JSON.stringify({
          type: "devicePropertiesUpdated",
          payload: updatePayload,
        });
        ws.send(d);
        return;
      }

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
        var msg = onLoadAutomationBuildResponse();
        ws.send(msg);
        break;

      case "loadDevices":
        var payload = buildNewDevicesPayload();
        var msg = JSON.stringify({ type: "devices", payload: payload });
        ws.send(msg);
        break;

      case "saveAutomation":
        console.log(obj.payload);
        break;

      case "loadBridgeFeatures":
        ws.send(featuresMsg);
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

function buildDeviceUpdatedPayload(s) {
  var func = updateDeviceMap[s.id];
  var payload = func(s);
  return payload;
}

function buildNewDevicesPayload() {
  var devices = [];
  settings.forEach((s) => {
    var func = newDeviceMap[s.id];
    var newDevicePayload = func(s); //
    devices.push(newDevicePayload);
  });
  return devices;
}

function currentTime() {
  var isoNow = moment().tz("Europe/London");
  return isoNow.format();
}

let settings = [
  {
    id: "0x00124b0029207763",
    friendlyName: "TH01",
    availability: "online",
    method: "mqtt",
    temperatureOffset: 0.6,
    humidityOffset: 11.3,
    delayInMs: 52000,
    humidity: humidityMin + 6,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
  {
    id: "0xa4c13894070052fc",
    friendlyName: "Human presence",
    availability: "online",
    method: "mqtt",
    luminance_lux_offset: 12,
    delayInMs: 35000,
    luminance_lux: luminance_luxMin,
    luminance_luxLastChanged: moment(),
    presenceLastChanged: moment(),
  },
  {
    id: "0x00124b00146c31cd",
    friendlyName: "Motion sensor 1",
    availability: "online",
    method: "mqtt",
    temperatureOffset: 1.2,
    delayInMs: 20000,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
  },
  {
    id: "0x123456",
    friendlyName: "weather node 1",
    availability: "offline",
    method: "http",
    delayInMs: 120000,
  },
];

let newDeviceMap = {};
newDeviceMap["0x123456"] = mockAddWeatherNode1v2;
newDeviceMap["0x00124b0029207763"] = mockAddTH01v2;
newDeviceMap["0xa4c13894070052fc"] = mockAddHumanPresencev2;
newDeviceMap["0x00124b00146c31cd"] = mockAddMotionSensorv2;

let updateDeviceMap = {};
updateDeviceMap["0x123456"] = mockUpdateWeatherNode1v2;
updateDeviceMap["0x00124b0029207763"] = mockUpdateTH01v2;
updateDeviceMap["0xa4c13894070052fc"] = mockUpdateHumanPresencev2;
updateDeviceMap["0x00124b00146c31cd"] = mockUpdateMotionSensorv2;

function mockAddWeatherNode1v2(settings) {
  var device = {
    id: "0x123456",
    friendlyName: "weather node 1",
    connection_type: "http",
    power_source: "battery",
    exposes: {
      humidity: {
        name: "humidity",
        unit: "%",
        data: 63.1,
        properties: {},
      },
      temperature: {
        name: "temperature",
        unit: "°C",
        data: 15,
        properties: {},
      },
    },
    properties: {
      availability: settings.availability,
      battery: 100,
      last_seen: currentTime(),
    },
  };

  return device;
}

function mockAddTH01v2(settings) {
  var device = {
    id: "0x00124b0029207763",
    friendlyName: "TH01",
    description: "Temperature and Humidity TH01 sensor",

    connection_type: "mqtt",
    power_source: "battery",
    exposes: {
      humidity: {
        name: "humidity",
        description: "Measured relative humidity",
        unit: "%",
        data: 83.91,
        properties: {},
      },
      temperature: {
        name: "temperature",
        description: "Measured temperature value",
        unit: "°C",
        data: 19.87,
        properties: {},
      },
    },
    properties: {
      availability: settings.availability,
      battery: 100,
      last_seen: currentTime(),
      linkquality: 32,
    },
  };

  return device;
}

function mockAddHumanPresencev2(settings) {
  var device = {
    id: "0xa4c13894070052fc",
    friendlyName: "Human presence",
    description: "Human presence sensor",
    connection_type: "mqtt",
    power_source: "mains",
    exposes: {
      illuminance_lux: {
        name: "illuminance_lux",
        description: "Measured illuminance in lux",
        unit: "lx",
        data: 3,
        properties: {},
      },
      presence: {
        name: "presence",
        description: "Indicates whether the device detected presence",
        data: false,
        properties: {},
      },
    },
    properties: {
      availability: settings.availability,
      last_seen: currentTime(),
      linkquality: 58,
    },
  };

  return device;
}

function mockAddMotionSensorv2(settings) {
  var device = {
    id: "0x00124b00146c31cd",
    friendlyName: "Motion sensor 1",
    description: "Motion sensor and temperature device",
    connection_type: "mqtt",
    power_source: "battery",
    exposes: {
      occupancy: {
        name: "occupancy",
        description: "Indicates whether the device detected occupancy",
        data: false,
        properties: {},
      },
      temperature: {
        name: "temperature",
        description: "Measured temperature value",
        unit: "°C",
        data: 23.75,
        properties: {},
      },
    },
    properties: {
      availability: settings.availability,
      battery: 7,
      last_seen: currentTime(),
      linkquality: 29,
    },
  };

  return device;
}

function mockUpdateHumanPresencev2(settings) {
  var device = {
    id: "0xa4c13894070052fc",
    last_seen: currentTime(),
    data: {
      illuminance_lux: 9,
      presence: true,
    },
  };

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
  return device;
}

function mockUpdateWeatherNode1v2(settings) {
  var device = {
    id: "0x123456",
    last_seen: currentTime(),
    data: {
      availability: getMockAvailability(settings),
    },
  };

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

  return device;
}

function getMockAvailability(settings) {
  if (settings.availability == "online") {
    settings.availability = "offline";
  } else {
    settings.availability = "online";
  }

  return settings.availability;
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
