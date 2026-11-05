# Tasks: Firmware zone write-back

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete (P5b pending)

- [x] T001 — Backend shape-aware zone role validation in `models/settings/spatial.go` with
  `spatial_test.go`; IDs: FR-01, AC-07; done: accepted/rejected combinations tested.
- [x] T002 — Polygon role and entity types in `types/spatial.ts`; IDs: FR-02; done: type-check.
- [x] T003 — Serialisation and plan in `domain/spatial/device-zones.ts` with tests; IDs:
  FR-02, AC-02, AC-03; done: order, bounds, count and blocking tested.
- [x] T004 — Polygon editor branch in `SpatialZoneDevice.vue` and role plumbing; IDs: FR-03,
  FR-04, AC-04–AC-06, AC-08; done: type-check and browser check.
- [x] T005 — Run checks and record evidence; gate: NH-05; done: results in plan.
- [x] T006 — Review diff against rules; gates: NH/BE/FE; done: findings resolved or raised.
