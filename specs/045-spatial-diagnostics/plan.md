# Plan: Spatial diagnostics

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Context and design

The replay recorder already holds up to five minutes of accepted raw positions in memory
(`replayFrames`). Diagnostics reuse it read-only; no new recording, route or persistence.

- `domain/spatial/diagnostics.ts`: per-slot statistics (count, mean interval, σx/σy, movement,
  stability), overlap frames and closest pair, using `SPATIAL_MERGE_MM`; thresholds exported.
- `SpatialDiagnostics.vue`: slot table with notes, recording span, overlap summary, and a
  measured clear-delay control driven by the derived zone signals (armed → occupied → cleared),
  with explicit timeout and grace wording.
- New Diagnostics view; no backend or store changes.

## Constitution Check

| Rule | Evidence | Before |
| --- | --- | --- |
| NH-01 | Spec 045, ACs, exclusions, questions | PASS |
| NH-02 | Measurements labelled session/upper-bound; never claim cause | PASS (design) |
| NH-03 | No route/auth/data change | PASS (design) |
| NH-04 | Pure maths; read-only derived signals | PASS (design) |
| NH-05 | Domain tests and checks | PASS (design) |
| FE-01/02 | Logic in domain; no wire change | PASS (design) |

## Traceability

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-01, AC-01–AC-05 | `diagnostics.ts` | `diagnostics.test.ts`: σ, intervals, sparse, thresholds, overlap summary, recording extent | pass |
| FR-02, AC-06–AC-08 | `SpatialDiagnostics.vue` | type-check/build; state machine (armed → occupied → cleared, timeout, reset) reviewed; devices spec passes; measurement is not browser-driven | pass (unit/type) |
| SC-02 | all | `npm test` (740 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` (18 passed) | pass |

## Completion

Implemented and checked: per-slot σ/interval/movement ratings with sparse slots labelled, overlap
evidence without a cause claim, and a guided clear-delay measurement against derived occupancy
with timeout/reset and explicit grace wording. The measurement state machine and view are
type-checked and unit-covered indirectly; they are not browser-driven and no real-device numbers
have been collected.
