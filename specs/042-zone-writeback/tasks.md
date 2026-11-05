# Tasks: Firmware zone write-back

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

- [x] T001 — Backend shape-aware zone role validation in `models/settings/spatial.go` with
  `spatial_test.go`; IDs: FR-01, AC-07; done: accepted/rejected combinations tested.
- [x] T002 — Polygon role and entity types in `types/spatial.ts`; IDs: FR-02; done: type-check.
- [x] T003 — Serialisation and plan in `domain/spatial/device-zones.ts` with tests; IDs:
  FR-02, AC-02, AC-03; done: order, bounds, count and blocking tested.
- [x] T004 — Polygon editor branch in `SpatialZoneDevice.vue` and role plumbing; IDs: FR-03,
  FR-04, AC-04–AC-06, AC-08; done: type-check and browser check.
- [x] T005 — Run checks and record evidence; gate: NH-05; done: results in plan.
- [x] T006 — Review diff against rules; gates: NH/BE/FE; done: findings resolved or raised.

- [x] T007 — Firmware mapping schema and validation in `models/settings/spatial.go` and
  `types/spatial.ts`/`workspace.ts`; IDs: FR-06, AC-14; done: Go and browser tests.
- [x] T008 — Firmware setting plans in `device-zones.ts` with native-unit bounds/step and
  switch handling; IDs: FR-05, AC-09–AC-12; done: domain tests.
- [x] T009 — `SpatialFirmwarePanel.vue` plus view wiring and workspace map setter; IDs:
  AC-09–AC-12, AC-15; done: type-check/build and browser spec.
- [x] T010 — Per-zone off-delay mapping/write in `SpatialZoneDevice.vue`; IDs: AC-13; done:
  domain test and type-check.
