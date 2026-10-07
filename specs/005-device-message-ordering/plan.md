# Plan: Ordered device message processing

Spec: [spec.md](spec.md) | Status: Draft (not reviewed or approved) | Author: Claude for mikeathan | Updated: 2026-10-07
Constitutions: parent 1.0.0 (proposed); backend 1.0.0 (proposed)

## Context and design

### Current behaviour (observed in code, 2026-10-07)

```
paho (SetOrderMatters(false): one goroutine per message)      mqtt.go:155, :64
  → HubController.processMessage (reads/writes responseHandlers map, unsynchronised)   hub-controller.go:577
    → utils.WorkerPool(4), unbuffered channel, no recover       hub-controller.go:88, utils/worker.go
      → deviceHandler → DeviceProcessor.CreateOrUpdateDevice
        → DeviceLifetimeService.Update  (dedupe → set exposes → automation → store → broadcast)
HTTP /data collector → HubController.Enqueue → same processMessage                 routes.go:366
Manual trigger (HTTP) → HubController.TriggerManual → engine.HandleManual          routes.go:555
Delayed action → own goroutine + time.NewTimer, guarded by MqttBaseAction.mut      action.go:372
```

Two reorder points exist (paho goroutines, then the 4-worker pool); ordering at the MQTT layer
alone would not hold. The `responseHandlers` map is also written from concurrent goroutines
(latent data race, found while planning).

### Proposed design (smallest suitable)

1. **Ordered ingress.** `SetOrderMatters(true)`; the paho callback only enqueues onto a single FIFO
   `ingress` lane and returns. Handlers that block (e.g. `AddTopic` → subscribe + wait) then run off
   paho's goroutine, so ordered delivery cannot deadlock (AC-16).
2. **One executor, injected.** New package `backend/internal/lanes` (name open to review)
   containing a `Executor` with per-key lanes:
   - `Submit(key, task)` (non-blocking), `Do(ctx, key, fn) error` (wait for result),
     `Shutdown(ctx)`.
   - A lane is a FIFO slice plus a goroutine that exists only while the lane has work, then exits
     and removes its map entry, so idle devices cost nothing and a busy device costs one
     goroutine. Per-lane cap with reject-and-log on overflow; per-task `recover` with logging so a
     panic cannot kill the lane or the process; depth/wait metrics via the existing logger.
   - Created once in `RegisterHubController` (composition root), passed to the processor and the
     automation engine through small consumer interfaces (BE-01, Go rules).
3. **Routing in one place.** `HubController.Ingest(source, topic, payload)` is the single entry for
   MQTT and HTTP. It runs on the `ingress` lane (so `responseHandlers` is single-owner and needs no
   lock), resolves the handler, and submits the work to the lane keyed by topic
   (`bridge/*` share one `bridge` lane). `Enqueue` becomes a thin caller of `Ingest`.
4. **Run-to-completion per device.** The lane task is exactly today's
   `CreateOrUpdateDevice` (dedupe, apply, automation, store, broadcast); nothing inside changes.
   The invariant — one writer per device, automations read that device's state only inside its lane —
   is documented on `DeviceLifetimeService` and tested (AC-04).
5. **Per-run automation context.** Replace mutation of the shared context with a decorator:
   `runContext{AutomationContext, payload, manual}` built per `Evaluate`/`EvaluateTrigger`, layered
   over the persistent `DeviceContext` (which keeps only `currentData`). Removes `SetManualTrigger`
   and `SetDevicePayload` from the interface; origin becomes part of the event (single source of
   truth). Fixes F-05 independently of ordering.
6. **Manual triggers through the lane.** `HubController.TriggerManual` → `executor.Do(ctx, deviceKey,
   engine.HandleManual…)`, key from the registrar (automation id → friendly name). A lane that cannot
   accept work before the request deadline returns `ErrLaneBusy`, mapped to 503 (today every error is
   412; existing 412 cases unchanged).
7. **Left unchanged on purpose (to avoid regressions):** `Update` semantics (BE-02), dedupe, debounce,
   step/rotate/TOGGLE maths, `MqttBaseAction` mutex protocol, delayed-action goroutine, scheduler
   (`SetEnabled` only), device creation queue from #51, WebSocket/HTTP/MQTT contracts, automation JSON.
8. **Determinism for tests only.** `executeBaseWithDelay` uses `time.NewTimer`; inject the existing
   `utils.Clock` (`AfterFunc`) behaviour-preservingly so delay/cancel tests use a fake clock (BE-03).
9. **Zigbee2MQTT stale snapshots.** No special handling: with order preserved the stale message equals
   the stored value and is skipped by the existing dedupe (`device_lifetime.go:134`), so it triggers
   no automation. Contingency if Q-01 shows `action` is re-published: filter event exposes on
   snapshot-only messages (separate spec amendment).

### Alternatives considered (full comparison in [ADR-005-01](adr-keyed-lanes.md))

| Option | Why not chosen |
| --- | --- |
| Order only in the MQTT layer, keep the 4-worker pool | Pool reorders again; proven by reading `AddTask` |
| Single global worker | Simplest, but one slow handler (disk store) delays every device incl. dial |
| Hash-striped fixed workers | Bounded goroutines, but unrelated devices can block each other |
| Last-write-wins sequence numbers in `Update` | Fixes state only; automations still see mixed values (F-04) |
| Fix Zigbee2MQTT config/version | Out of control of the repo; newest version is required (user decision) |

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 005 with reproduction (F-01–F-03), AC/FR traceability; Q-01–Q-04 listed, none blocks design | PASS | Evidence recorded (spec/plan/tasks updated with results); review pending |
| NH-02 | Events still bypass equality dedupe (AC-06); state vs command unchanged; no optimistic state shown as confirmed | PASS | AC-06 golden (identical events both processed) PASS; dedupe untouched; review pending |
| NH-03 | No new entry points or exposure; HTTP data collector and manual route keep current auth; new 503 body has no data | PASS | No new routes; 503 body is the error text only; review pending |
| NH-04 | New package behind small interfaces; contracts unchanged except internal `AutomationContext`; ADR-005-01 | PASS | `AutomationContext` and internal `AutomationTrigger` (ctx added) changed; HTTP/WS/MQTT/JSON unchanged; review pending |
| NH-05 | Traceability table below; replay harness fails on current code first; real-device checks only on user-approved targets (Q-04) | PASS (design) | Automated evidence recorded (failing-first for T004–T006); SC-02/T015 live checks unrun |
| NH-06 | Per-lane cap, reject+log, recover, `Shutdown(ctx)`; goroutines only while busy; NFR-02/04 | PASS (design) | Cap/reject/count/recover/shutdown tested; review pending |
| BE-01 | Transport adapts and delegates; executor and processor testable with fakes, no broker | PASS | Harness uses fake paho + memory store, no broker; review pending |
| BE-02 | `Update` untouched; regression evidence AC-04–AC-07 + existing identical-event, unchanged-state and debounced-read tests must still pass | PASS (design) | `Update` logic unchanged (doc comment only); services tests + golden PASS; review pending |
| BE-03 | Ownership/cleanup defined for lanes; fake clock; `go test -race` on affected packages; no sleeps | PASS (design) | `-race` PASS; fake clock for delays; a few bounded waits remain (100 ms absence checks, goroutine settle, shutdown poll); review pending |
| BE-04 | N/A: no storage/query/schema change | N/A | N/A |
| BE-05 | Manual-trigger route: auth, rate limit and 412 unchanged; new 503 documented | PASS | 503 mapped and tested; 412 cases unchanged; review pending |

## Contracts, data, rollout

- MQTT/WebSocket/HTTP/automation JSON: unchanged. One additive HTTP outcome: `POST
  /api/automation/trigger` (route auth and rate limit unchanged) may return 503
  `{"error": "device is busy, trigger not started: …"}` when the device lane is full, shut down, or
  does not start the trigger within 5 s; the trigger did not run and may be retried. All existing
  errors stay 412 with the same text. Consumer: `frontend/src/services/automation-trigger.service.ts`
  treats any non-OK status the same way (reads `error` from the JSON body), so no frontend change.
- Internal API: `AutomationContext` interface shrinks; update conditions, actions, mocks, tests.
- No persistence or configuration change. `utils.WorkerPool` has no other user than
  `hub-controller.go:88`; remove it (and its tests) only if nothing else adopts it, else leave and note.
- Rollout: backend image only; deploy to the live stack after the manual checks; no migration.
- Rollback: redeploy the previous backend image; nothing to repair. No runtime flag (rejected: keeps
  two dispatch paths alive forever).
- Rename edge: lanes key on the topic, so ordering between old and new topic names is not guaranteed
  during a rename; Zigbee2MQTT republishes afterwards. Documented, not covered by a test.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| AC-01, AC-02, FR-01 | `internal/controllers`, `internal/mqtt`, `internal/lanes`, `internal/services` | Replay harness with fake Zigbee2MQTT (stale+new pair), 2,000 randomised trials, both directions; run first on current code to show failure | Before: FAIL 1804/2000 wrong, identical pair 436/500 wrong. After: PASS 2000/2000, 0 extra broadcasts; 500/500 identical pairs; `-race -count=10` PASS |
| AC-03, FR-02, NFR-01, NFR-02 | `internal/lanes` | Slow-device isolation test; 1,000-message/20-device burst benchmark vs baseline; goroutine count before/after | `TestSlowKeyDoesNotBlockOtherKeys` PASS. p95 1.95–2.35 ms vs 4.04–4.56 ms baseline; allocs/op 47.3k vs 46.1–46.7k (+1.3–2.6%, ≤ +10%); hub idle goroutines 20 vs 25 (lower: pool removed); growth 0 (table below) |
| AC-04–AC-07, FR-05 | `internal/automations`, `internal/services` | Golden trace replay of dial/button traces captured read-only from the live feed; baseline first, then new code | Synthetic traces from the real JSON (no live capture: live access out of scope for this run). `TestDialGoldenSequence` PASS before and after (identical); burst FAIL before (44–49 commands for 40 events), PASS after |
| AC-08–AC-10, FR-03, FR-04 | `internal/automations`, `internal/controllers`, `internal/api` | Concurrent manual + device-message test on the unconditional trigger; error-mapping table test | T006 FAIL before, PASS after (`-race -count=50`); `TestTriggerManualKeepsEngineErrors`, `TestTriggerManualRunsOnDeviceLane` (busy → `ErrLaneBusy`, never runs), `TestAutomationTriggerHandler_ErrorMapping` (412 kept, busy → 503) PASS |
| AC-11, AC-12 | `internal/automations` | Fake-clock delay/cancel/cooldown and illuminance-condition tests using the real presence JSON | PASS (characterisation after T012 clock injection; `-race -count=20`) |
| AC-13 | `internal/automations` | Existing scheduler tests plus window test with fake clock | Existing scheduler/engine tests PASS (scheduler untouched); no new window test written |
| AC-14, FR-07, NFR-03, NFR-04 | `internal/lanes`, `internal/controllers` | Shutdown/leak test, overflow test, panic-isolation test, `-race` | Lane drain/deadline/idle-goroutine/overflow/panic tests and `TestHubStopsIngestingWhenContextEnds` PASS; `-race` PASS; burst produced no depth warning (ingress < 1,000, device lanes < 100) |
| AC-15, AC-16, FR-06 | `internal/controllers`, `internal/api`, `internal/mqtt` | HTTP and MQTT ordering parity test; blocking-subscribe handler cannot stall delivery | `TestConcurrentFirstMessagesFromMQTTAndHTTP` (20 new devices, MQTT+HTTP concurrently) `-race -count=10` PASS — before 005 the same situation crashed the process (T003 finding); `TestHTTPIngestionIsOrderedPerDevice` PASS (1,000 in order); `TestBridgeSubscribeDoesNotStallOrderedDelivery` PASS, and FAILS (subscribe timeout) when ingestion is made synchronous on the callback (negative control, reverted) |
| SC-02 | Live system | Manual: ten UI toggles each way on the approved light; read-only feed capture alongside | unrun (needs Q-04 approval) |

Checks to record with exact output (from `backend/`): `go test ./...`, `go build ./...`,
`go vet ./...`, `go test -race ./internal/...` for affected packages, benchmark comparison.
Mocked MQTT and temporary storage only; the live checks above are separately scoped manual
procedures.

### Recorded checks

Environment: `go version go1.26.1 darwin/arm64` (local toolchain; `go.mod` declares
`toolchain go1.24.8`, `GOTOOLCHAIN=auto`). Baseline tree: `f0e160af` (untouched).

T002 baseline (2026-10-07, untouched tree, from `backend/`):

- `go test -count=1 ./...` → exit 0; all 27 packages `ok` (`internal/mqtt` has no tests).
- `go vet ./...` → exit 0, no findings.
- `go test -race -count=1 ./internal/... ./utils/...` → exit 0; all packages `ok`, no race reports
  (the `responseHandlers` map race is latent: no existing test drives concurrent first messages).

Test seam added before any behaviour change: `mqtt.WithClientFactory` (default `mqttlib.NewClient`)
plus `mocks.FakePahoClient`, which reproduces paho v1.5.0 dispatch as read in its `router.go`
(`OrderMatters=false`: one goroutine per message; `true`: handlers inline on one router goroutine,
which also completes SUBACKs) and `client.go` (`go OnConnect`).

T003 baseline benchmark (untouched behaviour; `BenchmarkMixedBurst`, `internal/controllers`,
`-benchtime 20x -count 6`, Apple M4): one op = 1,000 messages over 20 devices through
fake paho → `MqttService` callback → hub → processor → `Update` → broadcast. Every message carries a
unique `action` event so each yields one broadcast; `state`/`linkquality` vary.

| run | ns/op (burst) | p50-ns/msg | p95-ns/msg | allocs/op | B/op | hub idle goroutines | growth |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | 3,533,740 | 1,742,041 | 4,035,459 | 46,723 | 2,178,535 | 25 | 0 |
| 2 | 3,477,581 | 1,936,875 | 4,561,125 | 46,213 | 2,113,874 | 25 | 0 |
| 3 | 3,389,267 | 1,752,334 | 4,219,500 | 46,369 | 2,130,416 | 25 | 0 |
| 4 | 3,387,169 | 1,893,083 | 4,196,791 | 46,249 | 2,116,288 | 25 | 0 |
| 5 | 3,195,823 | 1,804,875 | 4,060,625 | 46,102 | 2,100,867 | 25 | 0 |
| 6 | 3,420,215 | 1,931,708 | 4,230,208 | 46,154 | 2,105,980 | 25 | 0 |

NFR-01 target derived from this baseline: p95 ≤ 4,561,125 × 1.1 ≈ 5.02 ms and allocs/op ≤
46,723 × 1.1 ≈ 51,395 (worst baseline run + 10%). NFR-02: 25 idle goroutines per hub
(20 availability monitors + 4 pool workers + 1 pool supervisor); growth 0.
Finding: with concurrent *first* messages for 20 devices the current hub crashed the test
process with `fatal error: concurrent map read and map write` at `hub-controller.go:579`
(`responseHandlers`); the harness therefore seeds devices one at a time.

T004 on current code (`go test -count=1 -run TestReplay ./internal/controllers/`), FAIL as expected:

```
--- FAIL: TestReplayStaleThenNewPairEndsOnNewState (0.03s)
    ordering_test.go:109: seed=1791389074221569000
    ordering_test.go:152: trials=2000 wrongFinal=1804 extraStateBroadcasts=304 timeouts=0
--- FAIL: TestReplayIdenticalPairBroadcastsOnce (0.00s)
    ordering_test.go:193: trials=500 wrong=436 timeouts=0
```

A trial ends when a probe message (`power_on_behavior`, delivered after the pair on the same topic)
is broadcast; on current code the probe itself is also reordered, so "wrong" counts both a stale
final state and work still unhandled when a later message on the same device was already broadcast.

T005 on current code (`internal/automations/scenarios/dial_test.go`, real `0x001788010d7d9d3f.json`,
light confirmations played by the test, Q-02 maths unchanged):
`TestDialGoldenSequence` (slow/fast, left/right, clamp at 0 and 254, skip when unchanged, two
identical events, TOGGLE ×2, `button_1_press` ignored, presets 250→153→370→454→500→250 wrap)
PASS (`-count=5`), recorded as the baseline. `TestDialBurstEachEventUsesItsOwnValues` (40
alternating rotations, no confirmations in between) FAIL as expected, `-count=3`:

```
dial_golden_test.go:246: burst: 49 published for 40 events, 39 mismatched positions
dial_golden_test.go:246: burst: 44 published for 40 events, 39 mismatched positions
dial_golden_test.go:246: burst: 48 published for 40 events, 39 mismatched positions
```

More commands than events: conditions read the dial's live `action`, so one evaluation also fires
triggers for another event's action (F-04).

T006 on current code (`internal/automations/device_test.go`, interleaving forced by a trigger
that parks until released, no sleeps), FAIL as expected:

```
--- FAIL: TestManualTriggerRunsDespiteConcurrentDeviceMessage
    run_context_test.go:107: manual run executed the unconditional trigger 0 times, want 1
--- FAIL: TestDeviceMessageNeverRunsUnconditionalTrigger
    run_context_test.go:133: manual trigger ran the action 2 times, want 1
```

T014 after change (from `backend/`, 2026-10-07):

- `go build ./...` exit 0; `go vet ./...` exit 0; `go test -count=1 ./...` exit 0 (28 packages ok, incl. new `internal/lanes`).
- `go test -race -count=1 ./internal/lanes/ ./internal/mqtt/ ./internal/controllers/ ./internal/services/ ./internal/automations/ ./internal/api/ ./utils/ ./mocks/` exit 0, no races.
- Repeated: controllers ordering/dial/ingest/manual tests `-race -count=10`; `internal/lanes` `-race -count=20`;
  T006/T007 `-race -count=20` — all PASS.
- `internal/mqtt` still has no unit tests; its change is covered through `MqttService` in the controllers harness.

Benchmark after change (same command as T003):

| run | ns/op (burst) | p50-ns/msg | p95-ns/msg | allocs/op | B/op | hub idle goroutines | growth |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | 2,280,938 | 948,125 | 2,185,708 | 47,338 | 2,321,648 | 21 | 0 |
| 2 | 2,092,558 | 1,013,167 | 2,195,541 | 47,301 | 2,314,672 | 21 | 0 |
| 3 | 2,122,169 | 816,125 | 2,350,375 | 47,279 | 2,321,439 | 21 | 0 |
| 4 | 1,993,610 | 783,000 | 2,035,459 | 47,288 | 2,320,844 | 21 | 0 |
| 5 | 1,867,177 | 678,959 | 1,948,459 | 47,285 | 2,311,585 | 21 | 0 |
| 6 | 1,981,960 | 750,625 | 2,225,583 | 47,263 | 2,312,409 | 21 | 0 |

NFR-01 met (p95 ≤ 5.02 ms, allocs ≤ 51,395). B/op rose ~+9% (not an NFR). The 21 idle goroutines
include the fake paho router goroutine that exists only in ordered mode (real paho always runs its
own router); hub-owned idle goroutines fell from 25 to 20 (availability monitors only, no lanes), so
NFR-02 is met as "no more than baseline", not "equal".

### Implementation notes and deviations

- Executor is held by `HubController` only (routing and manual triggers live there); it is not
  passed to `DeviceProcessor` or the engine, which are unchanged. Two executor instances: ingress
  (single key, cap 10,000 — the plan named no cap for ingress) and device lanes (cap 1,000, Q-03).
- `AutomationTrigger.TriggerManual` gained a `context.Context` (models, mock, route, hub) to carry
  the request deadline; the route bounds it at 5 s. `ErrLaneBusy` lives in `models/automations`.
  Lane full, executor shut down, or not started before the deadline → `ErrLaneBusy` → 503.
- Origin travels with the event: `NewManualEvent` (manual) vs `NewDeviceEvent`; `EvaluateTrigger`
  takes the origin from the event, so tests that called it with `NewDeviceEvent` now use
  `NewManualEvent`. `DeviceContext` keeps only `currentData`; `NewRunContext` builds the per-run view.
- Delayed actions: `MqttBaseAction.SetClock` (nil = real clock); the timer goroutine/select and
  mutex protocol are unchanged, only `time.NewTimer` became `clock.AfterFunc` signalling a channel.
  Cloned automations (reload/edit) revert to the real clock.
- An unknown `bridge*` topic (reachable only via HTTP data collector) previously queued a nil
  handler (would panic a worker); it is now logged and ignored.
- `utils.WorkerPool`/`WorkerTask` and their tests removed (no other users); `mqttResponseTask` removed.
- Lane unit tests were written right after the executor, not seen failing against a stub first.
- Lane shutdown is wired with `context.AfterFunc(hub ctx)` (5 s bound, ingress then devices);
  the MQTT client is still not disconnected on shutdown (unchanged).

## Debt, risks, exceptions

- Debt touched and fixed: unsynchronised `responseHandlers` map; no panic recovery in workers;
  shared mutable context; `time.NewTimer` in delayed actions.
- Debt left alone: `MqttService.topics` is mutated by `AddTopic`/`RemoveTopic` (bridge lane) and
  read by `subscribeTopics` on reconnect without a lock (pre-existing; not reported by `-race` here
  because tests wait for initial subscriptions). #51 device-creation queue (redundant under lanes, harmless); `ctx.currentData`
  optimistic state (inert for commands, F-06); step actions use last *confirmed* target state.
- Risk R-01: automation runs inside the lane, so a slow synchronous step delays that device's
  later messages. Mitigation: only non-blocking publishes today; wait/depth logging; NFR-01.
- Risk R-02: cap overflow rejects a message. Mitigation: cap sized from the burst test; loud log.
- Risk R-03: Q-01 unresolved could re-trigger the dial from cached `action`. Mitigation: capture
  a dial trace before implementation (first preparation task in `tasks.md`, not yet written),
  contingency in design item 9.
- Exceptions: none requested.

## Completion

To be filled with acceptance evidence, review links and limitations. Update
`docs/sdd/repository-map.md` and `backend/docs/architecture.md` (flow guide currently omits
ordering) when implemented.
