# Spec: History preset, sidebar simplification and appearance palette

ID: 031 | Status: Draft | Owner: NodeHerder contributors | Updated: 2026-10-23
Constitution: proposed 1.0.0 | Components: frontend

## Problem, scope, exclusions

Three small, independent UI requests in one branch:

1. The history Period dropdown jumps from `24h` to `7d`; users wanting "the last few days"
   must use Custom bounds. Add a rolling 3-day preset.
2. The sidebar shows both **Devices** (cards) and **Device list** (table). Both pages already
   cross-link ("Device list" / "Cards" header buttons), and Overview tiles and `DevicePage`
   target the table, so one sidebar entry is enough. Remove the duplicate entry.
3. The appearance picker repeats a near-identical blue (`#0f8bd6` next to `#1b7fb0`). Drop the
   duplicate. Of the retro themes tried in review, keep only **Terminal Amber**; the user rejected
   **Teak, Avocado, Plum, Sunset Strip and Diner Mint**, so those palettes are removed outright
   (including the pre-existing Teak/Avocado/Plum). Replacement themes **Crimson**, **Emerald** and
   **Blueprint** plus accents `#c92a2a`, `#2b4fd8`, `#495057` were chosen in review; brown
   `#8a5a2b` and mustard `#c08a1e` are not offered.

Scope: `frontend/` only. Exclusions: no backend/WS/MQTT/contract change, no query or aggregation
change, no new persistence keys, no change to history defaults (`24h` stays the default).

Supersedes the "eight presets" wording of spec 030 AC-01/AC-04 (now nine) and the
"every route in the drawer" expectation of 030 T007 (the Device list route is retained but no
longer a sidebar destination). Route names, permissions and deep links are unchanged.

## Scenarios and acceptance

### US-01 — Pick a 3-day window (P1)

Independent verification: Jest range tests + HistoryToolbar render.

- AC-01: Given the history toolbar, when the Period picker opens, then it lists `1h, 6h, 24h, 3d,
  7d, 30d, Today, Yesterday, Custom` in that order.
- AC-02: Given `3d` is selected, when history loads, then the request window is `now − 72h … now`,
  the CSV export covers the same window, and the choice survives reload via
  `nh-history-pref:<deviceId>`.
- AC-03: Given a stored preference of a preset that no longer exists, when history loads, then the
  existing fallback (`24`) is used and no error surfaces.

### US-02 — One sidebar entry for devices (P2)

Independent verification: `e2e/layout.spec.ts`.

- AC-04: Given the navigation drawer, then "Devices" is a visible destination and "Device list" is
  not, at 320/390/768px.
- AC-05: Given `/devicelist` is open (reached from the Devices header button, an Overview tile or a
  deep link), then the "Devices" nav row is highlighted as active and the header title still reads
  "Device list".
- AC-06: Given keyboard/screen-reader use, then no navigation destination has disappeared from the
  tab order apart from the removed duplicate, and the remaining entries keep accessible names,
  focus rings and 44px targets on phones.

### US-03 — Retro palette without duplicates (P3)

Independent verification: Jest theme-settings tests + manual picker review.

- AC-07: Given the accent picker, then it offers ten swatches with no two visually
  indistinguishable: `#1b7fb0, #0e7c86, #2b8a3e, #8e3b7f, #7048e8, #c2255c, #d9480f, #c92a2a,
  #2b4fd8, #495057`; `#0f8bd6`, `#8a5a2b` and `#c08a1e` are gone; the Custom colour input still
  accepts any hex and a previously stored `#0f8bd6` still renders.
- AC-08: Given the theme picker, then it lists exactly Fleece, Daylight, Lagoon, Ember, Frost,
  Fjord, Graphite, Terminal Amber, Crimson, Emerald and Blueprint, and selecting any applies its
  palette immediately.
- AC-09: Given Terminal Amber, Crimson, Emerald or Blueprint, when mode is Light or Dark, then every
  themed token (`--nh-bg`, `--nh-surface`, `--nh-surface-2`, `--nh-border`, `--nh-border-strong`,
  `--nh-text`, `--nh-text-muted`, `--nh-accent`, `--nh-accent-contrast`, `--nh-nav-bg`,
  `--nh-header-bg`, `--dk-lamp`, `--dk-climate`, `--dk-glass`) resolves, and text/button contrast
  stays readable.
- AC-10: Given a stored theme id of `teak`, `avocado`, `plum`, `sunset` or `diner`, then the theme
  falls back to Fleece without an error, and no orphaned palette blocks remain in `themes.css`.

## Requirements and success

- FR-01: MUST add a rolling 72-hour preset labelled `3d` between `24h` and `7d`; acceptance: AC-01…03.
- FR-02: MUST remove the Device list sidebar entry while keeping the route reachable and the
  Devices entry active on `/devicelist` and `/device/:id`; acceptance: AC-04…06.
- FR-03: MUST replace `#0f8bd6` with `#c92a2a`, `#2b4fd8` and `#495057`, and MUST NOT offer
  `#8a5a2b` or `#c08a1e`; acceptance: AC-07.
- FR-04: MUST add `amber`, `crimson`, `emerald` and `blueprint` (light + dark each) and MUST remove
  `teak`, `avocado`, `plum`, `sunset` and `diner` from both `THEMES` and `themes.css`;
  acceptance: AC-08…10.
- NFR-01: Palette work is presentation only — no household data, no new storage keys, no network
  calls (FE-05).
- NFR-02: All new palette values MUST keep body text ≥ 4.5:1 against its own surface and
  `--nh-accent-contrast` readable on `--nh-accent` in both modes (FE-04).
- SC-01: Period dropdown shows nine options; sidebar shows seven destinations plus Panel mode;
  accent picker shows ten swatches and the theme picker eleven themes (verified by tests).

## Entities and external interfaces

- `HISTORY_PRESETS` (`src/domain/history/range.ts`) — single source for the picker, preference
  validation and export window. `value` remains an hour count for rolling presets, so `'72'`
  needs no special casing.
- `NAV_ITEMS` (`src/components/shell/navigation.ts`) — single source for sidebar, keyboard order
  and active-state matching via `includes`.
- `THEMES` / `ACCENTS` (`src/theme/theme-settings.ts`) — ids/names/swatches only; hex values live
  in `src/assets/styles/themes.css` (existing rule in that file header).
- No API/WS/MQTT/assistant contract touched.

## Questions and assumptions

- Q1: Should light variants of the retro themes be cool/mint or warm sand? Resolution: warm sand
  for Sunset Strip and Terminal Amber (70s/CRT), cool mint-grey for Diner Mint — matches their
  dark counterparts (confirmed by user: option (a), themes keep light+dark).
- Q2: Should the removed Device list nav item be replaced by anything? Assumption: no; the page
  stays reachable from three existing entry points.
- Q3: Which themes should replace the five rejected palettes, and which hue replaces the brown
  accent? Resolved in review: **Crimson**, **Emerald**, **Blueprint** plus accents `#c92a2a`,
  `#2b4fd8`, `#495057`; mustard and brown accents are removed.
- Assumption: existing `24h` default, `31d` max and 5-minute minimum history bounds are unchanged.
- Assumption: removing themes is acceptable loss for users who had them stored; `parseThemeSettings`
  already falls back to Fleece, and no migration is added.
