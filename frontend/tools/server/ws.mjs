/* eslint-disable no-console */
/* eslint-disable indent */
import expressWs from 'express-ws';
import {
  settings,
  buildDeviceUpdatedPayload,
  automationMap,
  getAutomations,
  appConfig,
  runPermitJoin,
} from './state.mjs';
import { sendMessage, sendOperationSuccess, sendOperationFailed, currentTime } from './utils.mjs';
import { getAllMetricsForDevice, getMetricsForDevice } from './metrics-mock.mjs';

/** Correlated reply used by the real backend; the dev server acknowledges the same way. */
const DEVICE_CONFIG_SAVE_RESULT = 'deviceConfigSaveResult';

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
            const exposes = getAllMetricsForDevice(deviceId, from, to);

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
          appConfig.hub.history = obj.payload;
          sendOperationSuccess(ws);
          break;
        case 'saveAssistantConfig':
          appConfig.hub.assistant = obj.payload;
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
          const { requestId, ...config } = obj.payload;
          appConfig.hub.devices.overrides[deviceId] = config;
          if (requestId) {
            sendMessage(ws, DEVICE_CONFIG_SAVE_RESULT, { requestId, deviceId, success: true });
          } else {
            sendOperationSuccess(ws);
          }
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
        case 'loadMCPStatus': {
          sendMessage(ws, 'mcpStatus', {
            running: appConfig.hub.mcp.enabled,
            version: '1.0.0-mock',
            connectedClients: appConfig.hub.mcp.enabled ? 1 : 0,
            name: 'Mock MCP Server'
          });
          break;
        }
        case 'restartMCP': {
          appConfig.hub.mcp.enabled = true;
          sendMessage(ws, 'mcpStatus', {
            running: true,
            version: '1.0.0-mock',
            connectedClients: 1,
            name: 'Mock MCP Server'
          });
          break;
        }
        case 'stopMCP': {
          appConfig.hub.mcp.enabled = false;
          sendMessage(ws, 'mcpStatus', {
            running: false,
            version: '1.0.0-mock',
            connectedClients: 0,
            name: 'Mock MCP Server'
          });
          break;
        }
        case 'startMCP': {
          appConfig.hub.mcp.enabled = true;
          sendMessage(ws, 'mcpStatus', {
            running: true,
            version: '1.0.0-mock',
            connectedClients: 1,
            name: 'Mock MCP Server'
          });
          break;
        }
        case 'pong':
          break;
        case 'saveObservationIdentity':
        case 'deleteObservationIdentity': {
          // The real backend persists the decision *and rebroadcasts the snapshot*. The mock used
          // to only acknowledge, so a triaged device never left a "needs triage" filter and the
          // control looked inert in the browser — the test could not tell a real defect from a
          // faithful-mock gap. Apply the decision to the fixture and resend.
          applyObservationDecision(obj.payload);
          sendOperationSuccess(ws);
          if (connected) sendMessage(ws, 'observations', buildObservationSnapshot());
          break;
        }
        default:
          console.log('ws unhandled type:', msg);
          break;
      }
    });

    // Wireless observation snapshot. The real hub broadcasts this whenever aggregated state
    // changes; the mock sends one so the viewer can be rendered and asserted in a browser.
    // Rebuilt per tick rather than captured once, so a spec that turns observation off actually
    // receives the empty broadcast the real service publishes.
    sendMessage(ws, 'observations', buildObservationSnapshot());
    const observationIntervalId = setInterval(function () {
      if (!connected) return;
      sendMessage(ws, 'observations', buildObservationSnapshot());
    }, 2000);
    intervals.push(observationIntervalId);

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

/**
 * One deterministic wireless observation snapshot for browser tests.
 *
 * Fixed data, not generated: assertions need stable keys. It covers the cases the viewer must
 * handle — a stable beacon with an operator label, a rotating address, a known device, and a
 * source that is connected but has delivered nothing.
 */
/**
 * Decisions the browser has made during this session.
 *
 * Kept because the mock has no store: without it, a rebroadcast would undo the decision that
 * triggered it, and the viewer would snap back to the pre-triage state.
 */
const observationDecisions = new Map();

/**
 * Whether the mock reports observation as running.
 *
 * The radar page has to distinguish "the integration is off" from "the radio hears nothing", and a
 * browser test can only assert that if the fixture can produce both. Defaults to on, so nothing
 * changes for specs that do not care.
 */
let observationEnabled = true;

/** Turn the fixture's observation on or off, for specs that assert the disabled state. */
export function setObservationEnabled(enabled) {
  observationEnabled = enabled;
}

/** Current fixture state, so a reset can restore the default. */
export function observationIsEnabled() {
  return observationEnabled;
}

/**
 * Clear every triage decision recorded in this process.
 *
 * The mock is one long-lived server shared by every browser spec, so a decision made in one test
 * was still applied in the next: desktop triaged a device and mobile then started with it already
 * triaged, which looked like a flaky failure rather than leaked state. Tests reset through the
 * `__test/reset` route.
 */
export function resetObservationDecisions() {
  observationDecisions.clear();
  observationEnabled = true;
}

/** Apply a triage decision to the fixture, so a rebroadcast reflects it. */
function applyObservationDecision(message) {
  if (!message || typeof message.key !== 'string' || typeof message.decision !== 'string') return;
  observationDecisions.set(message.key, message.decision);
}

function buildObservationSnapshot() {
  const now = Date.now();
  const at = (offsetSeconds) => new Date(now - offsetSeconds * 1000).toISOString();
  // Observed-off is a genuinely empty broadcast, which is what the real service publishes when the
  // gate closes. Returning the populated fixture with a false flag would let a page render a stale
  // census beside a "not observing" notice and still pass.
  if (!observationEnabled) {
    return { at: at(0), truncated: false, observing: false, sources: [], identities: [], arrivals: 0 };
  }
  return {
    at: at(0),
    truncated: false,
    observing: true,
    sources: [
      {
        id: 'mock-proxy-1',
        name: 'Mock hall proxy',
        media: ['ble.adv'],
        facets: ['kind', 'vendor'],
        diagnostics: { advertisements: 128, decoded: 96, dropped: 0, lastAdvertisement: at(1) },
      },
      {
        id: 'mock-proxy-2',
        name: 'Mock landing proxy',
        media: ['ble.adv'],
        facets: ['kind'],
        diagnostics: { advertisements: 0, decoded: 0, dropped: 0, lastAdvertisement: null },
      },
    ],
    identities: [
      {
        key: 'ibeacon:2f234454-cf6d-4a0f-adf2-f4911ba9ffa6:1:2',
        kind: 'ibeacon',
        vendor: 76,
        stable: true,
        label: 'Front door tag',
        decision: 'known',
        firstSeen: at(30),
        lastSeen: at(5),
        packets: 1841,
        rssi: -62,
        rssiSmoothed: -66,
        trend: 'approaching',
        trendDeltaDb: 5.5,
        signal: [-74, -72, -70, -68, -66, -64, -62],
        sources: [{ id: 'mock-proxy-1', rssi: -62 }],
        // A settled, visible class with its reasoning, so the badge and evidence text can be
        // asserted. Personal rather than infrastructure because this is a tag the operator placed,
        // and it matters that a settled class is visible by default.
        classification: {
          class: 'personal',
          showByDefault: true,
          confidence: 0.78,
          hoursObserved: 96,
          daysConsidered: 4,
          reason: 'Present for part of the day with a changing signal, which suggests a device someone carries.',
          advisory: true,
        },
      },
      {
        key: 'addr:aa:bb:cc:dd:ee:01',
        kind: 'adv',
        stable: true,
        decision: 'unknown',
        firstSeen: at(720),
        lastSeen: at(8),
        packets: 12,
        rssi: -78,
        sources: [{ id: 'mock-proxy-1', rssi: -78 }],
        // Still learning: shows the progress line rather than a bare class, so the "learning is
        // work in progress, not a failure" treatment is exercised.
        classification: {
          class: 'learning',
          showByDefault: true,
          confidence: 0,
          hoursObserved: 6,
          reason: '18 more observed hours needed before a class is asserted.',
          advisory: true,
        },
      },
      {
        key: 'fp:9c02ab41',
        kind: 'adv',
        stable: false,
        decision: 'unknown',
        firstSeen: at(900),
        lastSeen: at(20),
        packets: 3,
        rssi: null,
        sources: [{ id: 'mock-proxy-1', rssi: null }],
        // A provisional pre-filter result: shown, but marked as a guess rather than a verdict.
        classification: {
          class: 'mobile-unidentifiable',
          provisional: true,
          showByDefault: true,
          confidence: 0,
          hoursObserved: 0,
          reason: 'Address rotates, so this is a person\'s device rather than household equipment.',
          advisory: true,
        },
      },
      // A busy household: 14 rotating addresses, alongside the identifiable devices below, which
      // matches the proportions of the reported install (about half rotating) rather than an
      // extreme case.
      ...Array.from({ length: 14 }, (_, index) => {
        // Spread across the full range rather than a tight cluster: the report was one identifiable
        // device near the antenna and a dozen rotating ones far away, which collapsed the scale.
        const rssi = -62 - index * 1.6;
        return {
          key: `fp:crowd${String(index).padStart(2, '0')}`,
          kind: 'adv',
          stable: false,
          decision: 'unknown',
          firstSeen: at(120 + index * 7),
          lastSeen: at(2 + index),
          packets: 3 + index,
          rssi,
          rssiSmoothed: rssi,
          trend: index % 3 === 0 ? 'approaching' : 'steady',
          trendDeltaDb: index % 3 === 0 ? 3.5 : 0,
          signal: [rssi - 4, rssi - 2, rssi],
          sources: [{ id: 'mock-proxy-1', rssi }],
          classification: {
            class: 'mobile-unidentifiable',
            provisional: true,
            showByDefault: true,
            confidence: 0,
            hoursObserved: 0,
            reason: 'Address rotates, so this is a person\'s device rather than household equipment.',
            advisory: true,
          },
        };
      }),
      // Weak but stable devices: far from the antenna, so the pre-filter classes them as probably
      // outside. These are the ones the default filter used to hide, leaving the radar almost empty.
      ...Array.from({ length: 8 }, (_, index) => {
        const rssi = -78 - index * 1.5;
        return {
          key: `addr:aa:bb:cc:dd:ef:${String(index).padStart(2, '0')}`,
          kind: 'adv',
          stable: true,
          decision: 'unknown',
          firstSeen: at(600 + index * 30),
          lastSeen: at(6 + index),
          packets: 40 + index * 5,
          rssi,
          rssiSmoothed: rssi,
          trend: 'steady',
          trendDeltaDb: 0.5,
          signal: [rssi - 2, rssi, rssi],
          sources: [{ id: 'mock-proxy-1', rssi }],
          classification: {
            class: 'likely-neighbour',
            provisional: true,
            showByDefault: true,
            confidence: 0,
            hoursObserved: 0,
            reason: 'Always much weaker than devices in the room, so probably outside it.',
            advisory: true,
          },
        };
      }),
      {
        key: 'addr:aa:bb:cc:dd:ee:02',
        kind: 'adv',
        stable: true,
        decision: 'unknown',
        firstSeen: at(3600),
        lastSeen: at(60),
        packets: 4210,
        rssi: -58,
        rssiSmoothed: -59,
        trend: 'steady',
        trendDeltaDb: 0.4,
        signal: [-59, -60, -59, -58, -59, -59, -58],
        sources: [{ id: 'mock-proxy-1', rssi: -58 }],
        // Demoted by default: this is the row class the operator complained about, and the footer
        // must be able to say it is there without showing it.
        classification: {
          class: 'infrastructure',
          showByDefault: false,
          confidence: 0.83,
          hoursObserved: 168,
          daysConsidered: 7,
          reason: 'Present almost all day, including overnight, and barely moved.',
          advisory: true,
        },
      },
    ].map((identity) => {
      const decided = observationDecisions.get(identity.key);
      return decided ? { ...identity, decision: decided } : identity;
    }),
    arrivals: 1,
  };
}
