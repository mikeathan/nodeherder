/* 2 · Floorplan — the house plan is the home page; rooms open an inspector; wizard editor. */
window.NH_LAYOUT = function (c) {
  const { S, I, esc, btn, chip, toggle, seg, empty, dev, meta, fmt, tile, exposeRow, roomsOf, timeAgo, statusChip, protoChip, signalChip, powerChip, devIcon, isOnline, isDisabled, lqiOf, battOf, protocol, describeTrigger, autoStatus, canManual } = c;
  // Plan geometry: grid areas on a 12 × 9 plan. Garden sits outside the walls.
  const PLAN = { 'Attic room': 'attic', Bedroom: 'bed', 'Living room': 'living', Kitchen: 'kitchen', Garage: 'garage', Garden: 'garden' };
  const TABS = [['home', 'Plan', 'home'], ['overview', 'Health', 'overview'], ['devices', 'Devices', 'devices'], ['list', 'Find', 'search'], ['automations', 'Rules', 'automation'], ['assistant', 'Ask', 'sparkles'], ['console', 'Log', 'terminal'], ['settings', 'Settings', 'cog'], ['components', 'Parts', 'components'], ['all', 'All pages', 'stack']];
  S.room = S.room || 'Living room';
  S.wiz = S.wiz || { i: 0, step: 0 };
  const roomOfDevice = (id) => roomsOf().filter((r) => r.ids.indexOf(id) >= 0).map((r) => r.name);
  const roomStatus = (r) => {
    const ds = r.ids.map(dev).filter(Boolean);
    if (ds.some((d) => !isOnline(d))) return 'bad';
    if (ds.some((d) => (lqiOf(d) != null && lqiOf(d) < 50) || (battOf(d) != null && battOf(d) < 20))) return 'warn';
    return 'ok';
  };
  const reading = (r, n) => { const it = r.items.find((x) => x[1] === n || (n === 'temperature' && x[1] === 'local_temperature')); if (!it) return ''; const d = dev(it[0]); return d && isOnline(d) ? fmt(d.exposes[it[1]]) : ''; };

  function shell(view) {
    const r = c.route(), cur = r.name === 'device' ? 'devices' : r.name === 'automation' ? 'automations' : r.name;
    return '<button type="button" class="nh-skip" data-act="skip">Skip to content</button><div class="fp-app">' +
      '<header class="fp-top"><a class="fp-brand" href="#/home">' + I('hub') + '<span>NodeHerder</span></a>' +
      '<nav class="fp-tabs" aria-label="Main">' + TABS.map((t) => '<a href="#/' + t[0] + '" class="fp-tab' + (cur === t[0] ? ' is-active' : '') + '"' + (cur === t[0] ? ' aria-current="page"' : '') + '>' + I(t[2]) + '<span>' + t[1] + '</span></a>').join('') + '</nav>' +
      '<div class="fp-top-actions">' + (S.permit ? '<button type="button" class="nh-chip is-accent" data-act="permit">' + I('join') + '<span>Joining <span data-permit>' + c.fmtTimer(S.permit) + '</span></span></button>' : btn('Add device', 'permit', { icon: 'join', sm: true, title: 'Permit join (Zigbee)' })) +
      c.connBtn() + btn('', 'mode-toggle', { icon: document.documentElement.dataset.mode === 'dark' ? 'sun' : 'moon', title: 'Light or dark' }) + btn('', 'pop', { icon: 'palette', title: 'Appearance' }) + btn('', 'logout', { icon: 'logout', title: 'Sign out' }) + '</div></header>' +
      c.banners() + '<main class="nh-content fp-content" id="nh-content" tabindex="-1">' + view + '</main></div>' + c.overlays();
  }

  function planHtml(mini) {
    return '<div class="fp-plan' + (mini ? ' is-mini' : '') + '" role="group" aria-label="House plan">' + roomsOf().map((r) => {
      const area = PLAN[r.name] || 'extra', st = roomStatus(r), t = reading(r, 'temperature'), h = reading(r, 'humidity');
      const pins = mini ? '' : '<div class="fp-pins">' + r.items.map((it) => {
        const d = dev(it[0]), e = d && d.exposes[it[1]]; if (!e) return '';
        const ts = c.tileState(d, e), on = ts.on && isOnline(d) && !isDisabled(d), act = ts.actionable && c.canControl(d), pend = ts.target && S.pending[c.pkey(d.id, ts.target.name)] !== undefined;
        return '<button type="button" class="fp-pin kind-' + meta(it[1]).kind + (on ? ' is-on' : '') + (!isOnline(d) ? ' is-off' : '') + (pend ? ' is-pending' : '') + '" data-act="' + (act ? 'tile-toggle' : 'fp-room') + '" data-arg="' + esc(act ? c.pkey(d.id, ts.target.name) : r.name) + '" title="' + esc(d.friendly_name + ' · ' + meta(it[1]).label + ': ' + fmt(e)) + '" aria-label="' + esc((act ? 'Switch ' : '') + d.friendly_name + ' ' + meta(it[1]).label + ', ' + fmt(e)) + '">' + I(meta(it[1]).icon) + '</button>';
      }).join('') + '</div>';
      return '<div class="fp-room area-' + area + ' is-' + st + (S.room === r.name && !mini ? ' is-sel' : '') + '"><button type="button" class="fp-room-hit" data-act="fp-room" data-arg="' + esc(r.name) + '" aria-pressed="' + (S.room === r.name) + '"><span class="fp-room-name">' + esc(r.name) + '</span><span class="fp-room-read">' + (t ? '<b>' + esc(t) + '</b>' : '') + (h ? '<span>' + esc(h) + '</span>' : '') + (st !== 'ok' ? '<em>' + I(st === 'bad' ? 'offline' : 'warn') + (st === 'bad' ? 'device offline' : 'needs attention') + '</em>' : '') + '</span></button>' + pins + '</div>';
    }).join('') + '<div class="fp-door" aria-hidden="true"></div></div>';
  }
  function inspector() {
    const r = roomsOf().find((x) => x.name === S.room); if (!r) return '';
    const seen = {}, rows = [];
    r.ids.forEach((id) => { const d = dev(id); if (!d || seen[id]) return; seen[id] = 1;
      rows.push('<section class="fp-insp-dev"><a href="#/device/' + encodeURIComponent(id) + '" class="fp-insp-name">' + I(devIcon(d)) + '<span>' + esc(d.friendly_name) + '</span>' + (isOnline(d) ? '' : chip('Offline', 'danger', 'offline')) + (isDisabled(d) ? chip('Disabled', 'muted', 'disabled') : '') + '</a>' +
        '<div class="nh-rows">' + Object.values(d.exposes).filter((e) => e.category === 'measurement').map((e) => exposeRow(d, e, false)).join('') + '</div></section>'); });
    return '<aside class="fp-insp" aria-label="' + esc(r.name) + '"><div class="fp-insp-head"><h2>' + esc(r.name) + '</h2>' + btn('', 'fp-close', { icon: 'close', title: 'Close room' }) + '</div>' + rows.join('') + '</aside>';
  }

  const views = {};
  views.home = () => '<div class="fp-home' + (S.roomOpen ? ' has-sheet' : '') + '"><div class="fp-canvas"><div class="fp-canvas-head"><h1 class="nh-h1">Plan</h1><p class="nh-sub">Tap a room to open it. Tap a switch pin to turn it on or off.</p></div>' + planHtml(false) +
    '<ul class="fp-legend"><li><i class="is-on"></i>On</li><li><i></i>Off or reading</li><li><i class="is-off"></i>Offline</li><li><i class="is-warn"></i>Weak signal or battery</li></ul></div>' + inspector() + '</div>';

  views.overview = () => {
    const ds = S.devices, off = ds.filter((d) => !isOnline(d)), weak = ds.filter((d) => lqiOf(d) != null && lqiOf(d) < 50), low = ds.filter((d) => battOf(d) != null && battOf(d) < 20);
    const line = (n, label, href, kind) => '<a class="fp-count is-' + kind + '" href="' + href + '"><b>' + n + '</b><span>' + label + '</span></a>';
    return '<div class="fp-health"><div><h1 class="nh-h1">House health</h1><p class="nh-sub">Rooms are shaded by their worst device.</p>' + planHtml(true) +
      '<div class="fp-counts">' + line(ds.length - off.length + ' / ' + ds.length, 'online', '#/list', off.length ? 'warn' : 'ok') + line(weak.length, 'weak Zigbee links', '#/list/sort=lqi', weak.length ? 'warn' : 'ok') + line(low.length, 'low batteries', '#/list/sort=battery', low.length ? 'warn' : 'ok') + line(ds.filter((d) => d.connection_type === 'http').length, 'Wi-Fi devices', '#/list/proto=http', 'ok') + '</div></div>' +
      '<div class="fp-health-side"><h2>Needs a look</h2><ul class="fp-attn">' + ds.filter((d) => !isOnline(d) || isDisabled(d) || weak.includes(d) || low.includes(d)).map((d) => '<li><a href="#/device/' + encodeURIComponent(d.id) + '">' + I(devIcon(d)) + '<span><b>' + esc(d.friendly_name) + '</b><small>' + esc((roomOfDevice(d.id)[0] || 'No room') + ', ' + (!isOnline(d) ? 'offline since ' + timeAgo(d.last_seen) : isDisabled(d) ? 'disabled' : low.includes(d) ? 'battery ' + Math.round(battOf(d)) + ' %' : 'signal ' + lqiOf(d) + ' LQI')) + '</small></span></a></li>').join('') + '</ul>' +
      '<h2>Just happened</h2><div class="nh-feed" data-region="feed">' + c.feedHtml(8) + '</div></div></div>';
  };

  views.devices = () => {
    const rooms = roomsOf(), used = {};
    const lane = (name, ids) => '<section class="fp-lane"><h2>' + esc(name) + '<small>' + ids.length + '</small></h2><div class="fp-lane-row">' + ids.map((id) => c.dcard(dev(id))).join('') + '</div></section>';
    const out = rooms.map((r) => { const ids = r.ids.filter((id, i, a) => a.indexOf(id) === i && dev(id)); ids.forEach((id) => (used[id] = 1)); return lane(r.name, ids); });
    const rest = S.devices.filter((d) => !used[d.id]).map((d) => d.id);
    return c.pageHead('Devices by room', 'Scroll each room sideways. Devices not placed in a room are at the end.', '') + out.join('') + (rest.length ? lane('Not in a room', rest) : '');
  };

  views.list = (r) => {
    const L = S.list;
    if (r.a) { r.a.split('&').forEach((kv) => { const p = kv.split('='); if (p[0] === 'sort') { L.sort = p[1]; L.dir = 1; } else if (L[p[0]] !== undefined) L[p[0]] = p[1]; }); history.replaceState(null, '', '#/list'); }
    const match = (d, k, v) => v === 'all' || (k === 'proto' ? d.connection_type === v : k === 'power' ? d.power_source === v : v === 'online' ? isOnline(d) && !isDisabled(d) : v === 'offline' ? !isOnline(d) : isDisabled(d));
    const base = S.devices.filter((d) => !L.q || (d.friendly_name + ' ' + d.id + ' ' + d.description).toLowerCase().includes(L.q.toLowerCase()));
    const ds = base.filter((d) => ['proto', 'power', 'status'].every((k) => match(d, k, L[k])));
    const key = { friendly_name: (d) => d.friendly_name.toLowerCase(), lqi: (d) => (lqiOf(d) == null ? 999 : lqiOf(d)), battery: (d) => (battOf(d) == null ? 999 : battOf(d)), last_seen: (d) => -new Date(d.last_seen).getTime() }[L.sort];
    ds.sort((a, b) => (key(a) > key(b) ? 1 : key(a) < key(b) ? -1 : 0) * L.dir);
    const facet = (title, k, opts) => '<fieldset class="fp-facet"><legend>' + title + '</legend>' + opts.map((o) => { const n = base.filter((d) => match(d, k, o[0])).length; return '<label class="nh-check"><input type="radio" name="fp-' + k + '" data-act="list-filter-' + k + '" data-arg="' + o[0] + '"' + (L[k] === o[0] ? ' checked' : '') + '> ' + o[1] + ' <small class="nh-muted">' + n + '</small></label>'; }).join('') + '</fieldset>';
    return '<div class="fp-find"><aside class="fp-facets"><label class="nh-search">' + I('search') + '<input class="nh-input" type="search" placeholder="Name, address or model…" value="' + esc(L.q) + '" data-bind="list-q" data-focus-key="list-q" aria-label="Search devices"></label>' +
      facet('Connection', 'proto', [['all', 'Any'], ['mqtt', 'Zigbee'], ['http', 'Wi-Fi · HTTP']]) + facet('Power', 'power', [['all', 'Any'], ['mains', 'Mains'], ['battery', 'Battery']]) + facet('Status', 'status', [['all', 'Any'], ['online', 'Online'], ['offline', 'Offline'], ['disabled', 'Disabled']]) +
      '<fieldset class="fp-facet"><legend>Sort by</legend>' + seg([{ v: 'friendly_name', t: 'Name' }, { v: 'lqi', t: 'Signal' }, { v: 'battery', t: 'Battery' }, { v: 'last_seen', t: 'Recent' }], L.sort, 'list-sort', 'Sort') + '</fieldset></aside>' +
      '<section class="fp-results"><h1 class="nh-h1">' + ds.length + ' of ' + S.devices.length + ' devices</h1>' +
      (ds.length ? '<ul class="fp-res">' + ds.map((d) => '<li><a href="#/device/' + encodeURIComponent(d.id) + '"><span class="fp-res-ic">' + I(devIcon(d)) + '</span><span class="fp-res-t"><b>' + esc(d.friendly_name) + '</b><small>' + esc((roomOfDevice(d.id).join(', ') || 'No room') + ' — ' + d.description) + '</small></span></a><span class="fp-res-chips">' + protoChip(d) + (signalChip(d) || '') + powerChip(d) + statusChip(d) + '</span></li>').join('') + '</ul>' : empty('search', 'Nothing matches', 'Clear the search or pick "Any" in the filters.', btn('Clear filters', 'list-clear', { kind: 'primary' }))) + '</section></div>';
  };

  views.device = (r) => {
    const d = dev(r.a); if (!d) return empty('question', 'Device not found', 'It may have been removed or renamed.', '<a class="nh-btn is-primary" href="#/list">Find a device</a>');
    const sup = protocol(d.connection_type).supports, secs = [['about', 'About'], ['controls', 'Controls'], ['settings', 'Settings'], ['metrics', 'History']];
    return '<div class="fp-dev"><header class="fp-dev-hero"><span class="fp-dev-ic">' + I(devIcon(d)) + '</span><div><p class="fp-crumb"><a href="#/home">Plan</a> / ' + esc(roomOfDevice(d.id)[0] || 'No room') + '</p><h1 class="nh-h1">' + esc(d.friendly_name) + '</h1><p class="nh-sub">' + esc(d.description) + '</p><div class="nh-chips">' + statusChip(d) + protoChip(d) + signalChip(d) + powerChip(d) + '</div></div>' +
      '<div class="nh-page-actions">' + btn('Rename', 'dev-rename', { icon: 'edit', arg: d.id, disabled: !sup.rename }) + btn('Interview', 'dev-interview', { icon: 'refresh', arg: d.id, disabled: !sup.interview }) + btn('Remove', 'dev-remove', { icon: 'trash', arg: d.id, kind: 'danger', disabled: !sup.remove }) + '</div></header>' +
      '<div class="fp-dev-body"><nav class="fp-dev-nav" aria-label="Sections">' + secs.map((s) => '<a href="#" data-act="all-jump" data-arg="sec-' + s[0] + '">' + s[1] + '</a>').join('') + '</nav><div>' +
      secs.map((s) => '<section class="fp-dev-sec" id="sec-' + s[0] + '"><h2>' + s[1] + '</h2>' + c.DV[s[0]](d) + '</section>').join('') + '</div></div></div>';
  };

  views.automations = () => {
    const A = S.automations, byRoom = {};
    A.forEach((a, i) => { const k = roomOfDevice(a.id)[0] || 'No room'; (byRoom[k] = byRoom[k] || []).push([a, i]); });
    return c.pageHead('Rules', 'Each rule belongs to the device that starts it.', btn('New rule', 'auto-create', { icon: 'plus', kind: 'primary' })) +
      Object.keys(byRoom).map((k) => '<section class="fp-rules"><h2>' + esc(k) + '</h2>' + byRoom[k].map(([a, i]) => {
        const src = dev(a.id), targets = []; a.triggers.forEach((t) => t.actions.forEach((x) => targets.indexOf(x.id) < 0 && targets.push(x.id)));
        const st = autoStatus(a), m = a.triggers.findIndex(canManual);
        return '<article class="fp-rule"><a class="fp-rule-pic" href="#/automation/' + encodeURIComponent(a.id) + '" aria-label="Edit ' + esc(a.friendlyname) + '"><span class="fp-node">' + I(src ? devIcon(src) : 'devices') + '</span><span class="fp-wire"></span>' + targets.map((id) => { const t = dev(id); return '<span class="fp-node is-target" title="' + esc(t ? t.friendly_name : id) + '">' + I(t ? devIcon(t) : 'devices') + '</span>'; }).join('') + '</a>' +
          '<div class="fp-rule-t"><a href="#/automation/' + encodeURIComponent(a.id) + '"><b>' + esc(a.description || a.friendlyname) + '</b></a><p>' + esc(describeTrigger(a, a.triggers[0])) + (a.triggers.length > 1 ? ' And ' + (a.triggers.length - 1) + ' more.' : '') + '</p></div>' +
          '<div class="fp-rule-side">' + chip(st.text, st.kind, st.icon, st.title) + (m >= 0 ? btn('Run', 'auto-run', { icon: 'play', sm: true, arg: i + '|' + m }) : '') + toggle(a.enabled, 'auto-enable', i, { label: 'Enabled' }) + btn('', 'auto-delete', { icon: 'trash', sm: true, kind: 'danger', title: 'Delete rule', arg: i }) + '</div></article>';
      }).join('') + '</section>').join('');
  };

  // Wizard editor: real sequence, so steps are numbered.
  const STEPS = ['Start', 'Only if', 'Do this', 'Schedule', 'Review'];
  views.automation = (r) => {
    c.V.automation(r); // prepares S.draft
    if (!S.draft) return empty('question', 'Rule not found', '', '<a class="nh-btn is-primary" href="#/automations">Back to rules</a>');
    const P = c.edParts(), a = P.a, n = a.triggers.length;
    if (S.wiz.i >= n) S.wiz.i = Math.max(0, n - 1);
    const i = S.wiz.i, t = a.triggers[i], st = S.wiz.step;
    const stepBody = !t ? empty('automation', 'No triggers yet', 'Add a trigger to choose what starts this rule.', btn('Add trigger', 'ed-add-trig', { icon: 'plus', kind: 'primary' })) :
      st === 0 ? '<p class="fp-q">What should start this rule?</p>' + P.whenBody(i) :
      st === 1 ? '<p class="fp-q">Should it only run in some situations?</p>' + (t.conditions.map((x, j) => P.condRow(i, j)).join('') || '<p class="nh-muted">No conditions. The rule runs every time.</p>') + P.condAdd(i) :
      st === 2 ? '<p class="fp-q">What should happen? Actions run in order.</p>' + t.actions.map((x, k) => P.actRow(i, k)).join('') + P.err(P.tp(i) + '.actions') + P.actAdd(i) :
      st === 3 ? '<p class="fp-q">When is this rule allowed to run?</p><div class="fp-two"><div>' + P.schedule() + '</div><div>' + P.details() + '</div></div>' :
      '<p class="fp-q">Check the rule, then save.</p>' + a.triggers.map((x, k) => '<div class="fp-review"><b>Trigger ' + (k + 1) + '</b>' + P.sentence(k) + P.warnList(k) + '</div>').join('') + P.checks() + btn(S.jsonOpen ? 'Hide payload' : 'Show payload', 'ed-json', { icon: 'code', kind: 'ghost', sm: true }) + P.json();
    const errsAt = (s) => !t ? 0 : P.trigErrs(i).filter((e) => (s === 0 ? /\.name$/.test(e.p) && e.p.split('.').length === 3 : s === 1 ? e.p.indexOf('.conditions') > 0 : s === 2 ? e.p.indexOf('.actions') > 0 : false)).length;
    return '<div class="fp-wiz"><header>' + P.back() + '<h1 class="nh-h1">' + esc(a.friendlyname) + '</h1><p class="nh-sub">' + esc(a.description || 'No description') + '</p></header>' +
      '<div class="fp-trigtabs" role="tablist" aria-label="Triggers">' + a.triggers.map((x, k) => '<button type="button" role="tab" aria-selected="' + (k === i) + '" data-act="fp-trig" data-arg="' + k + '">Trigger ' + (k + 1) + (P.trigErrs(k).length ? ' ' + I('error') : '') + '</button>').join('') + btn('', 'ed-add-trig', { icon: 'plus', sm: true, title: 'Add trigger' }) + (t ? P.trigTools(i).replace(/data-act="ed-del"/, 'data-act="ed-del"') : '') + '</div>' +
      '<ol class="fp-steps">' + STEPS.map((s, k) => '<li><button type="button" data-act="fp-step" data-arg="' + k + '" aria-current="' + (k === st ? 'step' : 'false') + '" class="' + (k < st ? 'is-done' : '') + (errsAt(k) ? ' has-err' : '') + '"><span>' + (k + 1) + '</span>' + s + '</button></li>').join('') + '</ol>' +
      '<section class="fp-step-body">' + stepBody + '</section>' +
      '<div class="fp-wiz-nav">' + btn('Back', 'fp-step', { icon: 'chevronL', arg: Math.max(0, st - 1), disabled: st === 0 }) + (st < 4 ? btn('Next: ' + STEPS[st + 1], 'fp-step', { kind: 'primary', arg: st + 1 }) : '') + '</div>' + P.savebar() + '</div>';
  };

  const actions = {
    'fp-room': (n) => { S.room = n; S.roomOpen = true; c.render(); },
    'fp-close': () => { S.roomOpen = false; c.render(); },
    'fp-step': (k) => { S.wiz.step = +k; c.render(); },
    'fp-trig': (k) => { S.wiz.i = +k; c.render(); },
  };
  return { shell, views, actions };
};
