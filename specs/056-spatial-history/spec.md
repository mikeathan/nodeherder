# Spec: Spatial history recorder and recordings UI

ID: 056 | Status: In progress | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: backend + frontend

## Problem and scope

Presence history is charted from device metrics, but coordinates are only held in a browser tab
and lost on reload. Add a backend recorder gated by the existing per-device "Record history"
switch, plus a device History experience for playback and a persisted heatmap.

In scope:

- backend movement-gated session recording (≤1 Hz per visible target, ≥300 mm movement, 50 mm
  quantisation, 5-minute chunked values, no points while still, 3,000 points/day cap);
- sessions/points/heatmap authenticated APIs and a 7-day spatial purge independent of metrics;
- storage-size endpoint and Settings → System storage view;
- device History modes: Charts, Zone activity, Positions (Player + Heatmap);
- Presence navigation grouped as Live / Draw / Setup; remove Diagnostics, browser Replay/Heatmap
  and the browser recorder;
- drag/tap range focus on presence charts;
- multi-sensor presence summary only when ≥2 spatial sensors exist;
- minor version bump.

Exclusions: video, identity, cross-sensor path merging, raw-frame storage, public/MCP exposure,
configurable retention in v1 (constant 7 days).

## Acceptance

- AC-01: Recording follows the source device's `metricsEnabled` exactly; off ⇒ no points.
- AC-02: A stationary target produces no points; movement under 300 mm between ticks is skipped;
  at most one point per visible target per second; points are 50 mm quantised.
- AC-03: Duplicate slot reports within 350 mm become one point; slots are not identities.
- AC-04: Sessions group points and end on a cleared frame or source offline; chunks are flushed at
  least every five minutes; a daily cap marks the session capped.
- AC-05: Sessions/points/heatmap endpoints are authenticated and return only the requested device
  and range.
- AC-06: Spatial data is purged after 7 days, independently of the 10-day metrics purge.
- AC-07: Settings → System shows metrics.db, spatial store, logs and total data sizes.
- AC-08: Device History offers Charts, Zone activity and Positions; Positions has Player and
  Heatmap; Presence no longer shows Replay/Heatmap/Diagnostics.
- AC-09: Presence is grouped Live/Draw/Setup with contextual sub-views and remains usable at 375 px.
- AC-10: Presence/state charts support drag (mouse) and touch selection to focus a time range,
  with reset and keyboard/button access.
- AC-11: Multi-sensor summary appears only with at least two spatial-capable sensors and links to
  each device.
- AC-12: Version is bumped to 0.16.0; backend/frontend checks and affected browser specs pass.

## Questions

- Q-01: Storage is a separate `spatial.db` bbolt file with its own bucket and purge; same data
  directory, independent retention. Chosen for isolation and a different timeout.
- Q-02: Session close on cleared frame/offline; a still target produces no frames and keeps its
  session open until a clear/offline event. A 30-minute safety close bounds stale sessions.
