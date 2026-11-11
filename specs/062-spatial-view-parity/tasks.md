# Tasks: Spatial view parity and presence card states

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

## Preparation

- [x] T001 — Resolve the presence-source, replay-distance and centering questions in
  `spec.md` (Q1–Q3); IDs: NH-01; done: spec questions replaced with resolutions.
- [x] T002 — Check parent/frontend rules against the design; IDs: NH-01, FE-01;
  done: `plan.md` Constitution Check.

## Presence card state

- [x] T003 — Add `derivedPresenceDisplay` domain cases (null is not-derived; true/false
  map to occupied/clear) in `src/__tests__/domain/spatial/derived.test.ts`; IDs: FR-01,
  AC-01…AC-04; done: case passes.
- [x] T004 — Implement the helper and use it in `HomePresenceCard.vue`,
  `PresenceRailTile.vue` and `SpatialPanel.vue`; IDs: FR-01, FR-02, AC-01…AC-04; done:
  unit passes and the Home browser case sees not-derived → Clear.

## Replay distance

- [x] T005 — Add `showDistances` to the Positions player, bind the viewer checkbox and
  pass it to the canvas; IDs: FR-03, AC-05…AC-07; done: build/type-check pass.
- [x] T006 — Update the Positions browser case to expect and exercise the Distance
  checkbox, and assert the Heatmap has none; IDs: AC-05…AC-07; done: desktop/mobile pass.

## Sensor-centered replay

- [x] T007 — Add geometry cases for `centeredSpatialView` in
  `src/__tests__/domain/spatial/radar.test.ts` (symmetric, wide aspect, offset sensor,
  invalid aspect); IDs: FR-04, AC-08…AC-10; done: 4 cases pass.
- [x] T008 — Implement the helper and use it for the Positions base view and
  symbol/label scale; IDs: FR-04, AC-08…AC-10; done: unit and browser centering checks
  pass.
- [x] T009 — Extend the Positions browser case with a sensor-center assertion and the
  mobile/tablet overflow checks; IDs: AC-08, AC-11; done: desktop and mobile pass.

## Compact presence overview

- [x] T012 — Replace the rail tile's full-width preview and oversized type with a compact
  thumbnail/placeholder row using the shared Home type scale; IDs: FR-10, AC-12…AC-15;
  done: build/type-check pass.
- [x] T013 — Make the overview header adaptive: a single sensor renders the row as the
  card with no separate header; editing keeps a header for the remove control and hides
  the row name; multiple sensors use a section header and named rows; IDs: AC-14, AC-16;
  done: single and two-sensor browser cases pass.
- [x] T014 — Render the mini preview on any spatial frame (`hasFrameData`), keeping the
  icon placeholder only when no frame exists; IDs: AC-17; done: build and Home cases pass.
- [x] T015 — Assert the compact tile height (< 100/110 px) in the single and multiple Home
  overview browser cases; IDs: AC-13; done: 13 desktop + 13 mobile Home tests pass.

## Refactor

- [x] T016 — Extract presence copy, labels, zone chips, overview and server models into
  `domain/spatial/presence.ts`; move Positions fit/symbol math into
  `domain/spatial/positions.ts`; add `useInViewport` and `usePresenceTile`; IDs: NFR-05,
  NFR-06; done: components contain rendering only.
- [x] T017 — Add unit coverage for the presence models and position metrics; IDs: NFR-05;
  done: `presence.test.ts` and `positions.test.ts` pass.
- [x] T018 — Re-run the full affected matrix after the refactor; IDs: NFR-05, NFR-06;
  done: 146 Jest suites / 821 tests, 24 desktop + 24 mobile browser tests, build, lint and
  type-check pass.

## Handoff

- [x] T010 — Run selected checks and record exact results in `plan.md`; gate: NH-05;
  done: 146 Jest suites / 821 tests, build/lint/type-check and 24 desktop + 24 mobile
  browser tests recorded; backend untouched.
- [x] T011 — Final diff review for unrelated changes, stale comments and live-view
  regressions; gates: NFR-02, FE-01; done: review found only scoped presentation/domain
  changes; the live viewer and playback controls were not changed.
