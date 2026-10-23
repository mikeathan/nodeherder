# ADR-006: Chart interaction and statistics placement

Status: Proposed | Date: 2026-10-22 | Spec: [spec.md](spec.md)

## Context

History needs zoom, pan, reset and a readout that work with a mouse, a trackpad, a keyboard and a
touch screen, on a chart that must stay within the existing sample and render budgets. It also needs
a statistics reading, while the metrics API aggregates a **whole window** into one value
(`min/avg/max/count/last` in `backend/internal/metrics/query/aggregates.go`) and exposes no
interval/bucket parameter.

## Decision 1 — implement the interaction locally (no new dependency)

Add one small Chart.js plugin of our own (`components/chart/history/interactions.ts`, roughly
150–200 lines) registering `afterInit`/`beforeDestroy` on the existing chart, handling:

- wheel zoom around the pointer (mouse/trackpad),
- mouse drag-select to zoom a chosen range,
- single-finger drag that **only moves the readout** — never zooms, per AC-05,
- two-pointer pinch (tracked from pointer events, not a gesture library),
- keyboard `+ - 0` and arrows on a focusable container,
- `touch-action` rules so the plot never scrolls the page.

Pointer type decides the gesture, so one finger and one mouse button never mean two different things
on the same device. Touch range selection is covered by pinch and the buttons.

Range policy stays in `domain/history/range.ts`. The plugin does **not** write `scales.x.min/max`
directly: `options.ts` derives tick instants in `afterBuildTicks` and chooses label formatting from
the range duration, so the plugin reports the new range to `HistoryChart.vue`, which recomputes
`historyOptions(from, to, …)` and lets `HistoryCanvas` apply it. That keeps ticks, labels and bounds
consistent after every zoom, pan and reset.

**Rejected: `chartjs-plugin-zoom` 2.2.0.** Its peer is `chart.js >=3.2.0`, but its own
`dependencies` are `hammerjs ^2.0.8` and `@types/hammerjs ^2.0.45`, so adopting it adds an
unmaintained gesture library (last release 2017) to the tree, plus a runtime-declared types package.
Hammer duplicates gesture recognition we need in only one narrow form, and we would still write the
keyboard, readout, clamp policy and `touch-action` handling ourselves.

Rejected as well: a DOM overlay with its own canvas (duplicates rendering and hit-testing), and
server-side resampling per gesture (network chatter; the API has no bucket parameter).

Consequences: about 150–200 lines of gesture code we own and must harden for real input devices,
covered by Playwright with CDP-driven multi-touch; no supply-chain addition; the chart chunk grows
by roughly 2–4 kB gzip. The range is recomputed through the existing options factory rather than
patched on the scale, so zoom/pan/reset have a single rendering path and are tested for tick
positions and labels, not just axis bounds. If pinch turns out to be unreliable on a target browser, drag-select and the
buttons remain, and that deviation is recorded rather than hidden.

## Decision 2 — statistics are computed in the browser for this change

Statistics bucket the already-fetched series by local calendar day (`min/avg/max/last/count`) in
`domain/history/stats.ts`, reusing `numericStats`. Two constraints come straight from the code:

- Bucket the **raw decoded samples**, never the `decimate()` render set: decimation preserves extrema
  for drawing but would bias averages and sample counts.
- History mode is ascending-only, and `ApplyLimitSortBy` keeps `data[:limit]`, so an over-cap response
  keeps the **oldest** 10 001 samples and may be missing recent data. A partial series must therefore
  say which window is actually covered (first/last sample), keep buckets outside it as gaps, and
  disable export rather than presenting "today" as measured.

Rejected for now: N parallel per-day queries (the client serialises one request and the endpoint is
rate-limited to 4/s, so a 30-day view would be 30 serial round trips), and a new backend bucket
parameter (storage/query change with its own performance and migration analysis, out of scope here).

Consequences: no backend change and no new failure modes; long ranges stay limited by the 10&nbsp;000
sample cap and must say so. Server-side bucketing is recorded as follow-up work so a future spec can
replace the client computation without changing the view contract.

Review refinement (2026-10-23): each day uses a floating minimum–maximum bar, average circle and
last-value diamond, with distinct theme colours, a labelled HTML legend, named tooltips and the
existing accessible table. A measured-value axis avoids visually flattening temperature ranges
against zero; energy/event bar baselines remain unchanged. Empty terminal-midnight days are omitted,
while actual inclusive-end reports keep their day and explicit boundary label.

## Decision 3 — bottom navigation below 720px, drawer above it

`NAV_ITEMS` stays the single navigation source. The bottom bar contains Home, Devices, Automations,
Assistant and a More action opening the drawer. Below 720px, CSS hides primary links in the drawer;
at wider widths, the full sidebar/drawer remains available and the bottom bar is hidden. The same
route metadata drives both presentations. Safe areas use `--nh-safe-bottom` fed by
`env(safe-area-inset-bottom)` after adding `viewport-fit=cover`.

Rejected: a second navigation list (two sources of truth), and always-on bottom tabs on desktop
(steals vertical space from charts).

## Decision 4 — sheet drag-to-dismiss is scoped

The entity sheet dismisses on a downward drag only when its content is scrolled to the top, with a
120&nbsp;px threshold and an Escape/close fallback, and is disabled under
`prefers-reduced-motion`. Rejected: gesture on the whole dialog (fights inner scrolling and text
selection).
