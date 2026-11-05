# Spec: Replay starts on recorded positions

ID: 054 | Status: Fixed (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend presence tab

## Problem

After capturing a target, opening Replay showed a time and banner but appeared to contain
nothing, and Play was always offered even when no target position had been recorded.

The session buffer records every incoming frame, including the device's initial empty frame and
later cleared frames. Entering Replay seeked to the first buffered frame, which is often empty,
so the canvas rendered no positions. Play also had no guard on recorded positions, and the UI
never said where recordings live.

## Scope and acceptance

- AC-01: Entering Replay starts at the oldest frame that actually contains target positions,
  skipping empty/cleared frames.
- AC-02: Play is disabled, and no scrubber is shown, when no frame contains positions; the view
  distinguishes "frames arrived but no positions" from "nothing recorded".
- AC-03: The scrubber is bounded to the recorded position range; the status reports the current
  frame's slots, the number of frames with positions, and the recorded time span.
- AC-04: The copy states plainly that this is a rolling five-minute in-memory buffer, that there
  is no recordings list, that nothing is stored on the hub, and that reloading clears it.
- AC-05: Replay still never writes to the store/history/automations, and the replay canvas is
  labelled as replay with live tracking not shown.
- AC-06: Browser coverage proves recorded positions appear on the replay canvas after entering.

## Requirements and success

- FR-01: Pure helpers find first/last recorded frames and whether positions exist.
- FR-02: Playback state uses those helpers for initial time, start and enablement.
- SC-01: Domain tests for empty-frame skipping; devices e2e asserts a replay target and frame
  count; full checks pass.

## Assumptions

- Q-01: Empty frames are worth keeping for clear-state playback, so they are skipped when
  seeking rather than removed from the buffer.
- Q-02: There is intentionally no recordings library or persistence in this increment; adding
  clips/persistence is a separate feature.
