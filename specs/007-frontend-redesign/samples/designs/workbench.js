/* 3 · Workbench — IDE layout: activity bar, explorer tree, open tabs, property grids, status bar, log panel. */
window.NH_LAYOUT = function (c) {
  const { S, I, esc, btn, chip, toggle, empty, dev, meta, fmt, roomsOf, timeAgo, clock, statusChip, protoChip, signalChip, powerChip, devIcon, isOnline, isDisabled, lqiOf, battOf, protocol, capability, canControl, pkey, autoStatus, canManual, describeTrigger } = c;
  S.tree = S.tree || { areas: true, auto: true, hub: true, 'area:Living room': true };
  S.tabs = S.tabs || [];
  S.panel = S.panel === undefined ? true : S.panel;
  S.edSel = S.edSel || 'details';
  S.devTab = S.devTab || 'props';
  const hashNow = () => location.hash || '#/overview';

  function titleFor(h) {
    const p = h.replace(/^#\/?/, '').split('/').map(decodeURIComponent);
    if (p[0] === 'device') { const d = dev(p[1]); return d ? d.friendly_name : 'Device'; }
    if (p[0] === 'automation') { const a = S.automations.find((x) => x.id === p[1]); return (a ? a.friendlyname : 'Rule') + '.rule'; }
    const n = c.NAV.find((x) => x.id === p[0]); return n ? n.label : p[0];
  }
  function treeNode(key, label, icon, children, href, extra) {
    const open = S.tree[key], cur = href && hashNow() === href;
    const row = children ? '<button type="button" class="wb-tn" data-act="wb-tree" data-arg="' + esc(key) + '" aria-expanded="' + !!open + '">' + I(open ? 'chevronD' : 'chevronR', 'wb-car') + (icon ? I(icon) : '') + '<span>' + esc(label) + '</span>' + (extra || '') + '</button>'
      : '<a class="wb-tn is-leaf' + (cur ? ' is-cur' : '') + '" href="' + href + '"' + (cur ? ' aria-current="page"' : '') + '>' + (icon ? I(icon) : '') + '<span>' + esc(label) + '</span>' + (extra || '') + '</a>';
    return '<li>' + row + (children && open ? '<ul>' + children() + '</ul>' : '') + '</li>';
  }
  const dot = (d) => '<i class="wb-dot is-' + (!isOnline(d) ? 'bad' : isDisabled(d) ? 'off' : 'ok') + '" aria-label="' + (!isOnline(d) ? 'offline' : isDisabled(d) ? 'disabled' : 'online') + '"></i>';
  function explorer() {
    const hub = [['overview', 'Overview', 'overview'], ['home', 'Areas grid', 'home'], ['devices', 'All properties', 'devices'], ['list', 'Device table', 'list'], ['assistant', 'Assistant', 'sparkles'], ['console', 'Console', 'terminal'], ['settings', 'Settings', 'cog'], ['components', 'Components', 'components'], ['all', 'All pages', 'stack']];
    return '<ul class="wb-tree" role="tree">' +
      treeNode('hub', 'Hub', 'hub', () => hub.map((h) => treeNode('', h[1], h[2], null, '#/' + h[0])).join('')) +
      treeNode('areas', 'Areas', 'home', () => roomsOf().map((r) => treeNode('area:' + r.name, r.name, null, () => r.ids.filter((x, i, a) => a.indexOf(x) === i).map((id) => { const d = dev(id); return d ? treeNode('', d.friendly_name, devIcon(d), null, '#/device/' + encodeURIComponent(id), dot(d)) : ''; }).join(''), null, '<em>' + r.items.length + '</em>')).join('')) +
      treeNode('proto', 'By connection', 'zigbee', () => Object.keys(c.PROTOCOLS).map((k) => treeNode('proto:' + k, c.PROTOCOLS[k].label, c.PROTOCOLS[k].icon, () => S.devices.filter((d) => d.connection_type === k).map((d) => treeNode('', d.friendly_name, devIcon(d), null, '#/device/' + encodeURIComponent(d.id), dot(d))).join(''))).join('')) +
      treeNode('auto', 'Automations', 'automation', () => S.automations.map((a) => treeNode('', a.friendlyname + '.rule', a.enabled ? 'automation' : 'disabled', null, '#/automation/' + encodeURIComponent(a.id), '<em>' + a.triggers.length + '</em>')).join('') + '<li><button type="button" class="wb-tn is-leaf is-new" data-act="auto-create">' + I('plus') + '<span>New rule…</span></button></li>') +
      '</ul>';
  }

  function shell(view) {
    const h = hashNow();
    if (!S.tabs.some((t) => t === h)) { S.tabs.push(h); if (S.tabs.length > 7) S.tabs.shift(); }
    const on = S.devices.filter(isOnline).length;
    const act = [['explorer', 'Explorer', 'list', '#/overview'], ['rules', 'Automations', 'automation', '#/automations'], ['ask', 'Assistant', 'sparkles', '#/assistant'], ['parts', 'Components', 'components', '#/components'], ['all', 'All pages', 'stack', '#/all']];
    const logs = S.logs.slice(-40);
    return '<button type="button" class="nh-skip" data-act="skip">Skip to content</button><div class="wb-app' + (S.panel ? ' has-panel' : '') + '" data-nav-open="' + S.navOpen + '">' +
      '<nav class="wb-act" aria-label="Activity">' + '<span class="wb-logo">' + I('hub') + '</span>' + act.map((x) => '<a href="' + x[3] + '" title="' + x[1] + '" aria-label="' + x[1] + '">' + I(x[2]) + '</a>').join('') +
      '<span class="wb-act-gap"></span>' + btn('', 'wb-panel', { icon: 'terminal', title: (S.panel ? 'Hide' : 'Show') + ' log panel' }) + '<a href="#/settings" title="Settings" aria-label="Settings">' + I('cog') + '</a></nav>' +
      '<aside class="wb-side" aria-label="Explorer"><div class="wb-side-head"><span>Explorer</span>' + btn('', 'nav-close', { icon: 'close', title: 'Close explorer', sm: true }) + '</div>' + explorer() + '</aside><div class="nh-scrim" data-act="nav-close"></div>' +
      '<div class="wb-main"><div class="wb-tabs" role="tablist">' + btn('', 'nav-toggle', { icon: 'menu', title: 'Explorer', sm: true }) + S.tabs.map((t) => '<span class="wb-tab' + (t === h ? ' is-cur' : '') + '" role="tab" aria-selected="' + (t === h) + '"><a href="' + t + '">' + esc(titleFor(t)) + '</a><button type="button" data-act="wb-close" data-arg="' + esc(t) + '" aria-label="Close ' + esc(titleFor(t)) + '">' + I('close') + '</button></span>').join('') + '</div>' +
      c.banners() + '<main class="nh-content wb-content" id="nh-content" tabindex="-1">' + view + '</main>' +
      (S.panel ? '<section class="wb-panel" aria-label="Log"><div class="wb-panel-head"><b>Output</b><span>hub log</span><a href="#/console">Open console</a>' + btn('', 'wb-panel', { icon: 'close', sm: true, title: 'Hide panel' }) + '</div><div class="wb-panel-body" data-region="wblog">' + logs.map((l) => '<div class="lvl-' + l.level + '"><time>' + clock(l.timestamp) + '</time><span class="nh-lvl">' + l.level + '</span><span>' + esc(l.source) + '</span><span>' + esc(l.message) + '</span></div>').join('') + '</div></section>' : '') + '</div>' +
      '<footer class="wb-status"><button type="button" data-act="conn-cycle" class="wb-st is-' + S.conn + '">' + I(S.conn === 'connected' ? 'check' : S.conn === 'connecting' ? 'refresh' : 'offline') + esc(c.connText()) + '</button>' +
      '<a class="wb-st" href="#/list">' + I('devices') + on + '/' + S.devices.length + ' online</a>' +
      '<button type="button" class="wb-st' + (S.permit ? ' is-hot' : '') + '" data-act="' + (S.permit ? 'permit-stop' : 'permit') + '">' + I('join') + (S.permit ? 'Join open <span data-permit>' + c.fmtTimer(S.permit) + '</span> (stop)' : 'Join closed') + '</button>' +
      '<a class="wb-st" href="#/settings/mcp">' + I('stack') + 'MCP ' + (S.mcp.running ? S.mcp.connectedClients + ' client' : 'stopped') + '</a>' +
      '<span class="wb-st-gap"></span><button type="button" class="wb-st" data-act="mode-toggle">' + I(document.documentElement.dataset.mode === 'dark' ? 'sun' : 'moon') + 'Theme</button><button type="button" class="wb-st" data-act="pop">' + I('palette') + 'Appearance</button><span class="wb-st">v' + esc(c.DATA.version) + '</span><button type="button" class="wb-st" data-act="logout">' + I('logout') + 'Sign out</button></footer></div>' + c.overlays();
  }

  // property cell: the control for an expose, honouring pending vs confirmed
  function control(d, e) {
    const cap = capability(e), k = pkey(d.id, e.name), pend = S.pending[k], dis = !canControl(d);
    if (cap === 'switch') return toggle(e.data === e.values.on || e.data === true, 'expose-toggle', k, { disabled: dis || pend !== undefined, pending: pend !== undefined, label: meta(e.name).label }) + (pend !== undefined ? ' <small class="nh-muted">sending…</small>' : '');
    if (cap === 'slider') return '<input class="nh-slider wb-slider" type="range" min="' + e.attributes.value_min + '" max="' + e.attributes.value_max + '" value="' + (pend !== undefined ? pend : e.data) + '" data-bind="expose-slide" data-arg="' + esc(k) + '" aria-label="' + esc(meta(e.name).label) + '"' + (dis ? ' disabled' : '') + '><output>' + esc(fmt(e, pend !== undefined ? pend : e.data)) + '</output>';
    return '<span class="nh-muted">read only</span>';
  }
  const propHead = '<thead><tr><th scope="col">Property</th><th scope="col">Value</th><th scope="col">Type</th><th scope="col">Access</th><th scope="col">Category</th><th scope="col">Control</th></tr></thead>';
  const propRow = (d, e) => '<tr><td><span class="wb-prop kind-' + meta(e.name).kind + '">' + I(meta(e.name).icon) + esc(meta(e.name).label) + '</span><code>' + esc(e.name) + '</code></td><td class="wb-val">' + (S.pending[pkey(d.id, e.name)] !== undefined ? '<span class="nh-spin"></span>' : '') + esc(fmt(e)) + '</td><td><code>' + esc(e.type || 'text') + '</code></td><td><code>' + esc(e.access_mode) + '</code></td><td>' + esc(e.category) + '</td><td class="wb-ctl">' + control(d, e) + '</td></tr>';

  const views = {};
  views.overview = () => {
    const probs = [];
    S.devices.forEach((d) => {
      if (!isOnline(d)) probs.push(['error', d, 'Offline since ' + timeAgo(d.last_seen) + ' (availability timeout)']);
      if (isDisabled(d)) probs.push(['info', d, 'Disabled by device override']);
      if (lqiOf(d) != null && lqiOf(d) < 50) probs.push(['warn', d, 'Link quality ' + lqiOf(d) + ' is below 50']);
      if (battOf(d) != null && battOf(d) < 20) probs.push(['warn', d, 'Battery ' + Math.round(battOf(d)) + ' % is below 20 %']);
    });
    S.automations.forEach((a) => a.triggers.forEach((t, i) => { if (t.type === 'deviceTrigger' && !t.conditions.length) probs.push(['warn', null, a.friendlyname + ' trigger ' + (i + 1) + ' has no conditions; the hub will only run it manually']); }));
    const z = S.devices.filter((d) => d.connection_type === 'mqtt');
    const stats = [['Devices', S.devices.length], ['Online', S.devices.filter(isOnline).length], ['Zigbee routers', z.filter((d) => d.power_source === 'mains').length], ['Zigbee end devices', z.filter((d) => d.power_source === 'battery').length], ['Wi-Fi · HTTP', S.devices.length - z.length], ['Mean LQI', Math.round(z.map(lqiOf).filter((x) => x != null).reduce((a, b, _, r) => a + b / r.length, 0))], ['Automations', S.automations.length + ' (' + S.automations.filter((a) => a.enabled).length + ' enabled)']];
    return '<div class="wb-split"><section class="wb-box"><h1 class="wb-h">Problems <span class="wb-badge">' + probs.length + '</span></h1><table class="nh-table wb-grid"><thead><tr><th scope="col">Severity</th><th scope="col">Source</th><th scope="col">Message</th></tr></thead><tbody>' +
      probs.map((p) => '<tr><td><span class="wb-sev is-' + p[0] + '">' + I(p[0] === 'error' ? 'error' : p[0] === 'warn' ? 'warn' : 'info') + p[0] + '</span></td><td>' + (p[1] ? '<a href="#/device/' + encodeURIComponent(p[1].id) + '">' + esc(p[1].friendly_name) + '</a>' : 'automations') + '</td><td>' + esc(p[2]) + '</td></tr>').join('') + '</tbody></table></section>' +
      '<section class="wb-box"><h2 class="wb-h">Statistics</h2><table class="nh-table wb-grid"><tbody>' + stats.map((s) => '<tr><th scope="row">' + s[0] + '</th><td class="wb-num">' + s[1] + '</td></tr>').join('') + '</tbody></table></section></div>' +
      '<section class="wb-box"><h2 class="wb-h">Recent changes</h2><table class="nh-table wb-grid"><thead><tr><th scope="col">Time</th><th scope="col">Device</th><th scope="col">Property</th><th scope="col">From</th><th scope="col">To</th></tr></thead><tbody data-region="feed">' +
      S.feed.slice(0, 14).flatMap((f) => f.changes.map((ch) => '<tr><td class="wb-num">' + clock(f.t) + '</td><td><a href="#/device/' + encodeURIComponent(f.id) + '">' + esc(f.name) + '</a></td><td><code>' + esc(ch.n) + '</code></td><td class="nh-muted">' + esc(ch.from) + '</td><td><b>' + esc(ch.to) + '</b></td></tr>')).join('') + '</tbody></table></section>';
  };
  views.home = () => '<h1 class="wb-h">Areas</h1><p class="nh-sub">Dashboard groups as property grids. Edit groups in Settings or on the Hearth-style home.</p>' + roomsOf().map((r) => '<section class="wb-box"><h2 class="wb-h">' + I('home') + esc(r.name) + ' <span class="wb-badge">' + r.items.length + '</span></h2><table class="nh-table wb-grid"><thead><tr><th scope="col">Entity</th><th scope="col">Device</th><th scope="col">Value</th><th scope="col">Updated</th><th scope="col">Control</th></tr></thead><tbody>' +
    r.items.map((it) => { const d = dev(it[0]), e = d && d.exposes[it[1]]; if (!e) return ''; return '<tr><td><span class="wb-prop kind-' + meta(e.name).kind + '">' + I(meta(e.name).icon) + esc(e.description || meta(e.name).label) + '</span></td><td>' + dot(d) + ' <a href="#/device/' + encodeURIComponent(d.id) + '">' + esc(d.friendly_name) + '</a></td><td class="wb-val">' + esc(isOnline(d) ? fmt(e) : 'offline') + '</td><td class="nh-muted">' + esc(timeAgo(d.last_seen)) + '</td><td class="wb-ctl">' + (capability(e) === 'switch' || capability(e) === 'slider' ? control(d, e) : (c.stateExposeOf(d) ? control(d, c.stateExposeOf(d)) : '<span class="nh-muted">—</span>')) + '</td></tr>'; }).join('') + '</tbody></table></section>').join('');
  views.devices = () => '<h1 class="wb-h">All properties</h1><p class="nh-sub">Every device and every exposed property. Expand a device to edit.</p><table class="nh-table wb-grid wb-all">' + propHead + S.devices.map((d) => {
    const open = S.tree['dev:' + d.id] !== false;
    return '<tbody><tr class="wb-grp"><th colspan="6" scope="rowgroup"><button type="button" data-act="wb-tree" data-arg="dev:' + esc(d.id) + '" aria-expanded="' + open + '">' + I(open ? 'chevronD' : 'chevronR') + dot(d) + '<b>' + esc(d.friendly_name) + '</b></button> <code>' + esc(d.id) + '</code> ' + protoChip(d) + (signalChip(d) || '') + powerChip(d) + ' <a href="#/device/' + encodeURIComponent(d.id) + '">open</a></th></tr>' +
      (open ? Object.values(d.exposes).map((e) => propRow(d, e)).join('') : '') + '</tbody>';
  }).join('') + '</table>';
  views.device = (r) => {
    const d = dev(r.a); if (!d) return empty('question', 'Device not found', 'It may have been removed or renamed.');
    const sup = protocol(d.connection_type).supports, tabs = [['props', 'Properties'], ['settings', 'Settings'], ['metrics', 'History'], ['about', 'About'], ['raw', 'Raw JSON']];
    const body = S.devTab === 'props' ? '<table class="nh-table wb-grid">' + propHead + '<tbody>' + Object.values(d.exposes).map((e) => propRow(d, e)).join('') + '</tbody></table>' :
      S.devTab === 'raw' ? '<pre class="nh-json wb-code">' + esc(JSON.stringify(d, null, 2)) + '</pre>' : c.DV[S.devTab](d);
    return '<div class="wb-devhead">' + I(devIcon(d)) + '<div><h1 class="wb-h">' + esc(d.friendly_name) + '</h1><p class="nh-sub"><code>' + esc(d.id) + '</code> ' + esc(d.description) + '</p><div class="nh-chips">' + statusChip(d) + protoChip(d) + signalChip(d) + powerChip(d) + chip(timeAgo(d.last_seen), 'muted', 'clock') + '</div></div><div class="nh-page-actions">' +
      btn('Rename', 'dev-rename', { sm: true, icon: 'edit', arg: d.id, disabled: !sup.rename }) + btn('Interview', 'dev-interview', { sm: true, icon: 'refresh', arg: d.id, disabled: !sup.interview }) + btn('Remove', 'dev-remove', { sm: true, icon: 'trash', kind: 'danger', arg: d.id, disabled: !sup.remove }) + '</div></div>' +
      '<div class="wb-subtabs" role="tablist">' + tabs.map((t) => '<button type="button" role="tab" aria-selected="' + (S.devTab === t[0]) + '" data-act="wb-devtab" data-arg="' + t[0] + '">' + t[1] + '</button>').join('') + '</div>' + body;
  };
  views.automations = () => '<div class="wb-toolbar"><h1 class="wb-h">Automations</h1>' + btn('New rule', 'auto-create', { icon: 'plus', sm: true, kind: 'primary' }) + '</div><table class="nh-table wb-grid"><thead><tr><th scope="col">Rule</th><th scope="col">Source device</th><th scope="col">Triggers</th><th scope="col">First trigger</th><th scope="col">Status</th><th scope="col">Enabled</th><th scope="col"><span class="sr-only">Actions</span></th></tr></thead><tbody>' +
    S.automations.map((a, i) => { const st = autoStatus(a), m = a.triggers.findIndex(canManual); return '<tr><td><a href="#/automation/' + encodeURIComponent(a.id) + '"><b>' + esc(a.friendlyname) + '.rule</b></a><br><small class="nh-muted">' + esc(a.description) + '</small></td><td><code>' + esc(a.id) + '</code></td><td class="wb-num">' + a.triggers.length + '</td><td class="wb-sum">' + esc(describeTrigger(a, a.triggers[0])) + '</td><td>' + chip(st.text, st.kind, st.icon, st.title) + '</td><td>' + toggle(a.enabled, 'auto-enable', i, { label: 'Enabled' }) + '</td><td class="nh-row-actions">' + (m >= 0 ? btn('', 'auto-run', { icon: 'play', sm: true, title: 'Run', arg: i + '|' + m }) : '') + btn('', 'auto-delete', { icon: 'trash', sm: true, kind: 'danger', title: 'Delete', arg: i }) + '</td></tr>'; }).join('') + '</tbody></table>';

  // Editor: outline tree | inspector | live payload
  views.automation = (r) => {
    c.V.automation(r);
    if (!S.draft) return empty('question', 'Rule not found', '');
    const P = c.edParts(), a = P.a;
    const n = (key, label, icon, cnt, err) => '<li><button type="button" class="wb-tn is-leaf' + (S.edSel === key ? ' is-cur' : '') + '" data-act="wb-edsel" data-arg="' + key + '">' + I(icon) + '<span>' + label + '</span>' + (cnt != null ? '<em>' + cnt + '</em>' : '') + (err ? I('error', 'wb-err') : '') + '</button></li>';
    const outline = '<ul class="wb-tree">' + n('details', 'Details', 'info') + n('schedule', 'Schedule', 'calendar', a.schedules.length) +
      a.triggers.map((t, i) => { const errs = P.trigErrs(i); return '<li><span class="wb-tn is-grp">' + I('automation') + '<span>Trigger ' + (i + 1) + '</span></span><ul>' + n('t' + i + '.when', 'When: ' + (t.name ? meta(t.name).label : '…'), 'play', null, errs.some((e) => /\.name$/.test(e.p) && e.p.split('.').length === 3)) + n('t' + i + '.if', 'If', 'filter', t.conditions.length, errs.some((e) => e.p.indexOf('.conditions') > 0)) + n('t' + i + '.then', 'Then', 'bolt', t.actions.length, errs.some((e) => e.p.indexOf('.actions') > 0)) + '</ul></li>'; }).join('') +
      '<li><button type="button" class="wb-tn is-leaf is-new" data-act="ed-add-trig">' + I('plus') + '<span>Add trigger</span></button></li></ul>';
    const m = /^t(\d+)\.(when|if|then)$/.exec(S.edSel), i = m ? +m[1] : -1;
    let insp;
    if (m && !a.triggers[i]) S.edSel = 'details';
    if (S.edSel === 'schedule') insp = '<h2 class="wb-h">Schedule</h2>' + P.schedule();
    else if (m && a.triggers[i]) {
      const t = a.triggers[i];
      insp = '<div class="wb-insp-head"><h2 class="wb-h">Trigger ' + (i + 1) + ' · ' + { when: 'When', if: 'If', then: 'Then' }[m[2]] + '</h2>' + P.trigTools(i) + '</div>' + P.sentence(i) +
        (m[2] === 'when' ? P.whenBody(i) : m[2] === 'if' ? (t.conditions.map((x, j) => P.condRow(i, j)).join('') || '<p class="nh-muted">No conditions.</p>') + P.condAdd(i) : t.actions.map((x, k) => P.actRow(i, k)).join('') + P.err(P.tp(i) + '.actions') + P.actAdd(i)) + P.warnList(i);
    } else insp = '<h2 class="wb-h">Details</h2>' + P.details();
    const probs = P.v.errs.map((e) => '<li class="is-error">' + I('error') + '<code>' + esc(e.p) + '</code> ' + esc(e.m) + '</li>').concat(P.v.warns.map((w) => '<li class="is-warn">' + I('warn') + '<code>' + esc(w.p) + '</code> ' + esc(w.m) + '</li>')).join('');
    return '<div class="wb-ed"><aside class="wb-ed-tree"><div class="wb-side-head"><span>' + esc(a.friendlyname) + '.rule</span></div>' + outline + '</aside>' +
      '<section class="wb-ed-insp">' + insp + '<h3 class="wb-h3">Problems</h3><ul class="wb-probs">' + (probs || '<li class="is-ok">' + I('check') + 'No problems</li>') + '</ul>' + P.savebar() + '</section>' +
      '<section class="wb-ed-code"><div class="wb-side-head"><span>payload.json</span><small>read-only, sent as saveAutomation</small></div><pre class="nh-json wb-code" data-region="json">' + esc(JSON.stringify(a, null, 2)) + '</pre></section></div>';
  };

  const actions = {
    'wb-tree': (k) => { S.tree[k] = S.tree[k] === undefined ? !k.startsWith('dev:') ? true : false : !S.tree[k]; if (k.startsWith('dev:') && S.tree[k] === undefined) S.tree[k] = false; c.render(); },
    'wb-close': (h) => { S.tabs = S.tabs.filter((t) => t !== h); if (location.hash === h) c.go(S.tabs[S.tabs.length - 1] || '#/overview'); else c.render(); },
    'wb-panel': () => { S.panel = !S.panel; c.render(); },
    'wb-devtab': (t) => { S.devTab = t; c.render(); },
    'wb-edsel': (k) => { S.edSel = k; c.render(); },
  };
  // fix: device groups default open → first click closes
  actions['wb-tree'] = (k) => { const def = k.startsWith('dev:') ? true : false; const cur = S.tree[k] === undefined ? def : S.tree[k]; S.tree[k] = !cur; c.render(); };
  return { shell, views, actions, live: ['automations'] };
};
