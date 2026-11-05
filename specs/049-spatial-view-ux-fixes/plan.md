# Plan: Spatial view switching, replay position and calibration capture

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Changes

- `SpatialPanel.vue`: `showCanvas` excludes Activity/Firmware/Fusion/Diagnostics; replay banner
  above the workspace; entering Replay pauses and seeks to the first frame.
- `useSpatialReplay.ts`: stop resetting the clock to the newest frame on every update after the
  first initialisation, so paused replay cannot follow live data.
- `SpatialCanvas.vue`: computed aria label distinguishes replay from live.
- `SpatialCalibrationPanel.vue`: options include all reported slots (aged labelled), selection
  captures `measuredX/Y` once with a timestamp, re-selecting refreshes it, and stale rows block
  applying.
- `e2e/devices.spec.ts`: regression coverage for content-view canvas hiding, replay banner and
  replay canvas label, play/pause, and a slot surviving 3.5 s of ageing.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01, AC-06 | devices spec asserts no live canvas in Zone activity and its heading visible; live canvas still asserted elsewhere | pass |
| AC-02, AC-03 | devices spec asserts Replay canvas, banner and play/pause; replay clock no longer follows frames | pass |
| AC-04, AC-05 | devices spec selects slot 1, waits 3.5 s, asserts the selection persists and the stale row notice appears | pass |
| SC-02 | `npm test` (747 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` desktop + mobile (18 passed) | pass |

## Completion

Implemented and verified. Content views no longer show the live scene, Replay is pinned to the
scrubber with an explicit banner/label, and calibration selections survive ageing with captured
measurements and stale blocking. Browser coverage was added to `devices.spec.ts`.
