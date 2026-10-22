# Spec: History, charts and mobile shell

ID: 029 | Status: Draft | Owner: NodeHerder contributors | Updated: 2026-10-22
Constitution: proposed 1.0.0 | Components: both (frontend-led; no backend change in scope)

## Problem, scope, exclusions

History is hard to reach and hard to read. Verified in code:

- `components/device/DeviceMetrics.vue` puts the entity and period in two `<select>`s inside a
  device tab; only one expose can be shown; the chart is 235&nbsp;px inside a 64&nbsp;rem column.
- `components/chart/history/options.ts` forces exactly three x-axis ticks and there is **no zoom
  or pan**: the only chart dependency is `chart.js`, with no zoom plugin.
- `components/dialogs/EntityViewDialog.vue` shows a fixed "Last 4 hours" with no route to full
  history, so the sheet is a dead end.
- `domain/history/range.ts` supports `1h…30d`, `today` and `custom` but not **Yesterday**, and the
  display zone is always the browser's.
- Below 960&nbsp;px navigation collapses into a hamburger drawer; `--nh-bottom-nav-h` exists but is
  unused; `index.html` lacks `viewport-fit=cover`, so all three `env(safe-area-inset-*)` uses are
  zero; there is no web app manifest.
- Statistics are unavailable: the backend aggregates a whole window (`min/avg/max/count/last`), and
  no aggregation is requested by the browser, so long ranges are raw-sample limited (10&nbsp;000
  points, 31 days).

Deliver one coherent history experience: **focused and thumb-first on phones, full and detailed on
desktop**, on a shared chart component, without regressing the truthfulness rules established by
spec 007 (gaps stay gaps, unknown is not zero, partial data is labelled) and without touching device
commands, protocols, persistence or auth.

Design evidence: spec 007's follow-up design ([entities-and-history.md](../007-frontend-redesign/entities-and-history.md),
[history-contract.md](../007-frontend-redesign/history-contract.md), [ADR-004](../007-frontend-redesign/adr-004-history-and-quality.md))
and this spec's [ADR-006](adr-006-chart-interactions.md).

**In scope:** history toolbar and entity selection, chart sizing and interactions incl. touch and
keyboard, statistics mode with export, entity sheet → full history, mobile bottom navigation, safe
areas and web app manifest, retirement of the stale `charts_todo.md`, and closing spec 007
T209–T211 as superseded, with T212 and T213 left open for their non-history obligations.

**Out of scope, with reason:** cross-device comparison (the response contract rejects multi-device
groups: one device per series), automation-run markers (needs new backend recording), server-side
bucketed aggregation (see plan; client-side for this change), chart screenshot sharing, PWA offline
service worker, tablet split view, mixed-unit dual axis, and any change to stored history,
retention, MQTT or automations.

## Scenarios and acceptance

### US-01 — Find and read history without hunting (P1)

Independent verification: Jest for domain/state, Playwright for the flows at 1280×800 and 390×844.

- AC-01: Given a device with chartable exposes, when History opens on a ≥1024&nbsp;px viewport,
  then range presets (`1h`, `6h`, `24h`, `7d`, `30d`, `Today`, `Yesterday`, `Custom`) and every
  chartable expose are visible as labelled controls with `aria-pressed`, and the current range is
  readable without opening a `<select>`.
- AC-02: Changing the preset or expose updates the chart and the stats strip with no full-page
  reload; returning to the same device restores the last preset and expose, and the stored choice is
  ignored if that expose no longer exists. A persisted `Custom` range stores its validated bounds and
  restores them; if those bounds are missing, unparsable, out of order or longer than 31 days, the
  view falls back to `24h` and keeps the stored expose instead of showing an invalid range.
- AC-03: On ≥1024&nbsp;px the history chart is at least 360&nbsp;px tall and uses the available
  content width (the history route is not capped at 64&nbsp;rem) while keeping y-axis labels with
  the unit and truthful gaps.

### US-02 — Control the chart with any input (P1)

- AC-04: Given a rendered chart, zoom in, zoom out and reset are reachable by (a) buttons of at
  least 44×44&nbsp;px on touch layouts, (b) two-finger pinch, (c) mouse drag-select, (d) wheel, and
  (e) `+` `-` `0` and arrow keys when the chart has focus; each path applies the same range rules
  and clamps to between five minutes and 31 days.
- AC-05: The chart exposes an accessible name and the current range, has a visible focus ring, and
  one-finger drag on touch moves a readout showing time and value without changing zoom.
- AC-06: Zooming, panning or scrubbing never scrolls the page on touch (`touch-action` handles the
  gesture) and never sends a device command or a history request beyond the existing queue.

### US-03 — Read the same data two ways (P1)

- AC-07: A History/Statistics switch is present **for numeric series only**; Statistics shows per-day
  `min`, `average`, `max`, `last` and sample count computed from the loaded **raw decoded samples**
  (never the decimated render set), drawn as bars, and states how many samples it used. Binary and
  enum/event series keep their state timeline and recent-event list with no Statistics switch, and
  energy counters keep their existing observed-consumption summary instead of daily statistics.
- AC-08: Statistics buckets follow the browser's named zone and are DST-correct. There is exactly one
  bucket per intersecting local day, enumerated with no fixed cap: a valid 31-day range can touch
  **33** local dates across a DST transition (`Europe/London`, 1 Mar 2026 23:30 GMT → 2 Apr 2026
  00:30 BST, 744 h). Ownership is `[midnight, next midnight)`, a sample at exactly midnight belongs to
  the day starting there, and the first/last buckets are marked partial when the range does not start
  or end at midnight, so no sample is dropped or counted twice. An empty bucket is a gap (not a zero).
  A partial series states the window it actually covers and that newer samples may be missing, keeps
  uncovered buckets empty, disables export, and offers a one-tap narrower range instead of changing
  the user's selection by itself.
- AC-09: CSV export follows the **selected** range (preset or custom), not the zoomed viewport, and
  uses the schema of the active mode: history = `timestamp,value,unit`; statistics =
  `bucket_start,bucket_end,min,average,max,last,samples,unit`. Timestamps are ISO-8601 with offset,
  gaps and empty buckets leave value cells empty (never `0`), and export is unavailable with an
  explanation while the series is partial or empty.

### US-04 — Use it one-handed on a phone (P1)

- AC-10: Below 720&nbsp;px a bottom bar offers the four primary destinations that already exist as
  routes — Home, Devices, Automations, Assistant — plus **More**, which opens the existing drawer
  (Overview, Device list, Console, Settings, Panel mode, permit join, sign out). History is
  deliberately not a destination: it is device-scoped and reached from a device or the entity sheet.
  Tabs have 44&nbsp;px targets, safe-area padding and no horizontal page scroll; a tab is
  `aria-current="page"` when the route matches it or its `includes` list, More is marked current when
  the active route is not one of the four tabs and reports `aria-expanded`.
- AC-11: The entity sheet shows the confirmed value, one primary control and a compact chart; "View
  full history" opens `/devicepage/:id?tab=history&expose=<name>`, where the URL expose wins over the
  saved preference (which in turn wins over the first chartable expose); the sheet closes by drag
  (more than 120&nbsp;px from the top), Escape, or its close control, and its actions are never
  hidden behind the safe area.
- AC-12: `viewport-fit=cover` is set, safe-area insets are either applied or explicitly zero, and
  the app declares a manifest (name, short name, theme colour, icon) while remaining fully usable
  when the manifest cannot be fetched.

### US-05 — Change nothing that is currently correct (P1)

- AC-13: Gaps remain gaps, unknown remains distinct from measured zero, the latest-sample marker
  appears only on real samples, and every statistic states its sample basis.
- AC-14: Existing behaviour is preserved: the entity dialog keeps a compact history entry, no route,
  prop, store key or state message is removed, and offline/loading/error/partial states keep their
  current wording. Controls that change shape (the Entity and Period `<select>`s in
  `e2e/DeviceMetrics.spec.ts`) are migrated to the new controls **in the same change**, with the e2e
  assertions updated deliberately; no selector is dropped without its replacement test.
- AC-15: Budgets hold: one concurrent history request, at most 1&nbsp;000 rendered points, statistics
  computed once per series (memoised), and pointer work bounded per move.

## Requirements and success

- FR-01: Provide a history toolbar (range presets incl. `Yesterday`, expose chips, stats strip) with
  persisted per-device selection; acceptance: AC-01, AC-02.
- FR-02: Provide responsive chart geometry: full width and ≥360&nbsp;px on desktop, compact and
  thumb-operable below 720&nbsp;px; acceptance: AC-03, AC-10.
- FR-03: Provide zoom, pan, reset and readout across pointer, touch and keyboard with no new runtime
  dependency (one local plugin inside the lazily loaded chart chunk); acceptance: AC-04, AC-05, AC-06.
- FR-04: Provide a statistics mode and CSV export over the loaded series; acceptance: AC-07…AC-09.
- FR-05: Provide mobile navigation, safe-area handling and a manifest; acceptance: AC-10…AC-12.
- FR-06: Preserve truthfulness and compatibility; acceptance: AC-13…AC-15.
- NFR-01: No new runtime dependency; the initial entry is unchanged and the lazily loaded chart
  chunk grows by no more than 4&nbsp;kB gzip.
- NFR-02: Crosshair and zoom stay responsive on a 1&nbsp;000-point chart (no dropped frames during a
  drag on a mid-range phone); `prefers-reduced-motion` disables chart and sheet animation.
- NFR-03: History and sheet meet WCAG 2.1 AA in both themes (contrast, focus visible, keyboard
  operation, 44&nbsp;px targets, chart described by text).
- SC-01: Every acceptance item has a Jest or Playwright test (gestures driven through real input
  events), plus a recorded desktop/mobile manual pass; no unrun check is reported as passed.

## Entities and external interfaces

- History transport is unchanged: `POST /api/metrics/query` with `history: true`, `exposes[]`,
  `limit` = 10&nbsp;001, one device per request; the response decoder keeps enforcing identity,
  bounds, types and the sample limit.
- New browser-only persistence: `nh-history-pref:<deviceId>` =
  `{ version: 1, preset, expose, custom?: { from, to } }`, validated on read (known preset, expose
  exists on the device, custom bounds finite/ordered/≤31 days) with a `24h` + stored-expose fallback;
  no household values beyond that window are stored. Keys are namespaced and versioned.
- New static interface: `public/manifest.webmanifest` plus an icon; `index.html` gains
  `viewport-fit=cover` and `theme-color`.
- Data volume is unchanged and already bounded end to end: 250 000 scanned rows per query with an 8 s
  timeout, 10 001 points per request, 10 000 decoded samples with a `partial` flag, and at most 1 000
  rendered points through extrema-preserving decimation. This change adds no new dropping mechanism;
  it fixes **provenance** (statistics read the raw samples, never the decimated set) and **surfacing**
  (covered window, uncovered buckets, export refusal).
- No HTTP, WS, MQTT, MCP, persistence or configuration contract changes; no new auth surface.

## Questions and assumptions

No blocking questions. Assumptions to validate during implementation:

- A local Chart.js plugin can cover wheel, drag-select, two-pointer pinch and keyboard without adding
  a dependency; if pinch proves unreliable on a target browser, the fallback is buttons plus
  drag-select, recorded as a plan deviation rather than a silent omission.
- Statistics over already-fetched samples is acceptable for 24&nbsp;h–30&nbsp;d with reporting
  cadences in use today; if a household cadence yields more than 10&nbsp;000 samples, the view must
  say the series is partial instead of widening the query.
- The bottom bar can derive its destinations from the existing `NAV_ITEMS` without a second
  navigation source of truth.
