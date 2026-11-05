# Spec: Paired-point spatial calibration

ID: 043 | Status: Implemented (unreviewed) | Owner: unreviewed draft | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend (with existing firmware write path)

## Problem, scope, exclusions

A room can be drawn and zones placed, but the sensor pose (`x`, `y`, `rotation`) is entered by
hand. A user with a tape measure has no way to align the measured radar frame with marked room
positions, so zones and heatmaps can sit visibly off.

In scope: capture two to five reference pairs (a live target's measured position and its known
room position in millimetres), solve the rigid transform (rotation + translation, no scale),
show the residual, apply the result to the room's sensor pose as an unsaved draft, and offer an
explicit opt-in write of the computed angle to a mapped firmware installation-angle entity.

Exclusions: automatic detection of reference points, scale/skew correction, multi-sensor
fusion, any automatic firmware write, and persistence beyond the existing workspace save.

## Scenarios and acceptance

### US-01 — Calibrate the room from measured positions (priority P1)

- AC-01: Given at least two live slots and marked room positions, when the user selects a slot
  and enters its room X/Y, then the pair is captured using that slot's raw measured position;
  up to five pairs are supported.
- AC-02: Given two or more valid, distinct pairs, when solved, then the result contains rotation
  in degrees, translation in millimetres, RMS residual, and the worst pair's residual.
- AC-03: Given fewer than two pairs, non-finite values, or measured positions that collapse to a
  single point, then solving is blocked with a reason and nothing is applied.
- AC-04: Given a solution whose RMS residual exceeds 500 mm, then applying is blocked; between
  200 mm and 500 mm it may be applied with a visible warning.
- AC-05: Given a valid solution, when applied, then the room's sensor pose becomes an unsaved
  draft (existing save/undo semantics); no firmware write happens implicitly.
- AC-06: Given a mapped installation-angle entity, then an explicit "write angle" action sends
  the normalised angle once; it is labelled unverified because firmware angle semantics can
  differ. Without a mapping the action is unavailable.
- AC-07: Given a slot stops reporting, then its pair is marked stale and blocked until
  recaptured; a captured value is never silently refreshed.
- AC-08: The calibration view is keyboard operable and fits 375/768/1440 px.

## Requirements and success

- FR-01: Calibration maths MUST be pure and deterministic in `domain/spatial/calibration.ts`.
  Acceptance: AC-02, AC-03.
- FR-02: Applying MUST only mutate the draft sensor pose. Acceptance: AC-05.
- FR-03: Firmware angle writing MUST be explicit, optional and mapped-only. Acceptance: AC-06.
- FR-04: Residual thresholds MUST block unsafe applications. Acceptance: AC-04.
- NFR-01: No new persistence field, route or background work.
- SC-01: Domain tests cover identity, rotation, translation, noise, degenerate and invalid input.
- SC-02: Full suites, build, lint, type-check and the devices browser spec pass; results recorded.

## Questions and assumptions

- Q-01: The workspace sensor pose maps measured millimetres with the same convention as
  `rotatePoint`: `room = R(rotation)·measured + (x, y)`. Verified by unit round-trip; a
  real-device pass still decides whether users find the result aligned.
- Q-02: Users can obtain room coordinates from the room outline/grid; this increment does not
  add a room-point picker on the canvas.
