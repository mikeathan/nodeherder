# Redesign design samples (spec 007)

## Quick view — single files
`standalone/` has **one self-contained HTML file per design** (all CSS/JS inlined, ~200 KB).
Download or open any of them on its own — no other files needed:

| File | Design |
| --- | --- |
| `standalone/hearthpanel.html` | 1. Hearth + Panel — recommended merge (Hearth app + wall-panel mode, logo) |
| `standalone/hearth.html` | 2. Hearth — areas of tiles, sidebar |
| `standalone/deck.html` | 3. Deck — full wall-tablet layout, kept for comparison |
| `standalone/index.html` | Gallery linking the five |

Each opens on **All pages** (`#/all`): every screen stacked on one scroll — overview, home,
devices, device list, device page (about/controls/settings/metrics), automations, automation
editor, assistant, console, settings (appearance/defaults/MCP), components & states, login.
Use the navigation to open any page on its own; the palette icon opens the customiser.

## Source
- `hearthpanel.html`, `hearth.html`, `deck.html`, `index.html` — same samples using the shared files.
- `shared/themes.css` — the ten themes shared by all designs; `shared/logo.js` — embedded logo.
- `designs/<id>.js` — each design's own shell, page layouts and editor (Hearth uses the shared defaults).
- `make-pages.py` — regenerates the five HTML pages.
- `shared/` — mock data mirroring real contracts, icons, renderer, structural CSS.
- `designs/<id>.css` — each design's tokens and styles.
- `build-standalone.py` — regenerates `standalone/` after editing the source.

Fonts load from Google Fonts when online; system fonts are used offline. These are review
prototypes: not built, linted or shipped, and they never contact a hub.
