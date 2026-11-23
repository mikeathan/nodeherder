# Plan: Positions player — continuous playback across recordings

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft (unreviewed) | Updated: 2026-11-20
Constitutions: parent 1.0.0 (proposed); backend 1.0.0 (proposed); frontend 1.0.0 (proposed)
Contract: [contracts/sessions-samples.md](contracts/sessions-samples.md)

## Context and design

### Observed current state

- `frontend/src/composables/useSpatialPlayback.ts` owns one `time`, `playing`, `speed`
  and a 200 ms `setInterval`. `advance()` clamps to `bounds().to` and stops. `play()`
  restarts at `from` when `time >= to`.
- `SpatialPositions.vue` passes `bounds = { from, to }` of the **confirmed** session, so
  playback cannot leave a recording. `loadPoints()` always `pause()`s and `seek()`s to
  the first point.
- `PositionsTransport.vue` receives one `session` and its `points`, and derives `Next
  movement` from `movementEpisodes(points)`.
- `PositionsCoverageBar.vue` receives `sessions` summaries and renders one absolutely
  positioned button per session from `sessionSegments()` (left/width %). It has no
  per-session movement data.
- `backend/internal/spatialhistory/service.go` `Sessions()` already iterates every point
  of every chunk to count `Points`/`Slots`, filtered by `[from,to]`.
- `backend/internal/api/spatial_history.go` serialises `[]spatialhistory.Session` for
  `mode=sessions`, capped at 500.

### Design

**Backend (additive read model).** `spatialhistory.Session` gains
a `samples` int64 array (ascending, omitted when empty). In the existing single pass over chunks,
each session's filtered timestamps are collected and reduced with `samplesFrom(times)` ->
ascending, evenly reduced to `MaxSamplesPerSession = 240`. These are the recorded positions
themselves, not episode starts: an earlier draft exposed "where movement resumed", which made
a continuously moving recording look empty (real data: 18 samples, only 10 marks). Because a
session's chunks are read in key order the scan is already ascending; a small helper sorts
defensively. No new route, mode, storage or index. Payload/consumers documented in the
contract.

**Frontend time model.** New pure helpers in `frontend/src/domain/spatial/playback.ts`:

- `recordingOrder(sessions)` → chronologically ordered ids (the API returns ascending
  already; the helper makes ordering explicit and testable).
- `recordingNeighbours(sessions, id)` → `{ index, count, previous, next }`, used by both
  the transport labels and the buttons. Previous/Next move exactly one filtered recording
  and are disabled at the ends; an earlier draft restarted the current recording, which
  made a press right after Next unpredictable, so it was dropped.
- `sessionSegments(sessions, from, to)` / `playbackRecordings(sessions)` → each recording
  placed on the unified timeline with its absolute `from`/`to` and its recorded-position
  activity placed proportionally inside it, with the gaps between consecutive recordings
  addressed explicitly. A gap is reported as "skipped", never as recorded time.
- `movementChapters(points, gapMs)` → labelled chapters derived from
  `movementEpisodes`, sampled deterministically to `MAX_TRANSPORT_MARKS = 40` (a bound on
  focus stops) with the true count returned, so the scrubber states "24 movements, 12
  shown" rather than pretending 12 exist.
- `sanitiseSamples(value)` → runtime check for the additive field (array of finite,
  non-negative numbers, deduplicated, sorted, bounded to `MaxSamplesPerSession`) at the
  browser ingress; invalid values are ignored.

**Continuous controller.** `useSpatialPlayback` becomes the single owner of "advance
within a recording and cross into the next one". Its public API becomes
`useSpatialPlayback({ bounds, next })` where `bounds()` is current `{from,to}` and
`next()` returns the next recording's bounds or `null`; `advance()` sets the crossing
boundary and calls a new `onAdvance` hook that the component supplies asynchronously
(`loadRecording(nextId)`), so the composable stays free of transport code. To avoid an
interval callback awaiting network I/O, the composable sets `crossing` and the component
performs the load; the interval is paused (`playing` stays true) for the in-flight load
and resumes on the confirmed bounds. A failed load calls `failCrossing()`, which stops the
clock and clears the pending flag so a later Play can resume; the failure message comes from
the page's single `pointsError` announcement, not a second composable-owned error ref.

Gap skipping: when the next recording starts more than `POSITIONS_GAP_NOTICE_MS` after the
current ends, the loader reports the crossed `{ from, to }`, which the transport renders for
a few seconds and then clears; shorter gaps cross directly.

**Transport.** `PositionsTransport.vue` gains `Previous recording` / `Next recording`
using `recordingNeighbours`, a `2 of 5` position label plus the local range (AC-16), a
"Playing continuously" affordance only when a next recording exists, the skipped-gap
line, and labelled chapter markers instead of bare spans (`button.nh-transport-mark`
with an accessible name; gap stripes keep `aria-hidden` and gain a text summary). Play
keeps one button (Play/Pause) and the existing speed/Path/distance controls.

**Range timeline.** `PositionsCoverageBar.vue` draws one focusable mark per recorded
position, placed proportionally inside its recording, with nothing drawn where nothing was
recorded. Each mark names its local time and recording and emits `play-from: { id, at }`.
Clicking elsewhere in the segment also emits `play-from` at the recorded position nearest
the pointer, so the whole row stays a single-gesture target. Dense recordings use the
sampled `activity` list (`MAX_RECORDING_MARKS`), so the tab order is bounded. Segments
narrower than 28 px are drawn at 28 px and the row is rescaled to stay inside the track:
the strip is an index of recordings, not a proportional chart of a 24 h window. The
component receives `samples` from the session summary, so this is one request for the
whole filter (NFR-01).

**SpatialPositions wiring.** `usePositionsHistory.loadPoints` owns the fenced points read;
`usePositionsPlayback.select(id, at?)` confirms a recording and places the playhead (at `at`,
or the recording's first sample). Initial load, refresh, preset and device changes go through
the same clock: `useSpatialPlayback` places the playhead inside changed bounds even while
paused. Continuous advance loads the successor, then `resumeCrossing()` moves the playhead to
its start. A busy flag marks the crossing load; the previously confirmed recording stays
rendered until the new points arrive (FE-03), and a failure keeps it and announces the error.

### Files and ownership

| Path | Change |
| --- | --- |
| `backend/internal/spatialhistory/model.go` | additive `Samples` field + `MaxSamplesPerSession` |
| `backend/internal/spatialhistory/service.go` | derive activity samples in the existing session pass |
| `backend/internal/spatialhistory/service_test.go` | derivation, ordering, cap, range filter |
| `backend/internal/api/spatial_history_test.go` | payload includes additive samples |
| `frontend/src/services/spatial-history.service.ts` | optional `samples` on `SpatialSession` |
| `frontend/src/domain/spatial/playback.ts` | recording order/neighbours, playback segments, chapters, ingress check |
| `frontend/src/__tests__/domain/spatial/playback.test.ts` | unit coverage |
| `frontend/src/composables/useSpatialPlayback.ts` | continuous advance + crossing hook + failure |
| `frontend/src/__tests__/composables/useSpatialPlayback.test.ts` | composable coverage |
| `frontend/src/components/device/spatial/PositionsTransport.vue` | prev/next, position, chapters, skipped gap |
| `frontend/src/components/device/spatial/PositionsCoverageBar.vue` | per-movement mark controls + play-from |
| `frontend/src/components/device/spatial/SpatialPositions.vue` | continuous wiring, entry time, busy/failure |
| `frontend/tools/server/routes.mjs` | fixture samples + a longer multi-chunk day |
| `frontend/e2e/devices.spec.ts` | Positions continuous/navigation/marks coverage |

No ADR: additive optional field, reversible, frontend-local presentation, no storage or
trust change, no new dependency. The contract file records the interface.

### Alternatives rejected

- Per-recording points fan-out to mark activity on the range timeline: N requests for a
  24 h filter and a stalled UI. Rejected; the API approach the requester chose is used.
- One synthetic "megasession" by concatenating all recordings silently: hides the gaps
  between recordings and would render a continuous room from data that does not exist
  between them. Rejected; gaps are explicit.
- Making the recording list itself the timeline (selecting implies playing): loses the
  possibility of browsing without playing and duplicate controls. Rejected; transport
  keeps Play/Pause and adds explicit neighbour buttons, as comparable players do.
- Auto-advancing on load/first render: starts playback the operator did not ask for.
  Rejected (spec Q-05).
- Reusing `UiRange` for cross-recording scrubbing: its `min/max` are one linear range;
  recordings are disjoint intervals. The scrubber stays per-recording, and the range
  timeline carries cross-recording navigation.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 067 states problem, scope, AC-01…AC-20, FR-01…FR-11, NFR, SC; Q-07 records the requester's direction, no approval claimed. | PASS (design) | PASS — spec/plan/tasks/contract present; review still required |
| NH-02 | Read-only viewer; no device commands or automation input. Gaps are "skipped", never measured zero; failures stop playback instead of fabricating a transition. | PASS (design) | PASS — diff reads history only |
| NH-03 | Existing authenticated route only; no new trust boundary; `samples` carries operational positions already exposed by `mode=points`; no secrets or persistence. | PASS (design) | PASS — no route, auth or storage change |
| NH-04 | Business rules in `domain/spatial/playback.ts`; clock ownership in the composable; transport functions unchanged. Additive optional field documented in `contracts/`; no new dependency. | PASS (design) | PASS — contract + Types' optional field + fixture |
| NH-05 | AC→test map below; backend unit+handler tests, domain+composable Jest, browser suite; actual commands recorded below. | PASS (design) | PASS — commands and results recorded |
| NH-06 | Units/timezone unchanged (epoch ms, local labels); no stored data or retention change; interval cleared in scope disposal; a failed continuous load is announced and stops. | PASS (design) | PASS — no stored data touched |
| BE-04 | Read models keep time-boundary inclusivity and ordering; `Sessions`/`Points`/`Heatmap` share the straddling-chunk lookback; `Slots` has a defined empty shape (`[]`, not `null`); no schema/key/retention change. | PASS (design) | PASS — service tests cover overlap, range filter, ordering, empty and heatmap agreement |
| BE-05 | Existing authenticated `GET /api/spatial/history` route; success payload documented in the contract; error/empty responses unchanged. | PASS (design) | PASS — handler test asserts the additive key on the wire |
| FE-01 | State stays component-local; playback clock and crossing in the composable; no competing device truth. | PASS (design) | PASS |
| FE-02 | `SpatialSession` ingress gains an optional validated field; existing fields unchanged; mocks updated with the fixture. | PASS (design) | PASS — `sanitiseSamples` is the only ingress |
| FE-03 | Confirmed recording stays rendered while crossing loads; `aria-busy`; failure announced and playback stopped; empty/end states distinct. | PASS (design) | PASS — browser failure and empty scenarios |
| FE-04 | Named buttons and marks, keyboard activation, visible focus, non-colour cues ("2 of 5", "skipped"), 44 px transport targets, 390/768/1440 checks. | PASS (design) | PASS — desktop and mobile suites, Shift+Arrow test |
| FE-05 | No secrets, assistant content or browser persistence added; positions stay in memory. | PASS (design) | PASS — none present in the diff |

## Contracts, data, rollout

- HTTP: additive `samples` on an existing authenticated response
  ([contract](contracts/sessions-samples.md)); producers/consumers listed; older hubs omit
  it and the browser renders no marks (one fallback selection control per recording).
- Persistence/configuration: none. Rollback: revert the additive field and the frontend;
  no stored data involved, so an older frontend against a newer hub ignores the field.
- Ordering: deploy backend before frontend to see marks; frontend-first is safe and
  simply renders none.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| AC-01, AC-03, AC-04, AC-20, FR-01, FR-11 | `useSpatialPlayback.ts`, `SpatialPositions.vue` | composable unit tests (advance, crossing, end); browser `Play all recordings` from night → morning and end text | pass — 8 composable tests; desktop+mobile e2e |
| AC-02, FR-02 | `useSpatialPlayback.ts` | unit: pause then play resumes at the same time; browser Play/Pause/Play | pass |
| AC-05, FR-04 | `SpatialPositions.vue`, composable | existing browser failure scenario keeps the confirmed segment and alerts; composable `failCrossing` test | pass |
| AC-06…AC-09, FR-03 | `playback.ts`, `PositionsTransport.vue` | unit neighbours/boundaries; browser Next/Previous, disabled ends, Shift+Arrow | pass |
| AC-10…AC-13, FR-05, FR-06 | `service.go`, `PositionsCoverageBar.vue` | Go derivation/range/cap tests; handler payload test; browser per-movement mark names + keyboard activation; domain strip/label tests | pass |
| AC-14, AC-15, FR-07 | `playback.ts`, `PositionsTransport.vue` | unit chapter derivation and sampling; browser mark names | pass |
| AC-16, FR-08 | `PositionsTransport.vue` | browser "Recording 2 of 4" and local range assertions | pass |
| AC-17, AC-19, FR-09 | `SpatialPositions.vue` | existing empty/error/Heatmap/Presence assertions remain green | pass — full suites |
| AC-18, FR-10 | styles | 390 px sheet/overflow assertions; transport targets are CSS 44 px | pass — mobile suite; target size not measured |
| NFR-01…NFR-05 | all | single sessions request per filter asserted by the design; no dependency/persistence added | pass — diff review |
| SC-01 | `service_test.go`, `spatial_history_test.go` | `go test ./internal/spatialhistory ./internal/api` | pass |
| SC-02, SC-03 | `playback.test.ts`, `useSpatialPlayback.test.ts` | `npx jest <files> --runInBand` | pass — 15 + 8 tests |
| SC-04 | `devices.spec.ts` | `npm run test:e2e -- --project=desktop e2e/devices.spec.ts`; same for mobile | pass — 11/11 both projects |
| SC-05 | all | commands below | pass |

### Verification (recorded)

Environment: macOS, repository on 2026-11-20. No real hub was used.

- `backend`: `go test ./...` → **pass** (exit 0); `go build ./...` → **pass**.
- `frontend`: `npm test -- --runInBand` → **154 suites, 1021 tests passed**.
- `npm run lint` → **0 errors, 23 pre-existing warnings** (none in changed files).
- `npm run build` → **pass** (`vue-tsc` + `vite build`).
- `npm run type-check:e2e` → **pass**.
- `npm run test:e2e -- --project=desktop e2e/devices.spec.ts` → **11 passed**.
- `npm run test:e2e -- --project=mobile e2e/devices.spec.ts` → **11 passed**.
- `npm run test:e2e -- --project=desktop` (full) → **199 passed, 9 skipped**.
- `npm run test:e2e -- --project=mobile` (full) → **208 passed** on the settling run.

Not covered by automation (explicit unrun checks):

- a real-household recording (the 10 s episode constant and 60 s gap-notice threshold are
  validated only against the fixture);
- the 500-session × 500-movement payload stress;
- reduced-motion visuals (unchanged from spec 057; no new animation was added);
- 44 px target measurement (set in CSS, covered indirectly).
- One order-dependent mobile `LoggerSettings` failure occurred in a full run, then passed
  alone and in the next full run; the spec and its module are untouched by this diff.

## Polish pass (2026-11-20, second review)

The mechanics worked but the UI read as a form: a mostly empty 24 rem canvas outranked the
timeline, every string in the transport was 0.72-0.85 rem, and four overlays competed for
29 px of range timeline. An independent design critique plus full-page captures at 390, 1024
and 1440 px in both themes drove this pass.

| Defect | Change | IDs |
| --- | --- | --- |
| Axis label contradicted the 24h button (axis is the recorded envelope) | Head states the span it draws and the movement total; the misleading `24h filter` suffix was removed | AC-21 |
| Empty grid dominated the first screen | ~~Canvas narrowed to `clamp(11rem, 26vh, 17rem)`~~ **reverted (reviewer/owner feedback): shrinking the map and keeping a full-size transport simply inverted the disproportion. The canvas keeps spec 057's 16-24 rem; the transport was made proportionate instead (see the proportion row below).** | AC-21, FR-12 |
| Ticks were episode starts, not activity | The bar drew "movement resumed after a 10 s pause", so a continuously moving recording showed almost none: real data had 18 samples and 10 ticks, and operators could see movement where there was no tick. The payload now carries every recorded timestamp as `samples` (THE movement, because the recorder only stores on movement) | AC-10, AC-22 |
| Unstyled per-sample tabs were unhittable and unreadable | Each recorded position is a real focusable control with a padded target and a named time; dense recordings use the sampled list so the tab order stays bounded. Clicking elsewhere in the segment plays from the nearest recorded position. Segments narrower than 28 px are drawn at 28 px | AC-22, AC-26 |
| Selected, dense, played and playhead all looked alike | Neutral outlines for recordings, one solid accent block for the selection, accent outline for dense, played fill and second playhead deleted | AC-24 |
| Five type sizes in 0.13 rem; clock not leading | Three tiers: clock 1.15 rem/600 (1 rem on phones), recording label 0.9 rem/600, everything else 0.78 rem muted; Play is the only primary button | AC-23 |
| Transport as tall as the map (six stacked rows, 239 px vs a 256 px canvas at 620 px tall) | Idle / skipped-gap / movement-count / nothing-to-scrub collapsed into ONE status line; card padding and gaps tightened; scrubber 1.9 rem with the marks row tightened. Guarded by `devices.spec.ts` at 1440x800 and at 1280x600 — the latter is the strictest case because the canvas has reached its 16 rem floor while the transport keeps a fixed control height; phones keep mandated 44 px targets. | AC-23, FR-12 |
| Seven-item flat control row, clock between buttons, `of 3 m` orphan | Two groups (`controls`, right-aligned `metrics`); one distance node; axis row is three times, not a duration | AC-23 |
| "4" repeated five times, two different movement totals 40 px apart | Filter total in the head only; per-recording count worded "... in this recording" | AC-23 |
| Rail printed `0 s - 1 points - slot 1` | `1 sample - slot 1`, pluralised slots, seconds when the filter is under an hour | AC-25 |
| Chevrons wrapped away from the title on phones | Label wraps; chevrons pinned as siblings | AC-18 |
| Single-sample recording drew a dead scrubber | Transport says "One recorded sample at ... Nothing to play or scrub" and omits the scrub row; default selection unchanged (spec 057 AC-04) | AC-25 |

Verification after the pass: `npm test -- --runInBand` 154 suites / 1022 tests;
`npm run lint` 0 errors (23 pre-existing warnings); `npm run build` pass;
`npm run type-check:e2e` pass; `devices.spec.ts` 11/11 desktop and 11/11 mobile;
full desktop 199 passed / 9 skipped; full mobile 208 passed. Screenshots captured at
390/1024/1440 px dark and 1440 px light, plus 1:1 crops of the range timeline, transport
and canvas, before and after each change.

## Debt, risks, exceptions

- The 10 s episode gap is a presentation constant (spec Q-01) and the 60 s gap-notice
  threshold is new (Q-03); both are validated against fixtures, not a real recording.
- Up to 500 sessions × 240 samples bounds the payload; a pathological hub that records
  movement every few minutes for a week is bounded by the cap, and the scrubber states
  when it sampled its marks.
- Continuous playback streams recordings over the network as it advances; on a slow hub a
  crossing load is visibly busy and a failure stops rather than skips. Prefetching is
  deliberately out of scope (NFR-01).
- Movement marks are small by construction (proportional to the recording); they remain
  focusable and named, and the segment text carries the full information. AC-18 therefore
  scopes 44 px to the transport buttons, not to the proportional marks.
- Spec 057's "recording list" and this spec's "recording" are the same `SpatialSession`;
  the naming follows the product ("Recordings"), not the API type.
- No exception or deviation is requested; no rule is marked N/A.

## Completion

Implementation and frontend/backend verification are complete; artifact results and
unrun checks are recorded above. This plan does not claim review approval.
