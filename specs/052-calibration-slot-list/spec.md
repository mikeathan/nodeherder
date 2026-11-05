# Spec: Calibration lists configured slots

ID: 052 | Status: Fixed (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend presence tab

## Problem

After deploying the retained-position fix, the Calibration "Live slot" dropdown could still be
impossible to use on a real device: it only listed slots that had reported since the tab opened.
Without a recent sample (still target, sensor idle, or no frame yet) the list was empty, so there
was nothing to choose even though firmware slots exist.

## Scope and acceptance

- AC-01: The dropdown always lists every configured mapping slot (1..row count), whether or not
  it has reported, labelled `no position yet` when it has none.
- AC-02: Slots that have reported show their latest raw position and time; an explicit
  `Capture slot N` button refreshes the captured value without reselecting.
- AC-03: Selecting a slot and capturing works with no live target; the panel tells the user to
  move into view when no position exists.
- AC-04: Applying is blocked with a per-row message when a selected slot has no recorded
  position, instead of silently producing a NaN solution error.
- AC-05: Existing retained-position, ageing and capture behaviour is unchanged.
- AC-06: Browser coverage asserts the configured slots are present in the option list.

## Requirements and success

- FR-01: Slot options are the union of configured slots and any recently reported slots.
- FR-02: Capture reads the retained latest sample; missing data is explicit, not invented.
- SC-01: Devices browser spec counts the options and captures slot 1; full checks pass.

## Assumptions

- Q-01: Mapping row count equals the firmware slot count the tab can read; this is the same
  mapping the engine uses.
