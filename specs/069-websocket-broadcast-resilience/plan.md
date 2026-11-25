# Plan: WebSocket broadcast resilience

Spec: [spec.md](spec.md) | Status: Implemented (self-reviewed) | Author: mikeathan | Updated: 2026-11-25
Constitutions: parent 1.0.0 (proposed); backend 1.0.0 (proposed)

## Context and design

Spec 068 on this branch is complete history; this is new work on the same branch, so it
gets its own spec set (`docs/sdd/README.md` branch-scope rule).

Ownership stays in `backend/internal/ws`, `backend/utils`, `backend/internal/observations`
and `backend/internal/api`. No frontend source, recipe, storage or wire contract changes.
`frontend/components.d.ts` drift produced by `npm run build` was reverted.

### Defect 1 — re-entrant broadcast deadlock (root cause)

Observed facts (verified against the pinned dependency):

- `melody@v1.2.1/hub.go:76-79` creates four **unbuffered** channels; `hub.run()`
  (`:83-114`) is their only consumer.
- `melody.go:203-212` `BroadcastFilter` sends `m.hub.broadcast <- message`.
- `hub.go:93-100` iterates `sessions.each` under `sessionSet.mu.RLock` and calls
  `session.writeMessage`; `session.go:24-27` invokes `errorHandler` synchronously on
  **that same hub goroutine** for a closed session.
- `backend/utils/logger.go` `RemoteHook.Fire` calls `emitter.Broadcast`; logrus
  `entry.go:242-252` fires hooks *before* `entry.write()`.
- So: WS error -> `utils.LogErrorf` -> remote hook -> `eventHub.Broadcast` ->
  `webSocketImpl.Broadcast` -> `BroadcastFilter` -> send on a channel whose only receiver
  is the goroutine currently inside `errorHandler`. Permanent self-deadlock. All later
  `register`/`unregister`/`exit`/`broadcast` sends block; new `/ws` upgrades answer 101
  before registration and then hang (`101 0`); all logging blocks in `fireHooks`; HTTP
  reads that neither broadcast nor log still return 200.

**Design (smallest suitable, at the owning WS layer):**

1. `webSocketImpl.Broadcast` stops using `BroadcastFilter`. It marshals once, enforces
   `maxBroadcastBytes` (8 MiB, deliberately larger than the 1 MiB inbound
   `maxMessageSize`, because `loadMetrics` results are server-generated and unbounded by
   the client request), takes a session snapshot with `melody.Sessions()`, filters with
   `auth.SessionValid` **before** taking `fanoutMu`, and calls `session.Write` per client.
   `Session.Write` is a non-blocking enqueue onto melody's existing bounded 256-message
   per-client queue. A closed session returns `ErrSessionClosed`; a full queue invokes
   melody's synchronous `errorHandler` on the **calling** goroutine (`session.go`
   `writeMessage`), which is why that handler must be hook-free and must never call
   `Broadcast`. No producer ever waits on melody's hub goroutine. `fanoutMu` preserves the
   global ordering that `hub.run` used to provide; its invariant is that nothing reachable
   from the locked region may emit a hook-bearing log or call `Broadcast` (the mutex is
   not reentrant). `WebSocket.Diagnostics()` exposes counted failures. A rejected result
   frame is answered with `operationFailed` by `eventHubImpl.execute`, so a pending
   browser request cannot hang (`loadMetrics` uses `ReportError=false`).
2. All logs emitted from melody callbacks and the WS request path become **hook-free local
   logs**. `utils.LogLocal{Debug,Debugf,Warnf,Errorf}` is backed by a second logger with
   the same output but no hooks, sharing a mutex-protected writer so lines cannot
   interleave; `SetLevel` keeps both levels in sync. This satisfies the unimplemented
   `specs/006-event-processing-hardening/plan.md:270` requirement and removes the
   error -> log -> frame -> error amplification loop.
3. WS session failures are counted (`BroadcastDiagnostics`) with a bounded hook-free
   diagnostic every `sessionErrorLogEvery` events. On `ErrMessageBufferFull` and
   `ErrWriteClosed` the raw connection is closed so readPump returns and melody
   unregisters the session.
4. `WsHandler.ServeHTTP` logged "client connected" *after* `HandleRequest` returned, i.e.
   on disconnect, through the remote hook. It now emits an accurate hook-free local debug
   line; its upgrade-error line is likewise local. `handleHubEvents`, `execute`, the
   observation publish error path and the observation drop callback are local too.
   Domain errors on the controller/MQTT path (`EmitBridgeConfig`,
   `reportDeviceOutputError`) stay remote-delivered, so the console viewer keeps them.

No new goroutine, timer, business channel or global mutable business state (NFR-01).

Alternatives considered and rejected:

- **Async `RemoteHook.Fire` with a bounded queue + draining goroutine.** Fixes all logging
  paths generically but changes synchronous delivery asserted by `utils/logger_test.go`
  and controller tests, adds a lifecycle to manage, and hides the fact that the WS layer
  is the component that must not re-enter itself.
- **Wrapper queue in front of `BroadcastFilter`.** Keeps the single hub goroutine as a
  mandatory hop, so a wedged hub still stalls the drain goroutine and every producer
  behind the queue; more moving parts for a weaker guarantee.
- **Forking/replacing melody.** Large scope, new dependency, unnecessary.

### Defect 2 — continuous observation snapshot flood (trigger/amplifier)

Observed facts: `service.go` `renderSignature` includes
`SourceDiagnostics.Advertisements/Decoded/Dropped`, `identity.Packets`, raw/smoothed RSSI,
`TrendDeltaDB` and every `Signal` sample. `hub.go` broadcasts a complete snapshot (up to
`MaxSnapshotIdentities = 256`) from a 250 ms ticker. Any BLE batch changes the key, so
"publish only on change" becomes "publish every tick" whenever a source is subscribed.

**Design:** keep `renderSignature` as the full key (byte-identical to pre-fix) and derive
`structuralSignature` from the same `snapshotSignature(snapshot, telemetry bool)` encoder
with telemetry fields omitted. A structural change (identity set, decision, label, class,
source ID, gate, classification progress) publishes at the existing `publishGap`; a
telemetry-only change publishes at the new `telemetryGap` (`defaultTelemetryGap = 1s`,
injectable via `Deps.TelemetryGap`, clamped to at least `publishGap`). Payload values stay
exact. `lastStructural` is updated only on publish, so a structural change that misses its
gap stays structural on the next tick instead of being throttled. Source names/media/facets
and identity kind/vendor/stable were absent from the original key and remain payload-only;
a NUL separator keeps source IDs from aliasing identity keys when telemetry is off.
`steady-state` snapshot rate drops from ~4/s to ~1/s while structural changes stay fast.

Alternatives considered and rejected: dropping diagnostics from the key entirely (breaks
spec 064 D-4 `TestDiagnosticsAloneTriggerAPublish` and hides a real fault state);
quantising counters in the signature (still republishes on RSSI churn); sending deltas
instead of snapshots (contract change, breaks the "one missed snapshot reconverges"
guarantee).

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | spec 069 scenarios/ACs precede edits; no approval claimed | PASS (design) | PASS (spec/plan/tasks) |
| NH-02 | repeated events, unchanged state, dedup and automation inputs untouched; spec 068 guard preserved | PASS (design) | PASS (53-package suite incl. automations/services/lanes) |
| NH-03 | no new entry point, secret or data exposure; hook-free logs stay local | PASS (design) | PASS (diff reviewed, no contract change) |
| NH-04 | established `ws`/`utils`/`observations` boundaries; additive local-log API | PASS (design) | PASS (architecture.md updated) |
| NH-05 | regression + acceptance tests named below; full component checks in Completion | PASS (design) | PASS (checks recorded below) |
| NH-06 | bounded per-client queue, counted failures, no new resources; rejected frame answered | PASS (design) | PASS (Diagnostics + operationFailed tests) |
| BE-01 | domain tests use fakes; no live broker/device | PASS (design) | PASS |
| BE-02 | device update semantics unchanged; existing lifetime/lane suites pass | PASS (design) | PASS |
| BE-03 | no new goroutine/timer/shared business state; race checks on affected packages | PASS (design) | PASS (`go test -race`) |
| BE-04/05 | no storage/query/route/assistant change | N/A (no change) | N/A |

## Contracts, data, rollout

`EventMessage` wire shape, all event names, HTTP/WS/MQTT/MCP routes and the
`observations` snapshot payload are unchanged. `WebSocket.Diagnostics()` and
`utils.LogLocal*` are additive. `operationFailed` is emitted for a rejected result frame;
it uses an existing event name, not a new one. No migration, config change or deployment
ordering. Rollback is a code revert; no data to restore. Deployed `enableRemoteLogger`
state is unconfirmed (Q-01); both fixes are unconditional, so the outcome does not depend
on it.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, AC-02 | `internal/ws/websocket_test.go` | `TestRemoteLoggingDoesNotWedgeBroadcastOnClosedSession` (continuous broadcasts while a session closes), `TestSlowClientDoesNotBlockFanoutAndErrorsStayLocal` | PASS |
| FR-02, AC-03 | `internal/ws/websocket_test.go`, `utils/logger_test.go` | `TestSlowClientDoesNotBlockFanoutAndErrorsStayLocal` (no remote session-error frames), `TestLogLocalDoesNotFireRemoteHook` | PASS |
| FR-03, AC-04 | `internal/ws/websocket_test.go`, `internal/ws/eventhub_internal_test.go` | `TestBroadcastRejectsOversizedFrame`, `TestLargestObservationSnapshotFitsBroadcastBound`, `TestExecuteReportsRejectedResult` | PASS |
| FR-04, AC-05, AC-06 | `internal/observations/service_test.go` | `TestTelemetryChurnPublishesAtTelemetryCadence`, `TestStructuralChangePublishesAtPublishCadence`, `TestStructuralChangeRetriesAfterMissedPublishGap`, `TestTelemetryGapIsConfigurable`, `TestTelemetryGapNeverBelowPublishGap`, `TestStructuralSignatureSeparatesSourcesFromIdentities`, existing spec 064 diagnostics/classification tests | PASS |
| AC-07 | existing suites | `go test -count=1 ./...` (53 packages) incl. automations, services, controllers, api, ws, observations | PASS |
| AC-08 | scratch copies at `a5bd7bb9` + pre-fix-API-only reproductions | deadlock repro FAIL pre-fix ("broadcast wedged", 3.05 s) / PASS post-fix; flood repro 8 snapshots in 8 ticks pre-fix / 2 post-fix | PASS |
| SC-01 | `backend/` | `go build`, `go vet`, `gofmt`, `go test -count=1 ./...`, `go test -race` on affected packages, `-cpu=1` WS tests | PASS |

### Negative controls (AC-08)

- A scratch copy with **pre-fix production sources** and pre-fix-API-only repro tests:
  `TestPrefixReproDeadlock` FAILS ("REPRODUCED: broadcast wedged on the melody hub
  goroutine", 3.05 s); `TestPrefixReproFlood` reports 8 full snapshots in 8 ticks.
  The same tests on fixed sources: deadlock repro PASSES; flood repro FAILS its pre-fix
  expectation (2 snapshots in 8 ticks).
- A **hybrid** copy with only the pre-fix `BroadcastFilter` fan-out and pre-fix remote
  `HandleError` logging, keeping the new APIs so the committed tests compile: BOTH
  `TestRemoteLoggingDoesNotWedgeBroadcastOnClosedSession` and
  `TestSlowClientDoesNotBlockFanoutAndErrorsStayLocal` FAIL. This confirms the committed
  tests genuinely guard AC-01/AC-02/AC-03.
- `renderSignature` was compared byte-for-byte against `a5bd7bb9` on empty and fully
  populated snapshots; equal. `structuralSignature` distinguishes source/identity
  aliasing and gate changes.

## Debt, risks, exceptions

- With remote logging enabled, operators no longer see WS connect/disconnect, session
  error, upgrade-error, observation publish/drop and WS-handler error lines in the
  browser console; they remain in the backend log. This is deliberate: those callbacks can
  run on the hub goroutine. Controller-path domain errors stay remote-delivered.
- `hub.go` observation publish failure is now a local warning rather than a remote error.
- `melody` still uses unbuffered `register`/`unregister`/`exit` channels; a hub wedged by
  a future callback would still stall upgrades. Callbacks now contain no remote side
  effects, and the incident's only re-entrancy path is removed. `fanoutMu` is not
  reentrant; its invariant is documented at the field.
- The outbound bound (`8 * maxMessageSize`) is a ceiling for runaway payloads, not a
  metrics downsample. `loadMetrics` results remain unbounded in principle; a rejected
  result now answers `operationFailed` instead of hanging the UI. A test ties the bound
  to `MaxSnapshotIdentities` so a future identity-limit change cannot silently drop all
  observation updates.
- The steady-state observation snapshot rate drops from ~4/s to ~1/s. Structural changes
  stay fast.
- `TestTelemetryGapIsConfigurable` and `TestExecuteReportsRejectedResult` pass against a
  hybrid that keeps pre-fix change detection, so they add behavioural coverage but not a
  pre-fix failure. The flood and deadlock negative controls above are the AC-08 evidence.
- nginx `/ws` keeps the default `proxy_read_timeout`; the incident's `101 0` lines are the
  wedged hub, not a proxy timeout, so no proxy change is included.
- No exceptions requested.

## Completion

Checks run 2026-11-25 (macOS). Passed: `gofmt -l` clean on every touched Go file;
`go build ./...`; `go vet ./...`; `go test -count=1 ./...` (53 packages, 0 failures);
`go test -race -count=1 -timeout=10m` on ws, utils, observations, controllers, api,
automations, services (0 races); `go test -cpu=1` on the new WS tests; `npm run lint`
(0 errors, 113 pre-existing warnings), `npm test -- --ci --runInBand` (162 suites /
1097 tests), `npm run type-check:e2e`, `npm run build`; `actionlint`, `shellcheck
scripts/manage-stack.sh`; `git diff --check`. Unrun: browser e2e (no frontend source or
contract change; CI runs them on merge). No household devices were touched. An
independent reviewer pass found no remaining deadlock, lock-order inversion or blocking
path; its findings (test negative controls, `-cpu=1` flake, case-sensitive log assertion,
outbound bound, diagnostics API, invariant comments, uncovered retry branch, stale
traceability, generated-file drift) were all fixed and re-verified. No human approval,
ratification or household verification is claimed.

Second review pass (2026-11-25, independent agent; not human approval):

- Re-ran `gofmt -l .` (only pre-existing files listed; none touched here), `go build ./...`,
  `go vet ./...`, `go test -count=1 -timeout=10m ./...` (53 packages, 0 failures),
  `go test -race -count=1` on ws, utils and observations (0 races),
  `go test -count=5 -cpu=1` on ws and observations (no flakes), and
  `go test -run '^$' ./internal` (CI fixture compile check).
- Re-ran the negative controls from a scratch copy: pre-fix `BroadcastFilter` fan-out plus
  pre-fix remote `HandleError` logging makes
  `TestRemoteLoggingDoesNotWedgeBroadcastOnClosedSession` fail with "broadcast wedged on
  the melody hub goroutine"; pre-fix change detection makes
  `TestTelemetryChurnPublishesAtTelemetryCadence` publish 9 snapshots in one gap and
  `TestStructuralChangePublishesAtPublishCadence` publish 3, confirming both tests guard
  the fix rather than restating it.
- Reviewed the locking and callback paths against the pinned melody source. Correction:
  `Session.Write` invokes `errorHandler` synchronously for a full queue (design item 1 now
  says so). The committed `HandleError` is hook-free, so that callback cannot re-enter
  `Broadcast` while `fanoutMu` is held; `logrus` fires hooks before writing, so no logger
  lock is held across a broadcast. No remaining hub-goroutine re-entrancy path found.
- Corrected the architecture note's outbound-bound wording: a frame above
  `maxBroadcastBytes` is rejected (and a result request is answered with `operationFailed`),
  not "not silently dropped". No code change was needed.
