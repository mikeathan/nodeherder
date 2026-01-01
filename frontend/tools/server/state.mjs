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
    delayInMs: 7000, // Should show "now" most of the time
    luminance_: luminance_luxMin,
    luminance_LastChanged: moment(),
    presenceLastChanged: moment(),
    presence: false,
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
    delayInMs: 90000, // 1.5 minutes - rarely shows "now"
    presenceLastChanged: moment(),
  },
  {
    id: '0x00124b0029207763',
    friendlyName: 'TH01',
    availability: 'offline',
    method: 'mqtt',
    temperatureOffset: 0.6,
    humidityOffset: 11.3,
    delayInMs: 18000, // 18 seconds - moderate delay
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
    delayInMs: 8000, // Should show "now"
  },
  {
    id: '0xa4c138c383ac3fc8',
    friendlyName: 'Smoke alarm',
    availability: 'offline',
    method: 'mqtt',
    smoke: false,
    device_fault: false,
    silence: false,
    delayInMs: 120000, // 2 minutes - shows last seen time
  },
  {
    id: '0x70ac08fffefafeca',
    friendlyName: 'Attic Light',
    availability: 'offline',
    method: 'mqtt',
    brightness: 120,
    color_temp: 250,
    state: 'OFF',
    delayInMs: 35000, // 35 seconds - shows last seen time
  },
  {
    id: '0x001788010d7d9d3f',
    friendlyName: 'Living room switch dial',
    availability: 'offline',
    method: 'mqtt',
    battery: 95,
    brightness_dial: 128,
    lastActionTime: moment(),
    delayInMs: 55000, // 55 seconds - shows last seen time
  },
  {
    id: '0x00124b002fa5844e',
    friendlyName: 'Front door sensor',
    availability: 'offline',
    method: 'mqtt',
    contact: true, // closed
    battery: 87,
    battery_low: false,
    delayInMs: 6000, // Should show "now" - active door
  },
  {
    id: '0x70b3d52b60136661',
    friendlyName: 'Kitchen power socket',
    availability: 'offline',
    method: 'mqtt',
    state: 'ON',
    power: 0,
    current: 0,
    voltage: 230,
    energy: 125.5,
    delayInMs: 25000, // 25 seconds - moderate delay
  },
  {
    id: '0xa4c13801b36effff',
    friendlyName: 'TV power socket',
    availability: 'offline',
    method: 'mqtt',
    current: 0.35,
    voltage: 230,
    energy_today: 1.6,
    energy_yesterday: 2.8,
    energy_month: 42.5,
    delayInMs: 20000, // 20 seconds - moderate delay
  },
  {
    id: '0xa4c1381b6fd53fc4',
    friendlyName: 'Attic room power socket',
    availability: 'offline',
    method: 'mqtt',
    state: 'ON',
    power: 66,
    current: 0.36,
    voltage: 237,
    energy: 637.73,
    delayInMs: 22000, // 22 seconds - moderate delay
  },
  {
    id: '0xa4c1384582432edd',
    friendlyName: 'Garden temperature',
    availability: 'offline',
    method: 'mqtt',
    temperature: 18.5,
    humidity: 65.0,
    battery: 82,
    temperatureOffset: 0.4,
    humidityOffset: 0.7,
    temperatureLastChanged: moment(),
    humidityLastChanged: moment(),
    delayInMs: 45000, // 45 seconds - outdoor sensor, less frequent
  },
  {
    id: '0xa4c138e1b5658e68',
    friendlyName: 'Attic air sensor',
    availability: 'offline',
    method: 'mqtt',
    co2: 450,
    formaldehyd: 5,
    humidity: 55,
    temperature: 22,
    pm25: 12,
    voc: 120,
    delayInMs: 70000, // 70 seconds - air quality updates slowly
  },
];

const updateDeviceMap = {
  '0x00124b0029207763': mockUpdateTH01v2,
  '0xa4c13894070052fc': mockUpdateHumanPresencev2,
  '0x00158d0005a23c38': mockUpdateLivingRoomLight,
  '0xa4c1389b273366c3': mockUpdateAtticAlarm,
  '0xa4c138c383ac3fc8': mockSmokeAlarm,
  '0x70ac08fffefafeca': mockUpdateAtticLight,
  '0x001788010d7d9d3f': mockUpdateSwitchDial,
  '0x00124b002fa5844e': mockUpdateDoorSensor,
  '0x70b3d52b60136661': mockUpdateKitchenSocket,
  '0xa4c13801b36effff': mockUpdateTvSocket,
  '0xa4c1381b6fd53fc4': mockUpdateAtticSocket,
  '0xa4c1384582432edd': mockUpdateGardenTemp,
  '0xa4c138e1b5658e68': mockUpdateAirSensor,
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
  // Occasionally toggle alarm (5% chance)
  if (Math.random() < 0.05) {
    settings.alarm = !settings.alarm;
  }

  // Randomly vary linkquality (90-100)
  const linkquality = 90 + Math.floor(Math.random() * 11);

  return {
    id: '0xa4c1389b273366c3',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      alarm: settings.alarm,
      melody: settings.melody,
      duration: settings.duration,
      volume: settings.volume,
      linkquality,
    },
  };
}

function mockUpdateLivingRoomLight(settings) {
  // Randomly adjust brightness slightly (±5)
  let brightness = settings.brightness + (Math.random() > 0.5 ? 1 : -1) * Math.floor(Math.random() * 5);
  brightness = Math.max(1, Math.min(254, brightness));
  settings.brightness = brightness;

  // Occasionally change state (3% chance)
  if (Math.random() < 0.03) {
    settings.state = settings.state === 'ON' ? 'OFF' : 'ON';
  }

  // Randomly vary color_temp slightly (±10)
  let color_temp = settings.color_temp + (Math.random() > 0.5 ? 1 : -1) * Math.floor(Math.random() * 10);
  color_temp = Math.max(153, Math.min(500, color_temp));
  settings.color_temp = color_temp;

  // Randomly vary linkquality (85-100)
  const linkquality = 85 + Math.floor(Math.random() * 16);

  return {
    id: '0x00158d0005a23c38',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      brightness,
      color_temp,
      state: settings.state,
      linkquality,
    },
  };
}

function mockSmokeAlarm(settings) {
  // Very rarely trigger smoke alarm (1% chance)
  if (Math.random() < 0.01) {
    settings.smoke = !settings.smoke;
  }

  // Even more rarely trigger device fault (0.5% chance)
  if (Math.random() < 0.005) {
    settings.device_fault = !settings.device_fault;
  }

  // Randomly vary linkquality (88-100)
  const linkquality = 88 + Math.floor(Math.random() * 13);
  const battery = 85 + Math.floor(Math.random() * 16); // 85-100

  return {
    id: '0xa4c138c383ac3fc8',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      smoke: settings.smoke,
      device_fault: settings.device_fault,
      silence: settings.silence,
      battery,
      linkquality,
    },
  };
}

function mockUpdateHumanPresencev2(settings) {
  // Randomly change illuminance (0-150 lux with more realistic variation)
  const diff = moment().diff(settings.luminance_LastChanged);
  const duration = moment.duration(diff);

  if (duration.seconds() >= 3) {
    // Gradual illuminance changes
    const change = (Math.random() - 0.5) * 10; // ±5 lux
    settings.luminance_ = Math.max(0, Math.min(150, settings.luminance_ + change));
    settings.luminance_LastChanged = moment();
  }

  // Toggle presence more realistically (10% chance every update)
  const presenceDiff = moment().diff(settings.presenceLastChanged);
  const presenceDuration = moment.duration(presenceDiff);

  if (presenceDuration.seconds() >= 8 && Math.random() < 0.1) {
    settings.presence = !settings.presence;
    settings.presenceLastChanged = moment();
  }

  // Randomly vary linkquality (75-100)
  const linkquality = 75 + Math.floor(Math.random() * 26);

  return {
    id: '0xa4c13894070052fc',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      illuminance: Math.round(settings.luminance_),
      presence: settings.presence,
      linkquality,
    },
  };
}

function mockUpdateTH01v2(settings) {
  // Random battery fluctuation (80-100)
  const battery = 80 + Math.floor(Math.random() * 21);

  // Randomly vary linkquality (70-100)
  const linkquality = 70 + Math.floor(Math.random() * 31);

  return {
    id: '0x00124b0029207763',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      temperature: getMockTemperature(settings),
      humidity: getMockHumidity(settings),
      battery,
      linkquality,
    },
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

  // More realistic temperature variation with some randomness
  const randomChange = (Math.random() - 0.5) * 0.3; // ±0.15°C random variation
  const change = settings.temperatureOffset + randomChange;
  settings.temperature += change;

  // Keep temperature in realistic bounds
  if (settings.temperature > temperatureMax) {
    settings.temperature = temperatureMax - Math.random() * 2;
  } else if (settings.temperature < temperatureMin) {
    settings.temperature = temperatureMin + Math.random() * 2;
  }

  settings.temperatureLastChanged = moment();
  return parseFloat(settings.temperature.toFixed(1));
}

function getMockHumidity(settings) {
  if (settings.availability === 'offline') {
    return settings.humidity;
  }
  const diff = moment().diff(settings.humidityLastChanged);
  const duration = moment.duration(diff);
  if (duration.seconds() < humidityChangeDelaySec) return settings.humidity;

  // More realistic humidity variation with some randomness
  const randomChange = (Math.random() - 0.5) * 0.8; // ±0.4% random variation
  const change = settings.humidityOffset + randomChange;
  settings.humidity += change;

  // Keep humidity in realistic bounds
  if (settings.humidity > humidityMax) {
    settings.humidity = humidityMax - Math.random() * 3;
  } else if (settings.humidity < humidityMin) {
    settings.humidity = humidityMin + Math.random() * 3;
  }

  settings.humidityLastChanged = moment();
  return parseFloat(settings.humidity.toFixed(1));
}

function mockUpdateAtticLight(settings) {
  // Randomly adjust brightness slightly (±5)
  let brightness = settings.brightness + (Math.random() > 0.5 ? 1 : -1) * Math.floor(Math.random() * 5);
  brightness = Math.max(1, Math.min(254, brightness));
  settings.brightness = brightness;

  // Occasionally change state (4% chance)
  if (Math.random() < 0.04) {
    settings.state = settings.state === 'ON' ? 'OFF' : 'ON';
  }

  // Randomly vary color_temp slightly (±8)
  let color_temp = settings.color_temp + (Math.random() > 0.5 ? 1 : -1) * Math.floor(Math.random() * 8);
  color_temp = Math.max(153, Math.min(370, color_temp));
  settings.color_temp = color_temp;

  const linkquality = 82 + Math.floor(Math.random() * 19);

  return {
    id: '0x70ac08fffefafeca',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      brightness,
      color_temp,
      state: settings.state,
      linkquality,
    },
  };
}

function mockUpdateSwitchDial(settings) {
  // Battery slowly decreases
  if (Math.random() < 0.02) {
    settings.battery = Math.max(0, settings.battery - 1);
  }

  // Occasionally simulate dial rotation (8% chance)
  const actions = [
    'button_1_press_release',
    'dial_rotate_right_slow',
    'dial_rotate_left_slow',
    'button_2_press_release',
  ];

  let action = null;
  let action_time = null;
  let action_step_size = null;

  if (Math.random() < 0.08) {
    action = actions[Math.floor(Math.random() * actions.length)];
    action_time = Math.floor(Math.random() * 3) + 1;
    action_step_size = Math.floor(Math.random() * 20) + 5;
  }

  const linkquality = 78 + Math.floor(Math.random() * 23);

  const data = {
    battery: settings.battery,
    brightness: settings.brightness_dial,
    linkquality,
  };

  if (action) {
    data.action = action;
    data.action_time = action_time;
    data.action_step_size = action_step_size;
  }

  return {
    id: '0x001788010d7d9d3f',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data,
  };
}

function mockUpdateDoorSensor(settings) {
  // Toggle door contact state occasionally (12% chance - doors open/close frequently)
  if (Math.random() < 0.12) {
    settings.contact = !settings.contact;
  }

  // Battery slowly decreases
  if (Math.random() < 0.01) {
    settings.battery = Math.max(0, settings.battery - 1);
    settings.battery_low = settings.battery < 20;
  }

  const linkquality = 65 + Math.floor(Math.random() * 36);
  const voltage = 2800 + Math.floor(Math.random() * 300);

  return {
    id: '0x00124b002fa5844e',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      contact: settings.contact,
      battery: settings.battery,
      battery_low: settings.battery_low,
      voltage,
      linkquality,
    },
  };
}

function mockUpdateKitchenSocket(settings) {
  // Occasionally toggle state (2% chance)
  if (Math.random() < 0.02) {
    settings.state = settings.state === 'ON' ? 'OFF' : 'ON';
  }

  // Simulate realistic power consumption
  if (settings.state === 'ON') {
    // Random load between 0-150W
    settings.power = Math.floor(Math.random() * 150);
  } else {
    settings.power = 0;
  }

  // Calculate current from power (P = V * I)
  settings.current = settings.power > 0 ? parseFloat((settings.power / settings.voltage).toFixed(2)) : 0;

  // Energy accumulates slowly
  if (settings.state === 'ON' && settings.power > 0) {
    settings.energy += settings.power / 3600000; // Wh to kWh per update interval
  }

  // Voltage varies slightly
  settings.voltage = 228 + Math.floor(Math.random() * 7);

  const linkquality = 80 + Math.floor(Math.random() * 21);

  return {
    id: '0x70b3d52b60136661',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      state: settings.state,
      power: settings.power,
      current: settings.current,
      voltage: settings.voltage,
      energy: parseFloat(settings.energy.toFixed(2)),
      linkquality,
    },
  };
}

function mockUpdateTvSocket(settings) {
  // Simulate modest, steady TV load.
  const baseCurrent = 0.25;
  settings.current = parseFloat((baseCurrent + Math.random() * 0.25).toFixed(2));
  settings.voltage = 228 + Math.floor(Math.random() * 7);

  const power = settings.voltage * settings.current; // W
  settings.energy_today += power / 3600000;
  settings.energy_month += power / 3600000;

  const linkquality = 78 + Math.floor(Math.random() * 24);

  return {
    id: '0xa4c13801b36effff',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      current: settings.current,
      voltage: settings.voltage,
      energy_today: parseFloat(settings.energy_today.toFixed(2)),
      energy_yesterday: parseFloat(settings.energy_yesterday.toFixed(2)),
      energy_month: parseFloat(settings.energy_month.toFixed(2)),
      linkquality,
    },
  };
}

function mockUpdateAtticSocket(settings) {
  // More stable power consumption (fan or heater)
  const basePower = 66;
  settings.power = basePower + Math.floor(Math.random() * 10) - 5; // ±5W variation

  // Calculate current from power
  settings.current = parseFloat((settings.power / settings.voltage).toFixed(2));

  // Energy accumulates
  settings.energy += settings.power / 3600000;

  // Voltage varies slightly
  settings.voltage = 235 + Math.floor(Math.random() * 5);

  const linkquality = 75 + Math.floor(Math.random() * 26);

  return {
    id: '0xa4c1381b6fd53fc4',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      state: settings.state,
      power: settings.power,
      current: settings.current,
      voltage: settings.voltage,
      energy: parseFloat(settings.energy.toFixed(2)),
      linkquality,
    },
  };
}

function mockUpdateGardenTemp(settings) {
  // Use similar logic to TH01 but with outdoor-appropriate values
  const diff = moment().diff(settings.temperatureLastChanged);
  const duration = moment.duration(diff);

  if (duration.seconds() >= temperatureChangeDelaySec) {
    const randomChange = (Math.random() - 0.5) * 0.4;
    const change = settings.temperatureOffset + randomChange;
    settings.temperature += change;
    settings.temperature = Math.max(5, Math.min(35, settings.temperature));
    settings.temperatureLastChanged = moment();
  }

  const humidityDiff = moment().diff(settings.humidityLastChanged);
  const humidityDuration = moment.duration(humidityDiff);

  if (humidityDuration.seconds() >= humidityChangeDelaySec) {
    const randomChange = (Math.random() - 0.5) * 1.0;
    const change = settings.humidityOffset + randomChange;
    settings.humidity += change;
    settings.humidity = Math.max(30, Math.min(95, settings.humidity));
    settings.humidityLastChanged = moment();
  }

  // Battery slowly decreases
  if (Math.random() < 0.01) {
    settings.battery = Math.max(0, settings.battery - 1);
  }

  const linkquality = 55 + Math.floor(Math.random() * 46);
  const voltage = 2700 + Math.floor(Math.random() * 400);

  return {
    id: '0xa4c1384582432edd',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      temperature: parseFloat(settings.temperature.toFixed(1)),
      humidity: parseFloat(settings.humidity.toFixed(1)),
      battery: settings.battery,
      voltage,
      linkquality,
    },
  };
}

function mockUpdateAirSensor(settings) {
  // CO2 varies (400-1000 ppm)
  settings.co2 += Math.floor((Math.random() - 0.5) * 30);
  settings.co2 = Math.max(400, Math.min(1000, settings.co2));

  // Formaldehyd varies (0-15 µg/m³)
  settings.formaldehyd += (Math.random() - 0.5) * 2;
  settings.formaldehyd = Math.max(0, Math.min(15, settings.formaldehyd));

  // Humidity varies
  settings.humidity += (Math.random() - 0.5) * 3;
  settings.humidity = Math.max(30, Math.min(70, settings.humidity));

  // Temperature varies
  settings.temperature += (Math.random() - 0.5) * 0.5;
  settings.temperature = Math.max(18, Math.min(26, settings.temperature));

  // PM2.5 varies (0-50 µg/m³)
  settings.pm25 += Math.floor((Math.random() - 0.5) * 8);
  settings.pm25 = Math.max(0, Math.min(50, settings.pm25));

  // VOC varies (0-500 ppb)
  settings.voc += Math.floor((Math.random() - 0.5) * 40);
  settings.voc = Math.max(0, Math.min(500, settings.voc));

  const linkquality = 85 + Math.floor(Math.random() * 16);

  return {
    id: '0xa4c138e1b5658e68',
    last_seen: currentTime(),
    availability: setDeviceOnline(settings),
    data: {
      co2: settings.co2,
      formaldehyd: parseFloat(settings.formaldehyd.toFixed(1)),
      humidity: parseFloat(settings.humidity.toFixed(1)),
      temperature: parseFloat(settings.temperature.toFixed(1)),
      pm25: settings.pm25,
      voc: settings.voc,
      linkquality,
    },
  };
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
