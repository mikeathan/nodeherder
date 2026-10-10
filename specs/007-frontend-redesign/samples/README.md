# Redesign design samples (spec 007)

## Quick view — single files
`standalone/` has **one self-contained HTML file per design** (all CSS/JS inlined, ~200 KB).
Download or open any of them on its own — no other files needed:

| File | Design |
| --- | --- |
| `standalone/hearth.html` | 1. Hearth — Lovelace-inspired |
| `standalone/mesh.html` | 2. Mesh — Zigbee2MQTT-inspired |
| `standalone/phosphor.html` | 3. Phosphor — 80s terminal |
| `standalone/studio.html` | 4. Studio — Braun/Rams hardware |
| `standalone/platinum.html` | 5. Platinum — 90s desktop OS |
| `standalone/index.html` | Gallery linking the five |

Each opens on **All pages** (`#/all`): every screen stacked on one scroll — overview, home,
devices, device list, device page (about/controls/settings/metrics), automations, automation
editor, assistant, console, settings (appearance/defaults/MCP), components & states, login.
Use the navigation to open any page on its own; the palette icon opens the customiser.

## Source
- `hearth.html` … `platinum.html`, `index.html` — same samples using the shared files.
- `shared/` — mock data mirroring real contracts, icons, renderer, structural CSS.
- `designs/` — one skin (tokens + chrome) per design.
- `build-standalone.py` — regenerates `standalone/` after editing the source.

Fonts load from Google Fonts when online; system fonts are used offline. These are review
prototypes: not built, linted or shipped, and they never contact a hub.
