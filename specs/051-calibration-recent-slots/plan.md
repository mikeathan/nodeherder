# Plan: Calibration slots from retained positions

Spec: [spec.md](spec.md) | Status: Fixed (unreviewed) | Updated: 2026-11-05

## Change

- `useSpatialTracking.ts`: keep `recentTargets` (slot, raw x/y, sample time), updated for every
  accepted sample from frames or sampled coordinates and cleared on reset.
- `useSpatialView.ts`, `SpatialPanel.vue`: expose/pass it to the calibration panel.
- `SpatialCalibrationPanel.vue`: options come from `recentTargets`; capture stores the retained
  position/time; remove the live-only stale gate; show capture time and age; explain when no
  position has been recorded.
- Tests: composable retention test; browser test keeps the selection after 5 s and checks the
  capture text.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01, AC-02, AC-05 | `useSpatialTracking.test.ts` retains the cleared slot's last raw position | pass |
| AC-02, AC-03, AC-06 | devices browser spec selects a slot, waits 3.5 s, keeps the value and shows "Captured … at" | pass |
| AC-04 | empty state covered by the panel; no positions → explicit message | pass (review) |
| SC-01 | `npm test` (750 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` desktop + mobile (18 passed) | pass |

## Completion

Implemented and verified. Calibration now derives its slots from retained per-slot raw
positions, so a still or briefly cleared target stays selectable, and the captured time/age is
shown. Full checks are recorded above.
