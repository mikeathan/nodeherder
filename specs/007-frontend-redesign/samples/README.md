# Redesign design samples (spec 007)

Open `index.html` in a browser (double-click works; no build or server needed). Fonts load from
Google Fonts when online and fall back to system fonts offline.

- `hearth.html`, `mesh.html`, `phosphor.html`, `studio.html`, `platinum.html` — the five designs.
- `shared/` — mock data mirroring the real contracts, icons, renderer, structural CSS.
- `designs/` — one skin (tokens + chrome) per design.

These are review prototypes, not production code: they are not built, linted or shipped, and they
never contact a hub. See [plan.md](../plan.md) for how the chosen design becomes Vue components.
