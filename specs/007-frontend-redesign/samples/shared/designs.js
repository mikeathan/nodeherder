/* Registry of the five sample designs (used by the gallery and the in-page design switcher). */
window.NH_DESIGNS = [
  { n: 1, id: 'hearth', name: 'Hearth', file: 'hearth.html', short: 'Lovelace-inspired · warm, rounded, friendly',
    blurb: 'Home Assistant–style areas and tiles with a soft 70s living-room palette. Sidebar + top bar. Best for wall tablets and family use.',
    shell: 'sidebar', defaults: { mode: 'system', accent: 'preset', density: 'comfortable', radius: 14, fontScale: 100, preset: 'daylight', effects: true } },
  { n: 2, id: 'mesh', name: 'Mesh', file: 'mesh.html', short: 'Zigbee2MQTT-inspired · dense, data-first',
    blurb: 'Operations console: stat strip, activity feed, dashed device cards, compact tables with LQI/battery chips. Best for power users.',
    shell: 'sidebar', defaults: { mode: 'dark', accent: 'preset', density: 'compact', radius: 6, fontScale: 100, preset: 'night', effects: true } },
  { n: 3, id: 'phosphor', name: 'Phosphor', file: 'phosphor.html', short: '80s terminal · monospace, phosphor glow',
    blurb: 'CRT terminal look with function-key tabs, inverse-video tiles and ASCII framing. Amber, green or ice phosphor; light mode is green-bar paper.',
    shell: 'tabs', defaults: { mode: 'dark', accent: 'preset', density: 'comfortable', radius: 0, fontScale: 105, preset: 'amber', effects: true } },
  { n: 4, id: 'studio', name: 'Studio', file: 'studio.html', short: 'Braun/Rams hardware · warm neutrals, orange',
    blurb: 'Industrial-design calm: hardware rocker switches, inset LCD readouts, rail navigation that becomes a bottom dock on phones.',
    shell: 'dock', defaults: { mode: 'light', accent: 'preset', density: 'comfortable', radius: 10, fontScale: 100, preset: 'braun', effects: true } },
  { n: 5, id: 'platinum', name: 'Platinum', file: 'platinum.html', short: '90s desktop OS · bevels, title bars',
    blurb: 'Classic desktop windows: menu bar navigation, striped title bars, bevelled controls and pinstripe desktop. Nostalgic but crisp.',
    shell: 'menubar', defaults: { mode: 'light', accent: 'preset', density: 'comfortable', radius: 3, fontScale: 100, preset: 'platinum', effects: true } },
];
window.NH_DESIGN_FOR = function (id, extra) {
  const d = window.NH_DESIGNS.find((x) => x.id === id);
  return Object.assign({}, d, extra);
};
