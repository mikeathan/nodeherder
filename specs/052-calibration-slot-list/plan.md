# Plan: Calibration lists configured slots

Spec: [spec.md](spec.md) | Status: Fixed (unreviewed) | Updated: 2026-11-05

## Change

- `SpatialPanel.vue`: pass `mappings.length` as `slot-count` to calibration.
- `SpatialCalibrationPanel.vue`: options are the union of `1..slotCount` and recently reported
  slots; reported ones show raw position/time, absent ones show `no position yet`. Add per-row
  `Capture slot N`, explicit missing-position message and apply guard.
- Option labels stay static (`Slot 1`, `Slot 2`, …); live coordinates moved beside the select.
  Updating option text rebuilt the native dropdown while it was open and swallowed the click, so
  the selection never committed even though options were visible.
- `devices.spec.ts`: assert four options, a static `Slot 1` label, capture by label, and latest
  raw values beside the control.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01, AC-06 | browser spec counts options, checks static labels, captures by label | pass |
| AC-02, AC-03 | capture button and retained target; missing-data copy | pass (type/browser) |
| AC-04 | missingRows blocks apply with a row message | pass (type) |
| follow-up | live option labels rebuilt the open dropdown and swallowed clicks; labels are now static | pass (browser) |
| AC-05 | existing ageing/capture browser assertions unchanged | pass |
| SC-01 | `npm test` (750), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` desktop + mobile (18 passed) | pass |

## Completion

Implemented and verified: the dropdown always contains every configured slot, absent positions
are labelled, capture is explicit per row, and applying is blocked with a row message when a slot
has no recorded position.
