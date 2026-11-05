# Spec: Calibration slots from retained positions

ID: 051 | Status: Fixed (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend presence tab

## Problem

The Calibration "Live slot" dropdown could offer nothing even while a target was present a
moment earlier. Calibration read only the current display tracks; a still person emits no new
frames, so the track ages to `idle` and then disappears after the hold window. Standing still at
a marked position — exactly what calibration requires — therefore emptied the list, and a slot
captured after moving could become unselectable again.

## Scope and acceptance

- AC-01: Calibration keeps the most recent raw position per slot for the session, independent of
  the live track ageing/hold rules; slots remain selectable after the target clears.
- AC-02: Selecting a slot captures that retained position once and shows its timestamp and age.
- AC-03: Historical captures are usable; the view no longer requires the slot to be live at apply
  time, and says the position may be older than a few seconds.
- AC-04: With no position recorded at all, the view says so and keeps the select empty rather
  than inventing slots.
- AC-05: Recent-target retention is cleared on tracking reset (device/offline/mapping reset).
- AC-06: Unit test proves a cleared target's last raw position is retained; browser test proves a
  selected slot remains selected after ageing and reports its capture time.

## Requirements and success

- FR-01: Keep a small per-slot latest-raw map in the tracking composable, exposed read-only.
- FR-02: Calibration options and captures use that map; no new persistence or backend change.
- SC-01: Composable test and devices browser spec; full frontend checks pass.

## Assumptions

- Q-01: The last reported raw position is the right calibration reference for a person standing
  still; its age is displayed instead of blocking the measurement.
