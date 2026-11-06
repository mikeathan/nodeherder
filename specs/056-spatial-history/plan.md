# Plan: Spatial history recorder and recordings UI

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Backend

- `internal/spatialhistory`: bbolt `spatial.db` (root bucket `spatial`, child bucket per device),
  JSON chunk values keyed by timestamp+session; repository with range read, prune and close.
- Recorder: bounded queue/worker; per-source session state; duplicate merge 350 mm; 1 Hz/300 mm
  gate; 50 mm quantisation; 5-minute/300-point chunk flush; 3,000 points/day cap; clear/offline
  close; 7-day purge task.
- Wiring: controller `OnSpatialTargets` feeds the recorder; recorder checks
  `DeviceConfigCache.IsMetricsEnabled(sourceID)`; shutdown closes it.
- APIs: authenticated sessions, session points, heatmap, storage usage; list/system routes in
  `internal/hub.go`.
- Derived presence history inheritance: when the source history flag is on, the derived device's
  metrics gate follows it (targeted change in the metrics gate).

## Frontend

- Device History: mode switch Charts / Zone activity / Positions; Positions Player and Heatmap
  using existing canvas/room/zones; sessions from the API; no browser buffer.
- Presence: Live/Draw/Setup grouping; remove Replay, Heatmap, Diagnostics, browser recorder and
  Diagnostics domain/tests.
- Settings → System: storage table (metrics, spatial, logs, total).
- Chart brush: shared pointer/touch range selection for state/presence charts with reset and
  announcements.
- Multi-sensor summary: global Home/Overview, rendered only with ≥2 spatial sensors.
- Remove dead modules/tests; add domain/API/UI tests; update e2e.
- Version 0.16.0 in package.json/package-lock.

## Verification (recorded)

- Backend: `go build ./...`, `go vet ./...`, `go test ./...` pass; `-race` over
  spatialhistory/api/store passes; recorder/store/API handler tests added.
- Frontend: `npm test` **741 passed**; `npm run build` pass; lint 0 errors (28 pre-existing
  warnings); `npm run type-check:e2e` pass.
- Browser: `devices.spec.ts` + `home.spec.ts` desktop and mobile **38 passed**, including the
  grouped Presence navigation and device History modes.
- Version: `frontend/package.json` and lock at **0.16.0**.

Not covered by automation: live Positions playback against a real hub, brush selection on a
touch device, and storage numbers with a non-empty spatial DB. These need review/manual checks.
