# Tasks: Distance-to-target labels in the spatial viewer

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete (unreviewed)

## Preparation

- [x] T001 — Confirm the data source: the frame contract has no distance, and
  `target_distance` is a single scalar unrelated to the mapped axes. IDs: NH-01, NH-04;
  done: recorded in `spec.md` (Problem/facts) and `plan.md` alternatives.
- [x] T002 — Resolve source, precision and scope with the owner (computed from mapped x/y;
  0.1 m steps; viewer only). IDs: NH-01.

## Coverage and implementation

- [x] T003 — Add `spatialTargetLabel(label, millimetres, showDistance)` to
  `src/domain/spatial/radar.ts` with cases in `src/__tests__/domain/spatial/radar.test.ts`
  (quantisation, null handling, formatQuantity unit). IDs: FR-03, FR-04, FR-05, AC-05,
  AC-07; depends: T002.
- [x] T004 — `SpatialCanvas.vue`: compute the distance from the position actually drawn and
  add a `showDistances` prop gating the label. IDs: FR-02, AC-03, AC-04, AC-06, NFR-01;
  depends: T003.
- [x] T005 — `SpatialViewer.vue` + `SpatialPanel.vue`: Distance checkbox beside Grid and
  Field of view, default off, wired to both canvases. IDs: FR-01, AC-01, AC-02, AC-08;
  depends: T004.
- [x] T006 — Extend the spatial browser test with the toggle scenario (labels change,
  keyboard reachable, no placeholder when unknown). IDs: AC-01…AC-08; depends: T005.

## Handoff

- [x] T007 — Run and record: `npm test -- --runInBand`, `npm run lint`, `npm run build`,
  `npm run type-check:e2e`, and the spatial browser test on desktop and mobile.
  Gate: NH-05.
- [x] T008 — Self-review the diff: no duplication of the projection maths, no new reactive
  source, accessible name, `plan.md` evidence updated. Gates: NH-04, FE-04.

## Defect found during implementation

- [x] T009 — `SpatialCanvas.vue` never rendered track labels: the `showTargetLabels`
  guard compared against `false`, but Vue casts an absent Boolean prop to `false`, so the
  condition was always false. Give the prop a `true` default via `withDefaults`; preview
  surfaces keep opting out explicitly. IDs: AC-02, FR-01; done: label renders, and the
  live/Positions/Heatmap surfaces are covered by `e2e/devices.spec.ts` on both projects.
