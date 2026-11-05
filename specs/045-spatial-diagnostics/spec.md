# Spec: Spatial diagnostics

ID: 045 | Status: Implemented (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

Users cannot tell whether jitter, ghost slots or a long clear delay are expected, caused by
interference, or a configuration mistake. The presence tab shows a single jitter estimate and
current state, but no per-slot evidence and no way to measure how long a zone takes to clear.

In scope: (1) per-slot diagnostics over the session's recorded frames — sample count, rate,
positional standard deviation, movement and a stability note; (2) duplicate/overlap evidence;
(3) a guided off-delay measurement against a derived zone, reporting the measured clear delay
including the hub's exit grace.

Exclusions: automatic interference diagnosis, firmware setting changes, long-term statistics
(activity history covers persisted trends), and any claim that a measurement proves a cause.

## Scenarios and acceptance

### US-01 — Per-slot stability (P1)

- AC-01: For each slot seen in the recorded window, show sample count, mean sample interval,
  positional σx/σy in millimetres, total movement and whether it looks stable, unsettled or
  noisy using documented thresholds.
- AC-02: Slots with too few samples are marked insufficient rather than rated.
- AC-03: The view shows how many frames were recorded and over what span; empty recording shows
  a clear message.
- AC-04: Copy states these are session measurements, not firmware guarantees.

### US-02 — Duplicate/overlap evidence (P2)

- AC-05: Count frames where two concurrently active slots are within the merge distance and
  report the closest pair, so ghost or reflection behaviour is visible without claiming a cause.

### US-03 — Measured zone clear delay (P1)

- AC-06: Choose a zone and start a measurement; the view waits for it to become occupied, records
  the entry time, then records when the derived zone reports clear and shows the elapsed delay.
- AC-07: The result is labelled as measured clear delay including the hub exit grace, not the
  firmware off delay; timeouts and missing derived data are explicit.
- AC-08: Starting, resetting and leaving the view does not change settings or automations.

## Requirements and success

- FR-01: Diagnostics maths MUST be pure and tested in `domain/spatial/diagnostics.ts`.
- FR-02: Measurement MUST only read derived zone signals and never write.
- NFR-01: No unbounded state; no timers left behind on unmount.
- SC-01: Domain tests cover σ, rate, stability thresholds, sparse slots, overlaps.
- SC-02: Full frontend checks plus the devices browser spec pass; results recorded.

## Questions and assumptions

- Q-01: σ over the session window is a useful stability signal; thresholds (40 mm stable,
  100 mm unsettled) are heuristics labelled as such.
- Q-02: Derived zone occupancy includes the hub's 800 ms exit grace, so a measured delay is an
  upper bound relative to raw firmware off delay.
