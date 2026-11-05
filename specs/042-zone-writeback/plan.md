# Plan: Firmware zone write-back

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: unreviewed draft | Updated: 2026-11-05
Constitutions: parent 1.0.0; backend 1.0.0; frontend 1.0.0

## Context and design

The rectangle path already exists end to end: `SpatialZone.entities` holds four coordinate
entity names, `backend/models/settings/spatial.go` validates the roles, and
`frontend/src/domain/spatial/device-zones.ts` builds exact native-unit commands that
`SpatialZoneDevice.vue` sends one by one through `useDeviceCommand`.

Polygon support reuses every part of that path and adds one role.

- `backend/models/settings/spatial.go`: `maxSpatialZoneEntities` 5; `FirmwareEntities` with
  five known roles and a bound; per-zone `offDelay` accepted for both shapes; a shape-aware
  validator rejects cross-shape mappings. Cloning copies the firmware map. Tests in
  `spatial_test.go`.
- `frontend/src/types/spatial.ts`: `ZONE_POLYGON_ROLE = 'polygon'`, `ZoneEntityRole`, and the
  zone entity map widened. `ZoneCoordinateRole` stays rectangle-only.
- `frontend/src/domain/spatial/device-zones.ts`: `isZonePolygonEntity` (text, read-write,
  available), `zonePolygonOptions`, `serializeZonePolygon(points)` (3–20 integer points inside
  the firmware envelope, `x:y;...`), `zonePolygonPlan(zone, exposes)`, and generic command rows
  that format numbers with units and text verbatim. Remove the obsolete "polygon unsupported"
  message; add polygon-mode guidance.
- `SpatialZoneDevice.vue`: branch on shape — polygon mapping select plus serialised preview;
  rectangle branch unchanged; both shapes expose an optional off-delay entity and value.
  `useWorkspaceEditor.mapZone` and the panel event accept the wider role.
- `types/spatial.ts`, `workspace.ts`: `FIRMWARE_ROLES`, optional `firmwareEntities`, defaults,
  copy and validation.
- `domain/spatial/device-zones.ts`: `FIRMWARE_SETTINGS` (label/kind),
  `firmwareEntityOptions`, `firmwareCommandPlan` (range/step in native units, switches),
  `zoneOffDelayPlan`, `zoneNumericOptions`.
- `SpatialFirmwarePanel.vue`: explicit role mapping, value input or switch, reported value,
  one apply action; wired as a sixth spatial view with help text.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 042 with ACs, exclusions and questions | PASS | pending |
| NH-02 | Writes are explicit requests with requested-vs-reported display; no confirmed-state claim | PASS (design) | pending |
| NH-03 | No route/auth/data-exposure change | PASS (design) | pending |
| NH-04 | Existing command path and persistence shape reused; shape-aware validation documented | PASS (design) | pending |
| NH-05 | Domain and backend validation tests; browser check for the editor | PASS (design) | pending |
| NH-06 | No destructive operation; non-atomic write warning retained | PASS (design) | pending |
| BE-01 | Validation stays in models; no transport change | PASS (design) | pending |
| BE-02 | Existing command path unchanged; text commands already validated | PASS (design) | pending |
| BE-04 | Workspace schema unchanged; mapping role validated | PASS (design) | pending |
| FE-01 | UI in the existing zone editor; logic in `domain/spatial` | PASS (design) | pending |
| FE-02 | No new wire contract; uses existing set-value command | PASS (design) | pending |
| FE-03 | Requested vs reported, pending and failure states kept | PASS (design) | pending |
| FE-04 | Keyboard and viewport checks on the existing editor | PASS (design) | pending |

## Traceability and verification

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-02, AC-02, AC-03, SC-01 | `types/spatial.ts`, `domain/spatial/device-zones.ts` | `device-zones.test.ts`: serialisation order, bounds, point count, plan blocking, shape isolation | pass |
| FR-01, AC-07, SC-02 | `backend/models/settings/spatial.go` | `spatial_test.go`: polygon role accepted for polygon, rejected for rectangle and vice versa, and more than one entity rejected | pass |
| FR-03, FR-04, AC-04–AC-06, AC-08 | `SpatialZoneDevice.vue`, `useWorkspaceEditor.ts`, `SpatialPanel.vue` | type-check/build; Playwright devices spec exercises the zone editor and rectangle path at 375/768/1440 | pass for rectangle regression; polygon editor is type-checked and unit-tested, not yet browser-driven |
| FR-05, AC-09–AC-13, AC-15 | `SpatialFirmwarePanel.vue`, `SpatialZoneDevice.vue`, `device-zones.ts` | `device-zones.test.ts` covers native-unit bounds/step, switches, unmapped roles, off-delay; type-check/build; view renders in the existing devices spec | pass |
| FR-06, AC-14 | `models/settings/spatial.go`, `workspace.ts` | Go validation/clone test; `workspace.test.ts` defaults, unknown role, independent copy, polygon+offDelay admission | pass |
| SC-03 | all | `go test ./...`, `npm test` (725 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` (18 passed) | pass |

## Risks, debt, excluded

- Firmware envelope and parser are assumptions (Q-01); a real EPL pass is required before
  claiming the polygon applies. Recorded, not claimed.
- Firmware polygon mode must already be enabled; not written here (P5b).
- Text entity is optimistic, so read-back proves delivery, not parsing.

## Completion

Implemented and checked for polygon write-back and firmware settings: explicit mappings,
native-unit validation, switches, per-zone off delay, requested-versus-reported display and
additive backend persistence. Q-01 (firmware envelope/parser) and the exact setting entity
units remain assumptions until a real EPL pass, so no real-device claim is made.
