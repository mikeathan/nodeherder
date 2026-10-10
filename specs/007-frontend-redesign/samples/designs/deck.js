/* 5 · Deck — wall-tablet panel: clock strip, one room per page, big dials and buttons, node-canvas editor. */
window.NH_LAYOUT = function (c) {
  const { S, I, esc, btn, chip, toggle, empty, dev, meta, fmt, roomsOf, timeAgo, devIcon, isOnline, isDisabled, lqiOf, battOf, protocol, capability, canControl, pkey, stateExposeOf, describeTrigger, autoStatus, canManual, condText, actText, $$ } = c;
  S.deckRoom = S.deckRoom || 0; S.deckSel = S.deckSel || null; S.deckTrig = S.deckTrig || 0;
  const DOCK = [['home', 'Rooms', 'home'], ['overview', 'Status', 'gauge'], ['devices', 'Devices', 'devices'], ['list', 'List', 'list'], ['automations', 'Rules', 'automation'], ['assistant', 'Ask', 'sparkles'], ['console', 'Log', 'terminal'], ['settings', 'Settings', 'cog'], ['components', 'Parts', 'components'], ['all', 'All', 'stack']];
  const now = () => { const d = new Date(); return { t: d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }), d: d.toLocaleDateString([], { weekday: 'long', day: 'numeric', month: 'long' }) }; };
  setInterval(() => $$('[data-clock]').forEach((n) => (n.textContent = now().t)), 15000);

  // signature element: 270° dial
  function dial(value, min, max, label, sub, kind) {
    const v = Math.max(min, Math.min(max, value)), f = (v - min) / (max - min), R = 70, C = 2 * Math.PI * R, arc = C * 0.75;
    return '<figure class="dk-dial kind-' + kind + '"><svg viewBox="0 0 180 180" role="img" aria-label="' + esc(label + ' ' + sub) + '"><circle cx="90" cy="90" r="' + R + '" class="dk-track" stroke-dasharray="' + arc + ' ' + C + '" transform="rotate(135 90 90)"/><circle cx="90" cy="90" r="' + R + '" class="dk-arc" stroke-dasharray="' + arc * f + ' ' + C + '" transform="rotate(135 90 90)"/></svg><figcaption><b>' + esc(label) + '</b><span>' + sub + '</span></figcaption></figure>';
  }
  function shell(view) {
    const r = c.route(), cur = r.name === 'device' ? 'devices' : r.name === 'automation' ? 'automations' : r.name, g = c.dev('0x00158d000611e5f5'), n = now();
    return '<button type="button" class="nh-skip" data-act="skip">Skip to content</button><div class="dk-app"><header class="dk-strip"><div class="dk-clock"><b data-clock>' + n.t + '</b><span>' + esc(n.d) + '</span></div>' +
      '<div class="dk-out">' + I('thermometer') + '<span>Garden <b>' + esc(g && isOnline(g) ? fmt(g.exposes.temperature) : '—') + '</b></span></div><span class="dk-gap"></span>' +
      (S.permit ? '<button type="button" class="dk-pill is-hot" data-act="permit-stop">' + I('join') + 'Joining <span data-permit>' + c.fmtTimer(S.permit) + '</span></button>' : '<button type="button" class="dk-pill" data-act="permit">' + I('join') + 'Add device</button>') +
      c.connBtn() + btn('', 'mode-toggle', { icon: document.documentElement.dataset.mode === 'dark' ? 'sun' : 'moon', title: 'Light or dark' }) + btn('', 'pop', { icon: 'palette', title: 'Appearance' }) + btn('', 'logout', { icon: 'logout', title: 'Sign out' }) + '</header>' +
      c.banners() + '<main class="nh-content dk-content" id="nh-content" tabindex="-1">' + view + '</main>' +
      '<nav class="dk-dock" aria-label="Main">' + DOCK.map((d) => '<a href="#/' + d[0] + '"' + (cur === d[0] ? ' aria-current="page"' : '') + '>' + I(d[2]) + '<span>' + d[1] + '</span></a>').join('') + '</nav></div>' + c.overlays();
  }
  function bigButton(d, e, label) {
    const st = capability(e) === 'switch' ? e : stateExposeOf(d), k = st && pkey(d.id, st.name), pend = k && S.pending[k] !== undefined;
    const on = st && (st.data === st.values.on) && isOnline(d) && !isDisabled(d), ctl = st && canControl(d);
    const br = d.exposes.brightness, ring = br && on ? '<svg class="dk-ring" viewBox="0 0 100 100" aria-hidden="true"><circle cx="50" cy="50" r="44" pathLength="100" stroke-dasharray="' + Math.round((br.data / 254) * 100) + ' 100"/></svg>' : '';
    const stateTxt = !isOnline(d) ? 'Offline' : isDisabled(d) ? 'Disabled' : pend ? 'Sending…' : st ? (on ? 'On' : 'Off') + (br && on ? ', ' + fmt(br) : '') : fmt(e);
    return (ctl ? '<button type="button" class="dk-big kind-' + meta(e.name).kind + (on ? ' is-on' : '') + (pend ? ' is-pending' : '') + '" data-act="tile-toggle" data-arg="' + esc(k) + '" aria-pressed="' + !!on + '">' : '<a class="dk-big kind-' + meta(e.name).kind + (isOnline(d) ? '' : ' is-off') + '" href="#/device/' + encodeURIComponent(d.id) + '">') +
      '<span class="dk-big-ic">' + ring + I(st ? (br ? 'bulb' : devIcon(d)) : meta(e.name).icon) + '</span><b>' + esc(label) + '</b><span>' + esc(stateTxt) + '</span>' + (ctl ? '</button>' : '</a>');
  }
  const views = {};
  views.home = () => {
    const rooms = roomsOf(); if (!rooms.length) return empty('home', 'No rooms yet', 'Rooms come from dashboard groups.');
    const i = ((S.deckRoom % rooms.length) + rooms.length) % rooms.length, r = rooms[i];
    const t = r.items.find((x) => x[1] === 'temperature' || x[1] === 'local_temperature' || x[1] === 'current_heating_setpoint'), h = r.items.find((x) => x[1] === 'humidity');
    const td = t && dev(t[0]), hd = h && dev(h[0]);
    const climate = td ? (isOnline(td) ? dial(td.exposes[t[1]].data, 5, 35, fmt(td.exposes[t[1]]), hd && isOnline(hd) ? esc(fmt(hd.exposes.humidity)) + ' humidity' : esc(meta(t[1]).label), 'temp') : '<div class="dk-dial is-off">' + I('offline') + '<b>Offline</b><span>' + esc(td.friendly_name) + '</span></div>') : '';
    const seen = {}, buttons = [];
    r.items.forEach((it) => { if (t && it[0] === t[0] && it[1] === t[1]) return; if (h && it[0] === h[0] && it[1] === 'humidity') return; const d = dev(it[0]), e = d && d.exposes[it[1]]; if (!e) return; const st = stateExposeOf(d); const key = st ? d.id : it.join(); if (seen[key]) return; seen[key] = 1; buttons.push(bigButton(d, e, st ? c.shortName(d, r.name) : e.description || meta(it[1]).label)); });
    return '<div class="dk-pager" role="tablist" aria-label="Rooms">' + btn('', 'dk-room', { icon: 'chevronL', title: 'Previous room', arg: i - 1 }) + rooms.map((x, k) => '<button type="button" role="tab" aria-selected="' + (k === i) + '" data-act="dk-room" data-arg="' + k + '">' + esc(x.name) + '</button>').join('') + btn('', 'dk-room', { icon: 'chevronR', title: 'Next room', arg: i + 1 }) + '</div>' +
      '<section class="dk-room' + (climate ? '' : ' no-climate') + '" aria-label="' + esc(r.name) + '">' + (climate ? '<div class="dk-climate">' + climate + '</div>' : '') + '<div class="dk-buttons">' + buttons.join('') + '</div></section>';
  };
  views.overview = () => {
    const ds = S.devices, on = ds.filter(isOnline).length, z = ds.map(lqiOf).filter((x) => x != null), bat = ds.filter((d) => battOf(d) != null), auto = S.automations;
    const avg = Math.round(z.reduce((a, b) => a + b, 0) / z.length);
    const attn = ds.filter((d) => !isOnline(d) || (lqiOf(d) != null && lqiOf(d) < 50) || (battOf(d) != null && battOf(d) < 20));
    return '<div class="dk-gauges">' + dial(on, 0, ds.length, on + '/' + ds.length, 'devices online', 'ok') + dial(avg, 0, 255, String(avg), 'average link quality', 'signal') + dial(bat.filter((d) => battOf(d) >= 20).length, 0, bat.length, bat.filter((d) => battOf(d) >= 20).length + '/' + bat.length, 'batteries healthy', 'battery') + dial(auto.filter((a) => a.enabled).length, 0, auto.length, auto.filter((a) => a.enabled).length + '/' + auto.length, 'rules enabled', 'switch') + '</div>' +
      '<h2 class="dk-h2">Needs a look</h2><div class="dk-rows">' + attn.map((d) => '<a class="dk-row" href="#/device/' + encodeURIComponent(d.id) + '">' + I(devIcon(d)) + '<b>' + esc(d.friendly_name) + '</b><span>' + esc(!isOnline(d) ? 'Offline since ' + timeAgo(d.last_seen) : battOf(d) != null && battOf(d) < 20 ? 'Battery ' + Math.round(battOf(d)) + ' %' : 'Weak signal, ' + lqiOf(d) + ' LQI') + '</span>' + I('chevronR') + '</a>').join('') + '</div>';
  };
  views.devices = () => '<div class="dk-grid">' + S.devices.slice().sort((a, b) => a.friendly_name.localeCompare(b.friendly_name)).map((d) => { const st = stateExposeOf(d), m = Object.values(d.exposes).find((e) => e.category === 'measurement' && capability(e) !== 'switch');
    return '<a class="dk-sq' + (isOnline(d) ? '' : ' is-off') + (st && st.data === st.values.on && isOnline(d) ? ' is-on' : '') + '" href="#/device/' + encodeURIComponent(d.id) + '">' + I(devIcon(d)) + '<b>' + esc(d.friendly_name) + '</b><span>' + esc(!isOnline(d) ? 'Offline' : isDisabled(d) ? 'Disabled' : st ? (st.data === st.values.on ? 'On' : 'Off') : m ? fmt(m) : protocol(d.connection_type).label) + '</span></a>'; }).join('') + '</div>';
  views.list = () => { const ds = S.devices.slice().sort((a, b) => a.friendly_name.localeCompare(b.friendly_name, undefined, { sensitivity: 'base' })); let last = '';
    return '<div class="dk-rows">' + ds.map((d) => { const L = d.friendly_name[0].toUpperCase(), head = L !== last ? '<h2 class="dk-letter">' + L + '</h2>' : ''; last = L; return head + '<a class="dk-row" href="#/device/' + encodeURIComponent(d.id) + '">' + I(devIcon(d)) + '<b>' + esc(d.friendly_name) + '</b><span>' + esc(protocol(d.connection_type).label + ', ' + (lqiOf(d) != null ? lqiOf(d) + ' LQI' : d.exposes.rssi ? d.exposes.rssi.data + ' dBm' : '') + (battOf(d) != null ? ', battery ' + Math.round(battOf(d)) + ' %' : '') + ', ' + (isOnline(d) ? timeAgo(d.last_seen) : 'offline')) + '</span>' + I('chevronR') + '</a>'; }).join('') + '</div>'; };
  views.device = (r) => {
    const d = dev(r.a); if (!d) return empty('question', 'Device not found', '');
    const st = stateExposeOf(d), br = d.exposes.brightness, sup = protocol(d.connection_type).supports, k = st && pkey(d.id, st.name), pend = k && S.pending[k] !== undefined, ctl = canControl(d);
    const reads = Object.values(d.exposes).filter((e) => e.category === 'measurement' && capability(e) !== 'switch' && capability(e) !== 'slider');
    return '<div class="dk-dev"><header class="dk-dev-head"><a class="nh-btn is-icon" href="#/devices" aria-label="Back">' + I('chevronL') + '</a><div><h1>' + esc(d.friendly_name) + '</h1><p>' + esc(d.description) + ' — ' + esc(protocol(d.connection_type).label) + (isOnline(d) ? ', ' + esc(timeAgo(d.last_seen)) : ', offline') + '</p></div></header>' +
      '<div class="dk-dev-ctl">' + (st ? '<button type="button" class="dk-power' + (st.data === st.values.on ? ' is-on' : '') + (pend ? ' is-pending' : '') + '" data-act="tile-toggle" data-arg="' + esc(k) + '" aria-pressed="' + (st.data === st.values.on) + '"' + (ctl && !pend ? '' : ' disabled') + '>' + I('power') + '<span>' + (pend ? 'Sending…' : st.data === st.values.on ? 'On' : 'Off') + '</span></button>' : '') +
      Object.values(d.exposes).filter((e) => capability(e) === 'slider').map((e) => { const kk = pkey(d.id, e.name), p = S.pending[kk], v = p !== undefined ? p : e.data; return '<label class="dk-vslider kind-' + meta(e.name).kind + '"><span class="dk-vs-val">' + esc(fmt(e, v)) + '</span><input type="range" min="' + e.attributes.value_min + '" max="' + e.attributes.value_max + '" value="' + v + '" data-bind="expose-slide" data-arg="' + esc(kk) + '"' + (ctl ? '' : ' disabled') + ' aria-label="' + esc(meta(e.name).label) + '"><span>' + esc(meta(e.name).label) + '</span></label>'; }).join('') +
      '<div class="dk-reads">' + reads.map((e) => '<div class="dk-read kind-' + meta(e.name).kind + '">' + I(meta(e.name).icon) + '<b>' + esc(fmt(e)) + '</b><span>' + esc(e.description || meta(e.name).label) + '</span></div>').join('') + '</div></div>' +
      '<div class="dk-dev-more"><section><h2 class="dk-h2">History</h2>' + c.DV.metrics(d) + '</section><section><h2 class="dk-h2">Settings</h2>' + c.DV.settings(d) + '</section><section><h2 class="dk-h2">Device</h2>' + c.DV.about(d) + '<p class="dk-actions">' + btn('Rename', 'dev-rename', { arg: d.id, disabled: !sup.rename }) + btn('Interview', 'dev-interview', { arg: d.id, disabled: !sup.interview }) + btn('Remove', 'dev-remove', { arg: d.id, kind: 'danger', disabled: !sup.remove }) + '</p></section></div></div>';
  };
  views.automations = () => '<div class="dk-rules">' + S.automations.map((a, i) => { const src = dev(a.id), st = autoStatus(a), m = a.triggers.findIndex(canManual);
    return '<article class="dk-rule' + (a.enabled ? ' is-on' : '') + '"><a href="#/automation/' + encodeURIComponent(a.id) + '" class="dk-rule-main">' + I(src ? devIcon(src) : 'automation') + '<b>' + esc(a.description || a.friendlyname) + '</b><span>' + esc(describeTrigger(a, a.triggers[0])) + '</span></a><div class="dk-rule-foot">' + chip(st.text, st.kind, st.icon, st.title) + (m >= 0 ? btn('Run', 'auto-run', { icon: 'play', arg: i + '|' + m }) : '') + toggle(a.enabled, 'auto-enable', i, { label: 'Enabled' }) + '</div></article>'; }).join('') + '<button type="button" class="dk-rule is-new" data-act="auto-create">' + I('plus') + '<b>New rule</b></button></div>';

  // Node canvas: When → If nodes → Then nodes, wired with SVG; tap a node to edit it in the drawer.
  const NH = 88, GAP = 18;
  views.automation = (r) => {
    c.V.automation(r);
    if (!S.draft) return empty('question', 'Rule not found', '');
    const P = c.edParts(), a = P.a, n = a.triggers.length;
    if (S.deckTrig >= n) S.deckTrig = Math.max(0, n - 1);
    const i = S.deckTrig, t = a.triggers[i];
    let canvas = '';
    if (t) {
      const errs = P.trigErrs(i), cs = t.conditions, as = t.actions, rows = Math.max(1, cs.length + 1, as.length + 1), H = rows * (NH + GAP);
      const y = (k) => k * (NH + GAP) + NH / 2;
      const node = (key, cls, title, text, bad, col, k) => '<button type="button" class="dk-node ' + cls + (S.deckSel === key ? ' is-sel' : '') + (bad ? ' has-err' : '') + '" style="grid-column:' + col + ';grid-row:' + (k + 1) + '" data-act="dk-sel" data-arg="' + key + '"><b>' + title + '</b><span>' + esc(text) + '</span></button>';
      const ifN = cs.map((x, j) => node('c' + j, 'is-if', x.type === 'time' ? 'Time window' : 'Only if', condText(a, x), errs.some((e) => e.p.indexOf('.conditions.' + j) > 0), 3, j)).join('') + '<button type="button" class="dk-add" style="grid-column:3;grid-row:' + (cs.length + 1) + '" data-act="ed-add-cond" data-arg="' + P.tp(i) + '|expose">' + I('plus') + 'Condition</button>';
      const thenN = as.map((x, k) => node('a' + k, 'is-then', (k + 1) + '. ' + { trigger: 'Set', step: 'Step', preset: 'Cycle' }[x.type], actText(x), errs.some((e) => e.p.indexOf('.actions.' + k) > 0), 5, k)).join('') + '<button type="button" class="dk-add" style="grid-column:5;grid-row:' + (as.length + 1) + '" data-act="ed-add-act" data-arg="' + P.tp(i) + '">' + I('plus') + 'Action</button>';
      const wire = (fromYs, toYs) => '<svg class="dk-wires" viewBox="0 0 60 ' + H + '" preserveAspectRatio="none" aria-hidden="true" style="height:' + H + 'px">' + fromYs.flatMap((fy) => toYs.map((ty) => '<path d="M0 ' + fy + ' C 30 ' + fy + ', 30 ' + ty + ', 60 ' + ty + '"/>')).join('') + '</svg>';
      const ifYs = cs.length ? cs.map((x, j) => y(j)) : [y(0)], thenYs = as.length ? as.map((x, k) => y(k)) : [y(0)];
      canvas = '<div class="dk-canvas" style="grid-template-rows:repeat(' + rows + ',' + NH + 'px);row-gap:' + GAP + 'px">' +
        node('when', 'is-when', t.type === 'manualTrigger' ? 'Run by hand' : 'When', (t.name ? meta(t.name).label : 'choose…') + ' changes', errs.some((e) => /\.name$/.test(e.p) && e.p.split('.').length === 3), 1, 0) +
        '<div class="dk-wirecol" style="grid-column:2;grid-row:1 / span ' + rows + '">' + wire([y(0)], cs.length ? ifYs : [y(0)]) + '</div>' +
        (cs.length ? ifN : '<div class="dk-pass" style="grid-column:3;grid-row:1">always</div>' + '<button type="button" class="dk-add" style="grid-column:3;grid-row:2" data-act="ed-add-cond" data-arg="' + P.tp(i) + '|expose">' + I('plus') + 'Condition</button>') +
        '<div class="dk-wirecol" style="grid-column:4;grid-row:1 / span ' + rows + '">' + wire(cs.length ? [ifYs[ifYs.length - 1]] : [y(0)], thenYs) + '</div>' + thenN + '</div>';
    }
    const k = S.deckSel, m = k && /^(when|c(\d+)|a(\d+)|details|schedule)$/.exec(k);
    let drawer = '';
    if (m && (t || k === 'details' || k === 'schedule')) {
      const body = k === 'when' ? P.whenBody(i) : k === 'details' ? P.details() : k === 'schedule' ? P.schedule() : m[2] != null && t.conditions[+m[2]] ? P.condRow(i, +m[2]) : m[3] != null && t.actions[+m[3]] ? P.actRow(i, +m[3]) : '';
      if (body) drawer = '<aside class="dk-drawer" aria-label="Edit node"><div class="dk-drawer-head"><h2>' + esc(k === 'when' ? 'When' : k === 'details' ? 'Details' : k === 'schedule' ? 'Schedule' : m[2] != null ? 'Condition' : 'Action') + '</h2>' + btn('', 'dk-sel', { icon: 'close', title: 'Close', arg: '' }) + '</div>' + body + (t && !/details|schedule/.test(k) ? P.warnList(i) : '') + '</aside>';
      else S.deckSel = null;
    }
    return '<div class="dk-ed' + (drawer ? ' has-drawer' : '') + '"><header class="dk-ed-head">' + P.back() + '<h1>' + esc(a.description || a.friendlyname) + '</h1><div class="dk-ed-tools">' + btn('Details', 'dk-sel', { icon: 'info', arg: 'details' }) + btn('Schedule', 'dk-sel', { icon: 'calendar', arg: 'schedule' }) + toggle(a.enabled, 'ed-enabled', '', { label: 'Enabled' }) + '</div></header>' +
      '<div class="dk-trigs" role="tablist">' + a.triggers.map((x, kk) => '<button type="button" role="tab" aria-selected="' + (kk === i) + '" data-act="dk-trig" data-arg="' + kk + '">Trigger ' + (kk + 1) + (P.trigErrs(kk).length ? ' ' + I('error') : '') + '</button>').join('') + btn('Add trigger', 'ed-add-trig', { icon: 'plus' }) + (t ? '<span class="dk-gap"></span>' + P.trigTools(i) : '') + '</div>' +
      (t ? P.sentence(i) : '') + '<div class="dk-ed-body">' + (canvas || empty('automation', 'No triggers', 'Add a trigger to start.')) + drawer + '</div>' + P.checks() + P.savebar() + P.json() + '</div>';
  };
  const actions = {
    'dk-room': (k) => { S.deckRoom = +k; c.render(); },
    'dk-sel': (k) => { S.deckSel = k || null; c.render(); },
    'dk-trig': (k) => { S.deckTrig = +k; S.deckSel = null; c.render(); },
  };
  document.addEventListener('keydown', (ev) => { if (c.route().name === 'home' && !S.dialog && /^Arrow(Left|Right)$/.test(ev.key) && !/INPUT|SELECT|TEXTAREA/.test(document.activeElement.tagName)) { S.deckRoom += ev.key === 'ArrowLeft' ? -1 : 1; c.render(); } });
  return { shell, views, actions };
};
