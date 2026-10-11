/* 6 · Hearth + Panel — Hearth is the app (all pages, phone-friendly editor); Deck's look
   becomes a full-screen panel mode (#/panel) for a wall tablet. One editor, one set of themes. */
window.NH_LAYOUT = function (c) {
  const { S, I, esc, btn, chip, empty, dev, meta, fmt, roomsOf, timeAgo, devIcon, isOnline, isDisabled, lqiOf, battOf, capability, canControl, pkey, stateExposeOf, shortName, $$ } = c;
  S.deckRoom = S.deckRoom || 0; S.panelTab = S.panelTab || 'rooms';
  c.NAV.splice(1, 0, { id: 'panel', label: 'Panel mode', icon: 'gauge' });
  const logo = () => (window.NH_LOGO ? '<img class="nh-logo-img" src="' + window.NH_LOGO + '" alt="" width="28" height="41">' : I('hub'));
  const now = () => { const d = new Date(); return { t: d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }), d: d.toLocaleDateString([], { weekday: 'long', day: 'numeric', month: 'long' }) }; };
  setInterval(() => $$('[data-clock]').forEach((n) => (n.textContent = now().t)), 15000);

  function dial(value, min, max, label, sub, kind) {
    const v = Math.max(min, Math.min(max, value)), f = (v - min) / (max - min || 1), R = 70, C = 2 * Math.PI * R, arc = C * 0.75;
    return '<figure class="hp-dial kind-' + kind + '"><svg viewBox="0 0 180 180" role="img" aria-label="' + esc(label + ', ' + sub) + '"><circle cx="90" cy="90" r="' + R + '" class="hp-track" stroke-dasharray="' + arc + ' ' + C + '" transform="rotate(135 90 90)"/><circle cx="90" cy="90" r="' + R + '" class="hp-arc" stroke-dasharray="' + arc * f + ' ' + C + '" transform="rotate(135 90 90)"/></svg><figcaption><b>' + esc(label) + '</b><span>' + sub + '</span></figcaption></figure>';
  }
  function bigButton(d, e, label) {
    const st = capability(e) === 'switch' ? e : stateExposeOf(d), k = st && pkey(d.id, st.name), pend = k && S.pending[k] !== undefined;
    const on = st && st.data === st.values.on && isOnline(d) && !isDisabled(d), ctl = st && canControl(d), br = d.exposes.brightness;
    const ring = br && on ? '<svg class="hp-ring" viewBox="0 0 100 100" aria-hidden="true"><circle cx="50" cy="50" r="44" pathLength="100" stroke-dasharray="' + Math.round((br.data / 254) * 100) + ' 100"/></svg>' : '';
    const txt = !isOnline(d) ? 'Offline' : isDisabled(d) ? 'Disabled' : pend ? 'Sending…' : st ? (on ? 'On' : 'Off') + (br && on ? ', ' + fmt(br) : '') : fmt(e);
    return (ctl ? '<button type="button" class="hp-big kind-' + meta(e.name).kind + (on ? ' is-on' : '') + (pend ? ' is-pending' : '') + '" data-act="tile-toggle" data-arg="' + esc(k) + '" aria-pressed="' + !!on + '">' : '<a class="hp-big kind-' + meta(e.name).kind + (isOnline(d) ? '' : ' is-off') + '" href="#/device/' + encodeURIComponent(d.id) + '">') +
      '<span class="hp-big-ic">' + ring + I(st ? (br ? 'bulb' : devIcon(d)) : meta(e.name).icon) + '</span><b>' + esc(label) + '</b><span>' + esc(txt) + '</span>' + (ctl ? '</button>' : '</a>');
  }
  function roomsPage() {
    const rooms = roomsOf(); if (!rooms.length) return empty('home', 'No rooms yet', 'Rooms come from your Home areas.');
    const i = ((S.deckRoom % rooms.length) + rooms.length) % rooms.length, r = rooms[i];
    const t = r.items.find((x) => /^(temperature|local_temperature)$/.test(x[1])), h = r.items.find((x) => x[1] === 'humidity');
    const td = t && dev(t[0]), hd = h && dev(h[0]);
    const climate = td ? (isOnline(td) ? dial(td.exposes[t[1]].data, 5, 35, fmt(td.exposes[t[1]]), hd && isOnline(hd) ? esc(fmt(hd.exposes.humidity)) + ' humidity' : 'temperature', 'temp') : '<div class="hp-dial is-off">' + I('offline') + '<b>Offline</b><span>' + esc(td.friendly_name) + '</span></div>') : '';
    const seen = {}, buttons = [];
    r.items.forEach((it) => { if (t && it[0] === t[0] && it[1] === t[1]) return; if (h && it[0] === h[0] && it[1] === 'humidity') return; const d = dev(it[0]), e = d && d.exposes[it[1]]; if (!e) return; const st = stateExposeOf(d), key = st ? d.id : it.join(); if (seen[key]) return; seen[key] = 1; buttons.push(bigButton(d, e, st ? shortName(d, r.name) : e.description || meta(it[1]).label)); });
    return '<div class="hp-pager" role="tablist" aria-label="Rooms">' + btn('', 'dk-room', { icon: 'chevronL', title: 'Previous room', arg: i - 1 }) + '<div class="hp-pager-list">' + rooms.map((x, k) => '<button type="button" role="tab" aria-selected="' + (k === i) + '" data-act="dk-room" data-arg="' + k + '">' + esc(x.name) + '</button>').join('') + '</div>' + btn('', 'dk-room', { icon: 'chevronR', title: 'Next room', arg: i + 1 }) + '</div>' +
      '<section class="hp-room' + (climate ? '' : ' no-climate') + '" aria-label="' + esc(r.name) + '">' + (climate ? '<div class="hp-climate">' + climate + '</div>' : '') + '<div class="hp-buttons">' + buttons.join('') + '</div></section>';
  }
  function statusPage() {
    const ds = S.devices, on = ds.filter(isOnline).length, z = ds.map(lqiOf).filter((x) => x != null), bat = ds.filter((d) => battOf(d) != null), ok = bat.filter((d) => battOf(d) >= 20).length, auto = S.automations;
    const attn = ds.filter((d) => !isOnline(d) || (lqiOf(d) != null && lqiOf(d) < 50) || (battOf(d) != null && battOf(d) < 20));
    return '<div class="hp-gauges">' + dial(on, 0, ds.length, on + '/' + ds.length, 'devices online', 'ok') + dial(Math.round(z.reduce((a, b) => a + b, 0) / z.length), 0, 255, String(Math.round(z.reduce((a, b) => a + b, 0) / z.length)), 'average link quality', 'signal') + dial(ok, 0, bat.length, ok + '/' + bat.length, 'batteries healthy', 'battery') + dial(auto.filter((a) => a.enabled).length, 0, auto.length, auto.filter((a) => a.enabled).length + '/' + auto.length, 'rules enabled', 'switch') + '</div>' +
      '<h2 class="hp-h2">Needs a look</h2><div class="hp-rows">' + attn.map((d) => '<a class="hp-row" href="#/device/' + encodeURIComponent(d.id) + '">' + I(devIcon(d)) + '<b>' + esc(d.friendly_name) + '</b><span>' + esc(!isOnline(d) ? 'Offline since ' + timeAgo(d.last_seen) : battOf(d) != null && battOf(d) < 20 ? 'Battery ' + Math.round(battOf(d)) + ' %' : 'Weak signal, ' + lqiOf(d) + ' LQI') + '</span></a>').join('') + '</div>';
  }
  function panelShell() {
    const g = dev('0x00158d000611e5f5'), n = now();
    return '<div class="hp-panel"><header class="hp-strip"><a class="hp-logo" href="#/home" aria-label="Exit panel mode">' + logo() + '</a><div class="hp-clock"><b data-clock>' + n.t + '</b><span>' + esc(n.d) + '</span></div>' +
      '<div class="hp-out">' + I('thermometer') + '<span>Garden <b>' + esc(g && isOnline(g) ? fmt(g.exposes.temperature) : '—') + '</b></span></div><span class="hp-gap"></span>' + c.connBtn() +
      btn('', 'mode-toggle', { icon: document.documentElement.dataset.mode === 'dark' ? 'sun' : 'moon', title: 'Light or dark' }) + btn('', 'pop', { icon: 'palette', title: 'Appearance' }) + '</header>' + c.banners() +
      '<main class="nh-content hp-content" id="nh-content" tabindex="-1">' + (S.panelTab === 'status' ? statusPage() : roomsPage()) + '</main>' +
      '<nav class="hp-dock" aria-label="Panel">' + [['rooms', 'Rooms', 'home'], ['status', 'Status', 'gauge']].map((x) => '<button type="button" data-act="hp-tab" data-arg="' + x[0] + '" aria-pressed="' + (S.panelTab === x[0]) + '">' + I(x[2]) + '<span>' + x[1] + '</span></button>').join('') +
      '<a href="#/home">' + I('logout') + '<span>Exit panel</span></a></nav></div>' + c.overlays();
  }
  const views = {
    panel: () => '',
    all: (r) => c.V.all(r) + '<section class="nh-allpage" id="pg-panel"><div class="nh-allpage-t"><span>18</span>Panel mode (wall tablet)<a href="#/panel">open alone ↗</a></div><div class="nh-allpage-b"><div class="hp-panel is-embed"><header class="hp-strip"><span class="hp-logo">' + logo() + '</span><div class="hp-clock"><b data-clock>' + now().t + '</b><span>' + esc(now().d) + '</span></div></header><div class="hp-content">' + roomsPage() + '<hr class="hp-sep">' + statusPage() + '</div></div></div></section>',
    home: (r) => c.V.home(r).replace('<div class="nh-page-actions">', '<div class="nh-page-actions"><a class="nh-btn" href="#/panel">' + I('gauge') + '<span>Panel mode</span></a>'),
  };
  const shell = (view) => (c.route().name === 'panel' ? panelShell() : c.defaultShell(view));
  const actions = {
    'dk-room': (k) => { S.deckRoom = +k; c.render(); },
    'hp-tab': (t) => { S.panelTab = t; c.render(); },
  };
  document.addEventListener('keydown', (ev) => { if (c.route().name === 'panel' && S.panelTab === 'rooms' && !S.dialog && /^Arrow(Left|Right)$/.test(ev.key) && !/INPUT|SELECT|TEXTAREA/.test(document.activeElement.tagName)) { S.deckRoom += ev.key === 'ArrowLeft' ? -1 : 1; c.render(); } });
  return { shell, views, actions, live: ['panel'] };
};
