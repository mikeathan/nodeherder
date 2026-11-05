# Spec: Spatial view switching, replay position and calibration capture

ID: 049 | Status: Implemented (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend presence tab

## Problems

1. Selecting **Zone activity** (and other content views) looked like nothing happened: the live
   canvas stayed above the content, so the page still appeared to be the live view.
2. **Replay** continued to show live positions while paused and while playing: the replay clock
   followed every incoming frame, so the canvas tracked live data instead of the scrub position.
3. **Calibration** dropped the selected live slot: options were filtered to `live` tracks, so a
   still target ageing to `idle` after three seconds disappeared from the list, and the measured
   position was re-read from the moving track instead of being captured.

## Scope and acceptance

- AC-01: Content-only views (Zone activity, Firmware settings, Multi-sensor, Diagnostics) do not
  render the live canvas; switching to them is visibly different and the content is at the top.
- AC-02: Entering Replay pauses it at the oldest recorded frame; while paused, incoming live
  frames do not move the replay time; the canvas is labelled as replay and a banner states that
  live tracking is not shown.
- AC-03: Play/pause and speed operate on the replay clock only; leaving Replay restores the live
  canvas.
- AC-04: Calibration keeps the chosen slot across ageing ticks; the measured position is captured
  once and shown with its capture time; re-selecting refreshes it.
- AC-05: A slot that is no longer live is marked stale and blocks applying until recaptured;
  aged slots remain selectable rather than silently disappearing.
- AC-06: Existing live canvas behaviour, labels and controls are unchanged.
- AC-07: Browser coverage asserts the new view switching, replay banner/label and calibration
  persistence.

## Requirements and success

- FR-01: Canvas visibility is a view decision; content views stay canvas-free.
- FR-02: Replay time is owned by the scrubber after initialisation, never by live frames.
- FR-03: Calibration captures measured coordinates at selection time; staleness is explicit.
- SC-01: `devices.spec.ts` covers switching, replay banner/canvas and a 3.5 s ageing selection.
- SC-02: Full frontend checks pass; results recorded.

## Assumptions

- Q-01: Hiding the canvas for content views is preferable to leaving a live scene above them;
  Replay and Calibration keep the canvas because they act on positions.
