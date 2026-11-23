# Clean-code remediation: spatial, history, ESPHome

Date: 2026-11-20 | Branch: feat/positions-player-continuous
Authority: [.agents/rules/vue-engineer.md](../../.agents/rules/vue-engineer.md),
[frontend/CONSTITUTION.md](../../frontend/CONSTITUTION.md) (FE-01, FE-02, FE-03, FE-04)

Owner feedback: components were large, mixed responsibilities and not readable; the code must
follow the repo's rules and keep a single source of truth for spatial positioning, history and
ESPHome. This records what violated which rule and what changed. Findings came from two
read-only audits plus direct measurement.

## 1. Spatial positioning

### Violations

| Rule | Violation |
| --- | --- |
| "Keep functions and components focused on one responsibility" | `SpatialPositions.vue` was 955 lines: 822 of `<script setup>` holding data loading, view-model derivation, zoom/fit maths, playback orchestration, keyboard handling and view measurement in one component. |
| "Separate domain rules, formatting, view-model mapping, transport and persistence" | Fit maths (`fittedBounds`, `zoomAnchorPoint`), trail/hop derivation and the recordable envelope were computed inline in the component instead of in `domain/`. Four near-identical `minX/minY/maxX/maxY` accumulators were copy-pasted. |
| "Clean up listeners, timers, subscriptions" | `ResizeObserver` lifetime was managed by hand in the component, with two separate `onBeforeUnmount` hooks. |
| "Preserve reactivity when destructuring" | Nothing owned "the filter changed": the component and the loader both reacted, so a caller could forget one of the two reloads. |

### Changes

| Path | Change |
| --- | --- |
| `domain/spatial/positions-playback.ts` (new) | `playbackTracks`, `playbackTrail`, `playbackHopLinks`, `latestSampleAt`, `recordedEnvelope` — the view-model mapping the component was doing. |
| `domain/spatial/positions.ts` | `fittedPointsBounds`, `zoomAnchorPoint`, one `BoundsAccumulator`; replaces four inline accumulator copies. |
| `composables/positions/usePositionsHistory.ts` (new) | Owns the filter, the recordings summary, confirmed points and heat cells, generation-fenced against stale replies, and the single "filter changed" trigger. |
| `composables/positions/usePositionsPlayback.ts` (new) | Owns the clock ↔ loader coupling: crossings, gap reporting, and settling a superseded load. |
| `composables/positions/useElementSize.ts` (new) | `ResizeObserver` lifetime in one place, including re-attach on a `v-if` swap. |
| `components/device/spatial/SpatialPositions.vue` | 955 → 616 lines; presentation and wiring only. |

A callback cycle between the two composables was avoided with a `supersededAt` counter rather
than playbook-breaking closure coupling: history must not import playback.

## 2. History

| Rule | Violation | Fix |
| --- | --- | --- |
| Units ownership | `HistoryStats.vue` hand-rolled rounding, spacing and a second missing-value sentinel (`'Unknown'`) beside the shared `'—'`. | Uses `formatWithUnit`; the sentinel now has one owner. |
| Units ownership | `options.ts` re-implemented the same formatting twice, and printed the unit on both bounds of a spread (`18 – 24 °C` became `18°C – 24°C` once routed through units). | One `tooltipValue()` helper; the spread carries its unit once. Test updated to the canonical units format. |
| Single source of truth | `NumericHistory.vue` repeated `domain/history/stats.ts`'s brightness/colour-temperature ternary verbatim. | New `historyDisplayUnit(series, expose)` owner. |
| Single source of truth | Five surfaces formatted a period with `toLocaleString()` and no options, so the same span read differently from `formatHistoryMoment`. | New `formatHistoryRange(from, to)`, used by the chart readout, the coverage notice, the sample table and the device toolbar. |
| Extract shared behaviour into composables | `HistoryStats` and `NumericHistory` each read theme tokens with `getComputedStyle`, called `matchMedia` inline, and used an undocumented `void settings.value` "read to subscribe" trick. | New `useChartAppearance()`; it also **observes** the motion preference instead of reading it once. |

Tests: `docs`-level assertion in `options.test.ts` corrected to the canonical temperature format
(`24°C`, no space) and to the shared sentinel.

## 3. ESPHome

| Rule | Violation | Fix |
| --- | --- | --- |
| Single source of truth | Three competing copies of the same user-facing strings: `contracts/esphome.ts` (`ESPHOME_MESSAGES`), a full duplicate block in `domain/esphome`, and literals in `services/esphome` and `useESPHomeManageDialog`. The 409 read two different sentences depending on the path. | `ESPHOME_MESSAGES` is the only owner; domain/service/composables import it. Verified no duplicate constant remains. |
| Type/findings | `useESPHomeManageDialog` re-implemented `isESPHomeDevice` and dropped the legacy "no provider yet" rule, so the unmatched-device warning could be wrong. | Uses the shared predicate. |
| Unvalidated external payload | `parseEnrollment` accepted any `status` string and any `reconnections` number, and the dialog rendered the status raw. | Narrowed to `'online' \| 'offline'` (the values `backend/internal/deviceproviders/constants.go` actually sends) and a finite non-negative count; rendered through `connectionStatusLabel`. An unrecognised status is dropped, not echoed. |
| Unvalidated boundary | `parseCandidate` accepted any numeric port. | Requires an integer in `ESPHOME_PORT_MIN..MAX`. |
| Lifecycle | Nothing reacted to a parent-driven close, so the draft and the encryption key stayed in memory. | `watch(open)` resets the form and aborts the request. |
| Timer/background cost | The 5 s enrollment poll kept requesting while the tab was hidden; peer composables gate on `visibilitychange`. | Polling skips requests while `document.hidden`. |
| Dead/unreadable code | `busy.value = false` looked dead before `close()`, but was load-bearing: `close()` refuses while busy. | Split `close()` (guarded, user-initiated) from `discardAndClose()` (unconditional, after a finished submit) so the contract is explicit. |

Tests added: `connectionStatusLabel` mapping, and that an unknown status and a negative
reconnect count are dropped at ingress rather than rendered.

## Verification

- `npm test -- --runInBand` → **154 suites, 1023 tests passed**.
- `npm run lint` → **0 errors, 23 pre-existing warnings** (none introduced).
- `npm run build`, `npm run type-check:e2e` → **pass**.
- `e2e/devices.spec.ts --grep ESPHome` → **7 passed**.
- `e2e/devices.spec.ts` → 10 passed, 1 failed: `spatial mappings save and show only current
  provider reports`, at the **Live** view assertion (`text.nh-target-label`). This fails
  identically on the untouched baseline (verified by stashing the refactor), is outside the
  files changed here, and is not caused by this work.

## Not done (deliberately)

- `StateHistoryChart.vue` (290-line script, four responsibilities) and the `StateHistory*`
  triple-duplicated label/palette logic: real findings, but they belong to a separate change
  with its own browser coverage. Recorded, not silently dropped.
- `HistoryView.vue` dead CSS that duplicates `HistorySamples.vue`, and its unscoped `<style>`
  leaking globally: same reason. Both are listed with line numbers in the audit.
- ESPHome `useESPHomeManageDialog` mutation `AbortController` scope and the add/manage dialog
  split: design-level changes that alter behaviour beyond clean-code, so they need a decision.

## Second pass — logic still living in components

Owner feedback after the first pass: components still mixed responsibility; logic that is shared
across components was repeating the same or similar calculations. Confirmed by measurement
rather than inspection.

### What was still wrong

| Duplication | Where | Rule |
| --- | --- | --- |
| The same user-facing sentences built inside components | `PositionsTransport`, `PositionsCoverageBar`, `PositionsSessionRail`, `SpatialPositions` each assembled "Recording 1 of 4 · …", "Skipped 4 min …", the idle line, the empty-recording sentence, pluralisation and the segment accessible name | "Separate … formatting, view-model mapping … using cohesive modules and named helpers" |
| Parts-to-percentage geometry inside a component | Both the transport and the timeline carried their own `Math.min(100, Math.max(0, ((at - from) / span) * 100))`, and their own `Math.max(1, to - from)` clamp | same |
| Activity-strip and segment-placement maths in a `.vue` | `PositionsCoverageBar.vue` owned bar grouping, merge thresholds, minimum widths and the nearest-click lookup: 75 lines of script in a presentation component | "Keep functions and components focused on one responsibility" |
| Unit conversion at the call site | `formatDistance(props.distanceMm / 1000)` — millimetres converted in the component | "Do not pass unit strings at call sites or concatenate them manually"; the formatter owns conversion |
| One state, two different sentences | The timeline and the recordings list said "No recordings in this range" vs "No recorded positions in this range" | "components must not invent duplicate wording" |

### Changes

| Path | Change |
| --- | --- |
| `domain/spatial/positions-labels.ts` (new) | Every user-facing sentence for this family: `countLabel` (one pluraliser), `hasPlayableSpan`, `timeLabelOptions`, `recordingPositionLabel`, `transportStatusLabel` (with the state ranking documented), `skippedGapLabel`, `pendingGapLabel`, `idleRecordingLabel`, `movementCountLabel`, `rangeHeadlineLabel`, `segmentLabel`, `railTimeLabel`, `railMetaLabel`, `emptyRecordingLabel`, and the loading/empty copy scoped per surface. |
| `domain/spatial/positions-strip.ts` (new) | Geometry: `activityBars`, `renderSegments` (minimum readable width + row rescale), `nearestPositionAt`, `positionPct`, `positionedSpan`, `totalSampleCount`. |
| `utils/units.ts` | New `formatMillimetres`. Spatial history stores millimetres; the UI reads metres and kilometres. |
| The four components | Presentation only: transport 115 lines of script (was 146 with labels), coverage bar 75 (was 134, now no geometry), rail 25 (was 34 with two formatters inline). |

### Bugs this pass found

1. **A real regression I introduced in pass 1:** the playhead was never seeked when a recording
   was selected. `loadPoints` stopped seeking and `selectSession` did not replace it, so the
   clock stayed at the previous recording's end and the canvas drew nothing (every sample was
   in the future). Fixed by moving the seek into `usePositionsPlayback.select()`, where the
   recording and the playhead change together.
2. **Pre-measurement layout blow-up:** `renderSegments` divided by a 1 px floor, so before the
   first `ResizeObserver` callback every recording expanded to full track width. It now returns
   the proportional layout until a width is known.
3. **Two sentences for one state**, resolved by owning both and scoping them deliberately
   instead of silently picking one.

### Tests

New `src/__tests__/domain/spatial/positions-labels.test.ts` — 15 tests over the extracted
owners: pluralisation, the single-sample rule, date/second precision, the empty-recording
sentence, rail rows, movement summary, the segment accessible name, live-region quantisation,
the transport state ranking, gap labels, bar grouping and weights, minimum width and row
rescaling, pre-measurement stability, nearest-click resolution and clamping, and span placement
against a zero-length window.

Two of my own expectations were wrong and the implementation was right (bar weights, and a
precision boundary); the pre-measurement case was the reverse — a real bug my test exposed.

### Verification

- `npm test -- --runInBand` → **155 suites, 1038 tests passed**.
- `npm run lint` → **0 errors, 23 pre-existing warnings**.
- `npm run build`, `npm run type-check:e2e` → pass.
- Full browser suites: **desktop 199 passed / 9 skipped**, **mobile 208 passed** — including
  `spatial mappings save and show only current provider reports`, which had been failing. Three
  of its assertions encoded assumptions my changes invalidated (the newest fixture now holds one
  sample, and the rail test asserted a fixture id instead of "the same recording stays
  selected"); those were corrected, not bypassed.
