/* eslint-disable no-console */
import moment from 'moment';
import 'moment-timezone';
import { createRequire } from 'module';
import { paths } from './config.mjs';
import { currentTime, sendMessage } from './utils.mjs';

// --- Constants for mock generation ---
export const temperatureChangeDelaySec = 5;
export const temperatureMin = 10.0;
export const temperatureMax = 40.0;

export const humidityChangeDelaySec = 30;
export const humidityMin = 30.0;
export const humidityMax = 100.0;

export const luminance_luxMin = 10;

// --- Mutable state ---
export let hubStatePayload = null; // full hub state payload from docs
export let appConfig = null; // hubStatePayload.config reference
export let metricsMap = {}; // deviceId -> metrics payload

// In-memory automations used by the mock server
export const automationMap = new Map([
  [
    '0xa4c13894070052fc',
    {
      id: '0xa4c13894070052fc',
      friendlyname: 'Human presence',
      description: 'Attic light test automation',
      enabled: true,
      schedules: [
        { startAt: '11:30', type: 'enable' },
        { startAt: '05:00', type: 'disable' },
      ],
      triggers: [
        {
          name: 'presence',
          conditions: [{ type: 'expose', name: 'presence', value: false, equality: '=' }],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [{ name: 'state', data: 'OFF' }],
              delay: { value: 5, unit: 'minutes' },
              type: 'trigger',
            },
          ],
        },
        {
          name: 'presence',
          conditions: [],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [{ name: 'state', data: 'OFF' }],
              delay: { value: 5, unit: 'minutes' },
              type: 'trigger',
            },
          ],
        },
        {
          name: 'presence',
          conditions: [
            { type: 'expose', name: 'presence', value: true, equality: '=' },
            { type: 'expose', name: 'illuminance', value: 30, equality: '<=' },
          ],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [{ name: 'state', data: 'ON' }],
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
          conditions: [{ type: 'expose', name: 'action', value: 'button_2_press', equality: '=' }],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'color_temp',
              data: null,
              delay: { value: 0, unit: 'seconds' },
              steps: [],
              type: 'preset',
            },
          ],
        },
        {
          name: 'action',
          conditions: [{ type: 'expose', name: 'action', value: 'button_1_press_release', equality: '=' }],
          actions: [
            {
              id: '0x00158d0005a23c38',
              exposes: [{ name: 'state', data: 'TOGGLE' }],
              type: 'trigger',
            },
          ],
        },
        {
          name: 'action',
          conditions: [{ type: 'expose', name: 'action', value: 'dial_rotate_right_slow', equality: '=' }],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'brightness',
              data: 10,
              steps: [
                { property: 'brightness', operator: '-', id: '0x00158d0005a23c38' },
                { operator: '-', property: 'action_time', id: '0x001788010d7d9d3f' },
              ],
              type: 'step',
            },
          ],
        },
        {
          name: 'action',
          conditions: [{ type: 'expose', name: 'action', value: 'dial_rotate_left_slow', equality: '=' }],
          actions: [
            {
              id: '0x00158d0005a23c38',
              property: 'brightness',
              data: 10,
              steps: [
                { property: 'brightness', operator: '+', id: '0x00158d0005a23c38' },
                { operator: '+', property: 'action_time', id: '0x001788010d7d9d3f' },
              ],
              type: 'step',
            },
          ],
        },
        {
          name: 'action',
          conditions: [{ type: 'expose', name: 'action', value: 'button_2_press_release', equality: '=' }],
          actions: [{ id: '0x00158d0005a23c38', property: 'color_temp', steps: [], type: 'preset' }],
        },
      ],
    },
  ],
]);

// --- Device update simulation ---
export const settings = [
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

const updateDeviceMap = {
  '0x00124b0029207763': mockUpdateTH01v2,
  '0xa4c13894070052fc': mockUpdateHumanPresencev2,
  '0x00158d0005a23c38': mockUpdateLivingRoomLight,
  '0xa4c1389b273366c3': mockUpdateAtticAlarm,
  '0xa4c138c383ac3fc8': mockSmokeAlarm,
};

export function buildDeviceUpdatedPayload(s) {
  const func = updateDeviceMap[s.id];
  try {
    return func ? func(s) : undefined;
  } catch (error) {
    console.log('id:', s.id + ' error:' + error);
    return undefined;
  }
}

function mockUpdateAtticAlarm(settings) {
  return {
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
}

function mockUpdateLivingRoomLight(settings) {
  return {
    id: '0x00158d0005a23c38',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: { brightness: 61, color_temp: 370, state: 'ON' },
  };
}

function mockSmokeAlarm(settings) {
  return {
    id: '0xa4c138c383ac3fc8',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      smoke: !settings.smoke,
      device_fault: !settings.device_fault,
      silence: !settings.silence,
    },
  };
}

function mockUpdateHumanPresencev2(settings) {
  return {
    id: '0xa4c13894070052fc',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: { illuminance: 9, presence: true },
  };
}

function mockUpdateTH01v2(settings) {
  return {
    id: '0x00124b0029207763',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: { temperature: getMockTemperature(settings), humidity: getMockHumidity(settings), battery: 92 },
  };
}

function setDeviceOnline(settings) {
  if (settings.availability === 'offline') {
    settings.availability = 'online';
  }
  return settings.availability;
}

function getMockTemperature(settings) {
  if (settings.availability === 'offline') {
    return settings.temperature;
  }
  const diff = moment().diff(settings.temperatureLastChanged);
  const duration = moment.duration(diff);
  if (duration.seconds() < temperatureChangeDelaySec) return settings.temperature;
  if (settings.temperature > temperatureMax) settings.temperature = temperatureMin;
  settings.temperature += settings.temperatureOffset;
  settings.temperatureLastChanged = moment();
  return settings.temperature;
}

function getMockHumidity(settings) {
  if (settings.availability === 'offline') {
    return settings.humidity;
  }
  const diff = moment().diff(settings.humidityLastChanged);
  const duration = moment.duration(diff);
  if (duration.seconds() < humidityChangeDelaySec) return settings.humidity;
  if (settings.humidity > humidityMax) settings.humidity = humidityMin;
  settings.humidity += settings.humidityOffset;
  settings.humidityLastChanged = moment();
  return settings.humidity;
}

// --- Loaders ---
export function loadHubState() {
  const require = createRequire(import.meta.url);
  const data = require(paths.hubState);
  return data.payload;
}

export function loadMetrics() {
  return {
    [loadTemperatureMetrics().deviceId]: loadTemperatureMetrics(),
    [loadLightMetrics().deviceId]: loadLightMetrics(),
    [loadPresenceMetrics().deviceId]: loadPresenceMetrics(),
  };
}

function loadTemperatureMetrics() {
  const require = createRequire(import.meta.url);
  return require(paths.metrics.temperature);
}

function loadPresenceMetrics() {
  const require = createRequire(import.meta.url);
  return require(paths.metrics.presence);
}

function loadLightMetrics() {
  const require = createRequire(import.meta.url);
  return require(paths.metrics.light);
}

// Initialize in-memory state
export function initState() {
  hubStatePayload = loadHubState();
  if (!hubStatePayload) throw new Error('Failed to load hub state');
  hubStatePayload.devices.sort((a, b) => a.friendly_name.localeCompare(b.friendly_name));
  appConfig = hubStatePayload.config;
  metricsMap = loadMetrics();
}

export function getAutomations() {
  const items = [];
  automationMap.forEach((v) => items.push(v));
  return items;
}

// Bridge permit join mock impl
let permitJoinTimer = null;
export function runPermitJoin(ws, bridgeConfig) {
  if (appConfig.bridge.permitJoin === bridgeConfig.permitJoin) return;
  if (bridgeConfig.permitJoin) {
    appConfig.bridge = bridgeConfig;
    const timeout = appConfig.bridge.maxTimeAllowed.value * 1000;
    sendMessage(ws, 'appConfig', appConfig);
    console.log('permitjoin is true for', appConfig.bridge.maxTimeAllowed.value, 'seconds', timeout);
    permitJoinTimer = setTimeout(() => {
      console.log('permitjoin is false, stopping timer');
      appConfig.bridge.permitJoin = false;
      sendMessage(ws, 'bridgeConfig', appConfig.bridge);
    }, timeout);
  } else {
    setTimeout(() => {
      clearTimeout(permitJoinTimer);
      console.log('permitjoin is false, manually stopped');
      appConfig.bridge.permitJoin = false;
      appConfig.bridge.maxTimeAllowed.value = 0;
      sendMessage(ws, 'bridgeConfig', appConfig.bridge);
    }, 2000);
  }
}
