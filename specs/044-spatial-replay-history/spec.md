# Spec: Spatial replay and persisted zone history

ID: 044 | Status: Implemented (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

The heatmap is session-only and the presence tab shows the current instant, so a user cannot
answer "which zone was occupied over the last day" or "what did the last minute look like".
Raw radar positions are not persisted (privacy and storage), but the hub already records the
derived presence entities as device metrics.

In scope: (1) a persisted zone-activity view built from those existing metrics — occupancy
share, transitions, count peak and a timeline per zone; (2) a bounded, session-only replay of
the last five minutes of accepted frames on the single canvas, with seek/play and an explicit
"session only" label.

Exclusions: storing raw positions, server-side replay, cross-session replay, exporting data,
new routes or retention changes, and any automation input from replay.

## Scenarios and acceptance

### US-01 — Persisted zone activity (P1)

- AC-01: Given a derived presence device with zone occupancy entities, when the Activity view
  opens, then it requests only those existing derived entities for the selected range
  (1 h / 24 h / 7 d) through the existing history API.
- AC-02: Each zone row shows the share of known time occupied, the number of state transitions,
  the highest target count and the time of the last occupancy, computed from returned samples;
  never from assumed values.
- AC-03: The share denominator is known time only: before the first sample and after data ends
  the row is marked partial rather than counted as empty.
- AC-04: Without a derived device or zone entities the view says why and makes no request.
- AC-05: The view states that raw positions are not stored, so this is zone history, not a
  position replay.

### US-02 — Session replay (P1)

- AC-06: While frames arrive, the browser records accepted active positions in memory, capped at
  five minutes and 1,500 frames; older entries are dropped and the cap is enforced.
- AC-07: The Replay view provides a time scrubber, play/pause and 1×/2×/4× speed; it shows the
  frame timestamp and slot count, ends at the newest frame, and never loops.
- AC-08: In replay the single canvas draws recorded positions and suppresses live occupancy;
  a visible notice says the canvas is replay, not live. Leaving the view restores live tracks.
- AC-09: Replay never writes to the store, history or automations; a reload discards it.
- AC-10: Keyboard operable; fits 375/768/1440 px.

## Requirements and success

- FR-01: Activity maths MUST be pure and tested in `domain/spatial/activity.ts`.
- FR-02: Replay math and recording bounds MUST be pure/tested; no timers in domain code.
- FR-03: Only existing metrics/history APIs are used; no new endpoint or persistence.
- NFR-01: Memory bounded by frame and time caps; no unbounded arrays or timers left on unmount.
- SC-01: Domain tests cover share/unknown/transitions/peak and replay bounds/seek semantics.
- SC-02: Full frontend checks plus the devices browser spec pass; results recorded.

## Questions and assumptions

- Q-01: Derived zone entities are discoverable from the derived device's `zoneId` attribute;
  no name inference beyond the documented suffix used by the resolver.
- Q-02: Five minutes/1,500 frames is enough for tuning; configurable in code only.
