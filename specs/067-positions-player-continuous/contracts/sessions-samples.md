# Contract: `samples` on `GET /api/spatial/history?mode=sessions`

Status: Draft (spec 067) | Changed: additive | Consumers: browser only

## Producer

`spatialhistory.Service.Sessions` (`backend/internal/spatialhistory/service.go`) already
iterates every point of every chunk in the requested range to count points and slots. The same
iteration collects each recorded timestamp; no second read, index or stored copy is introduced.
The handler (`backend/internal/api/spatial_history.go`) serialises the existing
`[]spatialhistory.Session` unchanged, so the field appears wherever sessions already do.

## What a sample is

The recorder writes a point only when a target moves (`movementGateMM` = 30 cm) with a
`minSampleGap` of 1 s. A recorded timestamp therefore *is* evidence of movement, and the
timestamps are the raw activity of a recording.

An earlier draft exposed "movement episode starts" instead: the first sample and each sample
after a pause longer than 10 s. That marked where movement **resumed**, so a recording moving
continuously showed almost none of its samples, and the range timeline read as empty while the
operator could plainly see movement. Real data made the error obvious: 18 recorded samples
produced 10 marks. This contract exposes the samples.

## Payload

```json
[
  {
    "id": "8f14e45fceea167a",
    "from": 1737981000000,
    "to": 1737981600000,
    "points": 24,
    "slots": [1],
    "capped": false,
    "samples": [1737981000000, 1737981015000, 1737981030000]
  }
]
```

| Field | Type | Meaning |
| --- | --- | --- |
| `samples` | integer array, epoch ms, ascending | When the recording stored a position inside `[from,to]` |
| `samples` absent or `[]` | | Nothing recorded in the range |

`points` and `len(samples)` agree unless the sample cap reduced the list; `samples` is reduced
evenly to `MaxSamplesPerSession` (240) so it keeps the shape of the activity rather than only
its opening moments.

## Compatibility and cost

- Additive only; no migration, no stored-format change, no new route or mode.
- Replaces the draft-only `movements` field. Both are additive and optional; a browser that
  predates either renders the bar with no activity marks, which is what it did before spec 067.
- Bounded: at most 240 entries per session. A 500-session day adds at most 120,000 numbers;
  derivation is O(points) and retains one int64 per point instead of a copy of each Point.
- Consumers: `frontend/src/services/spatial-history.service.ts` (optional field) and the
  Positions range timeline. No other consumer reads `mode=sessions`.
- Recovery: if a deployment receives a payload with an unexpected `samples` shape, the client
  validates it (array of finite non-negative numbers) and ignores it rather than corrupting
  playback state (`sanitiseSamples`).
