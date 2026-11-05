# Plan: Calibration lists configured slots

Spec: [spec.md](spec.md) | Status: Fixed (unreviewed) | Updated: 2026-11-05

## Change

- `SpatialPanel.vue`: pass `mappings.length` as `slot-count` to calibration.
- `SpatialCalibrationPanel.vue`: options are the union of `1..slotCount` and recently reported
  slots; reported ones show raw position/time, absent ones show `no position yet`. Add per-row
  `Capture slot N`, explicit missing-position message and apply guard.
- `devices.spec.ts`: assert four options ("Not captured" + three slots) and that capture works.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01, AC-06 | browser spec counts options before/while capturing | pass |
| AC-02, AC-03 | capture button and retained target; missing-data copy | pass (type/browser) |
| AC-04 | missingRows blocks apply with a row message | pass (type) |
| AC-05 | existing ageing/capture browser assertions unchanged | pass |
| SC-01 | `npm test` (750), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` desktop + mobile (18 passed) | pass |

## Completion

Implemented and verified: the dropdown always contains every configured slot, absent positions
are labelled, capture is explicit per row, and applying is blocked with a row message when a slot
has no recorded position.
