# ADR-006: Chart interaction and statistics placement

Status: Proposed | Date: 2026-10-22 | Spec: [spec.md](spec.md)

## Context

History needs zoom, pan, reset and a readout that work with a mouse, a trackpad, a keyboard and a
touch screen, on a chart that must stay within the existing sample and render budgets. It also needs
a statistics reading, while the metrics API aggregates a **whole window** into one value
(`min/avg/max/count/last` in `backend/internal/metrics/query/aggregates.go`) and exposes no
interval/bucket parameter.

## Decision 1 — interaction is implemented locally, not with a zoom plugin

Add one small Chart.js plugin in `frontend/src/components/chart/history/` that owns:

- `wheel` (zoom around the pointer),
- pointer drag (rubber-band range select on mouse),
- two-pointer pinch (touch),
- a movement-thresholded scrub for one finger,
- keyboard (`+ - 0`, arrows) on a focusable container.

Range state stays in the existing pure layer (`domain/history/range.ts` helpers) and is applied by
setting the x-scale `min`/`max`, so truthfulness rules (bounds, partial data, gaps) remain in one
place.

Rejected: `chartjs-plugin-zoom` (adds a runtime dependency and version coupling to Chart.js for
~8&nbsp;kB and behaviour we would still need to constrain); a DOM overlay with its own canvas
(duplicates rendering and hit-testing); resampling server-side on every gesture (network chatter,
and the API has no bucket parameter).

Consequences: about 150 lines of chart code under test, no dependency, full control of gesture
semantics; the cost is that pan/pinch edge cases are ours to test, so gestures are covered by
Playwright with real input events.

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
