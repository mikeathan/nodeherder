# Spec: WebSocket broadcast resilience

ID: 069 | Status: Draft | Owner: mikeathan | Updated: 2026-11-25
Constitution: 1.0.0 | Components: backend

## Problem, scope, exclusions

The deployed service stalls after a few hours. The frontend still loads devices over HTTP,
but controls stop working, no new readings arrive, every `/ws` upgrade returns `101` with
zero bytes, and the backend stops emitting **any** log lines until the container is
restarted. It happened twice and a restart fixed it both times. The z2m frontend can
still control the same Zigbee devices, so MQTT, Zigbee and the LAN are healthy.

Evidence from `docker logs --tail 200 nodeherder-backend` (2026-11-24):

- Repeated `WebSocket: client id N Session error: ... i/o timeout` and
  `tried to write to closed a session`, plus `WsHandler: client connected` logged after
  those errors (the handler logs on *disconnect*, not connect).
- One nginx `/ws` connection transferred 323,328,709 bytes before the final
  `writev ... i/o timeout` at 21:22:35; every `/ws` after that is `101 0`.
- Backend logs stop completely at 21:22:35 while `/api/hubstate`, `/api/metrics/query`
  and `/api/integrations/instances` keep returning 200.
- Startup shows `ESPHome observation subscribed` / `received first batch`, i.e. the
  wireless observation source introduced in #110 (2026-11-18) is active.

Two defects combine:

1. **Re-entrant broadcast deadlock (backend).** `github.com/olahol/melody@v1.2.1`
   runs one `hub.run()` goroutine consuming four *unbuffered* channels. Its broadcast
   case holds `sessionSet.mu.RLock` and calls `session.writeMessage` for every client;
   a closed session invokes `errorHandler(ErrWriteClosed)` **synchronously on that same
   goroutine**. `backend/internal/ws/websocket.go:167-171` logs that error.
   `logrus.fireHooks()` runs *before* `entry.write()`, and `utils.RemoteHook.Fire`
   (`backend/utils/logger.go:179`) calls `emitter.Broadcast("logger", ...)` ->
   `eventHub.Broadcast` -> `webSocketImpl.Broadcast` -> `melody.BroadcastFilter` ->
   `m.hub.broadcast <- message`. The only receiver is the goroutine currently blocked
   inside `sessions.each`, so the hub wedges permanently. Every later `register`,
   `unregister`, `exit` and `broadcast` send then blocks forever; every log line blocks
   in `fireHooks` before being written, which explains the total log silence. HTTP
   handlers that neither broadcast nor log still answer 200, which explains the device
   list rendering. `specs/006-event-processing-hardening/plan.md:270` already required
   "avoid recursive remote logging/broadcast from the full queue handler; diagnostics
   must be local for this failure" — the buffer-full branch honours it, the general
   error branch does not.
2. **Continuous observation snapshot flood (trigger/amplifier).** #110 added
   `hub.go:172-176`, broadcasting a complete `observations` snapshot (up to
   `MaxSnapshotIdentities = 256` identities, each with a 20-sample signal window and
   classification) from a 250 ms ticker. `renderSignature`
   (`backend/internal/observations/service.go:870`) includes monotonically increasing
   `SourceDiagnostics.Advertisements/Decoded/Dropped` and per-identity `Packets`, plus
   per-packet RSSI, smoothed RSSI, trend delta and signal samples. Any BLE advertisement
   changes the key, so the "publish only on change" contract publishes a full snapshot
   on essentially every tick, to every client, forever. Each WS error then adds a log
   line which adds another `logger` frame, which can produce another WS error: a
   positive-feedback loop that fills a slow client's 256-message queue.

In scope: make WS fan-out independent of melody's hub goroutine; keep WS session
diagnostics local; bound outbound frames; give observation telemetry a bounded
publish cadence while structural changes stay prompt. Out of scope: frontend/UI,
nginx timeouts, recipe thresholds, the automation feedback-loop guard (spec 068),
persistence, and HTTP/WS/MQTT/MCP contract changes.

## Scenarios and acceptance

Independent verification: Go tests driving a real `httptest` websocket server with the
remote logger hook enabled, plus fake-clock observation tests. No household devices.

- AC-01: Given a slow client with a saturated outbound queue, when many broadcasts are
  made, then no producer blocks on the melody hub goroutine and healthy clients keep
  receiving frames.
- AC-02: Given remote logging enabled and a client whose session is closed but still
  registered, when broadcasts continue, then the hub never wedges: a new client can
  upgrade, register and receive a frame, and a log line completes promptly.
- AC-03: Given WS session errors/overflows, then diagnostics stay local (no `logger`
  frame is emitted from a melody callback) and failures are counted in
  `BroadcastDiagnostics`, readable through `WebSocket.Diagnostics()`.
- AC-04: Given an outbound frame larger than the named outbound bound, then `Broadcast`
  returns an error, publishes nothing, and the caller answers the pending request with
  `operationFailed`. The largest legitimate observations snapshot must fit with headroom.
- AC-05: Given continuous observation telemetry churn, then snapshots are published at
  most once per the telemetry cadence, not once per 250 ms tick.
- AC-06: Given a structural observation change (new/removed identity, triage decision,
  label, class, source descriptor, gate or classification progress), then a snapshot is
  published at the existing 250 ms cadence and carries exact (unquantised) values.
- AC-07: Given repeated physical events, unchanged state, deduplicated reports, manual
  runs and the spec 068 primary-property guard, then behaviour is unchanged.
- AC-08: Given the pre-fix sources, then the new regression tests fail; given the final
  diff, then all required checks pass.

## Requirements and success

- FR-01: `Broadcast` MUST NOT send on `melody`'s `hub.broadcast` channel or otherwise
  wait on its single hub goroutine; acceptance: AC-01, AC-02.
- FR-02: Melody session callbacks MUST NOT emit remotely delivered log lines; their
  diagnostics MUST be local and counted; acceptance: AC-02, AC-03.
- FR-03: Outbound frames MUST be bounded by a named outbound limit with documented
  headroom over the largest legitimate result (`loadMetrics` ranges are unbounded by the
  client request size), and a rejected frame MUST NOT leave a pending request unanswered;
  acceptance: AC-04.
- FR-04: Observation telemetry-only changes MUST be rate-limited to a named cadence,
  while structural changes keep the existing cadence and the payload keeps exact
  values; acceptance: AC-05, AC-06.
- NFR-01: MUST NOT add goroutines, timers, unbounded channels or new global mutable
  business state; memory stays bounded by melody's existing per-client queue.
- NFR-02: MUST NOT change HTTP/WS/MQTT/MCP payloads, config, storage or automation
  execution (spec 068 remains authoritative).
- SC-01: Before-fix regressions fail; `go test ./...`, `go build ./...`,
  `go vet ./...` and race checks on affected packages pass.

## Entities and external interfaces

`ws.WebSocket.Broadcast` stays the single fan-out owner with the same signature and
`EventMessage` wire shape. A hook-free local log path is additive to `utils`.
`observations.Service` gains an injectable telemetry cadence; the `observations` event
payload is unchanged. No new public HTTP/WS/MQTT/MCP contract, route or persisted
schema.

## Questions and assumptions

- Q-01: Remote logging state on the deployed hub is unconfirmed (the local Docker
  daemon is not running, so the deployed config could not be read). Resolution: both
  fixes MUST be safe with `enableRemoteLogger` true *or* false; the deadlock mechanism
  is removed unconditionally rather than depending on that setting.
- Q-02: `Seed` forwarding a full first snapshot (spec 068 limitation) is unchanged.
- Assumption: the last two merges (#115 CI YAML, #114 frontend + read-only
  `spatialhistory` samples) contain no backend WS/event change; the newest backend
  change on the broadcast path is #110 (2026-11-18). No maintainer ratification,
  exception or review approval is claimed.
