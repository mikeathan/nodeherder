# Plan: Paired-point spatial calibration

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Context and design

The canvas transform is `roomPoint(point, sensor) = rotate(point, sensor.rotation) + sensor`.
Calibration recovers exactly that pose from correspondences: with centred measured (`m`) and
room (`r`) points, `angle = atan2(Σ(mx·ry − my·rx), Σ(mx·rx + my·ry))`, translation
`t = r̄ − R·m̄`, and residuals per pair. Two distinct pairs determine the pose; more pairs
reduce noise. Scale is fixed at 1:1 because both frames are millimetres.

- `domain/spatial/calibration.ts`: pair/solution types, `solveCalibration` (2–5 pairs, finite
  checks, degenerate rejection, rotation normalisation, RMS/worst residual, 200 mm warning and
  500 mm block), `calibrationResiduals`.
- `SpatialCalibrationPanel.vue`: pair rows (slot select from current tracks, room X/Y inputs),
  solved pose and residuals, apply-to-draft, mapped-only angle write with explicit warning.
- `SpatialPanel.vue`, `view.ts`, `useWorkspaceEditor.ts`: a Calibration view, `setSensor`
  (single undo checkpoint) and wiring. Existing room editor and save path unchanged.

## Constitution Check

| Rule | Evidence | Before |
| --- | --- | --- |
| NH-01 | Spec 043, ACs, exclusions, questions | PASS |
| NH-02 | Angle write explicit and labelled unverified; no implicit firmware changes | PASS (design) |
| NH-03 | No route/auth/data change | PASS (design) |
| NH-04 | Pure maths in domain; draft-only apply; existing save | PASS (design) |
| NH-05 | Domain tests and run checks | PASS (design) |
| FE-01/02 | Logic in `domain/spatial`; no new wire contract | PASS (design) |
| FE-03 | Blocked/warning/applied states distinct; residuals shown | PASS (design) |
| FE-04 | Bootstrap/card layout rules; keyboard + viewport check | PASS (design) |

## Traceability

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-01, AC-02, AC-03, SC-01 | `domain/spatial/calibration.ts` | `calibration.test.ts`: identity, rotation, translation, noise, degenerate, non-finite, normalisation | pass |
| FR-02, AC-04, AC-05 | `SpatialCalibrationPanel.vue`, `useWorkspaceEditor.ts` | type-check/build; apply path calls `setSensor` (single checkpoint) and never saves | pass |
| FR-03, AC-06, AC-07 | `SpatialCalibrationPanel.vue` | mapped-only angle write labelled unverified; stale rows blocked by live-slot filtering | pass |
| SC-02 | all | `go test ./...`, `npm test` (730 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` (18 passed) | pass |

## Completion

Implemented and checked: pure rigid solver with residual thresholds, capture UI with live-slot
raw values, draft-only apply, and an explicit mapped-only angle write. Calibration is not
applied automatically, and no real-device alignment claim is made yet (Q-01). The canvas
room-point picker remains excluded (Q-02).
