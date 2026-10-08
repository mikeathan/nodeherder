# Plan: Event processing hardening

Spec: [spec.md](spec.md) | Tasks: [tasks.md](tasks.md) | Status: Draft for review
Author: Codex | Updated: 2026-10-08 | Reviewed code: `26f48d08`
Constitutions: parent/backend/frontend 1.0.0, proposed; none claimed ratified.

Progress: T002 test improvements and scoped T006/T007 production validation/error
reporting are authorized, implemented and verified. [Execution evidence](evidence.md) records actual checks and limits.
T005 counters, T008 publication outcomes and all other implementation remain pending;
full AC-03/11 coverage is not claimed. Q-01/02/03 have recommendations, not acceptance.
See [runtime decisions](#runtime-decisions-awaiting-review) and [design review](#consolidated-design-review).

## Context and design

This plan covers every recommendation in the event-processing investigation. It
prioritizes fault handling and bounded resources over speculative speedups. Existing
[005 ordering](../005-device-message-ordering/plan.md) and
[003 concurrency](../003-backend-race-repair/plan.md) artifacts remain historical.
The significant boundary decision is [ADR-001](adr-001-processing-boundaries.md).

Current flow:

```text
Paho ordered receive callback / HTTP collector
  → ingress FIFO (10,000 waiting items)
  → route → device-topic FIFO (1,000 waiting items; bridge topics share a lane)
  → DeviceProcessor → Seed or Update
  → apply state → synchronous automation → sampled metrics → store/UI callbacks
```

Repeated events bypass equality checks; state precedes debounce; seed differs from
updates; automations read live confirmed device state. Those are explicit invariants,
not cleanup opportunities. No asynchronous automation queue or value-equality event
filter is proposed. New abstractions must remove an identified ownership ambiguity,
duplicate validation or test limitation.

### 1. Establish reproducible evidence before changes

Paths: `backend/testing/hubharness/`, `backend/mocks/`,
`backend/internal/controllers/hub-controller_test.go`,
`backend/internal/automations/scenarios/`, `backend/internal/services/`.

1. Freeze representative traces for state echoes, repeated button/dial inputs,
   brightness steps, presets, presence cancellation, schedules, disabled devices,
   seed and manual triggers. Reuse existing cases rather than duplicating all tests.
2. Record the current harness limitation: cleanup disconnects fake Paho without a
   complete hub lifecycle. T002 baseline is complete with this limitation; T014
   adds full cancellation/join once runtime APIs exist. Every newly introduced
   worker must have local ownership/cleanup tests immediately. Rebaseline after
   harness cleanup; resource counts supplement global goroutine counts.
3. Run the current benchmark five times with the same ordinary-build toolchain and
   environment; retain raw output, platform, log level and commit with the PR evidence.
4. Add controlled fault fixtures: delayed/failing tokens, reconnect during registry
   mutation, blocked metrics, blocked WS output, initialization panic, timer/update
   interleaving and shutdown with queued/running work. Use barriers and fake clocks.
5. Demonstrate the targeted failure before each production fix where feasible.
   T026 closes the deferred controller save-without-execution and enabled-automation
   output-failure coverage gaps; T002 completion is not proof of these paths.

The earlier review baseline is evidence of the existing path, not a new-change test:
three runs with `-benchtime=3x` gave p95 2.79–3.52 ms, 47,239–47,327 allocations and
2,315,544–2,334,504 bytes per 1,000-message burst; reported burst goroutine growth 0,
idle hub goroutines 21. The latter does not prove teardown cleanup. The broadcast
observer is mocked; no production throughput or slow-client claim follows.

### 2. MQTT subscription reconciliation

Paths: `backend/internal/mqtt/mqtt.go`, proposed `subscriptions.go` and tests,
`backend/internal/controllers/handlers.go`, `backend/mocks/` fake Paho.

Reasoning: one authoritative desired set solves duplicate membership, reconnect
races, failed-subscription suppression and rename cleanup together. A mutex around
the existing network loop alone risks blocking callbacks and stale completion races.

Design:

- Store normalized fully-qualified topics in a set owned by the MQTT service.
  Centralize prefix addition/removal; use TrimPrefix semantics rather than ReplaceAll.
  Copy initial bridge topics into instance-owned state.
- Use one reconciler worker with a coalesced wakeup, a bounded retry timer, and
  connection-generation identifiers. Desired changes occur under a short lock;
  broker I/O runs outside it. Snapshot desired membership/version before I/O and
  recheck before recording success. Reconcile late results against current intent.
- Bridge/device registration defines desired subscriptions independently of the
  bridge-payload hash. Hash dedupe may skip domain registration only after its
  successful outcome; it must not erase pending subscription recovery. Return errors
  from registration where callers need to distinguish success from partial failure.
- On reconnect invalidate observed membership and reconcile the desired set. On
  rename establish the new topic before removing the old; unsubscribe with exactly
  the same full-topic format. Exclude unsupported/inactive bridge devices as today.
- Separate desired intent from broker-ack completion in the internal API. Prefer
  `EnsureTopic`/`ForgetTopic` returning validation/admission errors, plus reconciler
  status, over pretending an async operation is synchronously complete. Migrate all
  MqttClient implementations/fakes/callers together; no external MQTT wire change.
- Use cancellable, bounded token waiting for connect/subscribe/unsubscribe. A token
  deadline does not cancel the library operation by itself: retain its generation,
  reconcile late completion and disconnect the owned client on failed startup.
- Proposed operation/startup/backoff defaults and readiness boundaries are in
  [Q-01 recommendations](plan.md#q-01-bounded-mqtt-operations-and-application-shutdown).
  Startup requires connection and six mandatory bridge subscription acknowledgements,
  not arrival of device metadata. Cap each operation by the remaining total budget.
  Preserve Paho reconnect ownership and existing one-second network timeouts. Retry
  only subscriptions; preserve QoS 1 subscriptions, QoS 0 publication, retain flags
  and receive ordering.
- Remove dead commented locking and unused mutex state once ownership is real.

Tests include blocked acknowledgements while receive continues; concurrent
OnConnect/add/remove; failed subscribe followed by identical bridge announcement;
late subscribe success after removal; rename prefix correctness; reconnect during
shutdown; invalid topic; and startup deadline. Unknown-message routing is unchanged.

### 3. Stage errors, validation and diagnostics

Paths: controller device callbacks/bridge handlers, `backend/models/devices/requests.go`,
`backend/internal/mqtt/mqtt.go`, `backend/internal/automations/action.go`,
`backend/internal/lanes/lanes.go`, `backend/internal/ws/websocket.go`.

- Define a small in-process diagnostics observer with a no-op implementation and
  test recorder. Use atomic counters or short-lock snapshots, bounded stage/outcome
  enums and bucketed durations. Avoid one permanent metric label per topic/device.
- Record ingress and device-stage admission separately. Define terminal task outcomes
  so rejected/not-started, successful, failed, panic and discarded work reconcile;
  do not double-count a stage failure as a second terminal message. Expose current
  depth/oldest age/busy lanes through snapshots. Rate-limit repeated pressure logs
  without losing counters; snapshots and local structured summaries need no endpoint.
- Preserve /api/collect's admission-only 200 `Success` behavior. Document that later
  rejection is observable diagnostically, not returned retroactively to HTTP/Paho.
  Do not promise broker redelivery of a message rejected after the callback returns.
- Make callback error handling explicit. A first small patch can report returned
  errors at the wiring boundary; then use narrow error-returning sink interfaces
  where they improve tests. Keep automation before sampled outputs and preserve
  existing successful callback order. Attempt independent metrics/UI outputs even
  if one fails; never replay the entire update. Device storage is currently in-memory,
  so do not describe its error handling as a durability guarantee.
- Introduce command completion observation through a small publisher interface.
  Preserve immediate command submission order and normal-path timing. Track token
  completion using one owned, bounded tracker with timeout/close semantics, not one
  permanently waiting goroutine per publish. Paho submission itself can block up to
  its existing WriteTimeout: the tracker bounds observation, not submission latency.
  Test saturated submission separately and give the tracker Stop/Wait ownership now. If tracking is saturated, record loss
  of observation without dropping an otherwise valid command. Immediate rejection,
  completion error and timeout are distinct; timeout is indeterminate. Do not wait
  for every token inside the device lane or retry TOGGLE/step commands.
- Initially report asynchronous publication outcomes locally; no new browser success
  or acknowledgement claim. Keep existing immediate validation errors and user flows.
- Decode bridge envelopes into typed operation-specific data, validate required
  strings/status before mutation, and wrap errors with causes. Retain different
  rename/remove/interview/permit-join business actions. Shared helpers decode/validate;
  they do not hide operation-specific policy. Malformed JSON must return an error
  rather than log-and-return-nil in the device handler.
- Scope validation to touched inputs and use established authentication. Do not log
  raw household payloads, broker credentials or full command values.

### 4. One application lifecycle owner

Paths: `backend/main.go`, `backend/internal/hub.go`, `backend/internal/api/server.go`,
controller shutdown, device processor/lifetime, automation engine/handlers/actions,
`backend/store/initialiser.go`, `backend/store/tasks.go`, settings task ownership.

Reasoning: asynchronous context callbacks cannot prove that the application has
finished its work. Stop/cancel and join must be separate where reentrant callbacks
would deadlock. A lifecycle owner is a focused composition boundary, not a framework.

Proposed implementation:

1. Introduce a concrete runtime returned by hub composition, with Start/Wait/Shutdown
   responsibilities and narrow resource handles. Keep HTTP server ownership inside
   it. Main awaits runtime shutdown completion instead of returning on HTTP close.
2. Separate stop-intake from cancel-execution contexts. At shutdown mark admission
   closed synchronously, stop HTTP acceptance and MQTT callback ingestion, and freeze
   independent timer commitments. Keep the MQTT connection usable for immediate
   commands from already accepted work until the drain completes.
3. Drain ingress, then device lanes with one overall deadline. Retire pending delayed
   actions/schedules at the cutoff; drained events cannot rearm future delayed work.
   Preserve already committed immediate work. Describe the cutoff in tests, not just
   comments. Add a real commitment gate; a flag check followed by an unlocked
   processAction leaves a stop-versus-commit race. Test barriers before/after the gate.
   Avoid joining an action from its own Publish callback.
4. Stop/join availability monitors, scheduler registrations, delayed-action workers,
   subscription/token trackers and metrics cleanup. Add explicit wait handles where
   current Stop methods only cancel. Default scheduler construction must apply the
   supplied context before creating the scheduler (current constructor does it first).
5. Disconnect MQTT and close WS after effects drain; close repositories only when
   their users are joined. Track store-owned repositories and tasks explicitly and
   unwind partial construction in reverse order. Make Close idempotent and return
   contextual joined errors. Add matching lifecycle methods to mocks as required.
6. If deadline expires, account for queued discards and running work separately.
   Stop new effects where possible and return an explicit incomplete-shutdown error.
   Do not close storage underneath unfinished operations. Main records failure and
   exits nonzero; this is a documented forced-exit path, not successful graceful
   completion. The proposed five seconds covers all drain/join/cleanup stages;
   nested waits use remaining time. No automatic replay follows restart.
7. Replace lossy unbuffered done/error notification in ApiServer.Listen with a
   reliable result channel/return error and HTTP Shutdown using the remaining context.
   Cover bind/start failure as well as signal-driven shutdown.

No lifecycle operation may hold a registry lock while waiting for callbacks. A
shutdown test checks closure order, not only that goroutine count eventually falls.

### 5. Device availability and initialization

Paths: `backend/internal/services/device_lifetime.go`, `device_processor.go`,
`device_registrar.go`, controller routing and existing fake-clock utilities.

Availability:

- Extract a monitor owning one resettable Clock timer and instance generation. Keep
  existing one-second polling resolution initially; no shared timing wheel or
  deadline optimization is needed to solve correctness.
- Timer callbacks enqueue a generation-tagged timeout candidate onto the same device
  ordering boundary as state updates. Inside the lane reread last_seen, disabled
  status, availability and generation immediately before deciding/emitting offline.
- Never block a timer callback on its own lane. A full lane leaves one coalesced
  pending check for a later tick; count that pressure. Do not queue unlimited ticks.
- Preserve last_seen fallback/parsing semantics and online restoration; use the
  injected Clock for now/timers throughout. Route lifecycle configuration transitions
  consistently with device ordering and merged effective config, including overrides.
- Removal/replacement retires the matching processor entry and monitor, unregisters
  owned callbacks where needed and prevents stale callbacks reviving deleted devices.
  Validate name/ID mapping on rename so timers and manual triggers do not select a
  second concurrent lane. If changing the identity routing contract is necessary,
  amend 006 explicitly before implementing it; do not silently broaden spec 005.

Initialization:

- Replace the ambiguous creating flag with explicit initialization state/error.
  Install deferred failure cleanup around Seed and initial draining before invoking
  external callbacks. Count and release pending entries on failure; stop the monitor.
- Bound pending initialization events to the specified cap and report overload.
  Recover only at a boundary that can restore ownership invariants. Do not mark a
  partially initialized service ready merely to continue the lane.
- Proposed fail-closed state rejects subsequent device work with a bounded diagnostic
  until restart (Q-03). Latch by stable identity across enable, rename and remove/re-add.
  Cancel uncommitted delayed/scheduled work owned by this source, without cancelling
  unrelated automations targeting the device. Fault is not physical offline status.
  See [policy details](plan.md#q-03-initialization-panic-faults-only-its-source-device).
  Do not roll back in-memory values and claim already emitted
  commands were undone. Do not automatically reseed the same event. Registry locks
  remain released during Seed/Update and callbacks, preserving reentrant config tests.

### 6. WebSocket memory and reconnect recovery

Paths: `backend/internal/ws/websocket.go`, `eventhub.go`,
`frontend/src/store/modules/ws/index.ts`, `frontend/src/services/hubstate.service.ts`,
`frontend/src/store/modules/hub-state/`, contracts/types and matching Jest mocks/tests.

1. Resolve Q-02 in a focused design/test spike. Inventory existing snapshot freshness:
   /api/hubstate is cached; some accepted debounced state/availability changes do not
   mark it dirty. Initial page setup skips initialization when persisted state exists.
   Merely calling the existing fetch on reconnect is insufficient evidence of truth.
   Prefer a recovery read bypassing the cache; copy values under an ordering boundary
   rather than serializing live mutable pointers. See the
   [spike checklist](plan.md#q-02-prove-browser-recovery-before-reducing-queues).
2. Specify an authenticated fresh-state read and an ordering fence against concurrent
   deltas, including rename/removal and stale socket/request generations. Prefer an
   existing contract if it can meet the acceptance test. If additive per-device
   revisions and a snapshot revision are needed, write the producer/consumer contract
   addendum and update the ADR before source changes. Do not use wall-clock last_seen
   as a unique sequence number; equal timestamps and availability-only changes exist.
3. Test the chosen protocol with updates during snapshot acquisition, delayed older
   HTTP responses, rapid reconnects, backend restart, snapshot failure and added/
   removed/renamed devices. Vuex setDevices replaces membership: stale snapshots can
   lose new devices or resurrect removals. Specify epoch and membership fencing. Bound any
   temporary delta buffer and retry only state reads. Apply state via existing Vuex
   ownership; do not maintain a second device truth in the WS module.
4. Define separate max-frame-byte and max-queued-message constants/options. Candidate
   queue 256 messages; select final size from normal-burst/slow-client measurements.
   Tune read/write frame buffers only with evidence; do not conflate them with limits.
5. On Melody ErrMessageBufferFull, count and close that session idempotently using a
   bounded, nonblocking path. Avoid recursive remote logging/broadcast from the full
   queue handler; diagnostics must be local for this failure. Healthy clients continue.
6. Remove disconnected client entries instead of retaining false values forever.
   Frontend reuses its reconnect logic, cleans timers/listeners on replacement and
   distinguishes connected transport from completed state synchronization. Use existing
   status/alert patterns; no unrelated visual redesign or persisted household logs.
7. Ship consumer recovery before enabling the lower backend queue. Prove behavior
   with the existing frontend and a controlled local WS server, no household commands.

Queue reduction is explicitly BLOCKED until Q-02 is resolved and AC-10 passes.
That gate prevents a memory improvement from introducing silent stale UI state.

### 7. Focused cleanup and measurement-led optimization

- Extract `registerEventHubEvents` into cohesive request wiring files by responsibility
  (device/configuration/automation), retaining controller/service/repository boundaries.
  Prefer a constructor with required dependencies over partially initialized builders
  when migrating the device processor. Avoid rewriting unrelated MCP handlers.
- Use small consumer-owned interfaces for lookup, publication, sampled metrics and
  broadcasting; do not replicate the entire AppStore/MqttClient interface everywhere.
- Within Update, extract a result containing automation delta and sampled output only
  if it improves clarity after tests exist. Apply state before debounce; preserve
  availability-only notifications and keep Seed's different behavior explicit.
- Share typed bridge decoding, normalized topic formatting and lifecycle diagnostics.
  Remove stale TODOs/dead members only in touched code; do not rename the whole tree.
- Profile allocations, time spent in synchronous metrics/serialization/broadcast,
  and lane lock contention under controlled slow I/O. Report p50/p95/p99, queue peak,
  reject count, allocations, retained memory and goroutine teardown per scenario.
- A ring buffer for lane tasks, metric encoding reuse, timer scheduling changes or
  asynchronous sinks are conditional follow-ups, not promised improvements. Prove the
  hotspot first; additional timing/storage changes need an amended spec/ADR and tests.

## Runtime decisions awaiting review

Proposed runtime defaults and failure/recovery behavior, not repository governance.
All choices remain pending. T002/T006/T007 authorization did not accept them.
The pipeline invariants above apply throughout; no event replay or device-command retries.

### Q-01: bounded MQTT operations and application shutdown

| Boundary | Recommended default | Reason and implementation constraint |
| --- | --- | --- |
| MQTT token observation | 5 seconds | Bound waiting for subscription/unsubscription completion and observing publication completion. Use the smaller of this window and the remaining parent budget. Timeout means uncertain outcome, not proof of delivery failure. |
| MQTT startup readiness | 10 seconds total | From initiating MQTT connection until connected and all six mandatory bridge subscriptions acknowledge success. Device metadata arrival is not a readiness condition. No sequential operation may reset the total budget. |
| Subscription retry | 1, 2, 4, 8, 16, 30 seconds, then 30 seconds | Cancellable, coalesced reconciliation of desired membership. Reset after success/new connection generation; do not create a second connection-retry worker alongside Paho reconnect. |
| Graceful shutdown | 5 seconds total | One budget for intake closure, drain, worker joins and cleanup. Nested waits receive only the remaining time. Record incomplete work and exit unsuccessfully on expiry. |

Keep current one-second Paho ConnectTimeout/WriteTimeout/PingTimeout settings
separate from these application limits. Start with constructor options/defaults and
injectable clocks; avoid adding persisted configuration without demonstrated need.
The initial retry schedule is deterministic; jitter can be considered separately if
shared-broker measurements justify it.

Paho v1.5.0 `Publish` can itself wait for its outbound channel up to WriteTimeout.
A completion tracker bounds token observation after submission; it cannot promise
nonblocking submission. Test saturated submission separately, as well as immediate
error, late completion, reconnection and tracker saturation. QoS 0 completion is not
physical acknowledgement. Never turn an indeterminate result into a command retry.
Each new tracker/reconciler must own cancellation and join tests when introduced;
T014 later composes those lifecycles rather than retrofitting ownership.

On startup failure, return a contextual error, disconnect the owned client and unwind
partial construction within the separate shutdown cleanup budget. The ten-second
MQTT readiness limit is not a claim that the entire process and cleanup take ten
seconds. Main must report failure and exit nonzero.

At shutdown start, freeze new delayed/scheduled action commitments. Accepted queued
immediate events may finish; they cannot arm delayed work after the cutoff. Use an
explicit commitment gate: checking a cancellation flag then unlocking before an
effect is insufficient. Test barriers on both sides of the commitment point and
avoid joining a worker from its own callback. Already committed physical commands
cannot be undone. On deadline expiry, account for discarded queues and unfinished
effects; never close repositories while their users remain active. Let process exit
release remaining resources, without automatic replay on restart.

Alternatives: unbounded waits hide failure indefinitely; immediate cancellation of
all accepted work can truncate effects; longer defaults postpone feedback without
adding delivery guarantees. The proposed limits need deterministic timeout tests and
local broker-failure fixtures before acceptance for deployment.

### Q-02: prove browser recovery before reducing queues

Recommend a focused design/test spike, not approval of a wire protocol. Retain the
existing production queue default until recovery-capable clients and compatibility
handling pass the agreed contract tests. A 256-message queue is a measurement
candidate, not an accepted setting or a bound on total process memory.

The spike must deliver `contracts/reconnect.md`, with producer/consumer fixtures:

1. A fresh authenticated snapshot of copied device values, including availability
   and accepted debounced state. Copy under a coherent ordering boundary; serializing
   mutable device pointers later is not an atomic snapshot. Prefer bypassing the
   existing hubstate cache for recovery over a broad unrelated cache rewrite.
2. An ordering fence between snapshot and deltas. Compare a global stream revision
   with per-device revisions plus a membership fence; specify server epoch/restart
   behavior and socket/request generations. `last_seen` is not a unique revision.
3. Rename/removal membership semantics and tombstones or an equivalent fence. Vuex
   `setDevices` replaces the entire map: a stale response can lose a newly added
   device or resurrect a removed one, as well as overwrite fields.
4. Bounded temporary buffering, failed-refresh/read-only retry policy and a visible
   distinction between transport connected and state synchronized. Refresh snapshots
   must never enter the automation event-processing path.
5. Old-client rollout and rollback. Deploy consumers first; retain old limits during
   the compatibility window unless a reviewed capability policy safely separates
   sessions. Measure startup bursts, snapshots and log traffic before selecting caps.

Test updates during snapshot capture, delayed older responses, add/remove/rename,
rapid reconnect, backend restart, buffer overflow and refresh failure. Existing
cached `/api/hubstate` plus blind buffered-delta replay cannot establish correctness:
cache invalidation misses some state/availability changes and can race refresh.
Any additive revisions, fresh-read option, WS request or endpoint require an explicit
contract addendum and existing authentication/error conventions. No transport choice
is accepted yet. If the fence cannot be proven, defer queue reduction.

### Q-03: initialization panic faults only its source device

Recommend fail closed until service restart, with no automatic reseed/replay. Latch
the fault by stable device identity for the process lifetime; configuration enable,
rename and remove/re-add must not accidentally clear it. Keep other devices running.
Faulted processing is distinct from physical offline status; do not invent an offline
transition or new UI schema to communicate the fault.

At the fault boundary, release/count pending initialization work, retire monitoring,
and cancel uncommitted delayed actions/schedules owned by that source device's
automation lifecycle. Define ownership precisely: this is not a blanket cancellation
of healthy automations that target the same physical device. Check existing shared
scheduler ownership before implementing per-source cancellation.

Effects committed before panic remain possible/indeterminate; report them without
claiming rollback. Keep registry locks out of callbacks and preserve reentrant
configuration behavior. Test panic in Seed and pending drain, cancellation on both
sides of commitment, stable-ID replacement/rename, unrelated-device progress and
restart clearing the in-memory fault latch. Restart permits fresh initialization;
it does not replay failed events or guarantee exactly-once physical effects.

Alternatives: automatic retry risks duplicate toggle/step commands; a manual reset
would need a separate safety/reconciliation contract. Restart-only recovery is
simpler to reason about but sacrifices availability for that source until restart.

### Acceptance record

Q-01: pending. Q-02: spike recommended; protocol pending. Q-03: pending.
Record actual maintainer decisions, date and scope here and in the spec before
implementing dependent behavior. No approval is inferred from this document update.

## Constitution Check

PASS (design) means the stated design addresses the rule; it is not implementation
verification or maintainer approval. Open decisions block their dependent slices.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec acceptance and tasks; Q-01/02/03 | BLOCKED for dependent slices pending decisions | Pending |
| NH-02 | AC-01–03; no command retries; shutdown cutoff | PASS (design); cutoff review pending Q-01 | Pending golden/scenario evidence |
| NH-03 | Typed touched inputs; unchanged auth; local payload-free diagnostics | PASS (design) | Pending fault/contract checks |
| NH-04 | ADR-001; narrow interfaces; WS addendum if necessary | BLOCKED for WS contract, Q-02 | Pending boundary/contract review |
| NH-05 | Traceability and isolated test commands below | PASS (design) | Pending implementation checks |
| NH-06 | Bounded work, join before close, no migration/replay | BLOCKED for Q-01/03 failure policy | Pending lifecycle evidence |
| BE-01 | Transport decoding delegates to existing services | PASS (design) | Pending dependency review |
| BE-02 | Seed/update/event/dedupe/debounce cases explicitly retained | PASS (design) | Pending AC-01–03/13 |
| BE-03 | Ownership/join/clock/reconciler specification | BLOCKED for Q-01 timing | Pending races/lifecycle tests |
| BE-04 | Metrics meaning unchanged; writes observable; defaults documented | PASS (design) | Pending error/default checks |
| BE-05 | Existing auth/contracts required for any recovery transport; publication outcomes not physical ack | PASS (design) | Pending contracts; WS Q-02 |
| FE-01 | Existing Vuex modules/services own recovery | PASS (design) | Pending store tests |
| FE-02 | Snapshot/delta ordering contract must precede source edits | BLOCKED, Q-02 | Pending producer/consumer tests |
| FE-03 | Sync/offline distinction; read-only retries | PASS (design); Q-02 detail pending | Pending reconnect tests |
| FE-04 | Existing status patterns; verify text/keyboard at 390/1280 px if UI touched | PASS (design) | Pending UI check or justified N/A |
| FE-05 | Existing authenticated helpers; no new browser persistence/secrets | PASS (design) | Pending source review |

No exceptions requested or granted. Recheck before implementation, after design
amendments, and before merge. Constitutions remain proposals.

## Contracts, data, rollout

- MQTT topics/payload/QoS/retain semantics unchanged. Internal client/fake interfaces
  migrate together; desired subscription admission is distinct from broker completion.
- Existing HTTP collection response remains admission-only. No public diagnostics API.
- WS queue-overflow close behavior changes; existing events stay compatible. Q-02
  may introduce additive recovery metadata and fresh-snapshot semantics. Its addendum
  must name fields, errors, auth, ordering, old-client behavior and deployment order.
- No device/automation JSON, metrics key/value, units, retention or timezone migration.
  Recoverable failures produce missing observations, never invented measurements.
- Document proposed timeouts/queue capacities in backend configuration docs when
  selected. Prefer constructor options/constants over new persisted settings unless
  operators need runtime configuration; no unused configuration surface.
- Deliver independently reviewable patches in task order: baseline/diagnostics,
  MQTT, lifecycle/device ownership, WS recovery, buffer reduction, cleanup/performance.
  Do not deploy a smaller queue before the recovery-capable client is deployed and
  old-session behavior is explicitly addressed. Use a temporary opt-in capacity or
  retain old capacity during the compatibility window; specify removal in Q-02.
- Roll back an individual patch to the previous binary; there is no data migration.
  If WS recovery metadata is added, revert queue pressure behavior first and retain
  compatibility with deployed clients. Never restore operation by replaying physical
  commands. A faulted device currently requires restart under the proposed policy.
- Operator smoke checks are passive: connection health, subscription outcomes, queue
  age, rejection/failure counts and memory. Real-device commands require a separately
  scoped manual instruction; old spec authorizations are not assumed for this work.

## Traceability and verification

| Requirement / acceptance | Paths | Required evidence | Actual result |
| --- | --- | --- | --- |
| FR-01; AC-01–03 | controllers, services, automations/scenarios | 2,000 replay trials; unchanged golden commands; seed/disabled/manual/debounce | Existing traces and T002/T007 suites passed; controller save invariant pending T026 |
| FR-02; AC-04–06 | mqtt, controllers/handlers, mocks | Reconnect/rename/late-token/failure/backoff/startup tests with fake Paho | Pending |
| FR-03; AC-07–08 | main, hub, api, store, services, automations | Ordered close log, deadline/partial-startup tests, no use-after-close or late timer command | Pending |
| FR-04; AC-09–10 | ws, frontend WS/store/service | Slow-client fixture, reconnect races, fresh snapshot/removal and memory baseline | BLOCKED by Q-02 |
| FR-05; AC-11–12 | callbacks, mqtt, action, handlers | Injected output failures; malformed typed payloads; no duplicate publish | Callback/validation subset passed; counters/publication and enabled-automation failures pending T005/T008/T026 |
| FR-06; AC-13 | lifetime, processor, registrar, lanes | Fake-clock timeout/update/disable/remove/rename generation interleavings | Pending |
| FR-07; AC-14 | processor, lifetime | Seed/drain panic, bounded pending, observable fault, other-device progress | Policy Q-03 pending |
| FR-08; AC-15 | lanes, mqtt, ws, diagnostics | Accounting identities under admission failure, panic, overload, shutdown | Pending |
| FR-09; AC-16 | controller wiring, update/result helpers | Golden trace equality; no package cycles or new framework | Typed decoder/existing traces passed for scoped T007; broader T021 extraction pending |
| NFR-01 | existing burst benchmark + slow sink fixtures | Five-run same-machine comparison within +10% | T007 five-run mocked comparison met +10% thresholds; slow-sink/production assessment pending |
| NFR-02/04 | lifecycle/queues/monitors | Cap boundaries; 100 create/retire cycles; deterministic five-second deadline | Pending Q-01 |
| NFR-03; SC-01 | all affected packages | Required suites/build/lint/vet/race and recorded failures | T002/T007 required backend checks passed; final broader-slice checks pending |
| NFR-05 | ws/session fixture | Selected cap, ≥90% lower retained empty-session memory, normal burst no disconnect | Pending Q-02 |

Before editing Go, use backend/go.mod toolchain (Go 1.24.0, toolchain go1.24.8);
record actual `go version`. Frontend uses `.nvmrc` and the lockfile. Commands from
backend after implementation:

```sh
go test ./...
go build ./...
go vet ./...
go test -race -count=1 -timeout=5m ./internal/lanes ./internal/controllers ./internal/services ./internal/automations/... ./internal/ws ./internal/mqtt ./internal/api ./store ./models/settings ./repository ./internal/metrics/storage
go test ./internal/controllers -run '^$' -bench '^BenchmarkMixedBurst$' -benchtime=3x -count=5
```

Run targeted deterministic race tests repeatedly for reconnect/shutdown/availability
after adding them. Do not keep rerunning unrelated suites once required checks pass.
Profile ordinary builds separately from race tests, with no competing benchmark jobs.
Use temporary databases, fake Paho, fake clock, local WS clients and fake store errors.

For frontend changes, from frontend:

```sh
npm ci
npm test -- --runInBand
npm run lint
npm run build
```

`npm ci` is needed only when dependencies are absent/stale. Do not use the watch-mode
type-check command as a terminating gate. Verify offline/reconnect/sync failure and
existing controls at 390 px and 1280 px, keyboard accessible with textual status if
UI presentation is touched. Backend-only slices record frontend checks as N/A.

Actual pre-plan investigation on 2026-10-08:

- `go test -race -count=1 -timeout=5m ./internal/lanes ./internal/controllers
  ./internal/services ./internal/automations/... ./internal/ws ./internal/mqtt` passed;
  mqtt reported no runnable tests. Sandbox cache denial was resolved by an approved
  rerun. No household devices were commanded.
- Existing benchmark command above with `-count=3` passed; results recorded in §1.
- Full suites/build/vet/frontend checks were not run during the investigation or
  documentation-only planning. No new feature acceptance is claimed passed.
- Temporary logs: `/private/tmp/nodeherder-event-review-tests.log` and
  `/private/tmp/nodeherder-event-review-benchmark.log`; summarized here because temp
  files are not durable repository evidence. Future change evidence belongs in PRs
  or a committed results note with sensitive logs removed.

## Debt, risks, exceptions

- QoS subscriptions and in-memory lanes do not provide durable end-to-end delivery.
  Changing broker ACK handling or adding an outbox is excluded and would need a new
  delivery/idempotency design for non-idempotent physical events.
- Per-device caps do not bound arbitrary topic cardinality or HTTP payload sizes.
  Do not claim total process-memory bounds; separate ingress admission limits need
  client-contract evidence before introduction.
- Reconnect tests must exercise the real adapter over fake tokens, not only mocks
  that serialize away the race. Current passing suites omit this coverage.
- Moving availability checks onto lanes can delay the offline notification under
  load; record queue latency, recheck fresh state and preserve check resolution.
- Rename/bridge metadata mutation can cross device ordering boundaries. Include
  focused races and alias/generation tests; touched races must be fixed or the slice
  blocked, not dismissed because old tests passed.
- Late token completion and shutdown timeout are indeterminate outcomes. Reporting
  them as definite device-command failure invites unsafe manual/automatic replay.
- A reduced WS queue without a proven refresh fence can regress browser truth.
  Q-02 deliberately blocks that slice; adding sequence metadata expands contract work.
- No performance improvement is promised beyond measured WS allocation reduction.
  Callback extraction and SOLID terminology alone do not justify extra interfaces.

Exceptions: none. Scoped T002/T006/T007 completion is recorded; broader implementation
and maintainer policy acceptance remain pending.

## Completion

Done requires resolved Q-01/02/03 for shipped slices, spec → plan → task agreement,
targeted regression evidence, passing required checks, updated architecture/config
docs, producer/consumer compatibility, recorded benchmark results and actual review.
T002 changed tests; scoped T006/T007 changed controller validation/error reporting.
Both saved [execution evidence](evidence.md#t002-characterization-and-faster-tests).
Remaining implementation must update task boxes
and actual-result cells only with evidence, preserving this investigation baseline.

## Consolidated design review

Reviewed 2026-10-08: 006 artifacts/evidence, current backend pipeline, pinned Paho,
controller changes/failure tests, initialization/lifetime, delayed actions, hubstate
cache and frontend WS/store recovery. Includes uncommitted T002/T006/T007; not a
whole-repository audit. Historical 003/005 artifacts remain unchanged. No additional
confirmed regression found in the scoped implementation; T026 combined failure/save
coverage remains missing. Full acceptance, stronger delivery guarantees and production
throughput gains are not claimed. See [execution evidence](evidence.md).

### Findings and disposition

| Priority | Finding | Recommendation / disposition |
| --- | --- | --- |
| High | Cached hubstate and blind delta replay cannot prove recovery; whole-map replacement also risks lost additions/resurrected removals. | Keep queue reduction blocked. Q-02/T003 now require atomic copied values, ordering/membership fencing, restart generations and compatibility proof. No protocol is chosen. |
| High | Shutdown cancellation checked before an unlocked effect leaves a commitment race. Resource closure during unfinished work is unsafe. | Q-01/T012–T014 require a commitment gate, deterministic before/after barriers, one total budget, join-before-close and explicit unsuccessful forced exit. Implementation pending. |
| High | Initialization failure may leave pending work/resources or later duplicate side effects if retried. | Q-03/T017 specify stable-ID fault latch until restart and cancellation of source-owned uncommitted work; no replay or invented physical-offline state. Policy acceptance pending. |
| Medium | Token timeout cannot cancel Paho operations, and publication submission itself can block. | T008/T010 distinguish submission, observation, late results and physical acknowledgement. Preserve current network timeouts; new workers own Stop/Wait tests immediately. |
| Medium | Disabled-automation fault tests do not prove command behavior under enabled automation; save-without-execution coverage was deferred. | Added test-only T026. Full AC-03/11 remains unclaimed. Final validation depends on it. |
| Medium | Baseline harness disconnects fake Paho but does not join the full hub lifecycle. | T002 remains completed with explicit limitation; T014 owns composed teardown and rebaseline. Baseline goroutine counts are not leak-free evidence. |
| Low | Plan progress said no production changes despite completed scoped T006/T007; endpoint prohibition preempted the Q-02 spike. | Corrected progress and ADR/BE-05 matrix. Any new recovery contract must be reviewed/authenticated, rather than assumed unnecessary. |

### Abstractions and cleanup

Retain small consumer-owned interfaces and cohesive ownership: typed bridge decoding
already removes repeated parsing while business actions stay visible. Next abstractions
should be the MQTT desired-membership reconciler, bounded publication observer,
application runtime and clock-driven availability monitor. Each owns its resources,
cancellation and tests; orchestration composes them. Keep Seed and Update explicit.
Do not introduce a general event framework, asynchronous sinks, batching or broad
controller moves without measured need. T021 mechanical extraction follows behavior
fixes so review can separate movement from semantic changes.


Keep spec (outcomes), plan (design/decisions/review), tasks (sequence), ADR (significant
boundary alternatives) and evidence (actual results). Removed redundant standalone
reports/raw-output files; all unique execution data is consolidated in evidence.md.
Next independent candidate: T026 test-only coverage. Q-01/Q-03 acceptance and Q-02
protocol design still gate dependent behavior; no pending runtime behavior was implemented during consolidation. Constitutions remain proposed; actual maintainer review
is pending.
