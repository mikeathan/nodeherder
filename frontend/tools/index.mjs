/* eslint-disable no-console */
/* eslint-disable indent */
import moment from 'moment';
import 'moment-timezone';

import express from 'express';
import expressWs from 'express-ws';
import http from 'http';
import { createRequire } from 'module';
import { link } from 'fs';
import cors from 'cors';

const hubStateFullPath = '../../docs/hub_state.json';
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
              type: 'expose',
              name: 'presence',
              value: false,
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [
                {
                  name: 'state',
                  data: 'OFF',
                },
              ],
              delay: {
                value: 5,
                unit: 'minutes',
              },
              type: 'trigger',
            },
          ],
        },
        {
          name: 'presence',
          conditions: [
            {
              type: 'expose',
              name: 'presence',
              value: true,
              equality: '=',
            },
            {
              type: 'expose',
              name: 'illuminance',
              value: 30,
              equality: '<=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [
                {
                  name: 'state',
                  data: 'ON',
                },
              ],
              type: 'trigger',
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
              type: 'expose',
              name: 'action',
              value: 'button_2_press',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'color_temp',
              data: null,
              delay: {
                value: 0,
                unit: 'seconds',
              },
              steps: [],
              type: 'preset',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              type: 'expose',
              name: 'action',
              value: 'button_1_press_release',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [
                {
                  name: 'state',
                  data: 'TOGGLE',
                },
              ],
              type: 'trigger',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              type: 'expose',
              name: 'action',
              value: 'dial_rotate_right_slow',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'brightness',
              data: 10,
              steps: [
                {
                  property: 'brightness',
                  operator: '-',
                  id: '0x00158d0005a23c38',
                },
                {
                  operator: '-',
                  property: 'action_time',
                  id: '0x001788010d7d9d3f',
                },
              ],
              type: 'step',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              type: 'expose',
              name: 'action',
              value: 'dial_rotate_left_slow',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'brightness',
              data: 10,
              steps: [
                {
                  property: 'brightness',
                  operator: '+',
                  id: '0x00158d0005a23c38',
                },
                {
                  operator: '+',
                  property: 'action_time',
                  id: '0x001788010d7d9d3f',
                },
              ],
              type: 'step',
            },
          ],
        },
        {
          name: 'action',
          conditions: [
            {
              type: 'expose',
              name: 'action',
              value: 'button_2_press_release',
              equality: '=',
            },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'color_temp',
              steps: [],
              type: 'preset',
            },
          ],
        },
      ],
    },
  ],
]);

expressWs(app, server);

var hubStatePayload = loadHubState();
var appConfig = hubStatePayload.config;
var metricsMap = loadMetrics();

// Allow CORS from frontend origin
app.use(
  cors({
    origin: 'http://localhost:4100',
  })
);


// Register HTTP GET route for /hubstate
app.get('/api/hubstate', (req, res) => {
  console.log('hubstate GET request');
  res.json(hubStatePayload);
});

var connected = false;

// Register web socket events
app.ws('/ws', async function (ws) {
  console.log('client connected');
  connected = true;

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

      // case 'loadHubState':
      //   sendMessage(ws, 'hubState', hubStatePayload);
      //   connected = true;
      //   break;
      case 'deviceSetValue':
        // Respond back with update value to update UI
        const updatePayload = {
          id: obj.payload.id,
          data: {
            [obj.payload.name]: obj.payload.value,
          },
          properties: {
            availability: 'online',
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
      case 'loadDashboardGroups':
        sendMessage(ws, 'dashboardGroups', appConfig.hub.dashboardGroups);
        break;

      case 'deleteDashboardGroup':
        const name = obj.payload.groupName;
        delete appConfig.hub.dashboardGroups[name];
        break;

      case 'saveDashboardGroup':
        const dashboardGroup = obj.payload;
        appConfig.hub.dashboardGroups[dashboardGroup.name] = dashboardGroup;
        break;

      case 'importDashboardGroups':
        appConfig.hub.dashboardGroups = {};
        Object.entries(obj.payload).forEach(([name, dashboardGroup]) => {
          appConfig.hub.dashboardGroups[name] = dashboardGroup;
        });
        sendMessage(ws, 'dashboardGroups', appConfig.hub.dashboardGroups);

      case 'saveDeviceConfigOverrides':
        {
          const deviceId = obj.payload.id;
          if (deviceId in appConfig.hub.devices.overrides == false) {
            appConfig.hub.devices.overrides[deviceId] = {};
          }
          appConfig.hub.devices.overrides[deviceId] = obj.payload;
          sendOperationSuccess(ws);
        }
        break;
      case 'deleteDeviceConfigOverrides':
        {
          delete appConfig.hub.devices.overrides[obj.payload.id];
          sendOperationSuccess(ws);
        }
        break;
      case 'saveDeviceConfigDefaults':
        appConfig.hub.devices.defaults = obj.payload;
        sendOperationSuccess(ws);
        break;

      case 'saveLoggerConfig':
        appConfig.hub.logger = obj.payload;

        if (appConfig.hub.logger.enableRemoteLogger) {
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
    luminance_offset: 12,
    delayInMs: 35000,
    luminance_: luminance_luxMin,
    luminance_LastChanged: moment(),
    presenceLastChanged: moment(),
  },
  {
    id: '0xa4c1389b273366c3',
    friendlyName: 'Attic alarm',
    availability: 'offline',
    method: 'mqtt',
    alarm: false,
    melody: 6,
    duration: 1,
    volume: 'high',
    delayInMs: 10500,
    presenceLastChanged: moment(),
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
    id: '0x00158d0005a23c38',
    friendlyName: 'Living room Light',
    availability: 'offline',
    method: 'mqtt',
    brightness: 60,
    color_temp: 370,
    state: 'ON',
    delayInMs: 15000,
  },
  {
    id: '0xa4c138c383ac3fc8',
    friendlyName: 'Smoke alarm',
    availability: 'offline',
    method: 'mqtt',
    smoke: false,
    device_fault: false,
    silence: false,
    delayInMs: 9000,
  },
];

let updateDeviceMap = {};
updateDeviceMap['0x00124b0029207763'] = mockUpdateTH01v2;
updateDeviceMap['0xa4c13894070052fc'] = mockUpdateHumanPresencev2;
//updateDeviceMap['0x70ac08fffefafeca'] = mockUpdateAtticLight;
updateDeviceMap['0x00158d0005a23c38'] = mockUpdateLivingRoomLight;
updateDeviceMap['0xa4c1389b273366c3'] = mockUpdateAtticAlarm;
updateDeviceMap['0xa4c138c383ac3fc8'] = mockSmokeAlarm;
function mockUpdateAtticAlarm(settings) {
  var device = {
    id: '0xa4c1389b273366c3',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      alarm: settings.alarm,
      melody: settings.melody,
      duration: settings.duration,
      volume: settings.volume,
      linkquality: 100,
    },
  };
  return device;
}
function mockUpdateLivingRoomLight(settings) {
  var device = {
    id: '0x00158d0005a23c38',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      brightness: 61,
      color_temp: 370,
      state: 'ON',
    },
  };

  return device;
}

function mockUpdateAtticLight(settings) {
  var device = {
    id: '0x70ac08fffefafeca',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      brightness: 61,
      color_temp: 370,
      state: 'ON',
    },
  };
  return device;
}

function mockSmokeAlarm(settings) {
  var device = {
    id: '0xa4c138c383ac3fc8',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      smoke: !settings.smoke,
      device_fault: !settings.device_fault,
      silence: !settings.silence,
    },
  };

  return device;
}

function mockUpdateHumanPresencev2(settings) {
  var device = {
    id: '0xa4c13894070052fc',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      illuminance: 9,
      presence: true,
    },
  };

  return device;
}

function mockUpdateTH01v2(settings) {
  var device = {
    id: '0x00124b0029207763',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      temperature: getMockTemperature(settings),
      humidity: getMockHumidity(settings),
      battery: 92,
    },
  };

  return device;
}

function setDeviceOnline(settings) {
  if (settings.availability == 'offline') {
    settings.availability = 'online';
    return settings.availability;
  }

  return settings.availability;
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

function loadHubState() {
  const require = createRequire(import.meta.url);
  var data = require(hubStateFullPath);
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
      sendMessage(ws, 'bridgeConfig', appConfig.bridge);
    }, timeout);
  } else {
    setTimeout(() => {
      clearInterval(permitJoinTimer);
      console.log('permitjoin is false, manually stopped');
      // stop timer as we are currently running permitjoin
      appConfig.bridge.permitJoin = false;
      appConfig.bridge.maxTimeAllowed.value = 0;
      sendMessage(ws, 'bridgeConfig', appConfig.bridge);
    }, 2000);
  }
}
