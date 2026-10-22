# ADR-006: Chart interaction and statistics placement

Status: Proposed | Date: 2026-10-22 | Spec: [spec.md](spec.md)

## Context

History needs zoom, pan, reset and a readout that work with a mouse, a trackpad, a keyboard and a
touch screen, on a chart that must stay within the existing sample and render budgets. It also needs
a statistics reading, while the metrics API aggregates a **whole window** into one value
(`min/avg/max/count/last` in `backend/internal/metrics/query/aggregates.go`) and exposes no
interval/bucket parameter.

## Decision 1 — use `chartjs-plugin-zoom`, own the parts it does not cover

Adopt `chartjs-plugin-zoom` (v2, Chart.js 4 compatible) for wheel, drag-select, pinch and pan, and
keep in our code only what the plugin does not do: the range policy (clamp to 5 min…31 days and to
the loaded window via `limits`), the `−  +  Reset` buttons, the focusable container with keyboard
handling (`+ - 0`, arrows), the time/value readout, and `prefers-reduced-motion` behaviour.

Rationale, revised after checking the build: Chart.js is **lazily imported**, so the plugin lands in
the existing chart chunk (`chart-*.js`, 208 kB in the current build) and does not touch the 438.8 kB
initial entry. The pinned candidate is `chartjs-plugin-zoom` 2.2.0 (peer `chart.js >=3.2.0`, no
runtime dependencies of its own; ~94 kB unpacked, mostly source and maps). The
"no new dependency" argument therefore buys very little, while hand-written pinch/pointer code is
exactly where cross-browser bugs live (pointer capture, cancellation, two-finger tracking, iOS Safari
intercepting gestures, drag-vs-double-click conflicts, `touch-action` interplay with page scroll).

Rejected: a local interaction plugin (about 150 lines of gesture code we would own and have to harden
for every input device); a DOM overlay with its own canvas (duplicates rendering and hit-testing);
server-side resampling per gesture (network chatter, and the API has no bucket parameter).

Consequences: one small, actively maintained dependency in a lazy chunk; we adapt its config rather
than its internals; version coupling to Chart.js is checked when either moves. Acceptance is
unchanged: every gesture still has a button and keyboard equivalent, and range policy stays in our
pure domain layer so truthfulness rules do not live inside a plugin.

## Decision 2 — statistics are computed in the browser for this change

Statistics bucket the already-fetched series by local calendar day (`min/avg/max/last/count`) in
`domain/history/stats.ts`, reusing `numericStats`. The view states the sample count and marks the
result partial when the decoder flagged it.

Rejected for now: N parallel per-day queries (the client serialises one request and the endpoint is
rate-limited to 4/s, so a 30-day view would be 30 serial round trips), and a new backend bucket
parameter (storage/query change with its own performance and migration analysis, out of scope here).

Consequences: no backend change and no new failure modes; long ranges stay limited by the 10&nbsp;000
sample cap and must say so. Server-side bucketing is recorded as follow-up work so a future spec can
replace the client computation without changing the view contract.

## Decision 3 — bottom navigation below 720px, drawer above it

`NAV_ITEMS` stays the single navigation source. The shell renders a bottom bar containing the five
primary routes and keeps the drawer for the remainder, driven by CSS breakpoints so only one DOM
tree exists. Safe areas are applied through a new `--nh-safe-bottom` token fed by
`env(safe-area-inset-bottom)` after adding `viewport-fit=cover`.

Rejected: a second navigation list (two sources of truth), and always-on bottom tabs on desktop
(steals vertical space from charts).

## Decision 4 — sheet drag-to-dismiss is scoped

The entity sheet dismisses on a downward drag only when its content is scrolled to the top, with a
120&nbsp;px threshold and an Escape/close fallback, and is disabled under
`prefers-reduced-motion`. Rejected: gesture on the whole dialog (fights inner scrolling and text
selection).
