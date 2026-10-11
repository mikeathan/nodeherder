/* Registry of the sample designs (gallery + in-page design switcher). Round 3: Hearth + Panel (merged), Hearth and Deck — see ../design-directions.md. */
window.NH_DESIGNS = [
  { n: 1, id: 'hearthpanel', name: 'Hearth + Panel', file: 'hearthpanel.html', short: 'Recommended · Hearth app with a wall-panel mode',
    blurb: 'Hearth for every page (with the NodeHerder logo and a phone-friendly automation editor) plus Deck’s big-dial look as a full-screen panel mode for a wall tablet. Recent activity can be switched off.',
    shell: 'sidebar', defaults: { mode: 'system', accent: 'preset', density: 'comfortable', radius: 14, fontScale: 100, preset: 'fleece', effects: true } },
  { n: 2, id: 'hearth', name: 'Hearth', file: 'hearth.html', short: 'Areas of soft tiles · sidebar',
    blurb: 'Home Assistant–style areas and tiles. Sidebar and top bar; When / If / Then flow editor (now phone-friendly).',
    shell: 'sidebar', defaults: { mode: 'system', accent: 'preset', density: 'comfortable', radius: 14, fontScale: 100, preset: 'daylight', effects: true } },
  { n: 3, id: 'deck', name: 'Deck', file: 'deck.html', short: 'Wall-tablet panel · big dials · room pager',
    blurb: 'The full Deck layout from round 2, for comparison. Its look lives on as panel mode in Hearth + Panel.',
    shell: 'custom', defaults: { mode: 'dark', accent: 'preset', density: 'spacious', radius: 22, fontScale: 100, preset: 'lagoon', effects: true } },
];
window.NH_DESIGN_FOR = function (id, extra) {
  const d = window.NH_DESIGNS.find((x) => x.id === id);
  return Object.assign({}, d, extra);
};
