/* eslint-disable no-console */
/* eslint-disable indent */
import moment from 'moment';
import 'moment-timezone';

import express from 'express';
import expressWs from 'express-ws';
import http from 'http';
import { createRequire } from 'module';
const devicesFullPath = '../../docs/devices.json';
const lightMetricsFullPath = './metrics/light.json';
const temperatureMetricsFullPath = './metrics/temperature.json';
const presenceMetricsFullPath = './metrics/presence.json';
// temperature
const temperatureChangeDelaySec = 5;
const temperatureMin = 10.0;
const temperatureMax = 40.0;

const humidityChangeDelaySec = 30;
const humidityMin = 30.0;
const humidityMax = 100.0;

const luminance_luxMin = 10;

let port = 3000;

let consoleLogIntervalId = 0;
const logSeverity = ['info', 'warning', 'error', 'critical'];

// App and server
let app = express();
let server = http.createServer(app).listen(port);
console.log('[' + currentTime() + '] server listening at port ' + port);

var appConfig = {
  history: {
    sleepTimeout: { value: 12, unit: 'hours' },
    expireAt: { value: 10, unit: 'days' },
  },
  logger: {
    enableRemoteLogger: false,
  },
  bridge: {
    maxTimeAllowed: { value: 120, unit: 'seconds' },
    permitJoin: false,
  },
  devices: {
    '0xa4c13894070052fc': {
      id: '0xa4c13894070052fc',
      disabled: false,
      metricsEnabled: false,
      rateLimit: 10000,
    },
    '0x001788010d7d9d3f': {
      id: '0x001788010d7d9d3f',
      disabled: false,
      metricsEnabled: false,
      rateLimit: 50000,
    },
    '0x70ac08fffefafeca': {
      id: '0x70ac08fffefafeca',
      disabled: false,
      metricsEnabled: true,
      rateLimit: 50000,
    },
  },
};
var automationMap = new Map([
  [
    '0xa4c13894070052fc',
    {
      id: '0xa4c13894070052fc',
      friendlyname: 'Human presence',
      description: 'Attic light test automation',
      enabled: true,
      schedules: [
        {
          startAt: '11:30',
          type: 'enable',
        },
        {
          startAt: '05:00',
          type: 'disable',
        },
      ],

      triggers: [
        {
          name: 'presence',
          conditions: [
            {
              name: 'presence',
              value: false,
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'state',
              data: 'OFF',
              delay: 300000,
              type: 'TriggerAction',
            },
          ],
        },
        {
          name: 'presence',
          conditions: [
            {
              name: 'presence',
              value: true,
              equality: '=',
            },
            {
              name: 'illuminance_lux',
              value: 30,
              equality: '<=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'state',
              data: 'ON',
              type: 'TriggerAction',
            },
          ],
        },
      ],
    },
  ],
  [
    '0x001788010d7d9d3f',
    {
      id: '0x001788010d7d9d3f',
      friendlyname: 'Hue tap dial switch',
      description: 'Light switch automation',
      enabled: false,
      triggers: [
        {
          name: 'action_direction',
          conditions: [
            {
              name: 'action',
              value: 'button_2_press',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'color_temp',
              data: null,
              operation: 0,
              delay: null,
              steps: [],
              type: 'PresetRotationAction',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              name: 'action',
              value: 'button_1_press_release',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'state',
              data: 'TOGGLE',
              type: 'TriggerAction',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              name: 'action',
              value: 'dial_rotate_right_slow',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'brightness',
              data: 10,
              steps: [
                {
                  property: 'brightness',
                  operator: '-',
                  id: '0x70ac08fffefafeca',
                },
                {
                  operator: '-',
                  property: 'action_time',
                  id: '0x001788010d7d9d3f',
                },
              ],
              type: 'StepAction',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              name: 'action',
              value: 'dial_rotate_left_slow',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'brightness',
              data: 10,
              steps: [
                {
                  property: 'brightness',
                  operator: '+',
                  id: '0x70ac08fffefafeca',
                },
                {
                  operator: '+',
                  property: 'action_time',
                  id: '0x001788010d7d9d3f',
                },
              ],
              type: 'StepAction',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              name: 'action',
              value: 'button_2_press_release',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x70ac08fffefafeca',
              friendlyname: 'Attic light',
              property: 'color_temp',
              steps: [],
              type: 'PresetRotationAction',
            },
          ],
        },
      ],
    },
  ],
]);

expressWs(app, server);

var devicesPayload = loadDevices();
var metricsMap = loadMetrics();

var connected = false;
// Get the /ws websocket route
app.ws('/ws', async function (ws) {
  console.log('client connected');

  settings.forEach((s) => {
    setInterval(function () {
      if (!connected) {
        return;
      }

      var updatePayload = buildDeviceUpdatedPayload(s);
      var d = JSON.stringify({
        type: 'deviceUpdated',
        payload: updatePayload,
      });
      ws.send(d);
    }, s.delayInMs);
  });

  ws.on('message', async function (msg) {
    console.log('message received' + msg);

    const obj = JSON.parse(msg);
    switch (obj.type) {
      case 'loadAutomations':
        sendMessage(ws, 'automations', getAutomations());
        break;

      case 'loadHubState':
        const hubStatePayload = {
          config: appConfig,
          devices: devicesPayload,
        };
        sendMessage(ws, 'hubState', hubStatePayload);
        connected = true;
        break;
      case 'deviceSetValue':
        // Respond back with update value to update UI
        const updatePayload = {
          id: obj.payload.id,
          data: {
            [obj.payload.name]: obj.payload.value,
          },
          properties: {
            availability: true,
            last_seen: currentTime(),
          },
        };

        sendMessage(ws, 'deviceUpdated', updatePayload);

        break;

      case 'saveAutomation':
        var automation = obj.payload;
        automationMap.set(automation.id, automation);
        sendOperationSuccess(ws);
        break;

      case 'deleteAutomation':
        if (!automationMap.has(obj.payload.id)) {
          sendOperationFailed('Delete failed. Automation id ' + obj.payload.id + ' not found');
          return;
        }

        automationMap.delete(obj.payload.id);
        sendMessage(ws, 'automations', getAutomations());

        break;

      case 'deleteAutomationTrigger':
        var aId = obj.payload.automationId;
        var tId = obj.payload.triggerId;

        if (!automationMap.has(aId)) {
          sendOperationFailed('Delete trigger. automation id ' + aId + ' not found');
          return;
        }

        var automation = automationMap.get(aId);
        if (tId >= automation.triggers.length) {
          sendOperationFailed('trigger index' + tId + ' out of boudeviceRemovends.');
          return;
        }

        automation.triggers.splice(tId, 1);
        sendMessage(ws, 'automationUpdated', automation);
        break;

      case 'loadAppConfig':
        sendMessage(ws, 'appConfig', appConfig);
        break;
      case 'bridgePermitJoin':
        setTimeout(() => {
          if (appConfig.bridge.permitJoin != obj.payload.permitJoin) {
            runPermitJoin(ws, obj.payload);
          } else {
            console.log('permitjoin already set to ' + obj.payload.permitJoin);
            sendOperationFailed(ws, 'Permit join is already set to ' + obj.payload.permitJoin);
          }
        }, 2000);

        break;
      case 'loadMetrics':
        const payload = metricsMap[obj.payload.id];

        if (!payload) {
          console.log('Metrics for device id' + obj.payload.id + ' not found');
          // sendOperationFailed(
          //   ws,
          //   'Metrics for device id' +
          //     obj.payload.id +
          //     ' not found',
          // );
          return;
        }
        sendMessage(ws, 'metrics', payload);
        break;

      case 'saveHistoryConfig':
        appConfig.history = obj.payload;
        sendOperationSuccess(ws);
        break;

      case 'saveDeviceConfig':
        var deviceId = obj.payload.id;
        appConfig[deviceId] = obj.payload;
        sendOperationSuccess(ws);
        break;
      case 'saveLoggerConfig':
        console.log('saveLoggerConfig', obj.payload);
        appConfig.logger = obj.payload;

        if (appConfig.logger.enableRemoteLogger) {
          if (consoleLogIntervalId != 0) {
            console.log('consoleLogIntervalId already running');
            clearInterval(consoleLogIntervalId);
          }

          console.log('enableRemoteLogger');

          consoleLogIntervalId = setInterval(() => {
            const randomIndex = Math.floor(Math.random() * logSeverity.length);

            const severity = logSeverity[randomIndex];
            const msg = {
              level: severity,
              message: severity + ' message',
              timestamp: Date.now(),
            };
            sendMessage(ws, 'logger', msg);
          }, 1000);
        } else {
          console.log('disableRemoteLogger');
          clearInterval(consoleLogIntervalId);
          consoleLogIntervalId = 0;
        }
        break;

      case 'pong':
        break;

      default:
        console.log('ws unhandled type: ', msg);
        break;
    }
  });

  ws.on('error', function (error) {
    console.log('Cannot start server' + error);
  });

  ws.on('close', function (code, message) {
    console.log('Disconnection: ' + code + ', ' + message);
    connected = false;
    // clearInterval(pingTimer);
  });

  // try {
  //   pingTimer = setInterval(() => {
  //     var msg = JSON.stringify({ type: "ping", payload: "" });
  //     ws.send(msg);
  //   }, 30000);
  // } catch (err) {
  //   console.log("ping error ", err);
  // }
});

function getAutomations() {
  const items = [];
  automationMap.forEach((values) => {
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
    type: 'operationSuccess',
    payload: {},
  });

  ws.send(msg);
}

function sendOperationFailed(ws, message) {
  var msg = JSON.stringify({
    type: 'operationFailed',
    payload: message,
  });

  ws.send(msg);
}

function buildDeviceUpdatedPayload(s) {
  var func = updateDeviceMap[s.id];
  try {
    var payload = func(s);
    return payload;
  } catch (error) {
    console.log('id:', s.id + 'error:' + error);
  }
  return undefined;
}

function currentTime() {
  var isoNow = moment().tz('Europe/London');
  return isoNow.format();
}

let settings = [
  {
    id: '0xa4c13894070052fc',
    friendlyName: 'Human presence',
    availability: 'offline',
    method: 'mqtt',
    luminance_lux_offset: 12,
    delayInMs: 35000,
    luminance_lux: luminance_luxMin,
    luminance_luxLastChanged: moment(),
    presenceLastChanged: moment(),
  },
  {
    id: '0x00124b00146c31cd',
    friendlyName: 'Motion sensor 1',
    availability: 'offline',
    method: 'mqtt',
    temperatureOffset: 1.2,
    delayInMs: 2000,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
  },
  {
    id: '92fe86b7',
    friendlyName: 'weather node 1',
    availability: 'offline',
    method: 'http',
    delayInMs: 10000,
    temperatureOffset: 1.2,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
  },
  {
    id: '0x00124b0029207763',
    friendlyName: 'TH01',
    availability: 'offline',
    method: 'mqtt',
    temperatureOffset: 0.6,
    humidityOffset: 11.3,
    delayInMs: 5000,
    humidity: humidityMin + 6,
    temperature: temperatureMin,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
  },
  {
    id: '0x70ac08fffefafeca',
    friendlyName: 'Attic Light',
    availability: 'offline',
    method: 'mqtt',
    brightness: 60,
    color_temp: 370,
    state: 'ON',
    delayInMs: 15000,
  },
];

let updateDeviceMap = {};
updateDeviceMap['92fe86b7'] = mockUpdateWeatherNode1v2;
updateDeviceMap['0x00124b0029207763'] = mockUpdateTH01v2;
updateDeviceMap['0xa4c13894070052fc'] = mockUpdateHumanPresencev2;
updateDeviceMap['0x00124b00146c31cd'] = mockUpdateMotionSensorv2;
updateDeviceMap['0x70ac08fffefafeca'] = mockUpdateAtticLight;

function mockUpdateAtticLight(settings) {
  var device = {
    id: '0x70ac08fffefafeca',
    last_seen: currentTime(),
    data: {
      brightness: 61,
      color_temp: 370,
      state: 'ON',
    },
    properties: {},
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.properties.availability = availability;
  }
  return device;
}

function mockUpdateHumanPresencev2(settings) {
  var device = {
    id: '0xa4c13894070052fc',
    last_seen: currentTime(),
    data: {
      illuminance_lux: 9,
      presence: true,
    },
    properties: {},
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.properties.availability = availability;
  }
  return device;
}

function mockUpdateMotionSensorv2(settings) {
  var device = {
    id: '0x00124b00146c31cd',
    last_seen: currentTime(),
    data: {
      occupancy: true,
      temperature: getMockTemperature(settings),
    },
    properties: {},
  };
  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.properties.availability = availability;
  }

  return device;
}

function mockUpdateWeatherNode1v2(settings) {
  var device = {
    id: '92fe86b7',
    last_seen: currentTime(),
    data: {
      temperature: getMockTemperature(settings),
    },
    properties: {},
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.properties.availability = availability;
  }
  return device;
}

function mockUpdateTH01v2(settings) {
  var device = {
    id: '0x00124b0029207763',
    last_seen: currentTime(),
    data: {
      temperature: getMockTemperature(settings),
      humidity: getMockHumidity(settings),
    },
    properties: {},
  };

  var availability = setDeviceOnline(settings);
  if (availability != undefined) {
    device.properties.availability = availability;
  }
  return device;
}

function setDeviceOnline(settings) {
  if (settings.availability == 'offline') {
    settings.availability = 'online';
    return settings.availability;
  }

  return undefined;
}

function getMockTemperature(settings) {
  if (settings.availability == 'offline') {
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
  if (settings.availability == 'offline') {
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

function loadMetrics() {
  const t = loadTemperatureMetrics();
  const l = loadLightetrics();
  const p = loadPresenceMetrics();
  return {
    [t.deviceId]: t,
    [l.deviceId]: l,
    [p.deviceId]: p,
  };
}
function loadTemperatureMetrics() {
  const require = createRequire(import.meta.url);
  var data = require(temperatureMetricsFullPath);
  return data;
}

function loadPresenceMetrics() {
  const require = createRequire(import.meta.url);
  var data = require(presenceMetricsFullPath);
  return data;
}
function loadLightetrics() {
  const require = createRequire(import.meta.url);
  var data = require(lightMetricsFullPath);
  return data;
}
var permitJoinTimer = null;
function runPermitJoin(ws, bridgeConfig) {
  if (appConfig.bridge.permitJoin == bridgeConfig.permitJoin) {
    return;
  }

  if (bridgeConfig.permitJoin) {
    // send back response that we set permitjoin to true and started timer
    appConfig.bridge = bridgeConfig;

    const timeout = appConfig.bridge.maxTimeAllowed.value * 1000;
    sendMessage(ws, 'appConfig', appConfig);
    console.log('permitjoin is true for ', appConfig.bridge.maxTimeAllowed.value, ' seconds', timeout);

    permitJoinTimer = setTimeout(() => {
      // start timer for 10 seconds
      // then send message back with updated appConfig set permitjoin to false
      console.log('permitjoin is false, stopping timer');

      appConfig.bridge.permitJoin = false;
      sendMessage(ws, 'appConfig', appConfig);
    }, timeout);
  } else {
    setTimeout(() => {

      clearInterval(permitJoinTimer);
      console.log('permitjoin is false, manually stopped');
      // stop timer as we are currently running permitjoin
      appConfig.bridge.permitJoin = false;
      appConfig.bridge.maxTimeAllowed.value = 0;
      sendMessage(ws, 'appConfig', appConfig);
    }, 2000);
  }
}
