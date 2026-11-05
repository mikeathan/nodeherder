# Plan: Calibration guidance and semantics

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Change

`SpatialCalibrationPanel.vue`: replace the terse intro with an explanation and an open
"How to calibrate (and what a slot is)" block (numbered steps, slot semantics, room-coordinate
example, apply/save/residual explanation). Rename labels and the apply button. Browser spec
follows the renamed slot label.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01–AC-05 | copy/label review in the rendered component; type-check | pass |
| AC-06 | devices browser spec selects the renamed slot label and captures | pass |
| SC-01 | `npm test` (750), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` desktop + mobile (18 passed) | pass |

## Completion

Copy, labels and the browser flow are updated and verified. A canvas point-picker for room
coordinates remains the main usability improvement and is proposed as a follow-up.
