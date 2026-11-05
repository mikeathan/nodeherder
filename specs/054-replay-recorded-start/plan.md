# Plan: Replay starts on recorded positions

Spec: [spec.md](spec.md) | Status: Fixed (unreviewed) | Updated: 2026-11-05

## Change

- `domain/spatial/replay.ts`: `firstRecordedFrame`, `lastRecordedFrame`, `hasRecordedPositions`.
- `useSpatialReplay.ts`: initialise to the last recorded frame; start from the first recorded
  frame; expose `hasPositions`; refuse Play without positions.
- `SpatialPanel.vue`: entering Replay seeks the first frame with positions.
- `SpatialReplay.vue`: uses recorded range for the scrubber, reports frame count/span, disables
  Play without positions, and explains in-memory/no-list/no-hub storage.
- Tests: domain helpers and browser assertions for a rendered replay target and frame count.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01, AC-02 | `replay.test.ts` skips empty frames; UI guards Play | pass (unit) |
| AC-03, AC-04, AC-05 | `SpatialReplay.vue` copy and recorded-range status | pass (review/type) |
| AC-06 | devices e2e asserts `circle.nh-target` and "frame(s) with positions" | pass |
| SC-01 | `npm test` (751), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` desktop + mobile (18 passed) | pass |

## Completion

Implemented and verified: Replay opens on the oldest frame with positions, Play is disabled when
nothing was recorded, the scrubber covers the recorded range, and the UI explains the in-memory
buffer/no-recordings-list behaviour. Browser coverage asserts a rendered replay target.
