# Tasks: Positions viewer layout and playback legibility

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete (unreviewed)

Each task maps to requirement/acceptance IDs or a constitutional gate, with paths,
dependencies and completion evidence.

## Preparation

- [x] T001 — Resolve spec Q-01…Q-07 in `spec.md`. IDs: NH-01; done: gap threshold,
  rail order, colour, wording and no-interpolation assumptions recorded; direction
  confirmed by the requester, not a maintainer approval.
- [x] T002 — Complete the rule check in `plan.md` (NH-01…06, FE-01…05). Depends: T001;
  done: design and before-merge columns recorded; unrun items listed.

## Coverage and implementation

- [x] T003 — Add a three-session fixture with gaps to `frontend/tools/server/routes.mjs`.
  IDs: AC-03, AC-06, AC-10; done: `devices.spec.ts` Positions assertions pass.
- [x] T004 — Add a browser regression measuring `.nh-app-main.scrollTop` across a
  recording selection at 1280×600 in `frontend/e2e/devices.spec.ts`. IDs: AC-01,
  NFR-01; done: passes on desktop and mobile projects. The old component was not
  executed with the new selectors; the collapse mechanism is documented from source.
- [x] T005 — Implement `frontend/src/domain/spatial/playback.ts` and
  `src/__tests__/domain/spatial/playback.test.ts`. IDs: AC-06, AC-08, AC-10, AC-14,
  NFR-04, NFR-05; done: 8 unit tests pass.
- [x] T006 — Implement `frontend/src/composables/useSpatialPlayback.ts` (200 ms clock,
  speed, seek, pause, scope cleanup). IDs: AC-08, AC-10; done: used by the viewer;
  interval cleared in `onScopeDispose`.
- [x] T007 — Restructure `SpatialPositions.vue` state: requested vs confirmed session,
  in-place loading with `aria-busy`, newest default, stale-response guard, reversible
  failure. IDs: AC-01, AC-02, AC-04, AC-05, AC-20; done: scroll and failure browser
  checks pass.
- [x] T008 — Add `PositionsCoverageBar.vue` (proportional clipped segments, gaps,
  playhead, played fill, ticks, pending state, keyboard buttons). IDs: AC-06, AC-07,
  AC-09, NFR-05; done: click and Enter selection asserted.
- [x] T009 — Add `PositionsSessionRail.vue` with newest-first rows, local time range,
  duration/points/slots, capped and pending states; desktop rail plus mobile sheet.
  IDs: AC-03, AC-04, AC-15; done: sheet selection closes without scroll movement.
- [x] T010 — Add `PositionsTransport.vue`: `UiRange` scrubber, elapsed/total text,
  speed and path controls, played distance via `formatDistance`, movement ticks, idle
  gaps, `Next movement`, "no movement recorded" status. IDs: AC-08, AC-10, AC-11,
  AC-14, AC-18, AC-19; done: browser assertions pass.
- [x] T011 — Extend `SpatialCanvas.vue` with optional `recentTrails`, `hopLinks`,
  `pulseAt` and styles, including reduced-motion. IDs: AC-12, AC-13, AC-17; done:
  Positions cues render; live surfaces pass their existing assertions.
- [x] T012 — Responsive layout: two columns at ≥900 px, sticky rail, canvas height
  caps, in-flow transport, sheet below 900 px, 44 px targets. IDs: AC-15, AC-16,
  NFR-01, NFR-02, NFR-03; done: 390 px sheet/overflow checks pass. The initially
  planned sticky transport was removed after it covered the coverage timeline at 600 px.
- [x] T013 — Extend `frontend/e2e/devices.spec.ts` for scroll stability, segment
  selection, keyboard, next movement, hop/pulse, idle wording, failure, empty states,
  mobile sheet and overflow. IDs: AC-01…AC-20; done: devices spec 9/9 on both projects.
- [x] T014 — Verify unchanged surfaces: full desktop and mobile suites include the
  Heatmap, Presence canvas and `PresenceRailTile` assertions. IDs: AC-17; done:
  142 passed/9 skipped desktop, 151 passed mobile.

## Handoff

- [x] T015 — Planned checks from `frontend/`: `npm test -- --runInBand` (748 passed),
  `npm run lint` (0 errors, 28 pre-existing warnings), `npm run build` pass,
  `npm run type-check:e2e` pass, full e2e desktop and mobile pass. Gate: NH-05; done:
  commands and results recorded in `plan.md`. Backend suites not run (no backend change).
- [x] T016 — Record failure and viewport verification for 390, 768, 1280 and 1440 px plus
  keyboard activation in `plan.md`. IDs: AC-13, AC-16, SC-04; done: failure/viewport/
  keyboard recorded; reduced-motion remains an explicit unrun check.
- [x] T017 — Update spec/plan status, traceability and debt; final rule check.
  Gates: NH-01, NH-04, NH-05; done: unrun checks listed; no claimed approval.

## Follow-up (rail density and viewer size)

- [x] T018 — Make the recordings rail collapsible from its header and restorable from
  the toolbar; collapse the grid to one column while hidden. IDs: AC-21, FR-14; done:
  hide/show asserted, confirmed segment unchanged, rail header stays fixed while the
  list scrolls.
- [x] T019 — Replace the width-driven square with a full-width viewer clamped to
  `clamp(16rem, 40vh, 24rem)`; fit the room and session points at 1×. IDs: AC-22,
  NFR-07; done: width equals the column, height inside the clamp, rail collapse widens
  without changing the height.
- [x] T020 — Add the `1× 2× 3× 5×` toolbar group; zoom shrinks the viewBox around the
  recorded points and a computed symbol scale keeps labels/markers readable. IDs: AC-22;
  done: 2× halves the viewBox width, the point stays in view, panel height unchanged.
- [x] T021 — Extend `devices.spec.ts` with rail collapse/restore and zoom assertions.
  IDs: AC-21, AC-22, SC-05; done: passes desktop and mobile.
- [x] T022 — Re-run checks after the zoom change: unit 748 passed, build pass, lint 0
  errors/28 pre-existing warnings, type-check:e2e pass, full e2e 142/9 desktop and 151
  mobile, devices spec 9/9 both projects. Gate: NH-05.
- [x] T023 — Add a display-only “Field of view” checkbox driven by the configured angle,
  with a labelled 90° fallback when none is set; add `showFov`/`fovAngle` props. IDs:
  AC-23, FR-16; done: saved angle untouched.
- [x] T024 — Assert the enabled toggle and the 90° cone in `devices.spec.ts`; re-run
  checks after the toggle. IDs: AC-23, SC-05; done: results recorded in `plan.md`.
- [x] T025 — Fix the 5× label halo: pass a `labelStroke` derived from the symbol size
  instead of the fixed 8 user units. IDs: AC-22; done: the empty-state 5× view renders
  the sensor dot and a normal “Sensor” label without the surface-coloured blob.
- [x] T026 — Replace the device dot with a presence glyph in a filled disc, and pin it to
  the nearest view edge when zoom moves it off-screen. IDs: AC-22; done: device remains
  visible at 2× and 5×.
- [x] T027 — Anchor zoom above 1× on the active sample when the level changes or the
  playhead moves, not on every playback tick, so the target travels across a fixed view.
  IDs: AC-22; done: target visible at 2× and 5×.
- [x] T028 — Re-run checks after the device/zoom/FOV changes. Gate: NH-05.
- [x] T029 — Grow the target 25% per zoom level (capped at 2×), re-measure the viewer after
  rail/view layout changes so the viewBox aspect matches the box, and assert that playback
  advances inside a recorded gap. IDs: AC-22; done: aspect assertion and Play-in-gap
  assertion pass.
- [x] T030 — Create `SpatialViewer.vue` (collapsible View panel with zoom, Grid, FOV, an
  optional display slot, pointer/touch pan, Reset and `resetKey`), add `panMode` to
  `SpatialCanvas`, the `tune` icon and `SPATIAL_ILLUSTRATIVE_FOV`. IDs: AC-24, AC-25,
  AC-26; done: shared panel renders above the canvas.
- [x] T031 — Wire `SpatialPanel` (panel above the canvas, move Smoothing/Scale/Orientation/
  Raw into the extra slot, add zoom/FOV and the Pan tool) and `SpatialPositions` (remove
  toolbar zoom/FOV, pass the base viewBox and reset key). IDs: AC-24, AC-26; done: all
  full spatial surfaces share the panel.
- [x] T032 — Update `devices.spec.ts`: open the new panel instead of the Display summary,
  assert select baselines, and add drag-pan + Reset. IDs: AC-24, AC-26, SC-06; done:
  desktop and mobile spatial spec passes.
- [x] T033 — Re-run unit/build/lint/type-check and the full desktop/mobile suites after the
  shared viewer change. Gate: NH-05.

## Unrun before merge

- Reduced-motion visual verification.
- Real-hub playback and a 500-session coverage bar.
- 44 px target measurement is CSS-set and only indirectly covered.
