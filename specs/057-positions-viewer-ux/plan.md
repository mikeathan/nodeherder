# Plan: Positions viewer layout and playback legibility

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: agent draft (unreviewed) | Updated: 2026-11-06
Constitutions: parent 1.0.0 (proposed); frontend 1.0.0 (proposed); backend N/A (no backend change)

## Context and design

### Observed current behaviour

- `SpatialPositions.vue` is one component: toolbar, canvas, session list, transport. The
  template swaps to a loading paragraph on every session switch (`v-else-if="loading"`),
  which unmounts the viewer and collapses page height.
- Session data: `services/spatial-history.service.ts` (`SpatialSession`, `SpatialPoint`);
  the backend returns sessions ascending by `from`, and the component auto-selects
  `sessions[0]` (oldest).
- Playback: `time` advances by `200 * speed` every 200 ms; `trails`/`tracks` are
  recomputed from `points.filter(at <= time)`. `tracks` hides a slot after 30 s.
- `SpatialCanvas.vue` is shared with `SpatialPanel.vue` and `PresenceRailTile.vue`.
- No `@media` rules and no shared timeline primitive exist for this view.

### Design

**Layout.** One grid inside the Positions player:

- main column: toolbar → canvas → coverage timeline → transport;
- rail column (`≥900px`): recordings in a bordered panel with a header matching the View
  panel height, sticky and internally scrollable, collapsible from its header (`Hide
  recordings`) and restorable from the toolbar (`Show recordings (N)`);
  the grid falls back to one column while collapsed;
- below `900px`: the rail is hidden and the same list renders inside `UiDialog`
  (already a bottom sheet at ≤640 px), opened from a `Recordings (N)` toolbar button.

The viewer spans its column at a fixed `clamp(16rem, 40vh, 24rem)` height, so the
coverage timeline and transport stay reachable. A `ResizeObserver` feeds the panel aspect
into a fitted `viewBox`: 1× fits the room outline and the confirmed session's points (with
padding); 2×/3×/5× shrink that extent. Above 1× the centre is an anchor captured when the
level changes or the playhead is moved explicitly; playback ticks do not re-centre, so the
target travels across a fixed view. The target marker grows 25% per level (capped at 2×)
so the scale change is visible while labels stay constant. The device is not forced into
the frame but is pinned to the nearest view edge with a presence glyph, so zooming never
makes it vanish. A post-layout measure keeps the viewBox aspect matched to the viewer after
rail or view changes. A computed symbol scale targets ~14 px for labels
and markers, so only the map content scales and stays readable at every level. Label halos
are passed as `labelStroke` (~25% of the label font) because a fixed user-unit stroke grew
into a surface-coloured blob at 5×. Hiding the
rail widens the panel; the aspect and viewBox update to show more area at the same height.
The field-of-view cone is driven by `workspace.sensor.fov` (0–180°) and is explicitly
illustrative, never device-reported. The player's display-only "Field of view" checkbox
uses that angle when configured and otherwise shows a labelled 90° fallback; it never
changes the saved angle.
The transport stays in normal flow directly below the coverage timeline.
Implementation verification rejected an initially planned `position: sticky; bottom: 0`
transport: on a 600 px-tall viewport it pinned over the coverage timeline and hid the
range information. No part of the player uses `position: fixed` or an overlay. Rail
visibility and zoom are component state only, not persisted.

**Shared view controls and pan.** `SpatialViewer.vue` wraps every full spatial canvas. Its
panel carries zoom (1×/2×/3×/5×), Grid, Field of view and an optional `#extra` group
(Smoothing/Scale/Orientation/Raw in Live and Draw). It is a disclosure: expanded at
≥900 px, collapsed below, with 44 px targets on phones. The viewer applies zoom to the
caller's base viewBox around the active anchor and adds a pan offset. Read-only views pan
on drag; editing views only pan with the Pan tool or Space, and the canvas's `panMode`
suspends editing while panning. Reset returns zoom 1 and clears the pan; a `resetKey`
clears pan on view/session changes. Setup forms and the Home rail preview do not use the
viewer.

**Stable selection.** The load path is split into `sessions`, `points` and `heatmap`
states. The viewer always renders; `aria-busy` marks the in-flight part. `selected`
is the requested session and `confirmed` the one whose points are rendered. A stale
points response is ignored; a failed switch reverts `selected` to `confirmed` and
announces the error. Default selection becomes the newest session when nothing is
confirmed.

**Time model.** Two bars with distinct jobs:

- `PositionsCoverageBar`: the recorded envelope inside the preset window (the preset
  window itself when nothing is recorded), one proportional segment per session, explicit
  empty gaps, playhead, played fill inside the active segment. Fitting the envelope keeps
  a four-minute recording visible instead of turning it into a sub-pixel sliver in a
  24 h window.
- `PositionsTransport`: `UiRange` bounded to the confirmed session, with start/duration/
  end labels. Idle gaps and movement ticks are drawn in a marks row above the range so
  `UiRange` itself stays the single slider primitive.

**Movement legibility.**

- `movementEpisodes(points, gap)` → first sample after a gap > 10 s; rendered as ticks
  and driven by a `Next movement` action.
- `trailGroups(points, at, 15 s)` → full path plus a recent subset (including one
  preceding point for continuity), rendered as `.nh-trail` and `.nh-trail.is-recent`.
- `hopLinks` → previous/current sample per slot; a dashed connector when the gap exceeds
  the idle threshold or the jump exceeds 1,000 mm.
- `pulseAt` → the current sample timestamp; a keyed ring restarts its animation when the
  sample changes. `prefers-reduced-motion` disables the animation.
- Last-known target stays visible for the session; older than 30 s it uses the existing
  `held` track state (dashed, muted) and the label names the last sample time.
  The transport says "No movement recorded for …", never "idle", because the recorder
  stores no sample while still.

**Distance.** `playedDistanceMm(points, time)` per slot; formatted with
`formatDistance(mm / 1000)` from `utils/units.ts`. The session total is shown beside it.

### Files and ownership

| Path | Change |
| --- | --- |
| `src/domain/spatial/playback.ts` | new pure helpers: segments, episodes, idle spans, trail groups, played distance, clock/duration labels |
| `src/__tests__/domain/spatial/playback.test.ts` | new unit tests |
| `src/composables/useSpatialPlayback.ts` | new playback clock/timer ownership |
| `src/components/device/spatial/PositionsCoverageBar.vue` | new range timeline |
| `src/components/device/spatial/PositionsTransport.vue` | new transport + session scrubber |
| `src/components/device/spatial/PositionsSessionRail.vue` | new recordings list (rail and sheet body) |
| `src/components/device/spatial/SpatialPositions.vue` | layout, data flow, stable loading/error |
| `src/components/device/spatial/SpatialViewer.vue` | new shared View panel + pan shell |
| `src/components/device/spatial/SpatialPanel.vue` | wrap the canvas, move display controls into the panel, add zoom/FOV |
| `src/components/device/spatial/SpatialCanvas.vue` | optional `recentTrails`, `hopLinks`, `pulseAt`, `uiScale`, `labelStroke`, `panMode` props |
| `tools/server/routes.mjs` | multi-session mock fixture with gaps |
| `e2e/devices.spec.ts` | Positions regressions |

No backend, contract, store or dependency changes. `SpatialCanvas` additions are optional
and default to no rendering, so `SpatialPanel` and `PresenceRailTile` are unaffected.

### Alternatives rejected

- Fixed bottom dock or sticky transport: conflicts with mobile shell/safe-area behaviour,
  and on short viewports it pinned over the coverage timeline. Rejected after observing
  the 600 px-tall layout during implementation.
- A width-driven square viewer (first follow-up): it made the viewer taller than the
  viewport and pushed the timeline and transport far below the fold. Rejected in favour of
  a full-width, fixed-height panel with content zoom.
- Normal/Large size presets: rejected by the requester; zoom should scale the map contents,
  not resize the panel.
- Player above the canvas: pushes the map below the fold and separates the control from
  the thing it controls.
- Interpolating between samples: depicts movement the sensor did not report.
- Red progress fill: red is the error/exclusion token; progress is encoded with accent
  fill plus text/height.

No ADR: the decisions are reversible, frontend-local, add no dependency or boundary, and
do not alter stored data or contracts.

## Constitution Check

Design evidence plus recorded execution checks. Unrun items are listed explicitly below.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 057 states problem, scope, exclusions, AC-01…AC-20, FR/NFR/SC. Direction confirmed by requester 2026-11-06 (spec Q-07); no maintainer approval claimed. | PASS (design) | PASS — spec/plan/tasks present; review still required |
| NH-02 | Read-only viewer; no device commands, state writes or automation input. Missing samples are labelled "no movement recorded", never measured stillness. | PASS (design) | PASS — diff reads existing history only |
| NH-03 | No new entry points, secrets or persistence. Positions stay in component memory; no logging. Existing authenticated API only. | PASS (design) | PASS — no contract or storage change |
| NH-04 | Helpers live in `domain/spatial/playback.ts`, presentation in components, playback timing in a composable. No dependency, contract or storage change; `SpatialCanvas` props are optional. | PASS (design) | PASS — new files + optional props only |
| NH-05 | AC→test map in Traceability; domain unit tests plus extended `devices.spec.ts`; frontend checks listed below. | PASS (design) | PASS — 748 unit tests; 142/9 desktop, 151 mobile e2e |
| NH-06 | Units mm→m via `utils/units.ts`; local-time labels; 7-day retention untouched; no data mutation; playback interval cleared in `onScopeDispose`. | PASS (design) | PASS — no stored data touched |
| FE-01 | Shared behaviour moves to `domain/spatial/playback.ts` and `useSpatialPlayback`; state stays component-local; no competing device truth. | PASS (design) | PASS — no store or service change |
| FE-02 | Consumes existing `SpatialSession`/`SpatialPoint` only; no payload or mock contract shape change beyond fixture data volume. | PASS (design) | PASS — types and client unchanged |
| FE-03 | Requested vs confirmed session separated; loading is in-place with `aria-busy`; errors announced and confirmed content retained; empty session distinguished from no sessions. | PASS (design) | PASS — failure/empty browser checks |
| FE-04 | Buttons/range/segments have names, keyboard operation, visible focus and non-colour cues; viewport checks at 390/768/1280/1440; 44 px CSS targets below 760. | PASS (design) | PASS — Enter on segment and ArrowRight on range asserted |
| FE-05 | No secrets, assistant content or browser persistence introduced. | PASS (design) | PASS — none present in the diff |

## Contracts, data, rollout

- HTTP: no changes. `GET /api/spatial/history` consumers already exist; only the e2e
  mock returns more sessions and points.
- Persistence/configuration: none.
- Compatibility: `SpatialCanvas` new props are optional; `SpatialPanel.vue` and
  `PresenceRailTile.vue` pass none and render as today.
- Rollback: revert the frontend files; no stored data or migration is involved.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| AC-01, AC-02, NFR-01 | `SpatialPositions.vue` | `devices.spec.ts` scroll stability at 1280×600, powered by `.nh-app-main.scrollTop` | pass — desktop and mobile projects |
| AC-03, AC-04 | `PositionsSessionRail.vue`, `SpatialPositions.vue` | rail/sheet selection, newest-first rows, pending state | pass — 3-row fixture, `aria-pressed` checks |
| AC-05, AC-20 | `SpatialPositions.vue` | points route returns 500; confirmed segment stays selected and alert appears | pass |
| AC-06, AC-07, AC-09 | `playback.ts`, `PositionsCoverageBar.vue` | `playback.test.ts` proportional/clipped segments; segment click and Enter | pass — 8 unit tests, keyboard assertion |
| AC-08, AC-14 | `PositionsTransport.vue`, `units.ts` | unit distance test; transport labels; ArrowRight step | pass |
| AC-10 | `playback.ts`, `PositionsTransport.vue` | episode unit tests; Next movement advances the range | pass |
| AC-11, AC-12 | `SpatialCanvas.vue`, `SpatialPositions.vue` | "No movement recorded" text, `line.nh-hop`, `circle.nh-pulse` | pass |
| AC-13 | `SpatialCanvas.vue` | reduced-motion CSS disables the pulse animation | **unrun** — CSS-only; no automated reduced-motion assertion |
| AC-15, AC-16, NFR-02, NFR-03 | `SpatialPositions.vue`, `PositionsSessionRail.vue` | 390 px sheet, rail hidden, no horizontal overflow, 44 px CSS targets | pass — overflow and sheet asserted; target size is CSS-set, not measured |
| AC-17 | `SpatialCanvas.vue` | no horizontal overflow at 768/1440; existing Positions/Presence assertions; full desktop+mobile suites | pass — 142/9 desktop, 151 mobile |
| AC-18, AC-19 | `SpatialPositions.vue` | sessions `[]` on Refresh; points `[]` for one session | pass |
| AC-21 | `SpatialPositions.vue` | hide/show recordings; selection preserved; rail removed/restored | pass — desktop and mobile projects |
| AC-22, NFR-07 | `SpatialPositions.vue`, `SpatialCanvas.vue` | viewer width/height, 2× and 5× shrink the viewBox, active target and pinned device visible | pass — automated |
| AC-23 | `SpatialPositions.vue`, `SpatialCanvas.vue` | enabled toggle shows/hides the labelled 90° fallback; saved angle untouched | pass — automated |
| AC-24 | `SpatialViewer.vue`, `SpatialPanel.vue`, `SpatialPositions.vue` | panel present above Live, Draw, Positions and Heatmap; absent in Setup/preview | pass — automated |
| AC-25 | `SpatialViewer.vue` | expanded desktop, collapsed phone, 44 px targets, no overflow | pass — CSS + viewport suites |
| AC-26 | `SpatialViewer.vue`, `SpatialCanvas.vue` | drag pans, Reset restores, Pan tool/Space shared; view/session resets | pass — pan/reset automated; Pan tool manual |
| NFR-04, NFR-05 | `playback.ts` | unit tests for helper bounds; 500-session stress not measured | partial — helpers are O(points); 500-segment render unrun |
| SC-03 | all | commands below | pass |

## Verification (recorded)

Environment: macOS, repository `frontend/` on 2026-11-06. No backend source changed, so
backend suites were not run.

- `npm test -- --runInBand` → **137 suites, 748 tests passed**.
- `npm run build` → **pass** (`vue-tsc` + `vite build`).
- `npm run lint` → **0 errors, 28 pre-existing warnings** (none in changed files).
- `npm run type-check:e2e` → **pass**.
- `npm run test:e2e -- --project=desktop` → **142 passed, 9 skipped**, exit 0.
- `npm run test:e2e -- --project=mobile` → **151 passed**, exit 0.
- `npm run test:e2e -- --project=desktop e2e/devices.spec.ts` → **9 passed**;
  same with `--project=mobile` → **9 passed**.
- Playback inside a recorded gap is asserted: Play → Pause appears and the scrubber value
  increases even while no sample changes.
- Follow-up (collapsible rail, fixed-height viewer, 1×/2×/3×/5× content zoom, FOV toggle,
  shared View panel and pan): the full suites above include these changes (142 passed/9
  skipped desktop, 151 passed mobile); `e2e/devices.spec.ts` also passed **9/9 on both
  projects**.

Not covered by automation (explicit unrun checks):

- reduced-motion visual check (CSS rule exists; no browser assertion);
- real-hub playback with more than the mocked 5–6 point sessions;
- 500-session coverage-bar stress;
- 44 px target measurement (set in CSS, covered indirectly by existing layout specs).

The pre-change scroll collapse was confirmed from the old `v-else-if="loading"` branch in
source. The new browser test selects the redesigned rail/sheet, so it was not executed
against the old component.

## Debt, risks, exceptions

- The 10 s movement-gap threshold (spec Q-01) is a presentation constant; validate against
  real sessions and tune without changing the contract.
- With many very short sessions the proportional bar relies on a minimum segment width;
  the rail/sheet remains the complete accessible list. Recorded in NFR-05.
- Deep-linking range/session/time remains out of scope (spec Q-03).
- The transport is in normal flow, not sticky, after the short-viewport overlap was
  observed during implementation. Users on very short viewports still scroll between the
  map and the transport; the scroll reset they reported is fixed.
- Zoom above 1× crops the room by design; the centre is anchored on the active sample when
  the level changes or the playhead is moved, playback leaves it fixed, and the device is
  pinned to the edge when it would otherwise leave the frame.
- The 90° FOV fallback is illustrative only and never saved; set the device angle in
  Presence → Draw → Room when it is known.
- Labels and markers target ~14 px; at extreme panel sizes that target is approximate.
- Pan is a view-only offset: it clears on view/session changes and is not persisted.
  Pinch and wheel zoom remain out of scope; the buttons are the zoom control.
- The shared View panel is used by Live, Draw, Positions and Heatmap only; Setup forms and
  the Home rail preview intentionally render no controls.
- Connection/session notice simplification is explicitly out of scope for this change and
  is captured in `TODO.md` for a design-first follow-up.
- Rail visibility and zoom are component state only, so they reset when the mode unmounts.
  Persisting them would need a reviewed preference store and is not in scope.
- Reduced-motion and large-session behaviour need review before merge.
- No exception or deviation is requested; no rule is marked N/A.

## Completion

Implementation and frontend verification are complete; artifacts record actual results
and unrun checks. This plan does not claim review approval.
