/* Registry of the sample designs (gallery + in-page design switcher). Round 2: Hearth kept,
   four structurally different directions added — see ../design-directions.md. */
window.NH_DESIGNS = [
  { n: 1, id: 'hearth', name: 'Hearth', file: 'hearth.html', short: 'Areas of soft tiles · sidebar',
    blurb: 'Home Assistant–style areas and tiles with a warm palette. Sidebar and top bar; When / If / Then flow editor.',
    shell: 'sidebar', defaults: { mode: 'system', accent: 'preset', density: 'comfortable', radius: 14, fontScale: 100, preset: 'daylight', effects: true } },
  { n: 2, id: 'floorplan', name: 'Floorplan', file: 'floorplan.html', short: 'The house plan is the home page',
    blurb: 'Rooms drawn as a floor plan with switch pins; a room inspector slides in. Faceted device finder, rules as device-to-device diagrams, step-by-step wizard editor.',
    shell: 'custom', defaults: { mode: 'light', accent: 'preset', density: 'comfortable', radius: 8, fontScale: 100, preset: 'blueprint', effects: true } },
  { n: 3, id: 'workbench', name: 'Workbench', file: 'workbench.html', short: 'IDE-style explorer, tabs and inspector',
    blurb: 'For the owner who configures everything: explorer tree, open tabs, property grids with type and access mode, status bar and a live log panel. Outline + inspector + JSON editor.',
    shell: 'custom', defaults: { mode: 'dark', accent: 'preset', density: 'compact', radius: 4, fontScale: 100, preset: 'graphite', effects: false } },
  { n: 4, id: 'brief', name: 'Brief', file: 'brief.html', short: 'The house written as sentences · command bar',
    blurb: 'One calm column. Home reads like a note about the house, with inline switches; Ctrl K finds any device, room or action. Rules are edited as fill-in-the-blanks sentences.',
    shell: 'custom', defaults: { mode: 'light', accent: 'preset', density: 'comfortable', radius: 10, fontScale: 105, preset: 'paper', effects: false } },
  { n: 5, id: 'deck', name: 'Deck', file: 'deck.html', short: 'Wall-tablet panel · big dials · room pager',
    blurb: 'Readable from across the room: clock strip, one room per page, temperature dial and big light buttons, bottom dock. Rules are a node canvas with an inspector drawer.',
    shell: 'custom', defaults: { mode: 'dark', accent: 'preset', density: 'spacious', radius: 22, fontScale: 100, preset: 'lagoon', effects: true } },
];
window.NH_DESIGN_FOR = function (id, extra) {
  const d = window.NH_DESIGNS.find((x) => x.id === id);
  return Object.assign({}, d, extra);
};
