/* eslint-disable no-console */
/* eslint-disable indent */
import expressWs from 'express-ws';
import {
  settings,
  buildDeviceUpdatedPayload,
  automationMap,
  getAutomations,
  appConfig,
  metricsMap,
  runPermitJoin,
} from './state.mjs';
import { sendMessage, sendOperationSuccess, sendOperationFailed, currentTime } from './utils.mjs';
import { getMetricsForDevice } from './metrics-mock.mjs';

export function registerWebsocket(app, server) {
  expressWs(app, server);

  app.ws('/ws', function (ws) {
    console.log('client connected');
    let connected = true;
    const intervals = [];
    let consoleLogIntervalId = 0;
    const logSeverity = ['info', 'warning', 'error', 'critical'];

    // periodic device updates per connected client
    settings.forEach((s) => {
      const id = setInterval(function () {
        if (!connected) return;
        const updatePayload = buildDeviceUpdatedPayload(s);
        if (!updatePayload) return;
        sendMessage(ws, 'deviceUpdated', updatePayload);
      }, s.delayInMs);
      intervals.push(id);
    });

    ws.on('message', async function (msg) {
      console.log('message received', msg);
      const obj = safeParse(msg);
      if (!obj) return;
      switch (obj.type) {
        case 'loadAutomations':
          sendMessage(ws, 'automations', getAutomations());
          break;
        case 'deviceSetValue': {
          const updatePayload = {
            id: obj.payload.id,
            data: { [obj.payload.name]: obj.payload.value },
            properties: { availability: 'online', last_seen: currentTime() },
          };
          sendMessage(ws, 'deviceUpdated', updatePayload);
          break;
        }
        case 'saveAutomation': {
          const automation = obj.payload;
          automationMap.set(automation.id, automation);
          sendOperationSuccess(ws);
          break;
        }
        case 'deleteAutomation': {
          if (!automationMap.has(obj.payload.id)) {
            sendOperationFailed(ws, 'Delete failed. Automation id ' + obj.payload.id + ' not found');
            return;
          }
          automationMap.delete(obj.payload.id);
          sendMessage(ws, 'automations', getAutomations());
          break;
        }
        case 'deleteAutomationTrigger': {
          const aId = obj.payload.automationId;
          const tId = obj.payload.triggerId;
          if (!automationMap.has(aId)) {
            sendOperationFailed(ws, 'Delete trigger. automation id ' + aId + ' not found');
            return;
          }
          const automation = automationMap.get(aId);
          if (tId >= automation.triggers.length) {
            sendOperationFailed(ws, 'trigger index ' + tId + ' out of bounds.');
            return;
          }
          automation.triggers.splice(tId, 1);
          sendMessage(ws, 'automationUpdated', automation);
          break;
        }
        case 'loadAppConfig':
          sendMessage(ws, 'appConfig', appConfig);
          break;
        case 'bridgePermitJoin':
          setTimeout(() => {
            if (appConfig.bridge.permitJoin !== obj.payload.permitJoin) {
              runPermitJoin(ws, obj.payload);
            } else {
              console.log('permitjoin already set to', obj.payload.permitJoin);
              sendOperationFailed(ws, 'Permit join is already set to ' + obj.payload.permitJoin);
            }
          }, 2000);
          break;
        case 'loadMetrics': {
          console.log('loadMetrics request:', obj.payload);
          const request = obj.payload;
          const deviceId = request.id;
          const expose = request.expose;
          const from = request.from * 1000; // Convert seconds to milliseconds
          const to = request.to * 1000;

          // If expose is specified, return only that expose
          if (expose) {
            const metricsData = getMetricsForDevice(deviceId, expose, from, to);
            if (!metricsData) {
              console.log('Metrics for device id', deviceId, 'expose', expose, 'not found');
              return;
            }
            sendMessage(ws, 'metrics', {
              deviceId: deviceId,
              exposes: [metricsData],
            });
          } else {
            // Return all exposes for the device
            const illuminanceData = getMetricsForDevice(deviceId, 'illuminance', from, to);
            const presenceData = getMetricsForDevice(deviceId, 'presence', from, to);

            const exposes = [];
            if (illuminanceData) exposes.push(illuminanceData);
            if (presenceData) exposes.push(presenceData);

            if (exposes.length === 0) {
              console.log('No metrics for device id', deviceId);
              return;
            }

            sendMessage(ws, 'metrics', {
              deviceId: deviceId,
              exposes: exposes,
            });
          }
          break;
        }
        case 'saveHistoryConfig':
          appConfig.history = obj.payload;
          sendOperationSuccess(ws);
          break;
        case 'loadDashboardGroups':
          sendMessage(ws, 'dashboardGroups', appConfig.hub.dashboardGroups);
          break;
        case 'deleteDashboardGroup': {
          const name = obj.payload.groupName;
          delete appConfig.hub.dashboardGroups[name];
          break;
        }
        case 'saveDashboardGroup': {
          const dashboardGroup = obj.payload;
          appConfig.hub.dashboardGroups[dashboardGroup.name] = dashboardGroup;
          break;
        }
        case 'importDashboardGroups': {
          appConfig.hub.dashboardGroups = {};
          Object.entries(obj.payload).forEach(([name, dashboardGroup]) => {
            appConfig.hub.dashboardGroups[name] = dashboardGroup;
          });
          sendMessage(ws, 'dashboardGroups', appConfig.hub.dashboardGroups);
          break;
        }
        case 'saveDeviceConfigOverride': {
          const deviceId = obj.payload.id;
          if (!(deviceId in appConfig.hub.devices.overrides)) {
            appConfig.hub.devices.overrides[deviceId] = {};
          }
          appConfig.hub.devices.overrides[deviceId] = obj.payload;
          sendOperationSuccess(ws);
          break;
        }
        case 'deleteDeviceConfigOverride':
          delete appConfig.hub.devices.overrides[obj.payload.id];
          sendOperationSuccess(ws);
          break;
        case 'saveDeviceConfigDefaults':
          appConfig.hub.devices.defaults = obj.payload;
          sendOperationSuccess(ws);
          break;
        case 'saveLoggerConfig': {
          appConfig.hub.logger = obj.payload;
          if (appConfig.hub.logger.enableRemoteLogger) {
            if (consoleLogIntervalId) {
              console.log('consoleLogIntervalId already running');
              clearInterval(consoleLogIntervalId);
            }
            console.log('enableRemoteLogger');
            consoleLogIntervalId = setInterval(() => {
              const severity = logSeverity[Math.floor(Math.random() * logSeverity.length)];
              const msg = { level: severity, message: severity + ' message', timestamp: Date.now() };
              sendMessage(ws, 'logger', msg);
            }, 1000);
          } else {
            console.log('disableRemoteLogger');
            clearInterval(consoleLogIntervalId);
            consoleLogIntervalId = 0;
          }
          break;
        }
        case 'pong':
          break;
        default:
          console.log('ws unhandled type:', msg);
          break;
      }
    });

    ws.on('error', function (error) {
      console.log('Cannot start server', error);
    });

    ws.on('close', function (code, message) {
      console.log('Disconnection:', code, message);
      connected = false;
      intervals.forEach((id) => clearInterval(id));
      if (consoleLogIntervalId) clearInterval(consoleLogIntervalId);
    });
  });
}

function safeParse(message) {
  try {
    return JSON.parse(message);
  } catch (e) {
    console.log('Invalid JSON message', e);
    return null;
  }
}
