/* NodeHerder redesign samples — shared renderer.
   Static prototype: no build, no network, no dependencies. Each design page sets
   window.NH_DESIGN and loads its own skin CSS; this file renders every screen. */
(function () {
  'use strict';
  const DESIGN = window.NH_DESIGN;
  const DATA = JSON.parse(JSON.stringify(window.NH_DATA));
  const I = window.NH_ICON;
  const $ = (s, r) => (r || document).querySelector(s);
  const $$ = (s, r) => Array.from((r || document).querySelectorAll(s));
  const esc = (v) => String(v == null ? '' : v).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  const clone = (o) => JSON.parse(JSON.stringify(o));
  const jsonAttr = (v) => esc(JSON.stringify(v));

  /* ------------------------------------------------------------------ registries */
  // Protocol registry (device-support.md §Protocol registry)
  const PROTOCOLS = {
    mqtt: { label: 'Zigbee', short: 'ZB', icon: 'zigbee', diag: ['lqi', 'battery'], supports: { rename: true, interview: true, remove: true, permitJoin: true } },
    http: { label: 'Wi-Fi · HTTP', short: 'WiFi', icon: 'wifi', diag: ['rssi', 'battery', 'ip'], supports: {} },
  };
  const protocol = (t) => PROTOCOLS[t] || { label: t || 'Unknown', short: (t || '?').slice(0, 4), icon: 'globe', diag: ['battery'], supports: {}, generic: true };

  // Expose presentation table + generic fallback (capability registry)
  const META = {
    temperature: ['Temperature', 'thermometer', 'temp'], local_temperature: ['Temperature', 'thermometer', 'temp'],
    current_heating_setpoint: ['Setpoint', 'flame', 'temp'], humidity: ['Humidity', 'droplet', 'humidity'],
    pressure: ['Pressure', 'pressure', 'air'], co2: ['CO₂', 'co2', 'air'], voc: ['VOC', 'air', 'air'], pm25: ['PM2.5', 'air', 'air'],
    formaldehyd: ['Formaldehyde', 'leaf', 'air'], energy: ['Energy', 'bolt', 'energy'], power: ['Power', 'bolt', 'power'],
    current: ['Current', 'current', 'power'], voltage: ['Voltage', 'bolt', 'power'], state: ['State', 'power', 'switch'],
    brightness: ['Brightness', 'bulb', 'light'], color_temp: ['Colour temp', 'tune', 'light'], illuminance: ['Illuminance', 'sunny', 'light'],
    presence: ['Presence', 'presence', 'motion'], occupancy: ['Motion', 'motion', 'motion'], target_distance: ['Target distance', 'ruler', 'motion'],
    contact: ['Contact', 'door', 'contact'], alarm: ['Alarm', 'alarm', 'alarm'], smoke: ['Smoke', 'smoke', 'alarm'],
    smoke_concentration: ['Smoke density', 'smoke', 'alarm'], device_fault: ['Device fault', 'fault', 'alarm'], test: ['Self test', 'test', 'misc'],
    action: ['Last action', 'dial', 'misc'], battery: ['Battery', 'battery', 'battery'], battpercentage: ['Battery', 'battery', 'battery'],
    linkquality: ['Link quality', 'signal', 'signal'], rssi: ['Wi-Fi signal', 'wifi', 'signal'], ip: ['IP address', 'globe', 'signal'],
  };
  const WORDS = {
    contact: { true: 'Closed', false: 'Open' }, presence: { true: 'Detected', false: 'Clear' }, occupancy: { true: 'Detected', false: 'Clear' },
    smoke: { true: 'SMOKE', false: 'Clear' }, alarm: { true: 'Sounding', false: 'Silent' }, device_fault: { true: 'Fault', false: 'OK' },
    test: { true: 'Testing', false: 'Idle' }, ON: 'On', OFF: 'Off',
  };
  const titleCase = (s) => String(s).replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
  const meta = (n) => { const m = META[n]; return m ? { label: m[0], icon: m[1], kind: m[2] } : { label: titleCase(n), icon: 'question', kind: 'misc' }; };

  const isWritable = (e) => e.access_mode !== 'read';
  const isStateExpose = (e) => e.type === 'binary' && isWritable(e) && e.category === 'measurement';
  const isPreset = (e) => e.values && (e.type === 'enum' || e.type === 'numeric') && Object.keys(e.values).length > 0;
  function capability(e) {
    if (e.category === 'config') return 'config';
    if (e.type === 'binary' && isWritable(e)) return 'switch';
    if (e.type === 'numeric' && isWritable(e) && e.attributes) return 'slider';
    if (isPreset(e) && isWritable(e)) return 'preset';
    if (e.type === 'binary') return 'binary-sensor';
    if (e.name.indexOf('action') === 0) return 'event';
    if (e.category === 'diagnostic') return 'diagnostic';
    return 'reading';
  }
  function fmt(e, v) {
    v = v === undefined ? e.data : v;
    if (v === null || v === undefined || v === '') return '—';
    const w = WORDS[e.name];
    if (e.type === 'binary') {
      if (w) return w[String(v === (e.values && e.values.on) || v === true)];
      return WORDS[v] || (v === true ? 'On' : v === false ? 'Off' : String(v));
    }
    if (e.name === 'brightness') return Math.round((v / (e.attributes ? e.attributes.value_max : 254)) * 100) + ' %';
    if (e.name === 'color_temp') return Math.round(1e6 / v) + ' K';
    if (typeof v === 'number') {
      const s = Number.isInteger(v) ? String(v) : v.toFixed(Math.abs(v) < 10 ? 2 : 1).replace(/0$/, '');
      return s + (e.unit && e.unit !== 'lqi' ? ' ' + e.unit : '');
    }
    return WORDS[v] || titleCase(v);
  }
  const fmtRaw = (v) => (typeof v === 'string' ? (WORDS[v] || titleCase(v)) : v === true ? 'true' : v === false ? 'false' : String(v));

  /* ------------------------------------------------------------------ state */
  const S = {
    devices: DATA.devices, cfg: DATA.appConfig, automations: DATA.automations, logs: DATA.logs, convs: DATA.conversations,
    mcp: DATA.mcpStatus, conn: 'connected', attempt: 0, permit: 0, pending: {}, toasts: [], dialog: null, feed: [],
    navOpen: false, edit: false, pop: false, consolePaused: false, consoleLevel: 'debug', consoleQ: '',
    list: { q: '', proto: 'all', power: 'all', status: 'all', sort: 'friendly_name', dir: 1 },
    activeConv: 'c1', typing: false, period: 'Today', draft: null, original: null, jsonOpen: false, running: {}, collapsed: {},
    authed: true, loginError: false, feedOn: true,
  };
  try { S.feedOn = localStorage.getItem('nh-sample-feed') !== 'off'; } catch (e) { /* default on */ }
  const dev = (id) => S.devices.find((d) => d.id === id);
  const devCfg = (id) => S.cfg.hub.devices.overrides[id] || S.cfg.hub.devices.defaults;
  const isDisabled = (d) => devCfg(d.id).disabled === true;
  const isOnline = (d) => d.availability === 'online';
  const stateExposeOf = (d) => Object.values(d.exposes).find(isStateExpose);
  const lqiOf = (d) => (d.exposes.linkquality ? d.exposes.linkquality.data : null);
  const battOf = (d) => { const e = d.exposes.battery || d.exposes.battpercentage; return e ? e.data : null; };
  const canControl = (d) => S.conn === 'connected' && isOnline(d) && !isDisabled(d);

  function timeAgo(iso) {
    const s = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 1000));
    if (s < 10) return 'just now';
    if (s < 60) return s + ' s ago';
    if (s < 3600) return Math.floor(s / 60) + ' min ago';
    if (s < 86400) return Math.floor(s / 3600) + ' h ago';
    return Math.floor(s / 86400) + ' d ago';
  }
  const clock = (t) => new Date(t).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });

  /* ------------------------------------------------------------------ theme / customiser */
  const TKEY = 'nh-sample-theme-' + DESIGN.id;
  let T = Object.assign({}, DESIGN.defaults);
  try { Object.assign(T, JSON.parse(localStorage.getItem(TKEY) || '{}')); } catch (e) { /* private mode: defaults */ }
  function lum(hex) {
    const n = parseInt(hex.slice(1), 16), c = [(n >> 16) & 255, (n >> 8) & 255, n & 255].map((x) => { x /= 255; return x <= 0.03928 ? x / 12.92 : Math.pow((x + 0.055) / 1.055, 2.4); });
    return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2];
  }
  function applyTheme() {
    const r = document.documentElement;
    const mode = T.mode === 'system' ? (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') : T.mode;
    r.dataset.design = DESIGN.id; r.dataset.shell = DESIGN.shell; r.dataset.mode = mode; r.dataset.density = T.density;
    r.dataset.preset = T.preset; r.dataset.effects = T.effects ? 'on' : 'off';
    r.style.setProperty('--nh-r', T.radius);
    r.style.setProperty('--nh-scale', T.fontScale / 100);
    if (T.accent && T.accent !== 'preset') {
      r.style.setProperty('--nh-accent', T.accent);
      r.style.setProperty('--nh-accent-contrast', lum(T.accent) > 0.36 ? '#111111' : '#ffffff');
    } else { r.style.removeProperty('--nh-accent'); r.style.removeProperty('--nh-accent-contrast'); }
    try { localStorage.setItem(TKEY, JSON.stringify(T)); } catch (e) { /* ignore */ }
  }
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => T.mode === 'system' && applyTheme());

  /* ------------------------------------------------------------------ routing */
  const NAV = [
    { id: 'overview', label: 'Overview', icon: 'overview' }, { id: 'home', label: 'Home', icon: 'home' },
    { id: 'devices', label: 'Devices', icon: 'devices' }, { id: 'list', label: 'Device list', icon: 'list' },
    { id: 'automations', label: 'Automations', icon: 'automation' }, { id: 'assistant', label: 'Assistant', icon: 'sparkles' },
    { id: 'console', label: 'Console', icon: 'terminal' }, { id: 'settings', label: 'Settings', icon: 'cog' },
    { id: 'components', label: 'Components', icon: 'components' }, { id: 'all', label: 'All pages', icon: 'stack' },
  ];
  function route() {
    const p = (location.hash || '#/overview').replace(/^#\/?/, '').split('/').map(decodeURIComponent);
    return { name: p[0] || 'overview', a: p[1], b: p[2] };
  }
  const go = (h) => { location.hash = h; };
  let lastHash = location.hash, suppressHash = false;

  /* ------------------------------------------------------------------ UI helpers */
  const btn = (label, act, o) => {
    o = o || {};
    const cls = 'nh-btn' + (o.kind ? ' is-' + o.kind : '') + (o.sm ? ' is-sm' : '') + (!label ? ' is-icon' : '');
    return '<button type="button" class="' + cls + '"' + (act ? ' data-act="' + act + '"' : '') + (o.arg !== undefined ? ' data-arg="' + esc(o.arg) + '"' : '') +
      (o.disabled ? ' disabled' : '') + (o.title || !label ? ' title="' + esc(o.title || '') + '" aria-label="' + esc(o.title || label) + '"' : '') +
      (o.attrs || '') + '>' + (o.icon ? I(o.icon) : '') + (label ? '<span>' + esc(label) + '</span>' : '') + '</button>';
  };
  const chip = (text, kind, icon, title) => '<span class="nh-chip' + (kind ? ' is-' + kind : '') + '"' + (title ? ' title="' + esc(title) + '"' : '') + '>' + (icon ? I(icon) : '') + '<span>' + esc(text) + '</span></span>';
  const toggle = (on, act, arg, o) => {
    o = o || {};
    return '<button type="button" role="switch" class="nh-toggle' + (o.pending ? ' is-pending' : '') + '" aria-checked="' + (on ? 'true' : 'false') + '" data-act="' + act + '" data-arg="' + esc(arg) + '"' +
      (o.disabled ? ' disabled' : '') + ' aria-label="' + esc(o.label || 'Toggle') + '"><span class="nh-toggle-knob"></span><span class="nh-toggle-txt">' + (on ? 'ON' : 'OFF') + '</span></button>';
  };
  const seg = (opts, val, act, label) => '<div class="nh-seg" role="group" aria-label="' + esc(label || '') + '">' + opts.map((o) => {
    const v = typeof o === 'object' ? o.v : o, t = typeof o === 'object' ? o.t : o;
    return '<button type="button" data-act="' + act + '" data-arg="' + esc(v) + '" aria-pressed="' + (String(v) === String(val)) + '">' + esc(t) + '</button>';
  }).join('') + '</div>';
  const sel = (opts, val, attrs, ph) => { const m = /class="([^"]*)"/.exec(attrs || ''); if (m) attrs = attrs.replace(m[0], ''); return selRaw(opts, val, attrs, ph, m ? ' ' + m[1] : ''); };
  const selRaw = (opts, val, attrs, ph, extra) => '<select class="nh-select' + extra + '" ' + attrs + '>' + (ph ? '<option value=""' + (val === '' || val == null ? ' selected' : '') + '>' + esc(ph) + '</option>' : '') +
    opts.map((o) => '<option value="' + esc(o.v) + '"' + (String(o.v) === String(val) ? ' selected' : '') + '>' + esc(o.t) + '</option>').join('') + '</select>';
  const empty = (icon, title, text, action) => '<div class="nh-empty">' + I(icon) + '<h3>' + esc(title) + '</h3><p>' + esc(text) + '</p>' + (action || '') + '</div>';
  const pageHead = (title, sub, actions) => '<div class="nh-page-head"><div><h1 class="nh-h1">' + esc(title) + '</h1>' + (sub ? '<p class="nh-sub">' + sub + '</p>' : '') + '</div><div class="nh-page-actions">' + (actions || '') + '</div></div>';
  const statusChip = (d) => !isOnline(d) ? chip('Offline', 'danger', 'offline') : isDisabled(d) ? chip('Disabled', 'muted', 'disabled') : chip('Online', 'ok', 'check');
  const protoChip = (d) => { const p = protocol(d.connection_type); return chip(p.label, p.generic ? 'muted' : 'proto-' + d.connection_type, p.icon, 'connection_type: ' + d.connection_type); };
  function signalChip(d) {
    if (d.exposes.linkquality) { const v = lqiOf(d); return chip(v + ' LQI', v < 50 ? 'warn' : '', 'signal', v < 50 ? 'Weak Zigbee link (< 50)' : 'Zigbee link quality'); }
    if (d.exposes.rssi) { const v = d.exposes.rssi.data; return chip(v + ' dBm', v < -75 ? 'warn' : '', 'wifi', 'Wi-Fi RSSI'); }
    return '';
  }
  function powerChip(d) {
    const b = battOf(d);
    if (d.power_source === 'battery' && b != null) return chip(Math.round(b) + ' %', b < 20 ? 'danger' : '', b < 20 ? 'batteryLow' : 'battery', b < 20 ? 'Low battery' : 'Battery');
    return chip('Mains', '', 'plug', 'Mains powered' + (d.connection_type === 'mqtt' ? ' (Zigbee router)' : ''));
  }
  function devIcon(d) {
    const order = ['state', 'brightness', 'alarm', 'smoke', 'contact', 'presence', 'occupancy', 'co2', 'temperature', 'local_temperature', 'action', 'power'];
    for (const n of order) if (d.exposes[n]) return n === 'state' ? (d.exposes.brightness ? 'bulb' : 'plug') : meta(n).icon;
    return 'devices';
  }

  /* ------------------------------------------------------------------ charts (inline SVG) */
  function seeded(seed) { let s = 0; for (const c of seed) s = (s * 31 + c.charCodeAt(0)) >>> 0; return () => ((s = (s * 1664525 + 1013904223) >>> 0) / 4294967296); }
  function series(d, e, n) {
    const r = seeded(d.id + e.name), base = typeof e.data === 'number' ? e.data : 1, out = [];
    let v = base * (0.92 + r() * 0.08);
    for (let i = 0; i < n; i++) {
      v += (r() - 0.5) * Math.max(Math.abs(base) * 0.04, 0.4) + Math.sin(i / (n / 6)) * Math.abs(base) * 0.004;
      out.push(i > n * 0.55 && i < n * 0.62 && d.power_source === 'battery' ? null : +v.toFixed(2)); // gap = missing samples, not zero
    }
    out[n - 1] = base;
    return out;
  }
  function spark(d, e) {
    if (typeof e.data !== 'number') return '';
    const s = series(d, e, 24).map((v) => (v == null ? null : v)), f = s.filter((v) => v != null);
    const mn = Math.min.apply(null, f), mx = Math.max.apply(null, f), w = 100, h = 24;
    let p = '', pen = false;
    s.forEach((v, i) => { if (v == null) { pen = false; return; } const x = (i / (s.length - 1)) * w, y = h - 2 - ((v - mn) / (mx - mn || 1)) * (h - 4); p += (pen ? 'L' : 'M') + x.toFixed(1) + ' ' + y.toFixed(1); pen = true; });
    return '<svg class="nh-spark" viewBox="0 0 100 24" preserveAspectRatio="none" aria-hidden="true"><path d="' + p + '"/></svg>';
  }
  function lineChart(d, e, n) {
    const s = series(d, e, n), f = s.filter((v) => v != null), W = 600, H = 180, L = 44, B = 22;
    const mn = Math.min.apply(null, f), mx = Math.max.apply(null, f), pad = (mx - mn) * 0.1 || 1, lo = mn - pad, hi = mx + pad;
    const X = (i) => L + (i / (n - 1)) * (W - L - 8), Y = (v) => 8 + (1 - (v - lo) / (hi - lo)) * (H - B - 8);
    let path = '', area = '', seg0 = null, pen = false, gaps = '';
    const step = X(1) - X(0);
    s.forEach((v, i) => {
      if (v == null) {
        if (pen) area += 'L' + X(i - 1) + ' ' + (H - B) + 'L' + X(seg0) + ' ' + (H - B) + 'Z';
        gaps += '<rect class="ch-gap" x="' + (X(i) - step / 2) + '" y="8" width="' + step + '" height="' + (H - B - 8) + '"/>';
        pen = false; return;
      }
      if (!pen) { seg0 = i; area += 'M' + X(i) + ' ' + Y(v); } else area += 'L' + X(i) + ' ' + Y(v);
      path += (pen ? 'L' : 'M') + X(i).toFixed(1) + ' ' + Y(v).toFixed(1); pen = true;
    });
    if (pen) area += 'L' + X(n - 1) + ' ' + (H - B) + 'L' + X(seg0) + ' ' + (H - B) + 'Z';
    let grid = '';
    for (let k = 0; k <= 3; k++) { const v = lo + ((hi - lo) * k) / 3, y = Y(v); grid += '<line class="ch-grid" x1="' + L + '" x2="' + (W - 8) + '" y1="' + y + '" y2="' + y + '"/><text class="ch-axis" x="' + (L - 6) + '" y="' + (y + 4) + '" text-anchor="end">' + (+v.toFixed(1)) + '</text>'; }
    ['00:00', '06:00', '12:00', '18:00', 'now'].forEach((t, k) => { grid += '<text class="ch-axis" x="' + (L + (k / 4) * (W - L - 8)) + '" y="' + (H - 6) + '" text-anchor="middle">' + t + '</text>'; });
    const avg = f.reduce((a, b) => a + b, 0) / f.length;
    return '<figure class="nh-chart"><figcaption><span>' + I(meta(e.name).icon) + esc(meta(e.name).label) + (e.unit ? ' <small>(' + esc(e.unit) + ')</small>' : '') + '</span>' +
      '<span class="nh-chart-stats">' + chip('min ' + mn.toFixed(1), 'muted') + chip('avg ' + avg.toFixed(1), 'muted') + chip('max ' + mx.toFixed(1), 'muted') + chip('now ' + fmt(e), 'accent') + '</span></figcaption>' +
      '<svg viewBox="0 0 ' + W + ' ' + H + '" role="img" aria-label="' + esc(meta(e.name).label) + ' history, local time">' + grid + gaps + '<path class="ch-area" d="' + area + '"/><path class="ch-line" d="' + path + '"/></svg>' +
      '<div class="nh-chart-legend"><span class="lg-line"></span>measured <span class="lg-gap"></span>no data (not zero)</div></figure>';
  }
  function binaryChart(d, e) {
    const r = seeded(d.id + e.name + 'b'); let x = 0, out = '', on = r() > 0.5;
    while (x < 100) { const w = 3 + r() * 14; out += '<rect class="' + (on ? 'ch-on' : 'ch-off') + '" x="' + x + '%" y="0" width="' + Math.min(w, 100 - x) + '%" height="100%"/>'; x += w; on = !on; }
    const w = WORDS[e.name] || { true: 'On', false: 'Off' };
    return '<figure class="nh-chart is-binary"><figcaption><span>' + I(meta(e.name).icon) + esc(meta(e.name).label) + '</span><span class="nh-chart-stats">' + chip(w.true, 'accent') + chip(w.false, 'muted') + '</span></figcaption><svg class="nh-bin" preserveAspectRatio="none" role="img" aria-label="' + esc(meta(e.name).label) + ' timeline">' + out + '</svg><div class="nh-axis-row"><span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>now</span></div></figure>';
  }

  /* ------------------------------------------------------------------ tiles */
  const pkey = (id, n) => id + '|' + n;
  function tileState(d, e) {
    const st = stateExposeOf(d), isBin = e.type === 'binary';
    const on = e.name === 'contact' ? e.data === false : isBin ? (e.data === (e.values ? e.values.on : true) || e.data === true) : st ? st.data === st.values.on : typeof e.data === 'number' ? e.data !== 0 : true;
    const alert = (e.name === 'smoke' || e.name === 'alarm' || e.name === 'device_fault') && e.data === true;
    const actionable = (isBin && isWritable(e)) || (!!st && (!isBin || isWritable(e)));
    return { on, alert, actionable, target: isBin && isWritable(e) ? e : st };
  }
  function shortName(d, group) {
    let n = d.friendly_name;
    if (group && n.toLowerCase().indexOf(group.toLowerCase()) === 0) n = n.slice(group.length).trim();
    return n ? n.charAt(0).toUpperCase() + n.slice(1) : d.friendly_name;
  }
  function tile(id, name, o) {
    o = o || {};
    const d = dev(id); if (!d) return '';
    const e = d.exposes[name]; if (!e) return '';
    const m = meta(name), t = tileState(d, e), pend = t.target && S.pending[pkey(id, t.target.name)];
    const off = !isOnline(d), dis = isDisabled(d), ctl = canControl(d) && t.actionable;
    const cls = 'nh-tile kind-' + m.kind + (t.on && !off && !dis ? ' is-on' : '') + (t.alert ? ' is-alert' : '') + (pend ? ' is-pending' : '') + (off ? ' is-offline' : '') + (dis ? ' is-disabled' : '') + (o.edit ? ' is-edit' : '');
    const value = off ? I('offline') + 'Offline' : dis ? I('disabled') + 'Disabled' : pend ? '<span class="nh-spin" aria-hidden="true"></span>Sending ' + esc(fmt(t.target, pend)) + '…' : esc(fmt(e));
    const iconBtn = ctl
      ? '<button type="button" class="nh-tile-ic" data-act="tile-toggle" data-arg="' + esc(pkey(id, t.target.name)) + '" aria-label="Turn ' + esc(d.friendly_name) + ' ' + (t.target.data === t.target.values.on ? 'off' : 'on') + '"' + (pend ? ' disabled' : '') + '>' + I(m.icon) + '</button>'
      : '<span class="nh-tile-ic" aria-hidden="true">' + I(m.icon) + '</span>';
    return '<div class="' + cls + '">' + iconBtn +
      '<button type="button" class="nh-tile-body" data-act="entity" data-arg="' + esc(pkey(id, name)) + '" aria-label="' + esc(m.label + ', ' + d.friendly_name + ': ' + (off ? 'offline' : dis ? 'disabled' : fmt(e)) + '. Show history') + '">' +
      '<span class="nh-tile-label" title="' + esc(d.friendly_name) + '">' + esc(name === 'state' ? shortName(d, o.group) : (e.description || m.label)) + '</span><span class="nh-tile-value" aria-live="polite">' + value + '</span>' +
      (o.showDevice ? '<span class="nh-tile-dev">' + esc(d.friendly_name) + '</span>' : '') + '</button>' +
      (o.spark ? spark(d, e) : '') +
      (o.edit ? '<button type="button" class="nh-tile-del" data-act="tile-remove" data-arg="' + esc(o.group + '|' + id + '|' + name) + '" aria-label="Remove ' + esc(m.label) + ' from ' + esc(o.group) + '">' + I('close') + '</button>' : '') + '</div>';
  }

  const logo = (cls) => (window.NH_LOGO && DESIGN.useLogo ? '<img class="nh-logo-img' + (cls ? ' ' + cls : '') + '" src="' + window.NH_LOGO + '" alt="" width="28" height="41">' : I('hub'));
  /* ------------------------------------------------------------------ shell */
  function shell(view) {
    const r = route();
    const connText = S.conn === 'connected' ? 'Connected' : S.conn === 'connecting' ? 'Reconnecting… (attempt ' + S.attempt + '/10)' : 'Disconnected';
    const navItems = NAV.map((n, i) =>
      '<a class="nh-nav-item' + (r.name === n.id || (r.name === 'device' && n.id === 'list') || (r.name === 'automation' && n.id === 'automations') ? ' is-active' : '') + '" href="#/' + n.id + '"' + (r.name === n.id ? ' aria-current="page"' : '') + ' title="' + esc(n.label) + ' (Alt+' + (i + 1) + ')">' +
      '<span class="nh-nav-key">F' + (i + 1) + '</span><span class="nh-nav-ic">' + I(n.icon) + '</span><span class="nh-nav-label">' + esc(n.label) + '</span></a>').join('');
    return '<button type="button" class="nh-skip" data-act="skip">Skip to content</button>' +
      '<div class="nh-app" data-nav-open="' + S.navOpen + '">' +
      '<nav class="nh-nav" aria-label="Main">' +
      '<div class="nh-brand"><span class="nh-brand-logo' + (window.NH_LOGO && DESIGN.useLogo ? ' has-img' : '') + '">' + logo() + '</span><span class="nh-brand-name">Node<b>Herder</b></span><span class="nh-brand-ver">v' + esc(DATA.version) + '</span></div>' +
      '<div class="nh-nav-list">' + navItems + '</div>' +
      '<div class="nh-nav-foot">' +
      '<button type="button" class="nh-nav-item nh-permit' + (S.permit ? ' is-on' : '') + '" data-act="permit"><span class="nh-nav-key">J</span><span class="nh-nav-ic">' + I('join') + '</span><span class="nh-nav-label">' + (S.permit ? 'Joining <span data-permit>' + fmtTimer(S.permit) + '</span>' : 'Permit join') + '</span></button>' +
      '<button type="button" class="nh-nav-item" data-act="logout"><span class="nh-nav-key">Q</span><span class="nh-nav-ic">' + I('logout') + '</span><span class="nh-nav-label">Sign out</span></button>' +
      '</div></nav>' +
      '<div class="nh-scrim" data-act="nav-close"></div>' +
      '<div class="nh-main">' +
      '<header class="nh-header">' +
      '<button type="button" class="nh-btn is-icon nh-menu-btn" data-act="nav-toggle" aria-label="Open navigation" aria-expanded="' + S.navOpen + '">' + I('menu') + '</button>' + (window.NH_LOGO && DESIGN.useLogo ? '<a class="nh-header-logo" href="#/home" aria-label="NodeHerder home">' + logo() + '</a>' : '') +
      '<div class="nh-header-title"><span class="nh-window-dots" aria-hidden="true"><i></i><i></i><i></i></span><span>' + esc(viewTitle(r)) + '</span></div>' +
      '<div class="nh-header-actions">' +
      (S.permit ? '<button type="button" class="nh-chip is-accent nh-permit-chip" data-act="permit">' + I('join') + '<span>Join open · <span data-permit>' + fmtTimer(S.permit) + '</span></span></button>' : '') +
      '<button type="button" class="nh-conn is-' + S.conn + '" data-act="conn-cycle" title="Hub connection (click to simulate)">' + '<span class="nh-dot"></span><span class="nh-conn-txt">' + esc(connText) + '</span></button>' +
      '<button type="button" class="nh-btn is-icon" data-act="mode-toggle" aria-label="Toggle light/dark">' + I(document.documentElement.dataset.mode === 'dark' ? 'sun' : 'moon') + '</button>' +
      '<button type="button" class="nh-btn is-icon" data-act="pop" aria-label="Customise appearance" aria-expanded="' + S.pop + '">' + I('palette') + '</button>' +
      '<span class="nh-user" title="Signed in">' + I('user') + '<span>' + esc(DATA.user.name) + '</span></span>' +
      '</div></header>' +
      (S.conn !== 'connected' ? '<div class="nh-banner is-' + (S.conn === 'connecting' ? 'warn' : 'danger') + '" role="status">' + I(S.conn === 'connecting' ? 'refresh' : 'offline') + '<span><b>' + esc(connText) + '.</b> Live values may be stale and device controls are disabled until the hub reconnects.</span>' + (S.conn === 'disconnected' ? btn('Reconnect', 'conn-reconnect', { sm: true }) : '') + '</div>' : '') +
      (S.permit ? '<div class="nh-banner is-info" role="status">' + I('join') + '<span><b>Permit join is open</b> for Zigbee devices — <span data-permit>' + fmtTimer(S.permit) + '</span> remaining. Put the new device in pairing mode.</span>' + btn('Stop', 'permit-stop', { sm: true }) + '</div>' : '') +
      '<main class="nh-content" id="nh-content" tabindex="-1">' + view + '</main></div></div>' +
      customiser() + '<div class="nh-toast-host" aria-live="polite">' + S.toasts.map(toastHtml).join('') + '</div>' + dialogHtml();
  }
  function viewTitle(r) {
    if (r.name === 'device') { const d = dev(r.a); return d ? d.friendly_name : 'Device'; }
    if (r.name === 'automation') return 'Automation editor';
    const n = NAV.find((x) => x.id === r.name); return n ? n.label : 'NodeHerder';
  }
  const fmtTimer = (s) => String(Math.floor(s / 60)).padStart(2, '0') + ':' + String(s % 60).padStart(2, '0');

  /* ------------------------------------------------------------------ customiser */
  function customiser() {
    if (!S.pop) return '';
    const D = window.NH_DESIGNS || [];
    return '<aside class="nh-pop" role="dialog" aria-label="Appearance"><div class="nh-pop-head"><h2>Appearance</h2>' + btn('', 'pop', { icon: 'close', title: 'Close' }) + '</div>' + appearanceForm(D) + '</aside>';
  }
  function appearanceForm(D) {
    const r = location.hash;
    return '<div class="nh-form">' +
      '<div class="nh-field"><label>Design</label><div class="nh-design-pick">' + D.map((x) => '<a href="' + x.file + r + '" class="nh-design-opt' + (x.id === DESIGN.id ? ' is-active' : '') + '"' + (x.id === DESIGN.id ? ' aria-current="true"' : '') + '><b>' + esc(x.n + '. ' + x.name) + '</b><small>' + esc(x.short) + '</small></a>').join('') + '</div></div>' +
      '<div class="nh-field"><label>Mode</label>' + seg([{ v: 'system', t: 'System' }, { v: 'light', t: 'Light' }, { v: 'dark', t: 'Dark' }], T.mode, 'th-mode', 'Mode') + '</div>' +
      '<div class="nh-field"><label>Palette</label><div class="nh-swatches">' + DESIGN.presets.map((p) => '<button type="button" class="nh-swatch" data-act="th-preset" data-arg="' + p.id + '" aria-pressed="' + (T.preset === p.id) + '" title="' + esc(p.name) + '"><span style="background:' + p.swatch[0] + '"></span><span style="background:' + p.swatch[1] + '"></span><em>' + esc(p.name) + '</em></button>').join('') + '</div></div>' +
      '<div class="nh-field"><label>Accent</label><div class="nh-swatches is-accent"><button type="button" class="nh-swatch is-auto" data-act="th-accent" data-arg="preset" aria-pressed="' + (T.accent === 'preset') + '" title="Palette default"><em>Auto</em></button>' +
      DESIGN.accents.map((a) => '<button type="button" class="nh-swatch" data-act="th-accent" data-arg="' + a + '" aria-pressed="' + (T.accent === a) + '" title="' + a + '" style="--sw:' + a + '"><span style="background:' + a + '"></span></button>').join('') +
      '<label class="nh-swatch is-custom" title="Custom colour"><input type="color" data-bind="th-accent-custom" value="' + (T.accent !== 'preset' ? T.accent : '#3b82f6') + '" aria-label="Custom accent colour"></label></div></div>' +
      '<div class="nh-field"><label>Density</label>' + seg([{ v: 'compact', t: 'Compact' }, { v: 'comfortable', t: 'Comfortable' }, { v: 'spacious', t: 'Spacious' }], T.density, 'th-density', 'Density') + '</div>' +
      '<div class="nh-field"><label for="th-r">Corner radius <output>' + T.radius + ' px</output></label><input id="th-r" class="nh-slider" type="range" min="0" max="20" value="' + T.radius + '" data-bind="th-radius"></div>' +
      '<div class="nh-field"><label for="th-f">Text size <output>' + T.fontScale + ' %</output></label><input id="th-f" class="nh-slider" type="range" min="90" max="125" step="5" value="' + T.fontScale + '" data-bind="th-font"></div>' +
      '<div class="nh-field is-row"><label>Retro effects <small>' + esc(DESIGN.effectsLabel || 'texture') + '</small></label>' + toggle(T.effects, 'th-effects', '', { label: 'Retro effects' }) + '</div>' +
      '<div class="nh-form-foot">' + btn('Reset to defaults', 'th-reset', { kind: 'ghost', icon: 'refresh' }) + '</div>' +
      '<p class="nh-help">Stored in this browser only (presentation preferences, no household data).</p></div>';
  }

  /* ------------------------------------------------------------------ views */
  const V = {};

  V.login = (r, noFab) => '<div class="nh-login"><form class="nh-login-card" data-form="login" novalidate>' +
    '<div class="nh-login-brand' + (window.NH_LOGO && DESIGN.useLogo ? ' has-img' : '') + '">' + logo('is-lg') + '<div><b>NodeHerder</b><small>Smart-home hub · v' + esc(DATA.version) + '</small></div></div>' +
    '<h1>Sign in</h1>' + (S.loginError ? '<div class="nh-alert is-danger" role="alert">' + I('error') + '<span>Wrong username or password.</span></div>' : '') +
    '<div class="nh-field"><label for="lu">Username</label><input id="lu" class="nh-input" autocomplete="username" value="mike"></div>' +
    '<div class="nh-field"><label for="lp">Password</label><input id="lp" class="nh-input" type="password" autocomplete="current-password" placeholder="any value — type wrong to see error"></div>' +
    '<button type="submit" class="nh-btn is-primary is-block">' + I('lock') + '<span>Sign in</span></button>' +
    '<div class="nh-or"><span>or</span></div><button type="button" class="nh-btn is-block" data-act="login-ok">' + I('globe') + '<span>Continue with Google</span></button>' +
    '<p class="nh-help">Local network hub. Sessions use secure cookies.</p></form>' + (noFab ? '' : customiserFab()) + '</div>';
  const customiserFab = () => '<button type="button" class="nh-btn is-icon nh-fab" data-act="pop" aria-label="Customise appearance">' + I('palette') + '</button>' + customiser();

  V.overview = () => {
    const ds = S.devices, on = ds.filter(isOnline), z = ds.filter((d) => d.connection_type === 'mqtt'), w = ds.filter((d) => d.connection_type !== 'mqtt');
    const routers = z.filter((d) => d.power_source === 'mains'), ends = z.filter((d) => d.power_source === 'battery');
    const lowL = ds.filter((d) => lqiOf(d) != null && lqiOf(d) < 50), lowB = ds.filter((d) => battOf(d) != null && battOf(d) < 20);
    const stat = (n, label, icon, kind, link, hint) => '<a class="nh-stat' + (kind ? ' is-' + kind : '') + '" href="' + link + '"><span class="nh-stat-ic">' + I(icon) + '</span><span class="nh-stat-num">' + n + '</span><span class="nh-stat-label">' + esc(label) + '</span>' + (hint ? '<span class="nh-stat-hint">' + esc(hint) + '</span>' : '') + '</a>';
    const attention = ds.filter((d) => !isOnline(d) || isDisabled(d) || (lqiOf(d) != null && lqiOf(d) < 50) || (battOf(d) != null && battOf(d) < 20));
    const auto = S.automations, en = auto.filter((a) => a.enabled).length, sch = auto.filter((a) => a.schedules.length).length;
    return pageHead('Overview', 'Network health, recent activity and things that need attention.', btn('Permit join', 'permit', { icon: 'join', kind: 'primary' })) +
      '<div class="nh-stats">' +
      stat(ds.length, 'Devices', 'devices', '', '#/list', on.length + ' online') +
      stat(ds.length - on.length, 'Offline', 'offline', ds.length - on.length ? 'danger' : 'ok', '#/list/status=offline') +
      stat(routers.length, 'Zigbee routers', 'plug', '', '#/list/power=mains') +
      stat(ends.length, 'End devices', 'battery', '', '#/list/power=battery') +
      stat(w.length, 'Wi-Fi · HTTP', 'wifi', '', '#/list/proto=http') +
      stat(lowL.length, 'Low LQI', 'signal', lowL.length ? 'warn' : 'ok', '#/list/sort=lqi', '< 50') +
      stat(lowB.length, 'Low battery', 'batteryLow', lowB.length ? 'warn' : 'ok', '#/list/sort=battery', '< 20 %') +
      '</div><div class="nh-cols">' +
      '<section class="nh-card nh-col-wide nh-activity' + (S.feedOn ? '' : ' is-off') + '"><div class="nh-card-head"><h2 class="nh-card-title">' + I('wave') + 'Recent activity</h2><span class="nh-card-meta">' + (S.feedOn ? 'Live, last ' + S.feed.length + ' updates' : 'Off') + '</span><label class="nh-feed-switch"><span>Show</span>' + toggle(S.feedOn, 'feed-toggle', '', { label: 'Show recent activity' }) + '</label></div>' +
      (S.feedOn ? '<div class="nh-feed" data-region="feed">' + feedHtml(10) + '</div>' : '<p class="nh-feed-off">' + I('pause') + 'Recent activity is off. Turn it on to watch device changes as they happen. Nothing is collected while it is off.</p>') + '</section>' +
      '<div class="nh-col-side">' +
      '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">' + I('warn') + 'Needs attention</h2></div><ul class="nh-attn">' +
      (attention.length ? attention.map((d) => '<li><a href="#/device/' + encodeURIComponent(d.id) + '"><span class="nh-attn-ic">' + I(devIcon(d)) + '</span><span><b>' + esc(d.friendly_name) + '</b><small>' + esc(!isOnline(d) ? 'Offline since ' + timeAgo(d.last_seen) : isDisabled(d) ? 'Disabled in device settings' : battOf(d) != null && battOf(d) < 20 ? 'Battery ' + Math.round(battOf(d)) + ' %' : 'Weak link ' + lqiOf(d) + ' LQI') + '</small></span>' + I('chevronR') + '</a></li>').join('') : '<li class="nh-muted">All good.</li>') + '</ul></section>' +
      '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">' + I('hub') + 'Hub</h2></div><dl class="nh-kv">' +
      '<dt>Connection</dt><dd>' + chip(S.conn, S.conn === 'connected' ? 'ok' : 'warn') + '</dd><dt>Zigbee bridge</dt><dd>' + chip(S.permit ? 'Joining ' + fmtTimer(S.permit) : 'Join closed', S.permit ? 'accent' : 'muted') + '</dd>' +
      '<dt>MCP server</dt><dd>' + chip(S.mcp.running ? 'Running · ' + S.mcp.connectedClients + ' client' : 'Stopped', S.mcp.running ? 'ok' : 'muted') + '</dd>' +
      '<dt>Automations</dt><dd>' + en + ' enabled · ' + sch + ' scheduled</dd><dt>Version</dt><dd>v' + esc(DATA.version) + '</dd></dl></section></div></div>';
  };
  function feedHtml(n) {
    if (!S.feed.length) return '<p class="nh-muted nh-pad">Waiting for device updates…</p>';
    return S.feed.slice(0, n).map((f) => '<a class="nh-feed-row" href="#/device/' + encodeURIComponent(f.id) + '"><span class="nh-feed-ic">' + I(meta(f.changes[0].n).icon) + '</span><span class="nh-feed-dev">' + esc(f.name) + '</span><span class="nh-feed-chg">' +
      f.changes.map((c) => '<span>' + esc(meta(c.n).label.toLowerCase()) + ' <s>' + esc(c.from) + '</s> → <b>' + esc(c.to) + '</b></span>').join('') + '</span><time>' + clock(f.t) + '</time></a>').join('');
  }

  V.home = () => {
    const G = S.cfg.hub.dashboardGroups, names = Object.keys(G);
    const actions = S.edit
      ? btn('New area', 'group-new', { icon: 'plus' }) + btn('Export', 'group-export', { icon: 'download', disabled: !names.length }) + btn('Import', 'group-import', { icon: 'upload' }) + btn('Done', 'edit-toggle', { icon: 'check', kind: 'primary' })
      : btn('Edit dashboard', 'edit-toggle', { icon: 'edit' });
    if (!names.length) return pageHead('Home', '', actions) + empty('home', 'No areas yet', 'Create an area and add the sensors and switches you want at a glance.', btn('New area', 'group-new', { icon: 'plus', kind: 'primary' }));
    return pageHead('Home', S.edit ? 'Editing — changes are saved to the hub as you go.' : names.length + ' areas · tap an icon to switch, tap a tile for history.', actions) +
      '<div class="nh-areas' + (S.edit ? ' is-edit' : '') + '">' + names.map((g) => {
        const items = []; Object.values(G[g].deviceGroup).forEach((dg) => dg.exposes.forEach((x) => items.push([dg.deviceId, x])));
        return '<section class="nh-area"><div class="nh-area-head"><h2>' + esc(g) + '</h2><span class="nh-area-count">' + items.length + '</span>' +
          (S.edit ? '<span class="nh-area-tools">' + btn('', 'group-rename', { icon: 'edit', title: 'Rename ' + g, arg: g, sm: true }) + btn('', 'group-delete', { icon: 'trash', title: 'Delete ' + g, arg: g, sm: true, kind: 'danger' }) + '</span>' : '') + '</div>' +
          '<div class="nh-tiles">' + items.map((it) => tile(it[0], it[1], { edit: S.edit, group: g })).join('') +
          (S.edit ? '<button type="button" class="nh-tile is-add" data-act="group-add" data-arg="' + esc(g) + '">' + I('plus') + '<span>Add entity</span></button>' : '') + '</div></section>';
      }).join('') + '</div>';
  };

  V.devices = () => pageHead('Devices', S.devices.length + ' devices · measurements at a glance', '<a class="nh-btn" href="#/list">' + I('list') + '<span>Table view</span></a>') +
    '<div class="nh-dcards">' + S.devices.slice().sort((a, b) => a.friendly_name.localeCompare(b.friendly_name, undefined, { sensitivity: 'base' })).map(dcard).join('') + '</div>';
  function dcard(d) {
    const ms = Object.values(d.exposes).filter((e) => e.category === 'measurement'), off = !isOnline(d), dis = isDisabled(d);
    return '<article class="nh-dcard' + (off ? ' is-offline' : '') + (dis ? ' is-disabled' : '') + '"><a class="nh-dcard-head" href="#/device/' + encodeURIComponent(d.id) + '"><span class="nh-dcard-ic">' + I(devIcon(d)) + '</span><span class="nh-dcard-t"><b>' + esc(d.friendly_name) + '</b><small>' + esc(d.description) + '</small></span>' + I('chevronR') + '</a>' +
      (off || dis ? '<div class="nh-overlay">' + I(off ? 'offline' : 'disabled') + '<span>' + (off ? 'Device is offline · last seen ' + timeAgo(d.last_seen) : 'Device is disabled') + '</span></div>' :
        '<div class="nh-dcard-rows">' + ms.map((e) => exposeRow(d, e, true)).join('') + '</div>') +
      '<div class="nh-dcard-foot">' + chip(timeAgo(d.last_seen), 'muted', 'clock') + signalChip(d) + powerChip(d) + protoChip(d) + '</div></article>';
  }
  function exposeRow(d, e, compact) {
    const m = meta(e.name), cap = capability(e), k = pkey(d.id, e.name), pend = S.pending[k], dis = !canControl(d);
    let ctl = '<span class="nh-row-val" aria-live="polite">' + (pend !== undefined ? '<span class="nh-spin"></span>' + esc(fmt(e, pend)) + '…' : esc(fmt(e))) + '</span>';
    if (cap === 'switch') ctl = '<span class="nh-row-val">' + (pend !== undefined ? '<small class="nh-pending-txt">sending…</small>' : '') + toggle(e.data === e.values.on || e.data === true, 'expose-toggle', k, { disabled: dis || pend !== undefined, pending: pend !== undefined, label: m.label + ' ' + d.friendly_name }) + '</span>';
    else if (cap === 'slider') {
      const a = e.attributes, v = pend !== undefined ? pend : e.data;
      ctl = '<span class="nh-row-val is-slider"><input class="nh-slider" type="range" min="' + a.value_min + '" max="' + a.value_max + '" value="' + v + '" data-bind="expose-slide" data-arg="' + esc(k) + '" aria-label="' + esc(m.label) + '"' + (dis ? ' disabled' : '') + '><output>' + (pend !== undefined ? '<span class="nh-spin"></span>' : '') + esc(fmt(e, v)) + '</output></span>';
    } else if (cap === 'preset' && !compact) ctl = '<span class="nh-row-val">' + seg(Object.keys(e.values).map((x) => ({ v: k + '|' + x, t: titleCase(x) })), k + '|' + Object.keys(e.values).find((x) => e.values[x] === (pend !== undefined ? pend : e.data)), 'expose-preset', m.label) + '</span>';
    return '<div class="nh-row kind-' + m.kind + '"><span class="nh-row-ic">' + I(m.icon) + '</span><span class="nh-row-label">' + esc(e.description || m.label) + '</span>' + ctl + '</div>';
  }

  V.list = (r) => {
    const L = S.list;
    if (r.a) r.a.split('&').forEach((kv) => { const p = kv.split('='); if (p[0] === 'sort') { L.sort = p[1]; L.dir = 1; } else if (L[p[0]] !== undefined) L[p[0]] = p[1]; });
    if (r.a) { history.replaceState(null, '', '#/list'); }
    let ds = S.devices.filter((d) => (!L.q || (d.friendly_name + ' ' + d.id + ' ' + d.description).toLowerCase().includes(L.q.toLowerCase())) &&
      (L.proto === 'all' || d.connection_type === L.proto) && (L.power === 'all' || d.power_source === L.power) &&
      (L.status === 'all' || (L.status === 'online' && isOnline(d) && !isDisabled(d)) || (L.status === 'offline' && !isOnline(d)) || (L.status === 'disabled' && isDisabled(d))));
    const key = { friendly_name: (d) => d.friendly_name.toLowerCase(), lqi: (d) => (lqiOf(d) == null ? 999 : lqiOf(d)), battery: (d) => (battOf(d) == null ? 999 : battOf(d)), last_seen: (d) => -new Date(d.last_seen).getTime() }[L.sort];
    ds.sort((a, b) => (key(a) > key(b) ? 1 : key(a) < key(b) ? -1 : 0) * L.dir);
    const th = (k, t) => '<th scope="col" aria-sort="' + (L.sort === k ? (L.dir > 0 ? 'ascending' : 'descending') : 'none') + '"><button type="button" data-act="list-sort" data-arg="' + k + '">' + esc(t) + (L.sort === k ? I(L.dir > 0 ? 'up' : 'down') : '') + '</button></th>';
    const f = (lab, k, opts) => '<div class="nh-filter"><span>' + lab + '</span>' + seg(opts, L[k], 'list-filter-' + k, lab) + '</div>';
    return pageHead('Device list', ds.length + ' of ' + S.devices.length + ' devices', btn('Permit join', 'permit', { icon: 'join' })) +
      '<div class="nh-toolbar"><label class="nh-search">' + I('search') + '<input class="nh-input" type="search" placeholder="Search name, IEEE address, model…" value="' + esc(L.q) + '" data-bind="list-q" data-focus-key="list-q" aria-label="Search devices"></label>' +
      f('Protocol', 'proto', [{ v: 'all', t: 'All' }, { v: 'mqtt', t: 'Zigbee' }, { v: 'http', t: 'Wi-Fi' }]) +
      f('Power', 'power', [{ v: 'all', t: 'All' }, { v: 'mains', t: 'Mains' }, { v: 'battery', t: 'Battery' }]) +
      f('Status', 'status', [{ v: 'all', t: 'All' }, { v: 'online', t: 'Online' }, { v: 'offline', t: 'Offline' }, { v: 'disabled', t: 'Disabled' }]) + '</div>' +
      (ds.length ? '<div class="nh-table-wrap"><table class="nh-table"><thead><tr>' + th('friendly_name', 'Device') + '<th scope="col">Protocol</th><th scope="col" class="hide-sm">Address</th>' + th('lqi', 'Signal') + th('battery', 'Power') + th('last_seen', 'Last seen') + '<th scope="col">Status</th><th scope="col"><span class="sr-only">Actions</span></th></tr></thead><tbody>' +
        ds.map((d) => { const p = protocol(d.connection_type).supports; return '<tr><td><a class="nh-dev-link" href="#/device/' + encodeURIComponent(d.id) + '"><span class="nh-dev-ic">' + I(devIcon(d)) + '</span><span><b>' + esc(d.friendly_name) + '</b><small>' + esc(d.description) + '</small></span></a></td><td>' + protoChip(d) + '</td><td class="hide-sm"><code>' + esc(d.id) + '</code></td><td>' + (signalChip(d) || '<span class="nh-muted">—</span>') + '</td><td>' + powerChip(d) + '</td><td>' + esc(timeAgo(d.last_seen)) + '</td><td>' + statusChip(d) + '</td><td class="nh-row-actions">' +
          btn('', 'dev-rename', { icon: 'edit', title: p.rename ? 'Rename' : 'Rename not supported for ' + protocol(d.connection_type).label, arg: d.id, sm: true, disabled: !p.rename }) +
          btn('', 'dev-interview', { icon: 'refresh', title: p.interview ? 'Interview' : 'Interview is Zigbee only', arg: d.id, sm: true, disabled: !p.interview }) +
          btn('', 'dev-remove', { icon: 'trash', title: p.remove ? 'Remove' : 'Remove is Zigbee only', arg: d.id, sm: true, kind: 'danger', disabled: !p.remove }) + '</td></tr>'; }).join('') + '</tbody></table></div>'
        : empty('search', 'No devices match', 'Try clearing the search or filters.', btn('Clear filters', 'list-clear', { kind: 'primary' })));
  };

  V.device = (r) => {
    const d = dev(r.a); if (!d) return empty('question', 'Device not found', 'It may have been removed or renamed.', '<a class="nh-btn is-primary" href="#/list">Back to list</a>');
    const tab = r.b || 'about', p = protocol(d.connection_type).supports;
    const tabs = [['about', 'About'], ['controls', 'Controls'], ['settings', 'Settings'], ['metrics', 'Metrics']];
    return '<div class="nh-dev-head"><a class="nh-btn is-icon" href="#/list" aria-label="Back to device list">' + I('chevronL') + '</a><span class="nh-dev-hero">' + I(devIcon(d)) + '</span><div class="nh-dev-title"><h1 class="nh-h1">' + esc(d.friendly_name) + '</h1><p class="nh-sub">' + esc(d.description) + '</p><div class="nh-chips">' + statusChip(d) + protoChip(d) + signalChip(d) + powerChip(d) + chip(timeAgo(d.last_seen), 'muted', 'clock') + '</div></div>' +
      '<div class="nh-page-actions">' + btn('Rename', 'dev-rename', { icon: 'edit', arg: d.id, disabled: !p.rename, title: p.rename ? '' : 'Not supported for this protocol' }) + btn('Interview', 'dev-interview', { icon: 'refresh', arg: d.id, disabled: !p.interview }) + btn('Remove', 'dev-remove', { icon: 'trash', arg: d.id, kind: 'danger', disabled: !p.remove }) + '</div></div>' +
      '<div class="nh-tabs" role="tablist">' + tabs.map((t) => '<a role="tab" class="nh-tab" href="#/device/' + encodeURIComponent(d.id) + '/' + t[0] + '" aria-selected="' + (tab === t[0]) + '">' + t[1] + '</a>').join('') + '</div>' +
      '<div class="nh-tabpanel" role="tabpanel">' + (DV[tab] || DV.about)(d) + '</div>';
  };
  const DV = {
    about(d) {
      const pr = protocol(d.connection_type);
      return '<div class="nh-cols"><section class="nh-card nh-col-wide"><div class="nh-card-head"><h2 class="nh-card-title">Device</h2></div><dl class="nh-kv">' +
        '<dt>' + (d.connection_type === 'mqtt' ? 'IEEE address' : 'Device id') + '</dt><dd><code>' + esc(d.id) + '</code> ' + btn('', 'copy', { icon: 'copy', title: 'Copy', arg: d.id, sm: true }) + '</dd><dt>Friendly name</dt><dd>' + esc(d.friendly_name) + '</dd><dt>Description</dt><dd>' + esc(d.description) + '</dd>' +
        '<dt>Availability</dt><dd>' + statusChip(d) + '</dd><dt>Last seen</dt><dd>' + esc(timeAgo(d.last_seen)) + ' <small class="nh-muted">(' + esc(new Date(d.last_seen).toLocaleString()) + ')</small></dd><dt>Power source</dt><dd>' + powerChip(d) + '</dd>' +
        '<dt>Connection</dt><dd>' + protoChip(d) + ' <small class="nh-muted">connection_type = "' + esc(d.connection_type) + '"</small></dd></dl></section>' +
        '<section class="nh-card nh-col-side"><div class="nh-card-head"><h2 class="nh-card-title">' + I(pr.icon) + esc(pr.label) + '</h2></div><div class="nh-card-body"><p class="nh-help">' +
        (d.connection_type === 'mqtt' ? 'Reported through Zigbee2MQTT over MQTT. Bridge operations (rename, interview, remove, permit join) are available.' : d.connection_type === 'http' ? 'Pushes JSON to <code>POST /api/collect</code> (e.g. ESP32/ESPHome, Tasmota, Shelly scripts). Bridge operations do not apply.' : 'Unknown protocol — shown with generic presentation.') +
        '</p><dl class="nh-kv">' + Object.values(d.exposes).filter((e) => e.category === 'diagnostic').map((e) => '<dt>' + esc(meta(e.name).label) + '</dt><dd>' + esc(fmt(e)) + '</dd>').join('') + '</dl></div></section></div>';
    },
    controls(d) {
      const cats = ['measurement', 'diagnostic', 'config'];
      return (!canControl(d) ? '<div class="nh-alert is-warn">' + I('warn') + '<span>' + (!isOnline(d) ? 'Device is offline — controls are disabled. Values shown are the last known state.' : isDisabled(d) ? 'Device is disabled in its settings.' : 'Hub connection lost — controls are disabled.') + '</span></div>' : '<p class="nh-help">Controls show <b>sending…</b> until the device confirms the new state.</p>') +
        cats.map((c) => { const es = Object.values(d.exposes).filter((e) => e.category === c); return es.length ? '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">' + titleCase(c) + '</h2><span class="nh-card-meta">' + es.length + '</span></div><div class="nh-rows">' + es.map((e) => exposeRow(d, e, false)).join('') + '</div></section>' : ''; }).join('');
    },
    settings(d) {
      const ov = S.cfg.hub.devices.overrides[d.id], c = devCfg(d.id), cats = c.defaultDebounceByCategory, deb = c.debounceOverrides || {};
      return '<div class="nh-alert ' + (ov ? 'is-info' : '') + '">' + I('info') + '<span>' + (ov ? 'This device has its own settings (override).' : 'Using hub defaults from Settings → Device defaults.') + '</span>' + (ov ? btn('Revert to defaults', 'cfg-revert', { arg: d.id, sm: true }) : btn('Customise this device', 'cfg-override', { arg: d.id, sm: true, kind: 'primary' })) + '</div>' +
        '<section class="nh-card"><div class="nh-rows">' +
        '<div class="nh-row"><span class="nh-row-ic">' + I('disabled') + '</span><span class="nh-row-label">Disabled<small>Ignore updates and hide values</small></span><span class="nh-row-val">' + toggle(c.disabled, 'cfg-toggle', d.id + '|disabled', { disabled: !ov, label: 'Disabled' }) + '</span></div>' +
        '<div class="nh-row"><span class="nh-row-ic">' + I('chart') + '</span><span class="nh-row-label">Store metrics<small>Keep history for charts and the assistant</small></span><span class="nh-row-val">' + toggle(c.metricsEnabled, 'cfg-toggle', d.id + '|metricsEnabled', { disabled: !ov, label: 'Store metrics' }) + '</span></div></div></section>' +
        '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">Debounce by category</h2><span class="nh-card-meta">minimum time between stored samples</span></div><div class="nh-rows">' +
        Object.keys(cats).map((k) => '<div class="nh-row"><span class="nh-row-label">' + titleCase(k) + '</span><span class="nh-row-val nh-inline"><input class="nh-input is-num" type="number" min="0" value="' + cats[k].value + '"' + (ov ? '' : ' disabled') + ' aria-label="' + k + ' debounce"> ' + sel([{ v: 'seconds', t: 'seconds' }, { v: 'minutes', t: 'minutes' }, { v: 'hours', t: 'hours' }], cats[k].unit, (ov ? '' : 'disabled ') + 'aria-label="unit"') + '</span></div>').join('') + '</div></section>' +
        '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">Per-expose overrides</h2>' + btn('Add', 'toast', { icon: 'plus', sm: true, disabled: !ov, arg: 'Expose picker opens here' }) + '</div><div class="nh-rows">' +
        (Object.keys(deb).length ? Object.keys(deb).map((k) => '<div class="nh-row"><span class="nh-row-ic">' + I(meta(k).icon) + '</span><span class="nh-row-label">' + esc(meta(k).label) + '</span><span class="nh-row-val">' + chip(deb[k].value + ' ' + deb[k].unit, 'accent') + btn('', 'toast', { icon: 'trash', sm: true, title: 'Remove', arg: 'Override removed (sample)' }) + '</span></div>').join('') : '<p class="nh-muted nh-pad">No per-expose overrides.</p>') + '</div></section>' +
        '<p class="nh-help">Debounce only thins stored metrics and UI updates — automations always receive every state change.</p>';
    },
    metrics(d) {
      const periods = ['1 Hour', '6 Hours', 'Today', 'Yesterday', '7 Days', '30 Days'];
      const es = Object.values(d.exposes).filter((e) => e.category === 'measurement');
      return '<div class="nh-toolbar">' + seg(periods, S.period, 'period', 'Period') + '<span class="nh-muted">Times in hub local time (' + esc(Intl.DateTimeFormat().resolvedOptions().timeZone) + ')</span></div>' +
        (!devCfg(d.id).metricsEnabled ? empty('chart', 'Metrics are off for this device', 'Enable “Store metrics” in the Settings tab.') :
          '<div class="nh-charts">' + es.map((e) => (e.type === 'numeric' ? lineChart(d, e, 48) : e.type === 'binary' ? binaryChart(d, e) : '')).join('') + '</div>');
    },
  };

  V.automations = () => {
    const A = S.automations;
    return pageHead('Automations', A.length + ' automations · ' + A.reduce((n, a) => n + a.triggers.length, 0) + ' triggers', btn('Create', 'auto-create', { icon: 'plus', kind: 'primary' })) +
      (A.length ? '<div class="nh-auto-list">' + A.map((a, i) => {
        const d = dev(a.id), st = autoStatus(a), manual = a.triggers.findIndex(canManual);
        return '<article class="nh-auto"><a class="nh-auto-main" href="#/automation/' + encodeURIComponent(a.id) + '"><span class="nh-auto-ic">' + I(d ? devIcon(d) : 'automation') + '</span><span class="nh-auto-t"><b>' + esc(a.friendlyname) + '</b><small>' + esc(a.description || '—') + '</small>' +
          '<span class="nh-auto-sum">' + esc(describeTrigger(a, a.triggers[0])) + (a.triggers.length > 1 ? ' <em>+' + (a.triggers.length - 1) + ' more</em>' : '') + '</span></span></a>' +
          '<div class="nh-auto-side">' + chip(st.text, st.kind, st.icon, st.title) + chip(a.triggers.length + ' trigger' + (a.triggers.length === 1 ? '' : 's'), 'muted') +
          (manual >= 0 ? btn('Run', 'auto-run', { icon: 'play', sm: true, arg: i + '|' + manual, disabled: S.running[a.id + manual] || S.conn !== 'connected', title: 'Run “' + a.triggers[manual].name + '” now' }) : '') +
          toggle(a.enabled, 'auto-enable', i, { label: 'Enabled' }) + btn('', 'auto-delete', { icon: 'trash', kind: 'danger', title: 'Delete ' + a.friendlyname, arg: i, sm: true }) + '</div></article>';
      }).join('') + '</div>' : empty('automation', 'No automations', 'Create one from a device: when something happens, if conditions match, then do something.', btn('Create', 'auto-create', { icon: 'plus', kind: 'primary' })));
  };
  function autoStatus(a) {
    if (!a.enabled) return { text: 'Disabled', kind: 'muted', icon: 'disabled' };
    if (a.schedules.length) { const nxt = a.schedules.map((s) => s.type + ' ' + s.startAt).join(' · '); return { text: 'Scheduled', kind: 'info', icon: 'clock', title: nxt }; }
    return { text: 'Enabled', kind: 'ok', icon: 'check' };
  }
  const canManual = (t) => t.conditions.every((c) => c.type === 'time');

  /* ---------- automation helpers (selectors/automation.ts in the plan) */
  function condText(a, c) {
    if (c.type === 'time') { const s = c.timeRange.startAt || '??:??', e = c.timeRange.endAt || '??:??'; return 'time is ' + s + '–' + e + (s > e && c.timeRange.endAt ? ' (overnight)' : ''); }
    const d = dev(a.id), e = d && d.exposes[c.name];
    return (c.name ? meta(c.name).label.toLowerCase() : '‹expose›') + ' ' + c.equality + ' ' + (c.value === '' || c.value == null ? '‹value›' : e ? fmt(e, c.value) : fmtRaw(c.value));
  }
  function actText(x) {
    const d = dev(x.id), dn = d ? d.friendly_name : '‹device›';
    if (x.type === 'trigger') return 'set ' + dn + ' ' + (x.exposes.length ? x.exposes.map((r) => (r.name ? meta(r.name).label.toLowerCase() : '‹expose›') + ' = ' + (r.data === null || r.data === '' ? '‹value›' : d && d.exposes[r.name] ? fmt(d.exposes[r.name], r.data) : fmtRaw(r.data))).join(', ') : '‹values›') + (x.delay && x.delay.value ? ' after ' + x.delay.value + ' ' + x.delay.unit : '');
    if (x.type === 'step') return 'step ' + dn + ' ' + (x.property ? meta(x.property).label.toLowerCase() : '‹expose›') + ' ' + ((x.steps[0] && x.steps[0].operator) || '+') + ' ' + (x.data === null || x.data === '' ? '‹amount›' : x.data);
    return 'cycle ' + dn + ' ' + (x.property ? meta(x.property).label.toLowerCase() : '‹expose›') + ' presets';
  }
  function describeTrigger(a, t) {
    if (!t) return 'No triggers yet';
    const d = dev(a.id);
    const when = t.type === 'manualTrigger' ? 'When run manually' : 'When ' + (d ? d.friendly_name : a.friendlyname) + ' ' + (t.name ? meta(t.name).label.toLowerCase() : '‹expose›') + ' changes';
    const ifs = t.conditions.length ? ', if ' + t.conditions.map((c) => condText(a, c)).join(' and ') : '';
    const thens = t.actions.length ? ', then ' + t.actions.map(actText).join(', then ') : ', then ‹add an action›';
    return when + ifs + thens + '.';
  }
  function validate(a) {
    const errs = [], warns = [], P = (p, m) => errs.push({ p, m });
    if (!a.triggers.length) warns.push({ p: 'triggers', m: 'No triggers — this automation does nothing.' });
    a.triggers.forEach((t, i) => {
      const tp = 'triggers.' + i;
      if (!t.name) P(tp + '.name', 'Choose which expose fires this trigger.');
      if (t.type === 'deviceTrigger' && !t.conditions.length) warns.push({ p: tp, m: 'No conditions — the hub blocks device-triggered runs without conditions; this trigger can only run manually.' });
      t.conditions.forEach((c, j) => {
        const cp = tp + '.conditions.' + j;
        if (c.type === 'expose') { if (!c.name) P(cp + '.name', 'Pick an expose.'); else if (c.value === '' || c.value == null) P(cp + '.value', 'Enter a value to compare.'); }
        else { if (!/^\d\d:\d\d$/.test(c.timeRange.startAt)) P(cp + '.timeRange.startAt', 'Start time required.'); if (!/^\d\d:\d\d$/.test(c.timeRange.endAt)) P(cp + '.timeRange.endAt', 'End time required.'); }
      });
      if (!t.actions.length) P(tp + '.actions', 'Add at least one action.');
      t.actions.forEach((x, k) => {
        const ap = tp + '.actions.' + k, d = dev(x.id);
        if (!x.id) { P(ap + '.id', 'Choose a target device.'); return; }
        if (d && !isOnline(d)) warns.push({ p: ap, m: d.friendly_name + ' is offline — the command will not be delivered until it reconnects.' });
        if (x.type === 'trigger') {
          if (!x.exposes.length) P(ap + '.exposes', 'Add at least one value to set.');
          x.exposes.forEach((r, n) => { if (!r.name) P(ap + '.exposes.' + n + '.name', 'Pick an expose.'); else if (r.data === null || r.data === '') P(ap + '.exposes.' + n + '.data', 'Enter a value.'); if (x.id === a.id && r.name === t.name) warns.push({ p: ap, m: 'Sets the same expose that fires this trigger — the hub cooldown prevents loops, but check this is intended.' }); });
        } else if (x.type === 'step') { if (!x.property) P(ap + '.property', 'Pick a numeric expose.'); if (x.data === null || x.data === '' || isNaN(+x.data)) P(ap + '.data', 'Enter a step amount.'); }
        else if (!x.property) P(ap + '.property', 'Pick an expose with presets.');
      });
    });
    return { errs, warns };
  }

  V.automation = (r) => {
    const src = S.automations.find((a) => a.id === r.a);
    if (!S.draft || S.draft.id !== r.a) {
      if (!src && !(S.draft && S.draft.id === r.a)) return empty('question', 'Automation not found', '', '<a class="nh-btn is-primary" href="#/automations">Back</a>');
      S.original = clone(src); S.draft = clone(src); S.jsonOpen = false; S.collapsed = {};
    }
    return editorHtml();
  };
  const dirty = () => S.draft && JSON.stringify(S.draft) !== JSON.stringify(S.original);
  // Editor building blocks shared by every design's editor layout (same data-path contract).
  function edParts() {
    const a = S.draft, d = dev(a.id), v = validate(a), E = {};
    v.errs.forEach((e) => (E[e.p] = e.m));
    const err = (p) => (E[p] ? '<span class="nh-err" id="err-' + p.replace(/\./g, '-') + '">' + I('error') + esc(E[p]) + '</span>' : '');
    const inv = (p) => (E[p] ? ' aria-invalid="true" aria-describedby="err-' + p.replace(/\./g, '-') + '"' : '');
    const srcEx = d ? Object.values(d.exposes).filter((e) => e.category !== 'config') : [];
    const writableDevs = S.devices.filter((x) => Object.values(x.exposes).some((e) => isWritable(e) && e.category !== 'config'));
    const devOpts = (list) => list.map((x) => ({ v: x.id, t: x.friendly_name + (isOnline(x) ? '' : ' (offline)') }));
    const withPath = (html, p) => html.replace(/data-arg="/g, 'data-path="' + p + '" data-arg="');
    function valueInput(dv, exName, val, path) {
      const e = dv && dv.exposes[exName], attrs = 'data-path="' + path + '" data-focus-key="' + path + '"' + inv(path);
      if (!e) return '<input class="nh-input" disabled placeholder="value" ' + attrs + '>';
      if (e.type === 'binary' || (e.values && e.type === 'enum')) {
        const opts = e.values ? Object.keys(e.values).map((k) => ({ v: JSON.stringify(e.values[k]), t: e.type === 'binary' ? fmt(e, e.values[k]) : titleCase(k) })) : [{ v: 'true', t: fmt(e, true) }, { v: 'false', t: fmt(e, false) }];
        return sel(opts, val === null || val === '' ? '' : JSON.stringify(val), 'data-kind="json" ' + attrs + ' aria-label="Value"', 'Choose…');
      }
      if (e.type === 'numeric') return '<span class="nh-unit-input"><input class="nh-input is-num" type="number" data-kind="num" ' + attrs + ' value="' + esc(val == null ? '' : val) + '"' + (e.attributes ? ' min="' + e.attributes.value_min + '" max="' + e.attributes.value_max + '"' : '') + ' aria-label="Value">' + (e.unit ? '<em>' + esc(e.unit) + '</em>' : '') + '</span>';
      return '<input class="nh-input" data-kind="str" ' + attrs + ' value="' + esc(val == null ? '' : val) + '" aria-label="Value">';
    }
    const P = { a, d, v, E, err, inv, srcEx, valueInput, withPath };
    P.tp = (i) => 'triggers.' + i;
    P.trigType = (i) => withPath(seg([{ v: 'deviceTrigger', t: 'Device changes' }, { v: 'manualTrigger', t: 'Manual only' }], a.triggers[i].type, 'ed-trig-type', 'Trigger type'), P.tp(i) + '.type');
    P.srcSelect = (i) => sel(srcEx.map((x) => ({ v: x.name, t: meta(x.name).label + '  (now ' + fmt(x) + ')' })), a.triggers[i].name, 'data-path="' + P.tp(i) + '.name" data-kind="str"' + inv(P.tp(i) + '.name') + ' aria-label="Source expose"', 'Which expose?');
    P.srcChip = () => '<span class="nh-src">' + I(d ? devIcon(d) : 'devices') + esc(d ? d.friendly_name : a.friendlyname) + '</span>';
    P.whenBody = (i) => '<div class="nh-inline">' + P.trigType(i) + '</div><div class="nh-inline">' + P.srcChip() + P.srcSelect(i) + '<span class="nh-muted">changes</span></div>' + err(P.tp(i) + '.name');
    P.condTools = (i, j) => { const cp = P.tp(i) + '.conditions.' + j, n = a.triggers[i].conditions.length; return '<span class="nh-reorder">' + btn('', 'ed-move', { icon: 'up', sm: true, title: 'Move up', arg: cp + '|-1', disabled: j === 0 }) + btn('', 'ed-move', { icon: 'down', sm: true, title: 'Move down', arg: cp + '|1', disabled: j === n - 1 }) + btn('', 'ed-del', { icon: 'trash', sm: true, kind: 'danger', title: 'Remove condition', arg: cp }) + '</span>'; };
    P.condFields = (i, j) => {
      const c = a.triggers[i].conditions[j], cp = P.tp(i) + '.conditions.' + j;
      if (c.type === 'time') return '<span class="nh-inline"><span>between</span><input type="time" class="nh-input" data-path="' + cp + '.timeRange.startAt" data-kind="str" data-focus-key="' + cp + 's" value="' + esc(c.timeRange.startAt) + '"' + inv(cp + '.timeRange.startAt') + ' aria-label="From"><span>and</span><input type="time" class="nh-input" data-path="' + cp + '.timeRange.endAt" data-kind="str" data-focus-key="' + cp + 'e" value="' + esc(c.timeRange.endAt) + '"' + inv(cp + '.timeRange.endAt') + ' aria-label="To">' + (c.timeRange.startAt > c.timeRange.endAt && c.timeRange.endAt ? chip('overnight', 'info', 'moon') : '') + '</span>' + err(cp + '.timeRange.startAt') + err(cp + '.timeRange.endAt');
      const e = d && d.exposes[c.name], ops = !e || e.type === 'numeric' ? ['=', '<', '<=', '>', '>='] : ['='];
      return '<span class="nh-inline">' + sel(srcEx.map((x) => ({ v: x.name, t: meta(x.name).label + '  (now ' + fmt(x) + ')' })), c.name, 'data-path="' + cp + '.name" data-kind="str" data-reset="' + cp + '.value"' + inv(cp + '.name') + ' aria-label="Expose"', 'Expose…') +
        sel(ops.map((o) => ({ v: o, t: o })), c.equality, 'data-path="' + cp + '.equality" data-kind="str" aria-label="Operator" class="is-op"') + valueInput(d, c.name, c.value, cp + '.value') + '</span>' + err(cp + '.name') + err(cp + '.value');
    };
    P.condKind = (c) => '<span class="nh-cond-kind">' + I(c.type === 'time' ? 'clock' : c.name ? meta(c.name).icon : 'filter') + (c.type === 'time' ? 'Time' : 'Value') + '</span>';
    P.condRow = (i, j) => { const f = P.condFields(i, j), k = f.indexOf('<span class="nh-err"'); return '<div class="nh-cond">' + P.condKind(a.triggers[i].conditions[j]) + (k < 0 ? f + P.condTools(i, j) : f.slice(0, k) + P.condTools(i, j) + f.slice(k)) + '</div>'; };
    P.condAdd = (i) => '<div class="nh-inline">' + btn('Value condition', 'ed-add-cond', { icon: 'plus', sm: true, arg: P.tp(i) + '|expose' }) + btn('Time window', 'ed-add-cond', { icon: 'clock', sm: true, arg: P.tp(i) + '|time' }) + '</div>';
    P.actType = (i, k) => withPath(seg([{ v: 'trigger', t: 'Set' }, { v: 'step', t: 'Step' }, { v: 'preset', t: 'Cycle' }], a.triggers[i].actions[k].type, 'ed-act-type', 'Action type'), P.tp(i) + '.actions.' + k);
    P.actTarget = (i, k) => { const x = a.triggers[i].actions[k], ap = P.tp(i) + '.actions.' + k; return sel(devOpts(x.type === 'preset' ? S.devices.filter((z) => Object.values(z.exposes).some(isPreset)) : writableDevs), x.id, 'data-path="' + ap + '.id" data-kind="str" data-reset-act="1"' + inv(ap + '.id') + ' aria-label="Target device"', 'Target device…'); };
    P.actTools = (i, k) => { const ap = P.tp(i) + '.actions.' + k, n = a.triggers[i].actions.length; return '<span class="nh-reorder">' + btn('', 'ed-move', { icon: 'up', sm: true, title: 'Move up', arg: ap + '|-1', disabled: k === 0 }) + btn('', 'ed-move', { icon: 'down', sm: true, title: 'Move down', arg: ap + '|1', disabled: k === n - 1 }) + btn('', 'ed-del', { icon: 'trash', sm: true, kind: 'danger', title: 'Remove action', arg: ap }) + '</span>'; };
    P.actBody = (i, k) => {
      const x = a.triggers[i].actions[k], ap = P.tp(i) + '.actions.' + k, td = dev(x.id);
      const filt = x.type === 'step' ? (e) => e.type === 'numeric' && isWritable(e) && e.category === 'measurement' : x.type === 'preset' ? isPreset : (e) => isWritable(e) && e.category !== 'config';
      const opts = td ? Object.values(td.exposes).filter(filt).map((e) => ({ v: e.name, t: meta(e.name).label })) : [];
      if (x.type === 'trigger') {
        return x.exposes.map((r2, n) => { const rp = ap + '.exposes.' + n; return '<div class="nh-set-row"><span class="nh-inline">' + sel(opts, r2.name, 'data-path="' + rp + '.name" data-kind="str" data-reset="' + rp + '.data"' + inv(rp + '.name') + ' aria-label="Expose"', 'Expose…') + '<span>=</span>' + valueInput(td, r2.name, r2.data, rp + '.data') + btn('', 'ed-del', { icon: 'close', sm: true, title: 'Remove value', arg: rp }) + '</span>' + err(rp + '.name') + err(rp + '.data') + '</div>'; }).join('') + err(ap + '.exposes') +
          '<div class="nh-inline nh-act-opts">' + btn('Add value', 'ed-add-row', { icon: 'plus', sm: true, arg: ap, disabled: !x.id }) +
          (x.exposes.length > 1 ? '<span class="nh-muted">Publish</span>' + withPath(seg([{ v: 'batch', t: 'Together' }, { v: 'single', t: 'One by one' }], x.publishMode || 'batch', 'ed-publish', 'Publish mode'), ap + '.publishMode') : '') +
          '<label class="nh-check"><input type="checkbox" data-act-check="ed-delay" data-path="' + ap + '"' + (x.delay && x.delay.value ? ' checked' : '') + '> Wait first</label>' +
          (x.delay && x.delay.value ? '<input class="nh-input is-num" type="number" min="1" data-path="' + ap + '.delay.value" data-kind="num" data-focus-key="' + ap + 'dv" value="' + x.delay.value + '" aria-label="Delay">' + sel([{ v: 'seconds', t: 'seconds' }, { v: 'minutes', t: 'minutes' }, { v: 'hours', t: 'hours' }], x.delay.unit, 'data-path="' + ap + '.delay.unit" data-kind="str" aria-label="Delay unit"') : '') + '</div>';
      }
      if (x.type === 'step') return '<div class="nh-inline">' + sel(opts, x.property, 'data-path="' + ap + '.property" data-kind="str" data-sync-steps="1"' + inv(ap + '.property') + ' aria-label="Numeric expose"', 'Numeric expose…') +
        withPath(seg([{ v: '+', t: '+' }, { v: '-', t: '−' }, { v: '*', t: '×' }], (x.steps[0] && x.steps[0].operator) || '+', 'ed-step-op', 'Operator'), ap) +
        '<input class="nh-input is-num" type="number" data-path="' + ap + '.data" data-kind="num" data-focus-key="' + ap + 'd" value="' + esc(x.data == null ? '' : x.data) + '"' + inv(ap + '.data') + ' aria-label="Amount"></div>' + err(ap + '.property') + err(ap + '.data');
      const pe = td && td.exposes[x.property];
      return '<div class="nh-inline">' + sel(opts, x.property, 'data-path="' + ap + '.property" data-kind="str"' + inv(ap + '.property') + ' aria-label="Preset expose"', 'Expose with presets…') + (pe && pe.values ? '<span class="nh-muted">cycles</span>' + Object.keys(pe.values).map((k2) => chip(titleCase(k2), 'muted')).join('<span class="nh-arrow">→</span>') : '') + '</div>' + err(ap + '.property');
    };
    P.actRow = (i, k) => '<div class="nh-act"><div class="nh-act-head"><span class="nh-act-n">' + (k + 1) + '</span>' + P.actType(i, k) + P.actTarget(i, k) + P.actTools(i, k) + '</div>' + err(P.tp(i) + '.actions.' + k + '.id') + '<div class="nh-act-body">' + P.actBody(i, k) + '</div></div>';
    P.actAdd = (i) => '<div class="nh-inline">' + btn('Add action', 'ed-add-act', { icon: 'plus', sm: true, arg: P.tp(i) }) + (a.triggers[i].actions.length > 1 ? '<span class="nh-muted">Actions run in order.</span>' : '') + '</div>';
    P.trigErrs = (i) => v.errs.filter((e) => e.p.indexOf(P.tp(i)) === 0 && (e.p.length === P.tp(i).length || e.p[P.tp(i).length] === '.'));
    P.trigWarns = (i) => v.warns.filter((w) => w.p.indexOf(P.tp(i)) === 0 && (w.p.length === P.tp(i).length || w.p[P.tp(i).length] === '.'));
    P.trigTools = (i) => { const t = a.triggers[i], te = P.trigErrs(i); return (te.length ? chip(te.length + ' to fix', 'danger', 'error') : chip('Valid', 'ok', 'check')) +
      (canManual(t) ? btn(S.running[a.id + i] ? 'Running…' : 'Run', 'ed-run', { icon: 'play', sm: true, arg: i, disabled: S.running[a.id + i] || S.conn !== 'connected', title: 'Run this trigger now (POST /api/automation/trigger)' }) : '') +
      btn('', 'ed-dup', { icon: 'copy', sm: true, title: 'Duplicate trigger', arg: i }) + btn('', 'ed-del', { icon: 'trash', sm: true, kind: 'danger', title: 'Delete trigger', arg: P.tp(i) }); };
    P.warnList = (i) => { const tw = P.trigWarns(i); return tw.length ? '<ul class="nh-warn-list">' + tw.map((w) => '<li>' + I('warn') + esc(w.m) + '</li>').join('') + '</ul>' : ''; };
    P.sentence = (i) => '<p class="nh-sentence" data-region="sent-' + i + '">' + esc(describeTrigger(a, a.triggers[i])) + '</p>';
    P.details = () => '<div class="nh-form"><div class="nh-field"><label>Device</label><div>' + esc(a.friendlyname) + ' <code class="nh-muted">' + esc(a.id) + '</code></div></div>' +
      '<div class="nh-field"><label for="ed-desc">Description</label><input id="ed-desc" class="nh-input" data-path="description" data-kind="str" data-focus-key="desc" value="' + esc(a.description) + '"></div>' +
      '<div class="nh-field is-row"><label>Enabled</label>' + toggle(a.enabled, 'ed-enabled', '', { label: 'Enabled' }) + '</div></div>';
    P.schedule = () => {
      const sch = a.schedules, en = sch.find((s) => s.type === 'enable'), di = sch.find((s) => s.type === 'disable');
      const pct = (hm) => { const p = hm.split(':'); return ((+p[0] * 60 + +p[1]) / 1440) * 100; };
      let band = '';
      if (en && di) { const s = pct(en.startAt), e = pct(di.startAt); band = s < e ? '<span class="nh-tl-on" style="left:' + s + '%;width:' + (e - s) + '%"></span>' : '<span class="nh-tl-on" style="left:0;width:' + e + '%"></span><span class="nh-tl-on" style="left:' + s + '%;width:' + (100 - s) + '%"></span>'; }
      const nowP = ((new Date().getHours() * 60 + new Date().getMinutes()) / 1440) * 100;
      return '<div class="nh-timeline" aria-label="24 hour schedule">' + band + sch.map((s) => '<span class="nh-tl-mark is-' + s.type + '" style="left:' + pct(s.startAt) + '%" title="' + s.type + ' at ' + s.startAt + '"></span>').join('') + '<span class="nh-tl-now" style="left:' + nowP + '%" title="now"></span></div><div class="nh-axis-row"><span>00</span><span>06</span><span>12</span><span>18</span><span>24</span></div>' +
        ['enable', 'disable'].map((ty) => { const s = sch.find((x) => x.type === ty); return '<div class="nh-sched-row"><span class="nh-sched-t is-' + ty + '">' + (ty === 'enable' ? 'Enable at' : 'Disable at') + '</span>' + (s ? '<input type="time" class="nh-input" data-sched="' + ty + '" value="' + s.startAt + '" aria-label="' + ty + ' time">' + btn('', 'ed-sched-del', { icon: 'close', sm: true, title: 'Remove', arg: ty }) : btn('Add', 'ed-sched-add', { icon: 'plus', sm: true, arg: ty })) + '</div>'; }).join('') +
        '<p class="nh-help">Hub timezone: ' + esc(Intl.DateTimeFormat().resolvedOptions().timeZone) + '</p>';
    };
    P.checks = () => '<div data-region="checks">' + checksHtml(v) + '</div>';
    P.savebar = () => '<div class="nh-savebar" data-region="savebar">' + savebarHtml(v) + '</div>';
    P.json = () => (S.jsonOpen ? '<section class="nh-card nh-json-card"><div class="nh-card-head"><h2 class="nh-card-title">' + I('code') + 'saveAutomation payload</h2><span class="nh-card-meta">read-only · exact shape sent over WS</span></div><pre class="nh-json" data-region="json">' + esc(JSON.stringify(a, null, 2)) + '</pre></section>' : '');
    P.back = () => '<a class="nh-back" href="#/automations">' + I('chevronL') + 'Automations</a>';
    return P;
  }
  function editorHtml() {
    const P = edParts(), a = P.a, d = P.d;
    const trig = a.triggers.map((t, i) => {
      const col = S.collapsed[i], te = P.trigErrs(i);
      return '<section class="nh-trig' + (col ? ' is-collapsed' : '') + (te.length ? ' has-err' : '') + '" aria-label="Trigger ' + (i + 1) + '"><div class="nh-trig-head"><button type="button" class="nh-trig-toggle" data-act="ed-collapse" data-arg="' + i + '" aria-expanded="' + !col + '">' + I(col ? 'chevronR' : 'chevronD') + '<b>Trigger ' + (i + 1) + '</b></button>' +
        P.sentence(i) + '<span class="nh-trig-tools">' + P.trigTools(i) + '</span></div>' +
        (col ? '' : '<div class="nh-flow">' +
          '<div class="nh-block is-when"><div class="nh-block-label"><span>When</span></div><div class="nh-block-body">' + P.whenBody(i) + '</div></div>' +
          '<div class="nh-block is-if"><div class="nh-block-label"><span>If</span></div><div class="nh-block-body">' + (t.conditions.map((c, j) => P.condRow(i, j)).join('') || '<p class="nh-muted">No conditions — always continue.</p>') + P.condAdd(i) + '</div></div>' +
          '<div class="nh-block is-then"><div class="nh-block-label"><span>Then</span></div><div class="nh-block-body">' + t.actions.map((x, k) => P.actRow(i, k)).join('') + P.err(P.tp(i) + '.actions') + P.actAdd(i) + '</div></div></div>') +
        P.warnList(i) + '</section>';
    }).join('');
    return '<div class="nh-editor"><div class="nh-editor-main">' +
      '<div class="nh-page-head"><div>' + P.back() + '<h1 class="nh-h1">' + esc(a.friendlyname) + '</h1><p class="nh-sub">Source device · ' + (d ? protoChip(d) : '') + ' ' + esc(a.triggers.length + ' trigger' + (a.triggers.length === 1 ? '' : 's')) + '</p></div></div>' +
      (trig || empty('automation', 'No triggers', 'Add a trigger to decide what this device reacts to.')) +
      '<button type="button" class="nh-add-trig" data-act="ed-add-trig">' + I('plus') + '<span>Add trigger</span></button></div>' +
      '<aside class="nh-editor-side">' +
      '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">Details</h2></div><div class="nh-card-body">' + P.details() + '</div></section>' +
      '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">' + I('calendar') + 'Schedule</h2></div><div class="nh-card-body">' + P.schedule() + '</div></section>' +
      '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">' + I('check') + 'Checks</h2></div><div class="nh-card-body">' + P.checks() + '</div></section></aside>' +
      P.savebar() + P.json() + '</div>';
  }
  const checksHtml = (v) => (v.errs.length ? '<p class="nh-err-sum">' + I('error') + v.errs.length + ' issue' + (v.errs.length > 1 ? 's' : '') + ' to fix before saving</p>' : '<p class="nh-ok-sum">' + I('check') + 'Ready to save</p>') + (v.warns.length ? '<p class="nh-warn-sum">' + I('warn') + v.warns.length + ' warning' + (v.warns.length > 1 ? 's' : '') + '</p>' : '');
  const savebarHtml = (v) => '<span class="nh-dirty">' + (dirty() ? I('edit') + 'Unsaved changes' : I('check') + 'All changes saved') + '</span>' + btn(S.jsonOpen ? 'Hide JSON' : 'View JSON', 'ed-json', { icon: 'code', kind: 'ghost' }) + btn('Discard', 'ed-discard', { disabled: !dirty() }) + btn(v.errs.length ? 'Fix ' + v.errs.length + ' to save' : 'Save', 'ed-save', { kind: 'primary', icon: 'check', disabled: !dirty() || v.errs.length > 0 });

  V.assistant = () => {
    const c = S.convs.find((x) => x.id === S.activeConv) || S.convs[0];
    const sug = ['Which devices have weak signal?', 'How much energy did the attic socket use today?', 'Was the front door opened last night?', 'Summarise living room presence this week'];
    return '<div class="nh-chat"><aside class="nh-chat-side"><div class="nh-chat-side-head"><h2>History</h2>' + btn('New chat', 'chat-new', { icon: 'plus', sm: true, kind: 'primary' }) + '</div><ul>' +
      S.convs.map((x) => '<li><button type="button" class="nh-conv' + (x.id === c.id ? ' is-active' : '') + '" data-act="chat-sel" data-arg="' + x.id + '"' + (x.id === c.id ? ' aria-current="true"' : '') + '><b>' + esc(x.title) + '</b><small>' + esc(timeAgo(new Date(x.updated).toISOString())) + '</small></button>' + btn('', 'chat-del', { icon: 'trash', sm: true, title: 'Delete conversation', arg: x.id }) + '</li>').join('') + '</ul></aside>' +
      '<section class="nh-chat-main"><div class="nh-msgs" data-region="msgs">' +
      (c.messages.length ? c.messages.map((m) => '<div class="nh-msg is-' + m.role + '"><span class="nh-msg-av">' + I(m.role === 'user' ? 'user' : 'sparkles') + '</span><div class="nh-msg-b">' + esc(m.content) + '</div></div>').join('') + (S.typing ? '<div class="nh-msg is-assistant"><span class="nh-msg-av">' + I('sparkles') + '</span><div class="nh-msg-b nh-typing"><i></i><i></i><i></i></div></div>' : '')
        : '<div class="nh-chat-empty">' + I('sparkles') + '<h2>Ask about your home</h2><p>Answers come from device metrics through the MCP server. The assistant cannot control devices.</p><div class="nh-suggest">' + sug.map((s) => '<button type="button" class="nh-chip is-btn" data-act="chat-suggest" data-arg="' + esc(s) + '">' + esc(s) + '</button>').join('') + '</div></div>') +
      '</div><form class="nh-chat-input" data-form="chat"><textarea class="nh-input" rows="1" placeholder="Ask about temperatures, energy, presence…" data-focus-key="chat" aria-label="Message"></textarea><button type="submit" class="nh-btn is-primary is-icon" aria-label="Send"' + (S.typing ? ' disabled' : '') + '>' + I('send') + '</button></form></section></div>';
  };

  const LV = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'];
  V.console = () => {
    const min = LV.indexOf(S.consoleLevel), q = S.consoleQ.toLowerCase();
    const rows = S.logs.filter((l) => LV.indexOf(l.level) >= min && (!q || (l.source + ' ' + l.message).toLowerCase().includes(q)));
    return pageHead('Console', 'Live hub log stream · ' + S.logs.length + ' messages buffered', btn(S.consolePaused ? 'Resume' : 'Pause', 'log-pause', { icon: S.consolePaused ? 'play' : 'pause' }) + btn('Clear', 'log-clear', { icon: 'trash' }) + btn('Download', 'toast', { icon: 'download', arg: 'Log file download (POST /api/logfile)' })) +
      '<div class="nh-toolbar"><div class="nh-filter"><span>Min level</span>' + seg(LV, S.consoleLevel, 'log-level', 'Minimum level') + '</div><label class="nh-search">' + I('search') + '<input class="nh-input" type="search" placeholder="Filter messages" value="' + esc(S.consoleQ) + '" data-bind="log-q" data-focus-key="log-q" aria-label="Filter log"></label>' +
      '<div class="nh-filter"><span>Remote logger</span>' + toggle(S.cfg.hub.logger.enableRemoteLogger, 'log-remote', '', { label: 'Remote logger' }) + '</div></div>' +
      '<div class="nh-console" data-region="console" role="log" aria-live="off">' + (rows.length ? rows.map((l) => '<div class="nh-log-row lvl-' + l.level + '"><time>' + clock(l.timestamp) + '</time><span class="nh-lvl">' + l.level.toUpperCase() + '</span><span class="nh-log-src">' + esc(l.source) + '</span><span class="nh-log-msg">' + esc(l.message) + '</span></div>').join('') : '<p class="nh-muted nh-pad">No messages at this level.</p>') +
      (S.consolePaused ? '<div class="nh-log-paused">' + I('pause') + 'Paused — new messages are buffered</div>' : '') + '</div>';
  };

  V.settings = (r) => {
    const tab = r.a || 'appearance', H = S.cfg.hub;
    const tabs = [['appearance', 'Appearance'], ['devices', 'Device defaults'], ['history', 'History'], ['logger', 'Logger'], ['mcp', 'MCP server'], ['assistant', 'Assistant'], ['about', 'About']];
    const interval = (label, iv, help) => '<div class="nh-row"><span class="nh-row-label">' + label + (help ? '<small>' + help + '</small>' : '') + '</span><span class="nh-row-val nh-inline"><input class="nh-input is-num" type="number" min="0" value="' + iv.value + '" aria-label="' + label + '">' + sel([{ v: 'seconds', t: 'seconds' }, { v: 'minutes', t: 'minutes' }, { v: 'hours', t: 'hours' }, { v: 'days', t: 'days' }], iv.unit, 'aria-label="unit"') + '</span></div>';
    const P = {
      appearance: () => '<section class="nh-card"><div class="nh-card-body">' + appearanceForm(window.NH_DESIGNS || []) + '</div></section>',
      devices: () => { const c = H.devices.defaults; return '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">Defaults for every device</h2><span class="nh-card-meta">' + Object.keys(H.devices.overrides).length + ' devices override these</span></div><div class="nh-rows">' +
        '<div class="nh-row"><span class="nh-row-label">Store metrics<small>Default for new devices</small></span><span class="nh-row-val">' + toggle(c.metricsEnabled, 'toast', 'Saved (saveDeviceConfigDefaults)', { label: 'Store metrics' }) + '</span></div>' +
        Object.keys(c.defaultDebounceByCategory).map((k) => interval(titleCase(k) + ' debounce', c.defaultDebounceByCategory[k])).join('') + '</div><div class="nh-card-foot">' + btn('Save', 'toast', { kind: 'primary', arg: 'Saved device defaults' }) + '</div></section>'; },
      history: () => '<section class="nh-card"><div class="nh-rows">' + interval('Sleep timeout', H.history.sleepTimeout, 'Gap after which a series is treated as missing, not zero') + interval('Keep history for', H.history.expireAt, 'Older samples are pruned') + '</div><div class="nh-card-foot">' + btn('Save', 'toast', { kind: 'primary', arg: 'Saved history settings' }) + '</div></section>',
      logger: () => '<section class="nh-card"><div class="nh-rows"><div class="nh-row"><span class="nh-row-label">Stream logs to browser</span><span class="nh-row-val">' + toggle(H.logger.enableRemoteLogger, 'log-remote', '', { label: 'Remote logger' }) + '</span></div><div class="nh-row"><span class="nh-row-label">Level</span><span class="nh-row-val">' + seg(LV, H.logger.level, 'toast', 'Level') + '</span></div></div></section>',
      mcp: () => '<section class="nh-card"><div class="nh-card-head"><h2 class="nh-card-title">' + I('stack') + 'Model Context Protocol</h2>' + chip(S.mcp.running ? 'Running' : 'Stopped', S.mcp.running ? 'ok' : 'muted', S.mcp.running ? 'check' : 'pause') + '</div><dl class="nh-kv"><dt>Server</dt><dd>' + esc(S.mcp.name) + ' v' + esc(S.mcp.version) + '</dd><dt>Endpoint</dt><dd><code>/api/mcp</code> · <code>/api/mcp/events</code></dd><dt>Clients</dt><dd>' + S.mcp.connectedClients + '</dd></dl>' +
        '<div class="nh-alert is-warn">' + I('warn') + '<span>MCP routes are currently public — restrict network exposure (see docs/setup.md).</span></div><div class="nh-card-foot">' + (S.mcp.running ? btn('Restart', 'mcp', { icon: 'refresh', arg: 'restart' }) + btn('Stop', 'mcp', { icon: 'pause', kind: 'danger', arg: 'stop' }) : btn('Start', 'mcp', { icon: 'play', kind: 'primary', arg: 'start' })) + '</div></section>',
      assistant: () => '<section class="nh-card"><div class="nh-card-body nh-form"><div class="nh-field"><label for="as-url">Assistant service URL</label><input id="as-url" class="nh-input" value="' + esc(H.assistant.url) + '"><span class="nh-help">The assistant uses MCP for read-only device context. Provider keys live on the assistant service, never in the browser.</span></div></div><div class="nh-card-foot">' + btn('Save', 'toast', { kind: 'primary', arg: 'Saved assistant URL' }) + '</div></section>',
      about: () => '<section class="nh-card"><dl class="nh-kv"><dt>Version</dt><dd>v' + esc(DATA.version) + '</dd><dt>Design sample</dt><dd>' + esc(DESIGN.name) + '</dd><dt>Protocols</dt><dd>' + Object.keys(PROTOCOLS).map((k) => chip(PROTOCOLS[k].label, 'proto-' + k, PROTOCOLS[k].icon)).join(' ') + '</dd><dt>Spec</dt><dd>specs/007-frontend-redesign</dd></dl></section>',
    };
    return pageHead('Settings', '', '') + '<div class="nh-settings"><nav class="nh-vtabs" aria-label="Settings sections">' + tabs.map((t) => '<a href="#/settings/' + t[0] + '" class="nh-vtab"' + (t[0] === tab ? ' aria-current="page"' : '') + '>' + esc(t[1]) + '</a>').join('') + '</nav><div class="nh-settings-body">' + (P[tab] || P.appearance)() + '</div></div>';
  };

  V.components = () => {
    const ls = dev(DATA.devices[9].id), tv = DATA.devices[12].id, sec = (t, body, note) => '<section class="nh-card nh-kit"><div class="nh-card-head"><h2 class="nh-card-title">' + esc(t) + '</h2>' + (note ? '<span class="nh-card-meta">' + esc(note) + '</span>' : '') + '</div><div class="nh-card-body">' + body + '</div></section>';
    const fake = (cls, ic, label, val) => '<div class="nh-tile ' + cls + '"><span class="nh-tile-ic">' + I(ic) + '</span><span class="nh-tile-body"><span class="nh-tile-label">' + label + '</span><span class="nh-tile-value">' + val + '</span></span></div>';
    return pageHead('Components', 'Every building block and state used by the screens.', '') + '<div class="nh-kit-grid">' +
      sec('Buttons', '<div class="nh-inline">' + btn('Primary', 'toast', { kind: 'primary', icon: 'check', arg: 'Primary' }) + btn('Secondary', 'toast', { arg: 'Secondary' }) + btn('Ghost', 'toast', { kind: 'ghost', arg: 'Ghost' }) + btn('Danger', 'toast', { kind: 'danger', icon: 'trash', arg: 'Danger' }) + btn('Disabled', '', { disabled: true }) + btn('', 'toast', { icon: 'cog', title: 'Icon button', arg: 'Icon' }) + btn('Small', 'toast', { sm: true, arg: 'Small' }) + '</div>') +
      sec('Chips & badges', '<div class="nh-inline">' + chip('Online', 'ok', 'check') + chip('Offline', 'danger', 'offline') + chip('Disabled', 'muted', 'disabled') + chip('Scheduled', 'info', 'clock') + chip('Zigbee', 'proto-mqtt', 'zigbee') + chip('Wi-Fi · HTTP', 'proto-http', 'wifi') + chip('Thread', 'muted', 'globe', 'Unknown protocol fallback') + chip('14 LQI', 'warn', 'signal') + chip('-61 dBm', '', 'wifi') + chip('15 %', 'danger', 'batteryLow') + chip('Mains', '', 'plug') + '</div>') +
      sec('Tile states', '<div class="nh-tiles">' + tile(ls.id, 'brightness') + tile(tv, 'power', { spark: true }) + fake('kind-light is-on is-pending', 'bulb', 'Brightness', '<span class="nh-spin"></span>Sending 50 %…') + fake('kind-alarm is-alert is-on', 'smoke', 'Smoke', 'SMOKE') + fake('kind-temp is-offline', 'thermometer', 'Temperature', I('offline') + 'Offline') + fake('kind-motion is-disabled', 'motion', 'Motion', I('disabled') + 'Disabled') + fake('kind-temp is-stale', 'thermometer', 'Temperature', '18.5 °C <small>· 2 h old</small>') + fake('kind-switch', 'plug', 'State', 'Off') + '</div>', 'pending ≠ confirmed (NH-02)') +
      sec('Inputs', '<div class="nh-form"><div class="nh-field"><label for="k1">Text</label><input id="k1" class="nh-input" value="Living room"></div><div class="nh-field"><label for="k2">With error</label><input id="k2" class="nh-input" aria-invalid="true" value=""><span class="nh-err">' + I('error') + 'Enter a value.</span></div><div class="nh-field"><label>Select</label>' + sel([{ v: 1, t: 'Living Room Light' }, { v: 2, t: 'Attic room Light' }], 1, 'aria-label="Select"') + '</div><div class="nh-field is-row"><label>Toggle</label>' + toggle(true, 'kit-toggle', '', { label: 'Demo' }) + toggle(false, 'kit-toggle', '', { label: 'Demo off' }) + toggle(true, '', '', { disabled: true, label: 'Disabled' }) + '</div><div class="nh-field"><label>Segmented</label>' + seg(['Today', '7 Days', '30 Days'], 'Today', 'kit-seg', 'Period') + '</div><div class="nh-field"><label>Slider</label><input class="nh-slider" type="range" value="60" aria-label="Demo slider"></div></div>') +
      sec('Connection & banners', '<div class="nh-inline">' + btn('Connected', 'conn-set', { arg: 'connected', sm: true }) + btn('Reconnecting', 'conn-set', { arg: 'connecting', sm: true }) + btn('Disconnected', 'conn-set', { arg: 'disconnected', sm: true }) + btn('Permit join', 'permit', { icon: 'join', sm: true }) + '</div><div class="nh-alert is-info">' + I('info') + '<span>Info alert</span></div><div class="nh-alert is-warn">' + I('warn') + '<span>Warning alert</span></div><div class="nh-alert is-danger">' + I('error') + '<span>Error alert</span></div>', 'changes the header for real') +
      sec('Toasts', '<div class="nh-inline">' + btn('Success', 'toast-kind', { arg: 'ok|Device renamed' }) + btn('Error', 'toast-kind', { arg: 'danger|No confirmation from Hallway Shelly plug within 5 s' }) + btn('Info', 'toast-kind', { arg: 'info|Interview started' }) + '</div><p class="nh-help">Toggle the <b>Hallway Shelly plug</b> on Home to see a real command timeout.</p>') +
      sec('Dialogs', '<div class="nh-inline">' + btn('Confirm', 'dlg-demo', { arg: 'confirm' }) + btn('Rename device', 'dev-rename', { arg: DATA.devices[0].id }) + btn('Remove device', 'dev-remove', { arg: DATA.devices[0].id, kind: 'danger' }) + btn('Entity history', 'entity', { arg: pkey(DATA.devices[4].id, 'temperature') }) + btn('Add entity', 'group-add', { arg: 'Garden' }) + btn('Unsaved changes', 'dlg-demo', { arg: 'leave' }) + '</div>') +
      sec('Empty / loading / error', '<div class="nh-kit-states">' + empty('home', 'Nothing here', 'Explain and offer the next step.', btn('Primary action', '', { kind: 'primary', sm: true })) + '<div class="nh-skel-card"><span class="nh-skel"></span><span class="nh-skel"></span><span class="nh-skel is-short"></span></div>' + '<div class="nh-empty is-error">' + I('error') + '<h3>Could not load metrics</h3><p>The hub returned 429 (rate limited).</p>' + btn('Retry', 'toast', { sm: true, arg: 'Retrying…' }) + '</div></div>') +
      sec('Assistant not configured', '<div class="nh-empty">' + I('sparkles') + '<h3>Assistant is not set up</h3><p>Set the assistant service URL and enable the MCP server.</p><a class="nh-btn is-primary is-sm" href="#/settings/assistant">Open settings</a></div>') +
      sec('Colour tokens', '<div class="nh-tokens">' + ['bg', 'surface', 'surface-2', 'border', 'text', 'text-muted', 'accent', 'ok', 'warn', 'danger', 'info', 'on'].map((t) => '<span class="nh-token"><i style="background:var(--nh-' + t + ')"></i>--nh-' + t + '</span>').join('') + '</div><p class="nh-type-sample"><span class="nh-h1">Display 24</span> <b>Body bold</b> body text <code>mono 0x54ef</code> <span class="nh-muted">muted</span></p>') +
      sec('Icons', '<div class="nh-icon-grid">' + window.NH_ICON_NAMES.map((n) => '<span title="' + n + '">' + I(n) + '</span>').join('') + '</div>') +
      '</div>';
  };

  // Every page of this design on one long scroll (review aid).
  V.all = () => {
    const light = DATA.devices[2].id, pres = DATA.devices[10].id, auto = DATA.devices[11].id;
    const P = [
      ['Overview', 'overview'], ['Home (areas & tiles)', 'home'], ['Devices', 'devices'], ['Device list', 'list'],
      ['Device page · About', 'device', light, 'about'], ['Device page · Controls', 'device', light, 'controls'],
      ['Device page · Settings', 'device', pres, 'settings'], ['Device page · Metrics', 'device', pres, 'metrics'],
      ['Automations', 'automations'], ['Automation editor', 'automation', auto], ['Assistant', 'assistant'], ['Console', 'console'],
      ['Settings · Appearance', 'settings', 'appearance'], ['Settings · Device defaults', 'settings', 'devices'], ['Settings · MCP server', 'settings', 'mcp'],
      ['Components & states', 'components'], ['Login', 'login'],
    ];
    return '<div class="nh-allpages-intro">' + pageHead('All pages', 'Every screen of the ' + esc(DESIGN.name) + ' design, top to bottom. Everything stays interactive; use the navigation to open a page on its own.', '') +
      '<nav class="nh-allpages-toc">' + P.map((x, i) => '<a href="#" data-act="all-jump" data-arg="pg-' + i + '">' + (i + 1) + '. ' + esc(x[0]) + '</a>').join('') + '</nav></div>' +
      P.map((x, i) => '<section class="nh-allpage" id="pg-' + i + '"><div class="nh-allpage-t"><span>' + (i + 1) + '</span>' + esc(x[0]) + '<a href="#/' + x.slice(1).map(encodeURIComponent).join('/') + '">open alone ↗</a></div><div class="nh-allpage-b">' +
        (x[1] === 'login' ? V.login(null, true) : V[x[1]]({ name: x[1], a: x[2], b: x[3] })) + '</div></section>').join('');
  };

  /* ------------------------------------------------------------------ dialogs & toasts */
  function renderToasts() { const h = $('.nh-toast-host'); if (h) h.innerHTML = S.toasts.map(toastHtml).join(''); }
  function toast(kind, text) { const t = { id: Math.random(), kind, text }; S.toasts.push(t); renderToasts(); setTimeout(() => { S.toasts = S.toasts.filter((x) => x !== t); renderToasts(); }, 4200); }
  const toastHtml = (t) => '<div class="nh-toast is-' + t.kind + '" role="status">' + I(t.kind === 'ok' ? 'check' : t.kind === 'danger' ? 'error' : 'info') + '<span>' + esc(t.text) + '</span></div>';
  function dialogHtml() {
    const g = S.dialog; if (!g) return '';
    return '<div class="nh-dialog-backdrop" data-act="dlg-cancel"></div><div class="nh-dialog' + (g.wide ? ' is-wide' : '') + '" role="dialog" aria-modal="true" aria-labelledby="dlg-t"><div class="nh-dialog-head"><span class="nh-window-dots" aria-hidden="true"><i></i><i></i><i></i></span><h2 id="dlg-t">' + esc(g.title) + '</h2>' + btn('', 'dlg-cancel', { icon: 'close', title: 'Close' }) + '</div><div class="nh-dialog-body">' + g.body + '</div><div class="nh-dialog-foot">' + (g.foot || btn(g.cancel || 'Cancel', 'dlg-cancel') + btn(g.ok || 'OK', 'dlg-ok', { kind: g.danger ? 'danger' : 'primary' })) + '</div></div>';
  }
  function openDialog(g) { S.dialog = g; render(); setTimeout(() => { const f = $('.nh-dialog [data-autofocus]') || $('.nh-dialog .nh-dialog-foot .nh-btn:last-child'); if (f) f.focus(); }, 0); }
  function closeDialog() { S.dialog = null; render(); }

  /* ------------------------------------------------------------------ device commands (useDeviceCommand) */
  function sendCommand(id, name, value) {
    const d = dev(id), k = pkey(id, name);
    if (!canControl(d) || S.pending[k] !== undefined) return; // no duplicate sends while pending (FE-03)
    S.pending[k] = value; render();
    const fail = d.id === 'shelly-plug-hallway'; // simulated timeout
    setTimeout(() => {
      delete S.pending[k];
      if (fail) { toast('danger', 'No confirmation from ' + d.friendly_name + ' within 5 s — state unchanged (simulated timeout).'); return; }
      applyUpdate(id, { [name]: value }, true); if (name === 'state' && d.exposes.power) applyUpdate(id, { power: value === 'ON' ? 45 : 0 }, true);
      softRender();
    }, fail ? 5000 : 650 + Math.random() * 500);
  }
  const LIVE = ['overview', 'home', 'devices', 'list', 'device', 'components', 'all'];
  // Re-render for timers/live data; never while a dialog is open or an unbound input has focus.
  function softRender() {
    const a = document.activeElement;
    if (S.dialog) return;
    if (a && /^(INPUT|TEXTAREA|SELECT)$/.test(a.tagName) && !a.dataset.focusKey && !a.dataset.bind) return;
    render();
  }
  function applyUpdate(id, data, quiet) {
    const d = dev(id); if (!d) return;
    const changes = [];
    Object.keys(data).forEach((n) => { const e = d.exposes[n]; if (!e) return; if (e.data !== data[n]) changes.push({ n, from: fmt(e), to: fmt(e, data[n]) }); e.data = data[n]; });
    d.last_seen = new Date().toISOString();
    if (changes.length && S.feedOn) { S.feed.unshift({ id, name: d.friendly_name, changes, t: Date.now() }); S.feed.length = Math.min(S.feed.length, 200); }
    const r = route();
    if (quiet || LIVE.indexOf(r.name) < 0 || (r.name === 'device' && r.b === 'settings')) return;
    softRender();
  }

  /* ------------------------------------------------------------------ editor mutations */
  function getP(o, p) { return p.split('.').reduce((x, k) => (x == null ? x : x[k]), o); }
  function setP(o, p, v) { const ks = p.split('.'), last = ks.pop(), t = ks.reduce((x, k) => x[k], o); t[last] = v; }
  function parent(p) { const ks = p.split('.'); const idx = +ks.pop(); return { arr: getP(S.draft, ks.join('.')), idx }; }
  const newAction = (type) => (type === 'trigger' ? { id: '', type, exposes: [], publishMode: 'batch', delay: { unit: 'seconds', value: 0 } } : type === 'step' ? { id: '', type, steps: [], property: '', data: null } : { id: '', type, property: '' });

  /* ------------------------------------------------------------------ render */
  function render() {
    const r = route();
    applyTheme();
    const fk = document.activeElement && document.activeElement.dataset ? document.activeElement.dataset.focusKey : null;
    const selS = fk && document.activeElement.selectionStart, cont = $('.nh-content'), sc = cont ? cont.scrollTop : 0;
    const app = $('#app');
    if (!S.authed || r.name === 'login') app.innerHTML = V.login() + '<div class="nh-toast-host" aria-live="polite">' + S.toasts.map(toastHtml).join('') + '</div>';
    else app.innerHTML = shellFn((V[r.name] || V.overview)(r));
    const c2 = $('.nh-content'); if (c2) c2.scrollTop = sc;
    if (fk) { const el = $('[data-focus-key="' + fk + '"]'); if (el) { el.focus(); try { if (selS != null && el.setSelectionRange) el.setSelectionRange(selS, selS); } catch (e) { /* type=number */ } } }
  }
  function refreshDerived() {
    if (!S.draft) return;
    const v = validate(S.draft);
    S.draft.triggers.forEach((t, i) => { const n = $('[data-region="sent-' + i + '"]'); if (n) n.textContent = describeTrigger(S.draft, t); });
    const sb = $('[data-region="savebar"]'); if (sb) sb.innerHTML = savebarHtml(v);
    const ch = $('[data-region="checks"]'); if (ch) ch.innerHTML = checksHtml(v);
    const js = $('[data-region="json"]'); if (js) js.textContent = JSON.stringify(S.draft, null, 2);
  }

  /* ------------------------------------------------------------------ events */
  const A = {
    'nav-toggle': () => { S.navOpen = !S.navOpen; render(); }, 'nav-close': () => { S.navOpen = false; render(); },
    pop: () => { S.pop = !S.pop; render(); },
    'feed-toggle': () => { S.feedOn = !S.feedOn; if (!S.feedOn) S.feed = []; try { localStorage.setItem('nh-sample-feed', S.feedOn ? 'on' : 'off'); } catch (e) { /* ignore */ } render(); },
    'all-jump': (id) => { const el = document.getElementById(id); if (el) el.scrollIntoView({ behavior: 'smooth' }); },
    skip: () => { const c = $('.nh-content'); if (c) c.focus(); },
    'mode-toggle': () => { T.mode = document.documentElement.dataset.mode === 'dark' ? 'light' : 'dark'; render(); },
    'th-mode': (a) => { T.mode = a; render(); }, 'th-preset': (a) => { T.preset = a; render(); }, 'th-accent': (a) => { T.accent = a; render(); },
    'th-density': (a) => { T.density = a; render(); }, 'th-effects': () => { T.effects = !T.effects; render(); },
    'th-reset': () => { T = Object.assign({}, DESIGN.defaults); render(); },
    'conn-cycle': () => { A['conn-set'](S.conn === 'connected' ? 'connecting' : S.conn === 'connecting' ? 'disconnected' : 'connected'); },
    'conn-set': (a) => { S.conn = a; S.attempt = a === 'connecting' ? 3 : 0; render(); }, 'conn-reconnect': () => { S.conn = 'connecting'; S.attempt = 1; render(); setTimeout(() => { S.conn = 'connected'; render(); toast('ok', 'Reconnected — device state refreshed'); }, 1500); },
    permit: () => { if (S.permit) return; openDialog({ title: 'Permit join', body: '<p>Allow new <b>Zigbee</b> devices to join for <b>120 seconds</b>?</p><p class="nh-help">Applies to the Zigbee2MQTT bridge only. Wi-Fi/HTTP devices appear automatically when they first post to <code>/api/collect</code>.</p>', ok: 'Open for 120 s', onOk: () => { S.permit = 120; toast('info', 'Permit join enabled (bridgePermitJoin)'); } }); },
    'permit-stop': () => { S.permit = 0; toast('info', 'Permit join disabled'); },
    logout: () => { S.authed = false; go('#/login'); }, 'login-ok': () => { S.authed = true; S.loginError = false; go('#/overview'); },
    'tile-toggle': (a) => { const p = a.split('|'), d = dev(p[0]), e = d.exposes[p[1]]; sendCommand(p[0], p[1], e.data === e.values.on ? e.values.off : e.values.on); },
    'expose-toggle': (a) => A['tile-toggle'](a),
    'expose-preset': (a) => { const p = a.split('|'), e = dev(p[0]).exposes[p[1]]; sendCommand(p[0], p[1], e.values[p[2]]); },
    entity: (a) => { const p = a.split('|'), d = dev(p[0]), e = d.exposes[p[1]]; openDialog({ title: meta(e.name).label + ' · ' + d.friendly_name, wide: true, body: (e.type === 'numeric' ? lineChart(d, e, 48) : e.type === 'binary' ? binaryChart(d, e) : '<p>Current value: <b>' + esc(fmt(e)) + '</b></p>') + '<div class="nh-inline">' + chip('Last seen ' + timeAgo(d.last_seen), 'muted', 'clock') + protoChip(d) + '<a class="nh-btn is-sm" href="#/device/' + encodeURIComponent(d.id) + '/metrics" data-act="dlg-cancel">Open device</a></div>', foot: btn('Close', 'dlg-cancel', { kind: 'primary' }) }); },
    'edit-toggle': () => { S.edit = !S.edit; render(); },
    'group-new': () => openDialog({ title: 'New area', body: '<div class="nh-field"><label for="dlg-in">Name</label><input id="dlg-in" class="nh-input" data-autofocus placeholder="e.g. Bedroom"><span class="nh-err" hidden id="dlg-err"></span></div>', ok: 'Create', onOk: () => {
      const v = $('#dlg-in').value.trim(), G = S.cfg.hub.dashboardGroups; if (!v || G[v]) { const er = $('#dlg-err'); er.hidden = false; er.textContent = v ? 'An area with this name exists.' : 'Enter a name.'; return false; } G[v] = { name: v, deviceGroup: {} }; toast('ok', 'Area “' + v + '” created (saveDashboardGroup)'); } }),
    'group-rename': (g) => openDialog({ title: 'Rename area', body: '<div class="nh-field"><label for="dlg-in">Name</label><input id="dlg-in" class="nh-input" data-autofocus value="' + esc(g) + '"></div>', ok: 'Rename', onOk: () => { const v = $('#dlg-in').value.trim(), G = S.cfg.hub.dashboardGroups; if (!v || (G[v] && v !== g)) return false; const o = G[g]; delete G[g]; o.name = v; G[v] = o; toast('ok', 'Renamed (renameDashboardGroup)'); } }),
    'group-delete': (g) => { const n = Object.values(S.cfg.hub.dashboardGroups[g].deviceGroup).reduce((s, x) => s + x.exposes.length, 0); openDialog({ title: 'Delete area?', danger: true, body: '<p>Delete <b>' + esc(g) + '</b> and its ' + n + ' tiles from the dashboard?</p><p class="nh-help">Devices and their history are not affected. Export first if you may want it back.</p>', ok: 'Delete area', onOk: () => { delete S.cfg.hub.dashboardGroups[g]; toast('ok', 'Area deleted'); } }); },
    'group-add': (g) => openDialog({ title: 'Add to ' + g, wide: true, body: '<div class="nh-field"><label for="dlg-dev">Device</label>' + sel(S.devices.map((d) => ({ v: d.id, t: d.friendly_name })), '', 'id="dlg-dev" data-autofocus data-bind="dlg-dev"', 'Choose a device…') + '</div><div id="dlg-ex" class="nh-checks"></div>', ok: 'Add', onOk: () => {
      const id = $('#dlg-dev').value, ex = $$('#dlg-ex input:checked').map((x) => x.value); if (!id || !ex.length) return false; const G = S.cfg.hub.dashboardGroups[g]; G.deviceGroup[id] = G.deviceGroup[id] || { deviceId: id, exposes: [] }; ex.forEach((x) => G.deviceGroup[id].exposes.indexOf(x) < 0 && G.deviceGroup[id].exposes.push(x)); toast('ok', ex.length + ' added to ' + g); } }),
    'tile-remove': (a) => { const p = a.split('|'), dg = S.cfg.hub.dashboardGroups[p[0]].deviceGroup[p[1]]; dg.exposes = dg.exposes.filter((x) => x !== p[2]); if (!dg.exposes.length) delete S.cfg.hub.dashboardGroups[p[0]].deviceGroup[p[1]]; toast('ok', 'Tile removed (saveDashboardGroup)'); },
    'group-export': () => { const b = new Blob([JSON.stringify({ dashboardGroups: S.cfg.hub.dashboardGroups }, null, 2)], { type: 'application/json' }), u = URL.createObjectURL(b), l = document.createElement('a'); l.href = u; l.download = 'dashboard-groups.json'; l.click(); URL.revokeObjectURL(u); },
    'group-import': () => openDialog({ title: 'Import areas', danger: true, body: '<p>Importing replaces all current areas with the file contents.</p>', ok: 'Choose file…', onOk: () => toast('info', 'File picker opens here (importDashboardGroups)') }),
    'list-sort': (k) => { S.list.dir = S.list.sort === k ? -S.list.dir : 1; S.list.sort = k; render(); },
    'list-clear': () => { Object.assign(S.list, { q: '', proto: 'all', power: 'all', status: 'all' }); render(); },
    'dev-rename': (id) => { const d = dev(id); openDialog({ title: 'Rename device', body: '<div class="nh-field"><label for="dlg-in">Friendly name</label><input id="dlg-in" class="nh-input" data-autofocus value="' + esc(d.friendly_name) + '"><span class="nh-help">Also renames the Zigbee2MQTT topic. Automations keep working (keyed by IEEE address).</span></div>', ok: 'Rename', onOk: () => { const v = $('#dlg-in').value.trim(); if (!v) return false; d.friendly_name = v; toast('ok', 'Rename requested (deviceRename) — waiting for bridge'); } }); },
    'dev-interview': (id) => openDialog({ title: 'Interview device', body: '<p>Re-read capabilities of <b>' + esc(dev(id).friendly_name) + '</b>? Battery devices must be awake (press a button).</p>', ok: 'Interview', onOk: () => toast('info', 'Interview started (deviceInterview)') }),
    'dev-remove': (id) => openDialog({ title: 'Remove device', danger: true, body: '<p>Remove <b>' + esc(dev(id).friendly_name) + '</b> from the Zigbee network?</p><label class="nh-check"><input type="checkbox" id="rm-force"> Force remove <small>(device unreachable)</small></label><label class="nh-check"><input type="checkbox" id="rm-block"> Block from re-joining</label><p class="nh-help">History is kept. Automations referencing it will show a missing-device error.</p>', ok: 'Remove', onOk: () => toast('ok', 'Remove requested (deviceRemove force=' + $('#rm-force').checked + ' block=' + $('#rm-block').checked + ')') }),
    copy: (v) => { try { navigator.clipboard.writeText(v); } catch (e) { /* ignore */ } toast('ok', 'Copied'); },
    'cfg-override': (id) => { S.cfg.hub.devices.overrides[id] = Object.assign(clone(S.cfg.hub.devices.defaults), { id }); toast('ok', 'Override created (saveDeviceConfigOverride)'); },
    'cfg-revert': (id) => openDialog({ title: 'Revert to defaults?', body: '<p>Remove this device\'s own settings and use hub defaults?</p>', ok: 'Revert', onOk: () => { delete S.cfg.hub.devices.overrides[id]; toast('ok', 'Override removed (deleteDeviceConfigOverride)'); } }),
    'cfg-toggle': (a) => { const p = a.split('|'), o = S.cfg.hub.devices.overrides[p[0]]; if (o) { o[p[1]] = !o[p[1]]; render(); } },
    period: (a) => { S.period = a; render(); },
    'auto-create': () => openDialog({ title: 'New automation', body: '<div class="nh-field"><label for="dlg-dev">Source device</label>' + sel(S.devices.filter((d) => !S.automations.some((a) => a.id === d.id)).map((d) => ({ v: d.id, t: d.friendly_name })), '', 'id="dlg-dev" data-autofocus', 'Which device should trigger it?') + '<span class="nh-help">Automations belong to the device whose changes trigger them.</span></div>', ok: 'Create', onOk: () => {
      const id = $('#dlg-dev').value; if (!id) return false; const d = dev(id); S.automations.push({ id, friendlyname: d.friendly_name, type: 'device', description: '', enabled: true, schedules: [], triggers: [] });
      S.original = clone(S.automations[S.automations.length - 1]); S.draft = clone(S.original); S.draft.triggers.push({ name: '', type: 'deviceTrigger', conditions: [], actions: [] }); setTimeout(() => go('#/automation/' + encodeURIComponent(id)), 0); } }),
    'auto-enable': (i) => { const a = S.automations[+i]; a.enabled = !a.enabled; toast('ok', a.friendlyname + (a.enabled ? ' enabled' : ' disabled') + ' (saveAutomation)'); },
    'auto-delete': (i) => { const a = S.automations[+i]; openDialog({ title: 'Delete automation?', danger: true, body: '<p>Delete <b>' + esc(a.friendlyname) + '</b> with ' + a.triggers.length + ' trigger(s) and ' + a.schedules.length + ' schedule(s)? This cannot be undone.</p>', ok: 'Delete', onOk: () => { S.automations.splice(+i, 1); toast('ok', 'Deleted (deleteAutomation)'); } }); },
    'auto-run': (a) => { const p = a.split('|'), au = S.automations[+p[0]]; runTrigger(au, +p[1]); },
    'ed-run': (i) => runTrigger(S.draft, +i),
    'ed-collapse': (i) => { S.collapsed[i] = !S.collapsed[i]; render(); },
    'ed-add-trig': () => { S.draft.triggers.push({ name: '', type: 'deviceTrigger', conditions: [], actions: [] }); render(); },
    'ed-dup': (i) => { S.draft.triggers.splice(+i + 1, 0, clone(S.draft.triggers[+i])); render(); },
    'ed-del': (p) => { const x = parent(p); const doIt = () => { x.arr.splice(x.idx, 1); render(); }; if (/^triggers\.\d+$/.test(p)) openDialog({ title: 'Delete trigger?', danger: true, body: '<p>' + esc(describeTrigger(S.draft, S.draft.triggers[x.idx])) + '</p><p class="nh-help">Takes effect when you save.</p>', ok: 'Delete', onOk: doIt }); else doIt(); },
    'ed-move': (a) => { const q = a.split('|'), x = parent(q[0]), j = x.idx + +q[1]; const t = x.arr[x.idx]; x.arr[x.idx] = x.arr[j]; x.arr[j] = t; render(); },
    'ed-add-cond': (a) => { const q = a.split('|'); getP(S.draft, q[0]).conditions.push(q[1] === 'time' ? { type: 'time', timeRange: { startAt: '', endAt: '' } } : { type: 'expose', name: '', value: null, equality: '=' }); render(); },
    'ed-add-act': (tp) => { getP(S.draft, tp).actions.push(newAction('trigger')); render(); },
    'ed-act-type': (v, el) => { const p = el.dataset.path, x = getP(S.draft, p); if (x.type === v) return; const n = newAction(v); n.id = x.id; setP(S.draft, p, n); render(); },
    'ed-add-row': (ap) => { getP(S.draft, ap).exposes.push({ name: '', data: null }); render(); },
    'ed-publish': (v, el) => { setP(S.draft, el.dataset.path, v); render(); },
    'ed-step-op': (v, el) => { const x = getP(S.draft, el.dataset.path); x.steps = [{ id: x.id, property: x.property, operator: v }]; render(); },
    'ed-trig-type': (v, el) => { setP(S.draft, el.dataset.path, v); render(); },
    'ed-enabled': () => { S.draft.enabled = !S.draft.enabled; render(); },
    'ed-sched-add': (ty) => { S.draft.schedules.push({ startAt: ty === 'enable' ? '07:00' : '23:00', type: ty }); render(); },
    'ed-sched-del': (ty) => { S.draft.schedules = S.draft.schedules.filter((s) => s.type !== ty); render(); },
    'ed-json': () => { S.jsonOpen = !S.jsonOpen; render(); },
    'ed-discard': () => openDialog({ title: 'Discard changes?', danger: true, body: '<p>Revert to the last saved version?</p>', ok: 'Discard', onOk: () => { S.draft = clone(S.original); } }),
    'ed-save': () => { const i = S.automations.findIndex((a) => a.id === S.draft.id); S.automations[i] = clone(S.draft); S.original = clone(S.draft); toast('ok', 'Saved — saveAutomation sent; waiting for automationUpdated'); },
    'chat-new': () => { const id = 'c' + Date.now(); S.convs.unshift({ id, title: 'New conversation', updated: Date.now(), messages: [] }); S.activeConv = id; render(); },
    'chat-sel': (id) => { S.activeConv = id; render(); }, 'chat-del': (id) => openDialog({ title: 'Delete conversation?', danger: true, body: '<p>This removes the conversation history from the hub.</p>', ok: 'Delete', onOk: () => { S.convs = S.convs.filter((c) => c.id !== id); if (!S.convs.length) A['chat-new'](); S.activeConv = S.convs[0].id; } }),
    'chat-suggest': (t) => sendChat(t),
    'log-pause': () => { S.consolePaused = !S.consolePaused; render(); }, 'log-clear': () => { S.logs = []; render(); }, 'log-level': (a) => { S.consoleLevel = a; render(); },
    'log-remote': () => { S.cfg.hub.logger.enableRemoteLogger = !S.cfg.hub.logger.enableRemoteLogger; render(); },
    mcp: (a) => { S.mcp.running = a !== 'stop'; toast('ok', a + 'MCP sent'); },
    toast: (a) => toast('info', a || 'Done'), 'toast-kind': (a) => { const p = a.split('|'); toast(p[0], p[1]); },
    'kit-toggle': (a, el) => { el.setAttribute('aria-checked', el.getAttribute('aria-checked') !== 'true'); }, 'kit-seg': (a, el) => { $$('button', el.parentNode).forEach((b) => b.setAttribute('aria-pressed', b === el)); },
    'dlg-demo': (a) => (a === 'leave' ? openDialog(leaveDialog(() => {})) : openDialog({ title: 'Are you sure?', body: '<p>Confirmations always state what will change and whether it can be undone.</p>', onOk: () => toast('ok', 'Confirmed') })),
    'dlg-cancel': () => closeDialog(), 'dlg-ok': () => { const g = S.dialog; if (g.onOk && g.onOk() === false) return; if (S.dialog === g) closeDialog(); else render(); },
  };
  ['proto', 'power', 'status'].forEach((k) => (A['list-filter-' + k] = (a) => { S.list[k] = a; render(); }));
  const leaveDialog = (go2) => ({ title: 'Unsaved changes', danger: true, body: '<p>You have unsaved changes to this automation. Leave and discard them?</p>', cancel: 'Stay', ok: 'Discard and leave', onOk: go2 });
  function runTrigger(a, i) {
    const k = a.id + i; if (S.running[k]) return; S.running[k] = true; render();
    setTimeout(() => { delete S.running[k]; toast('ok', 'Ran “' + (a.triggers[i].name || 'trigger') + '” — POST /api/automation/trigger 200'); }, 1200);
  }
  function sendChat(text) {
    text = (text || '').trim(); if (!text || S.typing) return;
    const c = S.convs.find((x) => x.id === S.activeConv); c.messages.push({ role: 'user', content: text }); if (c.title === 'New conversation') c.title = text.slice(0, 40);
    S.typing = true; render();
    setTimeout(() => { S.typing = false; c.messages.push({ role: 'assistant', content: 'Three devices report weak Zigbee links (< 50 LQI): Attic room temperature sensor (14), Attic room power socket (36) and Attic smoke alarm (40). Bedroom radiator valve is offline (last 8 LQI). Adding a mains-powered router in the attic would likely help.' }); render(); const m = $('[data-region="msgs"]'); if (m) m.scrollTop = m.scrollHeight; }, 1400);
  }


  /* ------------------------------------------------------------------ design layouts
     A design may replace the shell, any view and add actions (designs/<id>.js). */
  let shellFn = shell;
  const overlays = () => customiser() + '<div class="nh-toast-host" aria-live="polite">' + S.toasts.map(toastHtml).join('') + '</div>' + dialogHtml();
  const connText = () => (S.conn === 'connected' ? 'Connected' : S.conn === 'connecting' ? 'Reconnecting… (attempt ' + S.attempt + '/10)' : 'Disconnected');
  const banners = () => '<div class="nh-banners">' + (S.conn !== 'connected' ? '<div class="nh-banner is-' + (S.conn === 'connecting' ? 'warn' : 'danger') + '" role="status">' + I(S.conn === 'connecting' ? 'refresh' : 'offline') + '<span><b>' + esc(connText()) + '.</b> Live values may be stale and device controls are disabled until the hub reconnects.</span>' + (S.conn === 'disconnected' ? btn('Reconnect', 'conn-reconnect', { sm: true }) : '') + '</div>' : '') +
    (S.permit ? '<div class="nh-banner is-info" role="status">' + I('join') + '<span><b>Permit join is open</b> for Zigbee devices — <span data-permit>' + fmtTimer(S.permit) + '</span> remaining.</span>' + btn('Stop', 'permit-stop', { sm: true }) + '</div>' : '') + '</div>';
  const connBtn = () => '<button type="button" class="nh-conn is-' + S.conn + '" data-act="conn-cycle" title="Hub connection (click to simulate)"><span class="nh-dot"></span><span class="nh-conn-txt">' + esc(connText()) + '</span></button>';
  const roomsOf = () => { const G = S.cfg.hub.dashboardGroups; return Object.keys(G).map((g) => { const items = []; const ids = []; Object.values(G[g].deviceGroup).forEach((dg) => { ids.push(dg.deviceId); dg.exposes.forEach((x) => items.push([dg.deviceId, x])); }); return { name: g, items, ids }; }); };
  if (typeof window.NH_LAYOUT === 'function') {
    const ctx = { S, DATA, DESIGN, NAV, I, esc, clone, $, $$, btn, chip, toggle, seg, sel, empty, pageHead, statusChip, protoChip, signalChip, powerChip, devIcon, meta, fmt, fmtRaw, titleCase,
      timeAgo, clock, tile, tileState, shortName, dcard, exposeRow, lineChart, binaryChart, spark, series, feedHtml, describeTrigger, condText, actText, validate, dirty: () => dirty(), autoStatus, canManual,
      edParts, editorHtml, V: Object.assign({}, V), DV, A, route, go, dev, devCfg, isOnline, isDisabled, lqiOf, battOf, canControl, capability, protocol, PROTOCOLS, viewTitle, customiser, appearanceForm,
      overlays, banners, connBtn, connText, fmtTimer, pkey, isWritable, isPreset, stateExposeOf, getP, setP, toast, openDialog, sendCommand, roomsOf, render: () => render(), defaultShell: shell };
    const L = window.NH_LAYOUT(ctx) || {};
    if (L.shell) shellFn = L.shell;
    Object.assign(V, L.views || {});
    Object.assign(A, L.actions || {});
    if (L.live) L.live.forEach((x) => LIVE.indexOf(x) < 0 && LIVE.push(x));
  }
  document.addEventListener('click', (ev) => {
    const el = ev.target.closest('[data-act]'); if (!el || el.disabled) return;
    const fn = A[el.dataset.act]; if (!fn) return;
    if (el.tagName === 'A' && el.dataset.act === 'dlg-cancel') { S.dialog = null; return; }
    ev.preventDefault(); fn(el.dataset.arg, el);
  });
  document.addEventListener('submit', (ev) => {
    const f = ev.target.dataset.form; ev.preventDefault();
    if (f === 'login') { const p = $('#lp').value; if (p && p.length < 3) { S.loginError = true; render(); return; } A['login-ok'](); }
    if (f === 'chat') sendChat($('textarea', ev.target).value);
  });
  document.addEventListener('keydown', (ev) => {
    if (ev.key === 'Escape') { if (S.dialog) closeDialog(); else if (S.pop) { S.pop = false; render(); } else if (S.navOpen) { S.navOpen = false; render(); } }
    if (ev.altKey && /^[1-9]$/.test(ev.key)) { ev.preventDefault(); go('#/' + NAV[+ev.key - 1].id); }
    if (ev.key === 'Enter' && !ev.shiftKey && ev.target.matches('.nh-chat-input textarea')) { ev.preventDefault(); sendChat(ev.target.value); }
  });
  document.addEventListener('input', (ev) => {
    const t = ev.target, b = t.dataset.bind;
    if (b === 'th-radius') { T.radius = +t.value; applyTheme(); t.previousElementSibling.querySelector('output').textContent = t.value + ' px'; return; }
    if (b === 'th-font') { T.fontScale = +t.value; applyTheme(); t.previousElementSibling.querySelector('output').textContent = t.value + ' %'; return; }
    if (b === 'list-q') { S.list.q = t.value; render(); return; }
    if (b === 'log-q') { S.consoleQ = t.value; render(); return; }
    if (b === 'expose-slide') { const e = dev(t.dataset.arg.split('|')[0]).exposes[t.dataset.arg.split('|')[1]]; t.nextElementSibling.textContent = fmt(e, +t.value); return; }
    if (t.dataset.path && S.draft && t.tagName === 'INPUT' && t.type !== 'checkbox') { setP(S.draft, t.dataset.path, t.dataset.kind === 'num' ? (t.value === '' ? null : +t.value) : t.value); refreshDerived(); }
  });
  document.addEventListener('change', (ev) => {
    const t = ev.target, b = t.dataset.bind;
    if (b === 'th-accent-custom') { T.accent = t.value; render(); return; }
    if (b === 'th-radius' || b === 'th-font') { render(); return; }
    if (b === 'expose-slide') { const p = t.dataset.arg.split('|'); sendCommand(p[0], p[1], +t.value); return; }
    if (b === 'dlg-dev') { const d = dev(t.value); $('#dlg-ex').innerHTML = d ? Object.values(d.exposes).filter((e) => e.category === 'measurement').map((e) => '<label class="nh-check"><input type="checkbox" value="' + esc(e.name) + '"> ' + I(meta(e.name).icon) + esc(meta(e.name).label) + ' <small class="nh-muted">' + esc(fmt(e)) + '</small></label>').join('') : ''; return; }
    if (t.dataset.sched && S.draft) { S.draft.schedules.find((s) => s.type === t.dataset.sched).startAt = t.value; render(); return; }
    if (t.dataset.actCheck === 'ed-delay') { const x = getP(S.draft, t.dataset.path); x.delay = t.checked ? { unit: 'minutes', value: 1 } : { unit: 'seconds', value: 0 }; render(); return; }
    if (t.dataset.path && S.draft && t.tagName === 'INPUT' && t.type !== 'checkbox') { refreshDerived(); return; }
    if (t.dataset.path && S.draft) {
      let v = t.value; if (t.dataset.kind === 'json') v = v === '' ? null : JSON.parse(v); else if (t.dataset.kind === 'num') v = v === '' ? null : +v;
      setP(S.draft, t.dataset.path, v);
      if (t.dataset.reset) setP(S.draft, t.dataset.reset, null);
      if (t.dataset.resetAct) { const ap = t.dataset.path.replace(/\.id$/, ''), x = getP(S.draft, ap), n = newAction(x.type); n.id = v; if (x.delay) n.delay = x.delay; setP(S.draft, ap, n); }
      if (t.dataset.syncSteps) { const ap = t.dataset.path.replace(/\.property$/, ''), x = getP(S.draft, ap); x.steps = [{ id: x.id, property: v, operator: (x.steps[0] && x.steps[0].operator) || '+' }]; }
      if (/\.name$/.test(t.dataset.path) && /conditions\.\d+\.name$/.test(t.dataset.path)) setP(S.draft, t.dataset.path.replace(/name$/, 'equality'), '=');
      render();
    }
  });
  window.addEventListener('hashchange', () => {
    if (suppressHash) { suppressHash = false; return; }
    const was = lastHash, r = route();
    const wasEditor = /^#\/automation\//.test(was || '');
    if (wasEditor && dirty() && location.hash !== was) {
      const target = location.hash; suppressHash = true; location.hash = was;
      openDialog(leaveDialog(() => { S.draft = null; S.original = null; lastHash = target; suppressHash = false; location.hash = target; }));
      return;
    }
    if (!/^#\/automation\//.test(location.hash)) { S.draft = null; }
    lastHash = location.hash; S.navOpen = false; S.dialog = null; render();
    const c = $('.nh-content'); if (c) { c.scrollTop = 0; if (r.name !== 'login') c.focus({ preventScroll: true }); }
  });
  window.addEventListener('beforeunload', (e) => { if (dirty()) { e.preventDefault(); e.returnValue = ''; } });

  /* ------------------------------------------------------------------ simulation */
  setInterval(() => {
    if (S.conn !== 'connected') return;
    const p = dev('0xa4c138f1c2d3e408'), r = Math.random();
    const upd = { target_distance: +(p.exposes.target_distance.data + (Math.random() - 0.5) * 0.18).toFixed(2) };
    if (r > 0.5) upd.linkquality = Math.max(90, Math.min(116, p.exposes.linkquality.data + Math.round((Math.random() - 0.5) * 12)));
    applyUpdate(p.id, upd);
    if (r > 0.7) { const s = dev('0xa4c1386b2f0c11aa'); if (s.exposes.state.data === 'ON') applyUpdate(s.id, { power: Math.round(80 + Math.random() * 20) }); }
    if (r < 0.15) applyUpdate('esp32-garage-weather', { temperature: +(9 + Math.random()).toFixed(1) });
  }, 2600);
  setInterval(() => {
    if (S.consolePaused || !S.cfg.hub.logger.enableRemoteLogger) return;
    const f = S.feed[0], lv = Math.random() > 0.85 ? 'debug' : 'info';
    S.logs.push({ level: lv, source: 'device', message: f ? f.name + ' › ' + f.changes.map((c) => c.n + '=' + c.to).join(', ') : 'heartbeat', timestamp: Date.now() });
    if (S.logs.length > 400) S.logs.shift();
    if (route().name === 'console' && !S.dialog && !(document.activeElement && document.activeElement.dataset.focusKey === 'log-q')) { softRender(); const c = $('[data-region="console"]'); if (c) c.scrollTop = c.scrollHeight; }
  }, 3000);
  setInterval(() => {
    if (!S.permit) return;
    S.permit--;
    if (!S.permit) { toast('info', 'Permit join closed'); softRender(); return; }
    $$('[data-permit]').forEach((n) => (n.textContent = fmtTimer(S.permit)));
  }, 1000);
  // seed a few feed rows so the overview is not empty on first load
  [['0xa4c138f1c2d3e408', { target_distance: 1.97 }], ['0xa4c138f1c2d3e408', { linkquality: 102, target_distance: 1.95 }], ['0xa4c1386b2f0c11aa', { power: 92 }], ['0x00158d0007f3a9b4', { contact: true }]].forEach((u) => applyUpdate(u[0], u[1], true));
  if (!location.hash) history.replaceState(null, '', DESIGN.defaultRoute || '#/overview');
  lastHash = location.hash;
  render();
})();
