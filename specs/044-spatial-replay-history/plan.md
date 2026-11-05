# Plan: Spatial replay and persisted zone history

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Context and design

`useHistoryQuery` already wraps the rate-limited `/metrics/query` endpoint and returns boolean
and numeric series. The derived presence device carries per-zone entities with a `zoneId`
attribute, so activity can be requested without new persistence.

- `domain/spatial/activity.ts`: resolve zone entity names, build `HistoryRequest`, and compute
  per-zone share over known time, transitions, peak count, last occupancy and 48-bucket
  timelines. Pure, tested.
- `SpatialActivityHistory.vue`: range selector, request via `useHistoryQuery`, per-zone card
  with timeline, loading/empty/error and the "no raw positions" statement. New Activity view.
- `useSpatialTracking`: bounded replay buffer of accepted raw active positions (5 min,
  1,500 frames), exposed read-only.
- `domain/spatial/replay.ts` + `useSpatialReplay`: seek/play state and synthetic display tracks;
  timers only in the composable, cleared on unmount.
- `SpatialReplay.vue`: scrubber, play/pause, speed, timestamp/notice. `SpatialPanel` swaps the
  single canvas to recorded tracks and empty occupancy only while the Replay view is active.

## Constitution Check

| Rule | Evidence | Before |
| --- | --- | --- |
| NH-01 | Spec 044 | PASS |
| NH-02 | Replay clearly labelled; no store/history/automation writes | PASS |
| NH-03 | No route/auth/data change; existing endpoint only | PASS |
| NH-04 | Bounded buffers, cleared timers, pure domain | PASS (design) |
| NH-05 | Domain tests and full checks | PASS (design) |
| FE-01/02/03 | Domain maths; no wire change; partial/unknown visible | PASS (design) |
| FE-04 | Keyboard + viewport checks | PASS (design) |

## Traceability

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-01, AC-01–AC-05 | `activity.ts`, `SpatialActivityHistory.vue` | `activity.test.ts`: request discovery, known-time share, transitions, peak, partial/late samples | pass |
| FR-02, AC-06, AC-07, AC-09 | `replay.ts`, `useSpatialReplay.ts`, tracking buffer | `replay.test.ts`: window/cap, seek clamping, synthetic slot labels; bounded recording in tracking; no store writes | pass |
| AC-08, AC-10 | `SpatialReplay.vue`, `SpatialPanel.vue` | canvas swaps to recorded tracks and empty occupancy only in the Replay view; notice states session-only; type-check/build and devices spec pass (replay controls are not browser-driven yet) | pass (unit/type) |
| SC-02 | all | `npm test` (737 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` (18 passed) | pass |

## Completion

Implemented and checked: persisted zone activity from existing metrics with unknown time kept
partial, bounded session replay with seek/play/speed and a canvas swap that suppresses live
occupancy, plus explicit "session only / no raw positions stored" copy. Not browser-driven:
activity timelines and replay controls are covered by domain tests and type-checking only, and
real sensor history has not been inspected on a device.
