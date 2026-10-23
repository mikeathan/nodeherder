# Spec: Responsive binary history chart

ID: 032 | Status: Draft | Owner: NodeHerder contributors | Updated: 2026-10-23
Constitution: proposed 1.0.0 | Components: frontend

## Problem, scope, exclusions

The full-history binary chart in `frontend/src/components/chart/history/StateHistory.vue` is a 2rem-high strip. In a 24-hour range, a short opening can be a one- or two-pixel sliver; the state labels and summary sit below it, and the current range has only three time labels. On narrow phones the usable plot width makes the slivers harder to see. Dense histories also need a view bounded by screen width rather than by the number of reported transitions.

User review selected **Option A** from the throwaway prototype at [`binary-chart-review.html`](../../binary-chart-review.html): labelled active/inactive lanes, visible transition markers, exact duration summary, and a higher inactive/closed lane. The prototype was checked at 320, 390 and 1180px and with up to 9,500 synthetic state intervals. That is design evidence, not production acceptance evidence.

**In scope:** replace the full History view's binary strip with Option A at desktop and phone widths; raise the lower/inactive lane; preserve the existing expose-specific state labels and palette; provide truthful dense-history aggregation and inspectable transitions; retain current statistics, sample access, and compact-dialog behavior.

**Out of scope:** backend/query/transport/retention changes, new dependencies, numeric or enum charts, binary state semantics or palette redesign, chart zoom/pan, and redesign of the compact entity-dialog chart. The dialog continues to use the current compact binary presentation; the accepted Option A applies to the full History view, including its mobile layout.

## Scenarios and acceptance

### US-01 — Understand the state timeline (P1)

Independent verification: existing Jest state-domain tests and `DeviceMetrics.spec.ts` browser coverage.

- AC-01: Given a full History view for a binary expose, the chart shows the expose's active label in the upper lane and inactive label in the lower lane, with a time axis, legend, and visible state changes. Labels come from expose metadata (`values.on`/`values.off`), not hard-coded `Open`/`Closed` text.
- AC-02: Given the contact fixture, the lower teal/Closed lane is visibly raised from the bottom of the plot: its target is 70% of plot height on wide layouts and 72% on compact-width full-history layouts, measured from the plot top. The upper Open lane remains at 30% and 34%, respectively. Lane labels and endpoints remain inside the available width.
- AC-03: Given any viewport from 320px through 1280px, the SVG/plot fits its container with no horizontal page overflow. Tick density adapts to plot width, labels do not collide or clip, and the displayed range remains readable. Verify at 320×844, 390×844, 768×1024 and 1280×800 in light and dark modes.
- AC-04: Given theme changes or different binary expose kinds, the chart uses existing `exposePalette()` and CSS tokens. Contact retains orange Open / teal Closed; alarm-like exposes retain danger / contact colors; other exposes retain their kind / muted mapping. Unknown remains patterned and separately labelled. No color mapping is hard-coded into the new renderer.

### US-02 — Read dense data without losing its meaning (P1)

- AC-05: Given a sparse history, render time-proportional stepped state intervals and transition markers. Given transition density too high to distinguish at the available plot width (more than roughly one transition per four CSS pixels, with a minimum sparse-mode threshold), render width-bounded columns instead of one graphical mark per transition. Each column represents active-duration as a percentage of its **known** duration; unknown portions remain visually distinct and are not counted as inactive.
- AC-06: Given a dense column, pointer/tap or keyboard inspection reports its time interval, active share, known coverage and transition count. In sparse mode it reports the actual state, interval and duration. Arrow keys move the focus/readout; visible focus remains present. The chart has an accessible name/summary; labels and patterns mean state is not communicated by color alone.
- AC-07: Given up to `HISTORY_LIMIT` (10,000) decoded samples, chart binning is linear in input samples plus rendered columns and creates no more than `min(plot width in CSS pixels, RENDER_LIMIT)` density columns. Exact active/inactive durations and transition summary continue to use the source segments, not the aggregated render bins. The existing sample table and recent reports remain available.

### US-03 — Preserve truthful history behavior (P1)

- AC-08: Given empty, partial, unknown, repeated, same-time, or inverted-boolean contact history, existing state semantics remain unchanged: unknown is not measured inactive, partial tails remain unknown, equal reports do not invent transitions, and contact's configured on/off mapping still yields Open/Closed. No new history request or device command is caused by chart inspection.
- AC-09: The compact EntityViewDialog binary chart remains visually and functionally unchanged; existing full-history navigation, loading/error/empty notices, sample inspection, and non-binary history behavior are unaffected.

## Requirements and success

- FR-01: MUST replace the full-history binary strip with Option A state lanes, readable labels, adaptive ticks, and visible transitions; acceptance: AC-01…03.
- FR-02: MUST preserve expose-specific palette, state values/labels and unknown treatment; acceptance: AC-04, AC-08.
- FR-03: MUST bound dense rendering by plot width and show active share over known duration without losing exact summary semantics; acceptance: AC-05…07.
- FR-04: MUST support pointer/tap and keyboard inspection with accessible naming/focus and no color-only status; acceptance: AC-06.
- FR-05: MUST leave transport, device control, compact dialog, numeric/enum history and existing sample access unchanged; acceptance: AC-08/09.
- NFR-01: At 10,000 source samples, dense rendering produces at most one density column per CSS plot pixel and never more than `RENDER_LIMIT` columns; binning is O(N + W), where N is input segments and W is output columns.
- NFR-02: At 320, 390, 768 and 1280px, no horizontal overflow; at both theme modes chart labels/focus/state markers remain readable.
- SC-01: A mocked 10,000-sample binary browser fixture verifies aggregated mode, bounded render marks, responsive geometry, tooltip/readout and exact summaries; no household hub/device is used.

## Entities and external interfaces

- `HistorySeries` / `HistoryPoint` remain the decoded source owned by `frontend/src/domain/history/types.ts`; existing limit is 10,000 samples and partial state remains authoritative.
- `stateSegments()` remains the source of interval semantics; `stateStats()` remains the source of exact totals. Add a pure width-binning helper in `frontend/src/domain/history/state.ts` if required; it must conserve known/active duration and leave unknown time explicit.
- `StateHistory.vue` owns full-history SVG layout, resizing and presentation. Its `compact` branch remains the current dialog view. `HistoryView.vue` keeps routing binary series to this component.
- Existing `exposePalette()` and `formatExposeValue()` own colors and labels. No API, WS, MQTT, assistant, persistence, route, or device-command contract changes.

## Questions and assumptions

- Assumption: Option A is selected for the full History route on phones and desktop; the compact entity dialog keeps its current strip, following the reviewed prototype's separation of full and compact treatments.
- Assumption: existing `HISTORY_LIMIT`/`RENDER_LIMIT` remain unchanged; dense chart columns are capped by both actual plot width and `RENDER_LIMIT`.
- No unresolved question blocks planning. If the compact dialog scope is changed, update this spec/plan/tasks before implementation.
