# Investigation: "clicking a recording disables Play" and sparse motion

Date: 2026-11-20 | Branch: feat/positions-player-continuous
Store inspected: `backend/data/spatial.db` (copied to a scratch file; the live hub holds a lock)

## Question 1 — is the stored data broken?

No. The store holds 32 recordings across 7 days for one device. Every chunk that exists has
points, and the point count, span and movement episodes derived from it are internally
consistent. Sample from the real store:

| recording (UTC) | points | span | distance moved | largest jump |
| --- | --- | --- | --- | --- |
| 11-20 16:18–16:21 | 10 | 3 min 08 s | 2.9 m | 361 mm |
| 11-20 16:52–16:53 | 5 | 1 min 06 s | 1.3 m | 391 mm |
| 11-21 15:31–15:35 | 18 | 3 min 26 s | 5.9 m | 552 mm |
| 11-18 12:43–12:48 | 26 | 5 min 11 s | ~4 m | 728 mm |

## Question 2 — why does clicking a listed recording disable Play?

Reproduced against the real store. `Sessions()` looks back one `sourceIdleClose` (30 min) so a
chunk that *started before* the filter but holds points *inside* it is still listed; `Points()`
did not apply the same lookback. The browser therefore:

1. listed the recording (correctly, e.g. "10 points"),
2. loaded its detail and received `[]`,
3. rendered "No recorded positions in this session" and disabled Play with no explanation.

Minimal reproduction (`internal/spatialhistory/service_test.go`):

    chunk starts 20 min before the window, holds points inside it
    Sessions(from, to) -> 1 session, 2 points
    Points(id, from, to) -> 0 points

Fix: `Points()` applies the same `sourceIdleClose` lookback, so the list and the detail agree
about which chunks belong to a window. Verified: 2 points listed, 2 returned.

Note this bug is window-relative, not recording-relative: the same recording loads correctly
when the filter starts before its chunk key, and returns nothing when the filter starts after
it while the recording is still overlapping. That is why it appeared to affect "the last one"
in particular.

## Question 3 — "most recordings barely have any motion"

Not a defect. The recorder is movement-gated (`movementGateMM = 300` in
`internal/spatialhistory/model.go`) and stores a point only when a target has moved at least
30 cm. That produces two shapes of recording, both correct:

- **Many single-sample recordings.** A target appears and then holds still: one point, zero
  span, zero distance. 5 of 32 recordings are like this, and the UI now says "1 sample" and
  "Nothing to play or scrub" rather than showing `0 s · 1 points`.
- **Short, low-travel recordings.** A person moving slowly through a room produces few samples
  because the gate is 30 cm; 2–5 points over a minute is expected, not truncated data.

Nothing is being dropped: the stored point counts match `MaxPointsPerDay = 3000` headroom by
three orders of magnitude, no chunk is flagged `capped`, and no recording exceeded the 500
movement cap.

## Follow-ups not taken (require a decision)

- The movement gate is a recorder constant, not a display problem. If the operator wants more
  detail, `movementGateMM` (30 cm) and the 1 s `minSampleGap` are the two knobs, and lowering
  them increases stored volume for every device with history enabled. Not changed here.
- A listed-but-empty recording can still occur if a chunk is pruned between the two requests.
  The UI now states the disagreement ("lists N positions in this range, but none could be
  loaded") instead of implying the recording is empty.
