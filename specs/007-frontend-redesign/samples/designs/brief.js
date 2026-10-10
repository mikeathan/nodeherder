/* 4 · Brief — one column of plain sentences, a command palette (Ctrl K), fill-in-the-blanks rules. */
window.NH_LAYOUT = function (c) {
  const { S, I, esc, btn, chip, toggle, sel, empty, dev, meta, fmt, roomsOf, timeAgo, clock, devIcon, isOnline, isDisabled, lqiOf, battOf, protocol, capability, canControl, pkey, stateExposeOf, describeTrigger, autoStatus, canManual, $ } = c;
  S.cmd = false; S.cmdQ = ''; S.cmdI = 0; S.open = S.open || {};
  const MENU = [['home', 'Home'], ['overview', 'Status'], ['devices', 'Rooms'], ['list', 'All devices'], ['automations', 'Rules'], ['assistant', 'Ask'], ['console', 'Log'], ['settings', 'Settings'], ['components', 'Parts'], ['all', 'All pages']];
  const link = (d) => '<a href="#/device/' + encodeURIComponent(d.id) + '">' + esc(d.friendly_name) + '</a>';
  const list = (xs) => (xs.length < 2 ? xs.join('') : xs.slice(0, -1).join(', ') + ' and ' + xs[xs.length - 1]);
  const sw = (d, e) => { const k = pkey(d.id, e.name), p = S.pending[k] !== undefined; return '<span class="br-sw">' + toggle(e.data === e.values.on || e.data === true, 'expose-toggle', k, { disabled: !canControl(d) || p, pending: p, label: d.friendly_name }) + (p ? '<small>sending…</small>' : '') + '</span>'; };

  function commands() {
    const q = S.cmdQ.trim().toLowerCase(), out = [];
    MENU.forEach((m) => out.push({ t: 'Open ' + m[1], k: 'page', href: '#/' + m[0], icon: 'chevronR' }));
    S.devices.forEach((d) => {
      out.push({ t: d.friendly_name, k: 'device', href: '#/device/' + encodeURIComponent(d.id), icon: devIcon(d), sub: protocol(d.connection_type).label + (isOnline(d) ? '' : ', offline') });
      const st = stateExposeOf(d); if (st && canControl(d)) out.push({ t: 'Turn ' + (st.data === st.values.on ? 'off ' : 'on ') + d.friendly_name, k: 'action', act: 'expose-toggle', arg: pkey(d.id, st.name), icon: 'power' });
    });
    S.automations.forEach((a, i) => { out.push({ t: 'Edit rule: ' + (a.description || a.friendlyname), k: 'rule', href: '#/automation/' + encodeURIComponent(a.id), icon: 'automation' }); const m = a.triggers.findIndex(canManual); if (m >= 0) out.push({ t: 'Run rule: ' + (a.description || a.friendlyname), k: 'action', act: 'auto-run', arg: i + '|' + m, icon: 'play' }); });
    out.push({ t: 'Let new Zigbee devices join', k: 'action', act: 'permit', icon: 'join' }, { t: 'Switch light or dark', k: 'action', act: 'mode-toggle', icon: 'moon' }, { t: 'Change appearance', k: 'action', act: 'pop', icon: 'palette' });
    return (q ? out.filter((x) => (x.t + ' ' + (x.sub || '')).toLowerCase().includes(q)) : out.filter((x) => x.k !== 'device' || true)).slice(0, 9);
  }
  const cmdList = () => { const r = commands(); if (S.cmdI >= r.length) S.cmdI = 0; return r.length ? r.map((x, i) => (x.href ? '<a href="' + x.href + '" data-act="br-cmd-close"' : '<button type="button" data-act="br-run" data-arg="' + i + '"') + ' class="br-res' + (i === S.cmdI ? ' is-cur' : '') + '" id="br-r' + i + '" role="option" aria-selected="' + (i === S.cmdI) + '">' + I(x.icon) + '<span>' + esc(x.t) + (x.sub ? '<small>' + esc(x.sub) + '</small>' : '') + '</span><em>' + x.k + '</em>' + (x.href ? '</a>' : '</button>')).join('') : '<p class="br-none">Nothing matches “' + esc(S.cmdQ) + '”. Try a room or device name.</p>'; };
  const palette = () => !S.cmd ? '' : '<div class="nh-dialog-backdrop" data-act="br-cmd-close"></div><div class="br-cmd" role="dialog" aria-modal="true" aria-label="Command palette"><label class="br-cmd-in">' + I('search') + '<input type="text" placeholder="Find a device, room, rule or action…" value="' + esc(S.cmdQ) + '" data-cmdq="1" data-focus-key="cmdq" role="combobox" aria-expanded="true" aria-controls="br-results" aria-activedescendant="br-r' + S.cmdI + '" autocomplete="off" spellcheck="false"></label><div class="br-results" id="br-results" role="listbox">' + cmdList() + '</div><p class="br-hint">↑ ↓ to move, Enter to open, Esc to close</p></div>';

  function shell(view) {
    const r = c.route(), cur = r.name === 'device' ? 'devices' : r.name === 'automation' ? 'automations' : r.name;
    return '<button type="button" class="nh-skip" data-act="skip">Skip to content</button><div class="br-app"><header class="br-top"><a class="br-brand" href="#/home">NodeHerder</a>' +
      '<button type="button" class="br-bar" data-act="br-cmd">' + I('search') + '<span>Search or type a command…</span><kbd>Ctrl&nbsp;K</kbd></button>' +
      '<nav class="br-menu" aria-label="Main">' + MENU.map((m) => '<a href="#/' + m[0] + '"' + (cur === m[0] ? ' aria-current="page"' : '') + '>' + m[1] + '</a>').join('') + '</nav></header>' +
      c.banners() + '<main class="nh-content br-content" id="nh-content" tabindex="-1"><div class="br-col">' + view + '</div></main>' +
      '<footer class="br-foot">' + c.connBtn() + (S.permit ? '<button type="button" class="nh-chip is-accent" data-act="permit-stop">' + I('join') + '<span>Joining, <span data-permit>' + c.fmtTimer(S.permit) + '</span> left. Stop</span></button>' : '') + '<span class="br-foot-gap"></span>' + btn('', 'mode-toggle', { icon: document.documentElement.dataset.mode === 'dark' ? 'sun' : 'moon', title: 'Light or dark', sm: true }) + btn('', 'pop', { icon: 'palette', title: 'Appearance', sm: true }) + btn('Sign out', 'logout', { sm: true, kind: 'ghost' }) + '</footer></div>' + palette() + c.overlays();
  }

  const greet = () => { const h = new Date().getHours(); return h < 5 ? 'Late night' : h < 12 ? 'Good morning' : h < 18 ? 'Good afternoon' : 'Good evening'; };
  const views = {};
  views.home = () => {
    const ds = S.devices, off = ds.filter((d) => !isOnline(d)), onSw = [], offSw = [];
    ds.forEach((d) => { const st = stateExposeOf(d); if (!st || isDisabled(d) || !isOnline(d)) return; (st.data === st.values.on ? onSw : offSw).push([d, st]); });
    const humid = [], rooms = roomsOf().map((r) => {
      const t = r.items.find((x) => x[1] === 'temperature' || x[1] === 'local_temperature'), h = r.items.find((x) => x[1] === 'humidity');
      const td = t && dev(t[0]), hd = h && dev(h[0]);
      if (!td || !isOnline(td)) return td ? esc(r.name) + ' has no reading right now (' + link(td) + ' is offline).' : '';
      const hv = hd && hd.exposes.humidity.data; if (hv > 75) humid.push(r.name);
      return esc(r.name) + ' is <b>' + esc(fmt(td.exposes[t[1]])) + '</b>' + (hd ? ' at ' + esc(fmt(hd.exposes.humidity)) + ' humidity' : '') + '.';
    }).filter(Boolean);
    const quiet = !off.length && !humid.length;
    return '<h1 class="br-h1">' + greet() + '. ' + (quiet ? 'The house is quiet.' : 'A couple of things to know.') + '</h1>' +
      '<p class="br-lead">' + (ds.length - off.length) + ' of ' + ds.length + ' devices are online. ' + (off.length ? list(off.map((d) => link(d) + ' went offline ' + timeAgo(d.last_seen))) + '. ' : '') + (humid.length ? 'The ' + list(humid.map(esc)).toLowerCase() + ' ' + (humid.length > 1 ? 'are' : 'is') + ' humid.' : '') + '</p>' +
      '<h2 class="br-h2">Lights and switches</h2><p class="br-p">' + (onSw.length ? 'On: ' + onSw.map(([d, e]) => '<span class="br-tok">' + link(d) + sw(d, e) + '</span>').join(' ') : 'Everything is off.') + '</p>' + (offSw.length ? '<p class="br-p br-dim">Off: ' + offSw.map(([d, e]) => '<span class="br-tok">' + link(d) + sw(d, e) + '</span>').join(' ') + '</p>' : '') +
      '<h2 class="br-h2">Rooms</h2><p class="br-p">' + rooms.join(' ') + '</p>' +
      '<h2 class="br-h2">Today</h2><ol class="br-log" data-region="feed">' + S.feed.slice(0, 10).map((f) => '<li><time>' + clock(f.t).slice(0, 5) + '</time><span>' + link(dev(f.id)) + ': ' + f.changes.map((ch) => esc(meta(ch.n).label.toLowerCase()) + ' changed to <b>' + esc(ch.to) + '</b>').join(', ') + '.</span></li>').join('') + '</ol>';
  };
  views.overview = () => {
    const ds = S.devices, z = ds.filter((d) => d.connection_type === 'mqtt'), weak = ds.filter((d) => lqiOf(d) != null && lqiOf(d) < 50), low = ds.filter((d) => battOf(d) != null && battOf(d) < 20);
    const auto = S.automations, sched = auto.filter((a) => a.schedules.length), noCond = [];
    auto.forEach((a) => a.triggers.forEach((t, i) => t.type === 'deviceTrigger' && !t.conditions.length && noCond.push(a.friendlyname + ' (trigger ' + (i + 1) + ')')));
    return '<h1 class="br-h1">Status</h1><h2 class="br-h2">Network</h2><p class="br-p">NodeHerder is <b>' + esc(c.connText().toLowerCase()) + '</b> and knows ' + ds.length + ' devices: ' + z.length + ' on Zigbee (' + z.filter((d) => d.power_source === 'mains').length + ' mains-powered routers, ' + z.filter((d) => d.power_source === 'battery').length + ' battery end devices) and ' + (ds.length - z.length) + ' on Wi-Fi over HTTP. ' +
      (weak.length ? list(weak.map(link)) + (weak.length > 1 ? ' have' : ' has') + ' a weak Zigbee link (under 50). A mains-powered device nearby would help.' : 'All Zigbee links are healthy.') + ' ' + (S.permit ? 'New Zigbee devices can join for <span data-permit>' + c.fmtTimer(S.permit) + '</span>.' : 'Joining is closed. <a href="#" data-act="permit">Let new devices join</a>.') + '</p>' +
      '<h2 class="br-h2">Batteries</h2><p class="br-p">' + (low.length ? list(low.map((d) => link(d) + ' is at ' + Math.round(battOf(d)) + ' %')) + '.' : 'No batteries are low.') + '</p>' +
      '<h2 class="br-h2">Rules</h2><p class="br-p">' + auto.filter((a) => a.enabled).length + ' of ' + auto.length + ' rules are enabled; ' + sched.length + ' follow a schedule. ' + (noCond.length ? list(noCond.map(esc)) + ' will only run by hand, because the hub ignores device triggers without conditions.' : '') + '</p>' +
      '<h2 class="br-h2">Assistant</h2><p class="br-p">The MCP server is ' + (S.mcp.running ? 'running with ' + S.mcp.connectedClients + ' client connected' : 'stopped') + '. <a href="#/assistant">Ask a question about the house</a>.</p>';
  };
  views.devices = () => '<h1 class="br-h1">Rooms</h1>' + roomsOf().map((r) => '<h2 class="br-h2">' + esc(r.name) + '</h2>' + r.ids.filter((x, i, a) => a.indexOf(x) === i).map((id) => {
    const d = dev(id); if (!d) return '';
    const reads = Object.values(d.exposes).filter((e) => e.category === 'measurement' && capability(e) !== 'switch').slice(0, 3).map((e) => esc(meta(e.name).label.toLowerCase()) + ' ' + '<b>' + esc(fmt(e)) + '</b>');
    const st = stateExposeOf(d);
    return '<details class="br-dev" data-open-key="' + esc(id) + '"' + (S.open[id] ? ' open' : '') + '><summary><span>' + link(d) + (isOnline(d) ? (isDisabled(d) ? ' is disabled.' : (reads.length ? ': ' + list(reads) + '.' : '.')) : ' is offline (last seen ' + esc(timeAgo(d.last_seen)) + ').') + '</span>' + (st && isOnline(d) ? sw(d, st) : '') + '</summary><div class="nh-rows">' + Object.values(d.exposes).filter((e) => e.category === 'measurement').map((e) => c.exposeRow(d, e, false)).join('') + '</div></details>';
  }).join('')).join('');
  views.list = (r) => {
    const L = S.list;
    if (r.a) { r.a.split('&').forEach((kv) => { const p = kv.split('='); if (p[0] === 'sort') { L.sort = p[1]; L.dir = 1; } else if (L[p[0]] !== undefined) L[p[0]] = p[1]; }); history.replaceState(null, '', '#/list'); }
    const ds = S.devices.filter((d) => (L.proto === 'all' || d.connection_type === L.proto) && (L.power === 'all' || d.power_source === L.power) && (L.status === 'all' || (L.status === 'online' && isOnline(d) && !isDisabled(d)) || (L.status === 'offline' && !isOnline(d)) || (L.status === 'disabled' && isDisabled(d))) && (!L.q || (d.friendly_name + d.id + d.description).toLowerCase().includes(L.q.toLowerCase())));
    const key = { friendly_name: (d) => d.friendly_name.toLowerCase(), lqi: (d) => (lqiOf(d) == null ? 999 : lqiOf(d)), battery: (d) => (battOf(d) == null ? 999 : battOf(d)), last_seen: (d) => -new Date(d.last_seen).getTime() }[L.sort];
    ds.sort((a, b) => (key(a) > key(b) ? 1 : key(a) < key(b) ? -1 : 0) * L.dir);
    const s = (k, opts) => sel(opts.map((o) => ({ v: o[0], t: o[1] })), L[k], 'class="br-inl" data-br-filter="' + k + '" aria-label="' + k + '"');
    return '<h1 class="br-h1">All devices</h1><p class="br-p br-filter">Show ' + s('status', [['all', 'every'], ['online', 'online'], ['offline', 'offline'], ['disabled', 'disabled']]) + ' device on ' + s('proto', [['all', 'any connection'], ['mqtt', 'Zigbee'], ['http', 'Wi-Fi']]) + ', powered by ' + s('power', [['all', 'anything'], ['mains', 'mains'], ['battery', 'battery']]) + ', sorted by ' + s('sort', [['friendly_name', 'name'], ['lqi', 'weakest signal'], ['battery', 'lowest battery'], ['last_seen', 'most recent']]) + '. <input class="br-inl br-q" type="search" placeholder="Filter by name…" value="' + esc(L.q) + '" data-bind="list-q" data-focus-key="list-q" aria-label="Filter by name"></p>' +
      '<p class="br-count">' + ds.length + ' of ' + S.devices.length + '</p><table class="br-table"><thead><tr><th scope="col">Device</th><th scope="col">Connection</th><th scope="col">Signal</th><th scope="col">Power</th><th scope="col">Seen</th></tr></thead><tbody>' +
      ds.map((d) => '<tr' + (isOnline(d) ? '' : ' class="is-off"') + '><td>' + link(d) + (isOnline(d) ? '' : ' <span class="br-flag">offline</span>') + (isDisabled(d) ? ' <span class="br-flag">disabled</span>' : '') + '</td><td>' + esc(protocol(d.connection_type).label) + '</td><td class="br-num">' + (lqiOf(d) != null ? lqiOf(d) + ' LQI' : d.exposes.rssi ? d.exposes.rssi.data + ' dBm' : '—') + '</td><td class="br-num">' + (battOf(d) != null ? Math.round(battOf(d)) + ' %' : 'mains') + '</td><td>' + esc(timeAgo(d.last_seen)) + '</td></tr>').join('') + '</tbody></table>';
  };
  views.device = (r) => {
    const d = dev(r.a); if (!d) return empty('question', 'Device not found', 'It may have been removed or renamed.');
    const sup = protocol(d.connection_type).supports, ms = Object.values(d.exposes).filter((e) => e.category === 'measurement');
    return '<p class="br-crumb"><a href="#/devices">Rooms</a></p><h1 class="br-h1">' + esc(d.friendly_name) + '</h1><p class="br-lead">' + esc(d.description) + '. Connected over ' + esc(protocol(d.connection_type).label) + ', ' + (d.power_source === 'battery' ? 'on battery (' + Math.round(battOf(d) || 0) + ' %)' : 'mains powered') + ', ' + (isOnline(d) ? 'last heard from ' + esc(timeAgo(d.last_seen)) : '<b>offline</b> since ' + esc(timeAgo(d.last_seen))) + '.</p>' +
      '<dl class="br-readings">' + ms.map((e) => '<div><dt>' + esc(e.description || meta(e.name).label) + '</dt><dd>' + (capability(e) === 'switch' ? esc(fmt(e)) + sw(d, e) : esc(fmt(e))) + '</dd></div>').join('') + '</dl>' +
      '<h2 class="br-h2">Adjust</h2><div class="nh-rows br-rows">' + ms.filter((e) => capability(e) === 'slider' || capability(e) === 'preset').map((e) => c.exposeRow(d, e, false)).join('') + '</div>' +
      '<h2 class="br-h2">Settings</h2>' + c.DV.settings(d) + '<h2 class="br-h2">History</h2>' + c.DV.metrics(d) +
      '<h2 class="br-h2">Device</h2>' + c.DV.about(d) + '<p class="br-p">' + btn('Rename', 'dev-rename', { arg: d.id, disabled: !sup.rename }) + ' ' + btn('Interview', 'dev-interview', { arg: d.id, disabled: !sup.interview }) + ' ' + btn('Remove from network', 'dev-remove', { arg: d.id, kind: 'danger', disabled: !sup.remove }) + '</p>';
  };
  views.automations = () => '<h1 class="br-h1">Rules</h1><p class="br-lead">Each rule reads as a sentence. Select a rule to change it.</p>' + S.automations.map((a, i) => { const st = autoStatus(a), m = a.triggers.findIndex(canManual);
    return '<article class="br-rule' + (a.enabled ? '' : ' is-off') + '"><div class="br-rule-head">' + toggle(a.enabled, 'auto-enable', i, { label: 'Enable ' + a.friendlyname }) + '<h2><a href="#/automation/' + encodeURIComponent(a.id) + '">' + esc(a.description || a.friendlyname) + '</a></h2><span class="br-meta">' + esc(st.text + (st.title ? ', ' + st.title : '')) + '</span></div>' + a.triggers.map((t) => '<p>' + esc(describeTrigger(a, t)) + '</p>').join('') + '<p class="br-rule-act">' + (m >= 0 ? btn('Run now', 'auto-run', { sm: true, icon: 'play', arg: i + '|' + m }) : '') + '<a class="nh-btn is-sm" href="#/automation/' + encodeURIComponent(a.id) + '">Edit</a>' + btn('Delete', 'auto-delete', { sm: true, kind: 'danger', arg: i }) + '</p></article>'; }).join('') + '<p>' + btn('Write a new rule', 'auto-create', { kind: 'primary', icon: 'plus' }) + '</p>';

  // Fill-in-the-blanks editor: the sentence IS the form.
  views.automation = (r) => {
    c.V.automation(r);
    if (!S.draft) return empty('question', 'Rule not found', '');
    const P = c.edParts(), a = P.a, sch = a.schedules;
    const paras = a.triggers.map((t, i) => {
      const conds = t.conditions.map((x, j) => '<span class="br-clause">' + (j ? 'and ' : 'only if ') + P.condFields(i, j) + btn('', 'ed-del', { icon: 'close', sm: true, kind: 'ghost', title: 'Remove this condition', arg: P.tp(i) + '.conditions.' + j }) + '</span>').join(' ');
      const acts = t.actions.map((x, k) => '<span class="br-clause">' + (k ? 'and then ' : 'then ') + P.actType(i, k) + ' ' + P.actTarget(i, k) + P.err(P.tp(i) + '.actions.' + k + '.id') + ' ' + P.actBody(i, k) + btn('', 'ed-del', { icon: 'close', sm: true, kind: 'ghost', title: 'Remove this action', arg: P.tp(i) + '.actions.' + k }) + '</span>').join(' ');
      return '<section class="br-mad"><p class="br-mad-p"><span class="br-clause">' + (t.type === 'manualTrigger' ? 'When I run it by hand' : 'When <b>' + esc(a.friendlyname) + '</b>’s ' + P.srcSelect(i) + ' changes') + P.err(P.tp(i) + '.name') + '</span>, ' + (conds ? conds + ', ' : '') + (acts || '<span class="br-missing">then… (add an action)</span>') + '.' + P.err(P.tp(i) + '.actions') + '</p>' +
        '<p class="br-mad-tools">' + btn('add a condition', 'ed-add-cond', { sm: true, kind: 'ghost', icon: 'plus', arg: P.tp(i) + '|expose' }) + btn('add a time window', 'ed-add-cond', { sm: true, kind: 'ghost', icon: 'clock', arg: P.tp(i) + '|time' }) + btn('add an action', 'ed-add-act', { sm: true, kind: 'ghost', icon: 'plus', arg: P.tp(i) }) + '<span class="br-gap"></span>' + P.trigType(i) + P.trigTools(i) + '</p>' + P.warnList(i) + '</section>';
    }).join('');
    const en = sch.find((s) => s.type === 'enable'), di = sch.find((s) => s.type === 'disable');
    return '<p class="br-crumb">' + P.back() + '</p><h1 class="br-h1">' + esc(a.description || a.friendlyname) + '</h1>' +
      '<p class="br-lead">Change any highlighted word. This rule belongs to ' + (P.d ? '<a href="#/device/' + encodeURIComponent(a.id) + '">' + esc(a.friendlyname) + '</a>' : esc(a.friendlyname)) + ' and is ' + toggle(a.enabled, 'ed-enabled', '', { label: 'Enabled' }) + ' ' + (a.enabled ? 'enabled' : 'disabled') + '.</p>' +
      (paras || '<p class="br-p">This rule has no triggers yet.</p>') + '<p>' + btn('Add another “when…”', 'ed-add-trig', { icon: 'plus' }) + '</p>' +
      '<h2 class="br-h2">Schedule</h2><p class="br-p">' + (en ? 'Turn the rule on at <input type="time" class="br-inl" data-sched="enable" value="' + en.startAt + '" aria-label="Enable at"> ' + btn('', 'ed-sched-del', { icon: 'close', sm: true, kind: 'ghost', title: 'Remove', arg: 'enable' }) : btn('turn on at a set time', 'ed-sched-add', { sm: true, kind: 'ghost', icon: 'plus', arg: 'enable' })) + ' ' + (di ? 'and off at <input type="time" class="br-inl" data-sched="disable" value="' + di.startAt + '" aria-label="Disable at"> ' + btn('', 'ed-sched-del', { icon: 'close', sm: true, kind: 'ghost', title: 'Remove', arg: 'disable' }) : btn('turn off at a set time', 'ed-sched-add', { sm: true, kind: 'ghost', icon: 'plus', arg: 'disable' })) + '. Times are in ' + esc(Intl.DateTimeFormat().resolvedOptions().timeZone) + '.</p>' +
      '<h2 class="br-h2">Description</h2><p class="br-p"><input class="br-inl br-wide" data-path="description" data-kind="str" data-focus-key="desc" value="' + esc(a.description) + '" aria-label="Description"></p>' +
      '<h2 class="br-h2">In plain words</h2>' + a.triggers.map((t, i) => P.sentence(i)).join('') + P.checks() + P.savebar() + P.json();
  };

  views.assistant = () => {
    const cv = S.convs.find((x) => x.id === S.activeConv) || S.convs[0];
    return '<h1 class="br-h1">Ask about the house</h1><p class="br-p">Earlier: ' + S.convs.map((x) => '<a href="#" data-act="chat-sel" data-arg="' + x.id + '"' + (x.id === cv.id ? ' aria-current="true"' : '') + '>' + esc(x.title) + '</a>').join(', ') + '. ' + btn('Start over', 'chat-new', { sm: true, kind: 'ghost' }) + '</p>' +
      '<div class="br-chat" data-region="msgs">' + (cv.messages.length ? cv.messages.map((m) => '<p class="br-msg is-' + m.role + '"><b>' + (m.role === 'user' ? 'You' : 'NodeHerder') + '</b> ' + esc(m.content) + '</p>').join('') : '<p class="br-dim">Try “Which devices have weak signal?” or “How warm was the garden last night?”. Answers come from stored metrics; the assistant cannot change devices.</p>') + (S.typing ? '<p class="br-msg is-assistant"><b>NodeHerder</b> thinking…</p>' : '') + '</div>' +
      '<form class="br-ask" data-form="chat"><textarea class="nh-input" rows="2" placeholder="Ask a question…" data-focus-key="chat" aria-label="Question"></textarea>' + '<button type="submit" class="nh-btn is-primary"' + (S.typing ? ' disabled' : '') + '>Ask</button></form>';
  };

  const actions = {
    'br-cmd': () => { S.cmd = true; S.cmdQ = ''; S.cmdI = 0; c.render(); setTimeout(() => { const i = $('[data-cmdq]'); if (i) i.focus(); }, 0); },
    'br-cmd-close': () => { S.cmd = false; c.render(); },
    'br-run': (i) => { const x = commands()[+i]; S.cmd = false; if (x && x.act) actions._run(x); else c.render(); },
  };
  actions._run = (x) => { c.render(); c.A[x.act](x.arg); };
  document.addEventListener('keydown', (ev) => {
    if ((ev.ctrlKey || ev.metaKey) && ev.key.toLowerCase() === 'k') { ev.preventDefault(); actions['br-cmd'](); return; }
    if (!S.cmd) return;
    const n = commands().length;
    if (ev.key === 'Escape') { ev.stopPropagation(); actions['br-cmd-close'](); }
    else if (ev.key === 'ArrowDown' || ev.key === 'ArrowUp') { ev.preventDefault(); S.cmdI = (S.cmdI + (ev.key === 'ArrowDown' ? 1 : n - 1)) % Math.max(n, 1); const box = $('#br-results'); if (box) box.innerHTML = cmdList(); }
    else if (ev.key === 'Enter') { ev.preventDefault(); const x = commands()[S.cmdI]; if (!x) return; S.cmd = false; if (x.href) { c.render(); c.go(x.href); } else actions._run(x); }
  }, true);
  document.addEventListener('input', (ev) => { if (ev.target.dataset && ev.target.dataset.cmdq) { S.cmdQ = ev.target.value; S.cmdI = 0; const box = $('#br-results'); if (box) box.innerHTML = cmdList(); } });
  document.addEventListener('change', (ev) => { const k = ev.target.dataset && ev.target.dataset.brFilter; if (k) { if (k === 'sort') { S.list.sort = ev.target.value; S.list.dir = 1; } else S.list[k] = ev.target.value; c.render(); } });
  document.addEventListener('toggle', (ev) => { const k = ev.target.dataset && ev.target.dataset.openKey; if (k) S.open[k] = ev.target.open; }, true);
  return { shell, views, actions };
};
