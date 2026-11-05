# Spec: Calibration guidance and semantics

ID: 053 | Status: Implemented (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend presence tab

## Problem

Users cannot tell what Calibration actually does, whether a slot records anything, or how to
obtain the room coordinates it asks for. The previous copy named "Live slot" and "Room X/Y"
without explaining that slots are firmware target IDs (not recordings), that room coordinates
are plan coordinates in millimetres, or what Apply changes.

## Scope and acceptance

- AC-01: The view states in plain language that calibration aligns measured positions with the
  room plan, that it does not record presence/video, and that it does not move zones or outline.
- AC-02: It explains slots: firmware targets 1–3, reused by different people, not recordings;
  the chosen slot keeps its latest captured position for the session.
- AC-03: It explains room coordinates as plan coordinates in millimetres, with a worked example
  (`(0, 2000)` two metres straight ahead when the sensor is assumed at the origin), and says two
  pairs are the minimum, three improve fit.
- AC-04: It explains Apply: updates sensor position/rotation in the unsaved draft; Save spatial
  settings persists; residual describes fit.
- AC-05: Labels say "Firmware slot (measured)", "Known room X/Y", "Apply pose to room draft".
- AC-06: Browser flow still selects a slot by label and captures; existing behaviour unchanged.

## Requirements and success

- FR-01: Copy and labels only; no solver or persistence change.
- SC-01: Type-check, full frontend checks and devices e2e pass.

## Assumptions

- Q-01: Users can read room coordinates from the room outline/grid. A canvas point-picker would
  remove this mental step and is proposed as a follow-up, not part of this change.
