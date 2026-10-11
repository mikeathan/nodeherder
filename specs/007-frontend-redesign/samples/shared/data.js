/* Mock household mirroring the real browser contracts (frontend/src/types/device.d.ts,
   settings.type.ts, automation.type.ts). Values taken from the review screenshots. */
(function () {
  const now = Date.now();
  const ago = (s) => new Date(now - s * 1000).toISOString();

  // Expose factory: type binary|numeric|enum|'' ; access read|write|readwrite ; category
  function ex(name, data, o) {
    o = o || {};
    return {
      name: name,
      description: o.description || '',
      unit: o.unit || '',
      data: data,
      type: o.type || (typeof data === 'boolean' ? 'binary' : typeof data === 'number' ? 'numeric' : 'enum'),
      category: o.category || 'measurement',
      access_mode: o.access || 'read',
      attributes: o.attrs || null,
      values: o.values || null,
    };
  }
  const lqi = (v) => ex('linkquality', v, { unit: 'lqi', category: 'diagnostic' });
  const batt = (v, n) => ex(n || 'battery', v, { unit: '%', category: 'diagnostic' });
  const onoff = (v) => ex('state', v, { type: 'binary', access: 'readwrite', values: { on: 'ON', off: 'OFF' } });
  const rssi = (v) => ex('rssi', v, { unit: 'dBm', category: 'diagnostic' });

  function dev(id, name, description, conn, power, lastSeenS, availability, exposes) {
    const m = {};
    exposes.forEach((e) => (m[e.name] = e));
    return {
      id, friendly_name: name, description, connection_type: conn, power_source: power,
      last_seen: ago(lastSeenS), availability, exposes: m, properties: {},
    };
  }

  const devices = [
    dev('0x54ef44100045a1b2', 'Attic air sensor', 'Smart air house keeper', 'mqtt', 'mains', 25, 'online', [
      ex('co2', 354, { unit: 'ppm' }), ex('formaldehyd', 1, { unit: 'µg/m³' }), ex('humidity', 84.9, { unit: '%' }),
      ex('pm25', 7, { unit: 'µg/m³' }), ex('temperature', 20.8, { unit: '°C' }), ex('voc', 6, { unit: 'ppb' }), lqi(91)]),
    dev('0xa4c138d0a7e4e9f1', 'Attic alarm', 'Alarm', 'mqtt', 'mains', 60, 'online', [
      ex('alarm', false, { type: 'binary', access: 'readwrite', values: { on: true, off: false } }), batt(100, 'battpercentage'), lqi(80)]),
    dev('0x001788010c5f3e21', 'Attic room Light', 'Automatically generated definition', 'mqtt', 'mains', 60, 'online', [
      onoff('ON'),
      ex('brightness', 254, { access: 'readwrite', attrs: { value_min: 0, value_max: 254 } }),
      ex('color_temp', 370, { unit: 'mired', access: 'readwrite', attrs: { value_min: 153, value_max: 500 } }), lqi(94)]),
    dev('0xa4c1386b2f0c11aa', 'Attic room power socket', 'Smart plug (UK)', 'mqtt', 'mains', 20, 'online', [
      ex('current', 0.5, { unit: 'A' }), ex('energy', 293.9, { unit: 'kWh' }), ex('power', 89, { unit: 'W' }), onoff('ON'), lqi(36)]),
    dev('0x00158d0008a1b2c3', 'Attic room temperature sensor', 'Temperature and humidity sensor', 'mqtt', 'battery', 180, 'online', [
      ex('humidity', 84.5, { unit: '%' }), ex('temperature', 19.8, { unit: '°C' }), batt(99.5), lqi(14)]),
    dev('0x00158d0005c4d1e2', 'Attic smoke alarm', 'Photoelectric smoke detector', 'mqtt', 'battery', 10800, 'online', [
      ex('device_fault', false, { type: 'binary' }), ex('smoke', false, { type: 'binary' }),
      ex('smoke_concentration', 0, { unit: 'ppm' }), ex('test', false, { type: 'binary' }), batt(100), lqi(40)]),
    dev('0x00158d0007f3a9b4', 'Front door sensor', 'Contact sensor', 'mqtt', 'battery', 1800, 'online', [
      ex('contact', true, { type: 'binary', values: { on: true, off: false } }), batt(100), lqi(116)]),
    dev('0x00158d000611e5f5', 'Garden temperature', 'Temperature and humidity sensor - with or without display', 'mqtt', 'battery', 660, 'online', [
      ex('humidity', 76, { unit: '%' }), ex('temperature', 12.2, { unit: '°C' }), batt(88), lqi(72)]),
    dev('0xa4c1384d9e7b2c06', 'kitchen power socket', 'Smart Plug', 'mqtt', 'mains', 55, 'online', [
      ex('energy', 280.3, { unit: 'kWh' }), ex('power', 12, { unit: 'W' }), onoff('ON'), lqi(109)]),
    dev('0x0017880109ab4c07', 'Living Room Light', 'Hue white ambiance E27 1100lm with Bluetooth', 'mqtt', 'mains', 840, 'online', [
      onoff('ON'),
      ex('brightness', 224, { access: 'readwrite', attrs: { value_min: 0, value_max: 254 } }),
      ex('color_temp', 300, { unit: 'mired', access: 'readwrite', attrs: { value_min: 153, value_max: 454 },
        values: { coolest: 153, cool: 250, neutral: 370, warm: 454 } }), lqi(76)]),
    dev('0xa4c138f1c2d3e408', 'Living room presence sensor', 'Mini human breathe sensor', 'mqtt', 'mains', 2, 'online', [
      ex('presence', true, { type: 'binary' }), ex('illuminance', 255, { unit: 'lx' }),
      ex('target_distance', 1.93, { unit: 'm' }), lqi(105)]),
    dev('0x001788010d1e2f09', 'Living room switch dial', 'Hue Tap dial switch', 'mqtt', 'battery', 2100, 'online', [
      ex('action', 'brightness_step_up', { type: 'enum', values: { button_1_press: 'button_1_press', button_2_press: 'button_2_press',
        brightness_step_up: 'brightness_step_up', brightness_step_down: 'brightness_step_down' } }), batt(100), lqi(65)]),
    dev('0xa4c1382a3b4c5d10', 'TV power socket', 'Zigbee smart plug', 'mqtt', 'mains', 1500, 'online', [
      ex('current', 0.2, { unit: 'A' }), ex('energy', 0.3, { unit: 'kWh', description: 'Energy today' }), ex('power', 45, { unit: 'W' }), onoff('ON'), lqi(109)]),
    dev('esp32-garage-weather', 'Garage weather station', 'ESP32 + BME280 (HTTP push)', 'http', 'mains', 40, 'online', [
      ex('temperature', 9.4, { unit: '°C' }), ex('humidity', 81, { unit: '%' }), ex('pressure', 1012, { unit: 'hPa' }), rssi(-61),
      ex('ip', '192.168.1.48', { type: '', category: 'diagnostic' })]),
    dev('shelly-plug-hallway', 'Hallway Shelly plug', 'Shelly Plus Plug S (HTTP script)', 'http', 'mains', 15, 'online', [
      onoff('OFF'), ex('power', 0, { unit: 'W' }), ex('energy', 12.4, { unit: 'kWh' }), rssi(-55),
      ex('ip', '192.168.1.52', { type: '', category: 'diagnostic' })]),
    dev('0x00158d000aa0b1c2', 'Bedroom radiator valve', 'Thermostatic radiator valve', 'mqtt', 'battery', 7200, 'offline', [
      ex('local_temperature', 18.5, { unit: '°C' }),
      ex('current_heating_setpoint', 20, { unit: '°C', access: 'readwrite', attrs: { value_min: 5, value_max: 30 } }), batt(23), lqi(8)]),
    dev('0x00158d0001e2f3a4', 'Hallway motion (spare)', 'Motion sensor', 'mqtt', 'battery', 86400, 'online', [
      ex('occupancy', false, { type: 'binary' }), batt(15), lqi(52)]),
  ];
  const byName = {};
  devices.forEach((d) => (byName[d.friendly_name] = d.id));
  const D = (n) => byName[n];

  const deviceDefaults = {
    id: 'defaults', disabled: false, metricsEnabled: true,
    defaultDebounceByCategory: { measurement: { value: 2, unit: 'seconds' }, diagnostic: { value: 5, unit: 'minutes' }, config: { value: 0, unit: 'seconds' } },
    debounceOverrides: {},
  };
  const overrides = {};
  overrides[D('Hallway motion (spare)')] = Object.assign({}, deviceDefaults, { id: D('Hallway motion (spare)'), disabled: true });
  overrides[D('Living room presence sensor')] = Object.assign({}, deviceDefaults, {
    id: D('Living room presence sensor'), debounceOverrides: { target_distance: { value: 10, unit: 'seconds' }, linkquality: { value: 10, unit: 'minutes' } } });

  function grp(name, entries) {
    const g = { name, deviceGroup: {} };
    entries.forEach(([dn, exn]) => {
      const id = D(dn);
      g.deviceGroup[id] = g.deviceGroup[id] || { deviceId: id, exposes: [] };
      g.deviceGroup[id].exposes.push(exn);
    });
    return g;
  }
  const dashboardGroups = {
    'Attic room': grp('Attic room', [['Attic room temperature sensor', 'temperature'], ['Attic room power socket', 'state'],
      ['Attic room temperature sensor', 'humidity'], ['Attic alarm', 'alarm'], ['Attic room Light', 'brightness'], ['Attic smoke alarm', 'smoke'],
      ['Attic air sensor', 'co2']]),
    'Garden': grp('Garden', [['Garden temperature', 'humidity'], ['Garden temperature', 'temperature']]),
    'Kitchen': grp('Kitchen', [['Front door sensor', 'contact'], ['kitchen power socket', 'energy'], ['kitchen power socket', 'state']]),
    'Living room': grp('Living room', [['Living Room Light', 'brightness'], ['Living room presence sensor', 'presence'],
      ['TV power socket', 'energy'], ['Living room presence sensor', 'illuminance'], ['Living Room Light', 'state'], ['TV power socket', 'power']]),
    'Garage': grp('Garage', [['Garage weather station', 'temperature'], ['Garage weather station', 'pressure'], ['Hallway Shelly plug', 'state']]),
    'Bedroom': grp('Bedroom', [['Bedroom radiator valve', 'current_heating_setpoint'], ['Bedroom radiator valve', 'local_temperature'], ['Hallway motion (spare)', 'occupancy']]),
  };

  const automations = [
    {
      id: D('Front door sensor'), friendlyname: 'Front door sensor', type: 'device', description: 'Front door triggers alarm', enabled: true, schedules: [],
      triggers: [{
        name: 'contact', type: 'deviceTrigger',
        conditions: [{ type: 'expose', name: 'contact', equality: '=', value: false }, { type: 'time', timeRange: { startAt: '23:00', endAt: '06:30' } }],
        actions: [{ id: D('Attic alarm'), type: 'trigger', exposes: [{ name: 'alarm', data: true }], publishMode: 'batch', delay: { unit: 'seconds', value: 0 } }],
      }],
    },
    {
      id: D('Living room switch dial'), friendlyname: 'Living room switch dial', type: 'device', description: 'Light switch automation', enabled: true, schedules: [],
      triggers: [
        { name: 'action', type: 'deviceTrigger', conditions: [{ type: 'expose', name: 'action', equality: '=', value: 'brightness_step_up' }],
          actions: [{ id: D('Living Room Light'), type: 'step', property: 'brightness', data: 25, steps: [{ id: D('Living Room Light'), property: 'brightness', operator: '+' }] }] },
        { name: 'action', type: 'deviceTrigger', conditions: [{ type: 'expose', name: 'action', equality: '=', value: 'brightness_step_down' }],
          actions: [{ id: D('Living Room Light'), type: 'step', property: 'brightness', data: 25, steps: [{ id: D('Living Room Light'), property: 'brightness', operator: '-' }] }] },
        { name: 'action', type: 'deviceTrigger', conditions: [{ type: 'expose', name: 'action', equality: '=', value: 'button_2_press' }],
          actions: [{ id: D('Living Room Light'), type: 'preset', property: 'color_temp' }] },
      ],
    },
    {
      id: D('Attic room Light'), friendlyname: 'Attic room Light', type: 'device', description: 'test', enabled: true,
      schedules: [{ startAt: '07:00', type: 'enable' }, { startAt: '23:30', type: 'disable' }],
      triggers: [{ name: 'state', type: 'manualTrigger', conditions: [{ type: 'time', timeRange: { startAt: '18:00', endAt: '23:00' } }],
        actions: [{ id: D('Attic room Light'), type: 'trigger', exposes: [{ name: 'state', data: 'ON' }, { name: 'brightness', data: 200 }], publishMode: 'batch' }] }],
    },
    {
      id: D('Living room presence sensor'), friendlyname: 'Living room presence sensor', type: 'device', description: 'Attic light test automation', enabled: true, schedules: [],
      triggers: [
        { name: 'presence', type: 'deviceTrigger', conditions: [{ type: 'expose', name: 'presence', equality: '=', value: true }, { type: 'expose', name: 'illuminance', equality: '<', value: 120 }],
          actions: [{ id: D('Attic room Light'), type: 'trigger', exposes: [{ name: 'state', data: 'ON' }], publishMode: 'batch' }] },
        { name: 'presence', type: 'deviceTrigger', conditions: [{ type: 'expose', name: 'presence', equality: '=', value: false }],
          actions: [{ id: D('Attic room Light'), type: 'trigger', exposes: [{ name: 'state', data: 'OFF' }], publishMode: 'batch', delay: { unit: 'minutes', value: 2 } }] },
      ],
    },
  ];

  const logs = [
    ['info', 'hub', 'NodeHerder v0.12.16-dev started'],
    ['info', 'mqtt', 'connected to mqtt://192.168.1.10:1883'],
    ['debug', 'bridge', 'bridge/devices received (13 devices)'],
    ['info', 'mcp', 'MCP server listening on /api/mcp'],
    ['info', 'automation', 'loaded 4 automations'],
    ['warn', 'device', 'Bedroom radiator valve availability timeout (offline)'],
    ['debug', 'lanes', 'lane 0xa4c138f1c2d3e408 drained 3 messages'],
    ['info', 'http', 'POST /api/collect esp32-garage-weather 200'],
    ['info', 'automation', 'Living room switch dial › action=brightness_step_up › step brightness +25'],
    ['error', 'bridge', 'deviceInterview 0x00158d000aa0b1c2 failed: device did not respond'],
    ['debug', 'metrics', 'sampled 42 values, 3 suppressed by debounce'],
  ].map((l, i) => ({ level: l[0], source: l[1], message: l[2], timestamp: now - (11 - i) * 47000 }));

  const conversations = [
    { id: 'c1', title: 'Attic humidity trend', updated: now - 3600e3, messages: [
      { role: 'user', content: 'Why is the attic humidity so high this week?' },
      { role: 'assistant', content: 'Attic room temperature sensor averaged 83.9 % RH over the last 7 days (min 78.2 %, max 88.1 %). The Attic air sensor agrees within 1 %. Humidity rises after 18:00 when the garden temperature drops below 13 °C, which suggests condensation from poor ventilation rather than a sensor fault.' },
    ] },
    { id: 'c2', title: 'Energy yesterday', updated: now - 86400e3, messages: [] },
    { id: 'c3', title: 'Which devices have weak signal?', updated: now - 3 * 86400e3, messages: [] },
  ];

  const appConfig = {
    hub: {
      devices: { defaults: deviceDefaults, overrides },
      history: { sleepTimeout: { value: 15, unit: 'minutes' }, expireAt: { value: 90, unit: 'days' } },
      logger: { enableRemoteLogger: true, level: 'info' },
      mcp: { enabled: true },
      assistant: { url: 'http://192.168.1.20:8080' },
      dashboardGroups,
    },
    bridge: { permitJoin: false, maxTimeAllowed: { value: 120, unit: 'seconds' } },
  };
  const mcpStatus = { running: true, enabled: true, name: 'nodeherder-mcp', version: '0.4.2', connectedClients: 1 };

  window.NH_DATA = { devices, appConfig, automations, logs, conversations, mcpStatus, version: '0.12.16-dev', user: { name: 'mike', email: 'owner@home.lan' } };
})();
