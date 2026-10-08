# Spec: Event processing hardening

ID: 006-event-processing-hardening | Status: Draft for review | Author: Codex | Updated: 2026-10-08
Constitutions: parent/backend/frontend 1.0.0, proposed, not ratified.
Components: backend; limited frontend reconnect recovery and contract verification.

## Problem, scope, exclusions

The user reports no current service problems and requests preventative reliability,
performance, and maintainability improvements. Preserving device and automation
behavior is the primary acceptance condition. This is a follow-up to
[005 ordered processing](../005-device-message-ordering/spec.md) and
[003 race repair](../003-backend-race-repair/spec.md), not a replacement of their history.
The initial request authorized planning only. On 2026-10-08 the user subsequently
authorized T002 baseline characterization and test speed/reliability improvements;
see [T002 evidence](evidence.md#t002-characterization-and-faster-tests). The user later authorized scoped T006/T007 validation
and output-error handling; see [T006/T007 evidence](evidence.md#t006t007-validation-and-output-errors). The user then
authorized T026 and the MQTT topic-mapping subset of FR-02; see
[T026 and topic mapping evidence](evidence.md#t026-and-mqtt-topic-mapping). Other
production changes and Q-01/02/03 decisions remain pending review.

Repository review at `26f48d08` found:

- MQTT subscription reads/writes are unsynchronized, unsubscribe uses a different
  topic format from subscribe, and a bridge announcement is marked handled even
  when subscriptions fail. Subscribe/connect waits have no application deadline.
- Lane draining is triggered asynchronously on cancellation; application exit does
  not join it. Availability monitors have independent background contexts.
- WebSocket queue capacity is 1,048,576 messages, using a byte-size constant.
  Melody 1.2.1 allocates that channel per session. Browser reconnect currently
  requests MCP status but does not refresh device state.
- Callback wiring discards metrics/device-store errors; publication completion is
  ignored. Availability timeout decisions can become stale before commitment.
- A panic during seed/draining can leave an entry permanently initializing and
  continue growing its pending list despite the outer lane recovering.
- Queue rejection is bounded and logged, as intentionally specified in 005, but
  ingress acceptance and final processing outcomes are not separately observable.

These are code-path findings, not claims that failures occurred in the household.
The existing FIFO lanes and event/state distinction are retained.

Scope includes subscription reconciliation, lifecycle ownership, bounded consumer
buffers, failure reporting, availability ordering, initialization recovery, useful
diagnostics, and focused extractions. Performance work begins with measurement.
No broker replacement, durable event log, exactly-once claim, MQTT QoS change,
automatic command replay, schema migration, general event bus, automation arithmetic
change, or unrelated frontend redesign is included. Async metrics storage/batching
requires a separate measured justification and spec amendment.

## Scenarios and acceptance

### US-01 — Existing device behavior remains stable (P1)

Verification: mocked hub replay, fake clock, existing automation scenarios.

- AC-01: Given identical consecutive physical button/dial events, every accepted
  event is evaluated; repeated unchanged state remains deduplicated.
- AC-02: Given stale-then-new state in arrival order, the final state and emitted
  final update are new in 2,000 randomized replay trials. Automation reads fresh
  state before metric/UI debounce; one slow device does not reorder another.
- AC-03: Dial step arithmetic, confirmed-target brightness basis, preset cycling,
  presence delay/cancel, unchanged-false delay behavior, manual-trigger isolation,
  schedule windows, seed, disabled-device and save-without-execution behavior match
  pre-change golden traces. No attempted recovery replays a device command.

### US-02 — Broker interruptions recover subscriptions (P1)

- AC-04: Concurrent reconnect, add, rename and remove reconcile to the current
  desired fully-qualified topic set without races, duplicate registry entries,
  stale resubscriptions after removal, or blocking the receive callback.
- AC-05: Subscription failure remains pending and is retried with bounded backoff;
  an identical bridge announcement does not suppress recovery. Each network wait
  is bounded and shutdown interrupts retry scheduling. A timed-out operation that
  completes later cannot publish stale subscription state.
- AC-06: Startup with an unavailable broker exits or exposes a defined startup
  failure within the configured overall deadline; successful reconnect restores
  desired subscriptions without changing QoS or command retry semantics.

### US-03 — Shutdown has a verifiable completion boundary (P1)

- AC-07: Shutdown rejects new ingestion, drains ingress then device work, retires
  timers/workers and transport resources, and only then closes repositories and
  returns to main. Repeated shutdown is safe; partial startup unwinds owned resources.
- AC-08: At deadline, queued discards and running unfinished work are separately
  reported. No repository is closed underneath running work. Delayed actions not
  committed at the documented shutdown cutoff cannot publish afterwards; already
  committed actions are joined within the remaining budget or reported unfinished.

### US-04 — A slow browser cannot consume unbounded memory (P1)

- AC-09: Each session has an explicit bounded message-count queue, independently
  of frame byte limits. A full queue disconnects the affected slow session once,
  records the reason and does not block device processing or other sessions.
- AC-10: After reconnect, the browser refreshes confirmed device state using
  authenticated reads and reconciles concurrent updates without stale rollback.
  Failed refresh remains visibly unsynchronized/offline, does not clear good state,
  and never resends device commands. Removed devices and availability are covered.
  Snapshot ordering must be resolved under Q-02 before queue reduction ships.

### US-05 — Failures are visible without duplicate physical effects (P1)

- AC-11: Injected metrics-store, device-store, broadcast and publication failures
  identify the failed stage and increment an outcome counter. One failed sampled
  output does not suppress independently valid downstream outputs. No whole-event
  replay occurs. Publish timeout is reported as indeterminate, not confirmed failure
  of the physical device or confirmed state.
- AC-12: Malformed bridge response fields and invalid device JSON return bounded,
  contextual validation errors without panic, state mutation or command execution.
  Unknown valid expose names retain the existing ignore behavior.

### US-06 — Device lifecycle remains consistent at boundaries (P2)

- AC-13: A timeout candidate queued before a fresh device update cannot mark the
  fresh device offline after that update commits. Disable/remove/shutdown retire
  monitor resources and stale timer callbacks cannot affect a replacement instance.
  Preserve one-second timeout-check resolution and current last_seen interpretation.
- AC-14: Injected panic during Seed or initial pending draining leaves no permanent
  initializing state or unbounded pending queue. The failed device is explicitly
  faulted; accepted pending work is counted as failed, subsequent work is rejected
  observably, and unrelated devices continue. No automatic reseed/replay is allowed
  where physical effects may already have happened; recovery policy is Q-03.

### US-07 — Operators and maintainers can assess pressure (P2)

- AC-15: Diagnostics distinguish received, admitted, rejected, started, completed,
  failed, shutdown-discarded and unfinished work by stage; queue depth, oldest wait,
  processing duration and slow-client disconnects are available without raw payloads.
  Under a deterministic overload fixture, every submitted item has an accounted
  outcome and queue limits hold. Existing ingress/device rejection policy remains.
- AC-16: Small focused interfaces/helpers replace duplicated transport parsing and
  lifecycle wiring. Seed and Update stay distinct; golden output sequences remain
  identical through mechanical extractions. No new generic framework is required.

## Requirements and success

| ID | Required outcome | Acceptance |
| --- | --- | --- |
| FR-01 | Preserve existing device/event/automation semantics and FIFO | AC-01–03 |
| FR-02 | Reconcile MQTT subscriptions safely with bounded recovery | AC-04–06 |
| FR-03 | Own and join application lifecycle resources | AC-07–08 |
| FR-04 | Bound WebSocket memory and recover client truth safely | AC-09–10 |
| FR-05 | Surface stage errors and validate touched transport inputs | AC-11–12 |
| FR-06 | Serialize availability decisions and retire device resources | AC-13 |
| FR-07 | Bound and terminate failed initialization | AC-14 |
| FR-08 | Account for processing and overload outcomes | AC-15 |
| FR-09 | Reduce responsibilities/duplication without semantic drift | AC-16 |

- NFR-01: Same-machine before/after 1,000-message, 20-device benchmark: median of
  five runs for p95 per-message latency and allocations per burst must be no worse
  than baseline +10%. Use identical Go version, flags, log level and fixture. Do
  not compare race builds with ordinary builds or use the review numbers as a CI SLA.
- NFR-02: Queue cap remains 10,000 at ingress and 1,000 waiting tasks per device;
  provisional initialization pending cap is 1,000. After 100 create/retire cycles,
  task/monitor/session registries return to baseline and owned workers are joined.
  Per-key caps do not constitute a global device-cardinality/memory limit.
- NFR-03: Affected race tests pass; no routine test connects to household devices.
- NFR-04: Proposed total graceful shutdown budget is five seconds, with deterministic
  deadline tests. Proposed MQTT operation timeout is five seconds, startup deadline
  ten seconds, subscription retry backoff 1/2/4/8/16/30 seconds capped at 30 seconds.
  Startup means connection plus six mandatory bridge subscription acknowledgements;
  nested operation waits use remaining total time. Shutdown includes drain/join/cleanup
  and returns failure on expiry without closing resources under active users.
  These are reviewable design defaults, not already accepted operational requirements.
- NFR-05: Select and record the WS queue capacity using the slow-client fixture;
  initial candidate 256 messages. Measured retained per-session memory after forced
  GC must be at least 90% lower than the existing empty-session baseline, without
  normal-burst disconnects. Record frame buffers separately; preserve 1 MiB frame
  limit unless an explicit contract amendment is made.
- SC-01: All applicable acceptance scenarios pass, old golden traces remain equal,
  failure injection demonstrates the targeted fixes, and all component merge checks
  have recorded outcomes. Timing-policy decisions are reviewed before dependent work.

Metrics units, timestamps, retention and storage format remain unchanged. Missing
metrics caused by a failed write remain missing, never synthesized as zero.

## Entities and external interfaces

- Inbound message: topic, transport and payload, accepted into an in-memory FIFO;
  acceptance is not a durability or processing acknowledgement.
- Subscription: desired topic membership plus current connection-generation result.
- Device lifecycle: uninitialized/initializing/ready/faulted/retired; initialization
  failure is distinct from physical offline availability.
- Publication outcome: local submission/completion/error/indeterminate; none alone
  confirms physical state. MQTT topics, QoS, retain and payloads remain unchanged.
- HTTP /api/collect retains its existing status/body and admission-only meaning;
  changing it to an end-to-end acknowledgement is excluded.
- WS existing event payloads remain compatible. Recovery contract details are gated
  by Q-02. Diagnostics use local structured logs/in-process snapshots initially;
  no new diagnostics endpoint or household-data persistence is proposed. Any recovery
  transport addition requires the Q-02 contract review.

## Questions and assumptions

[Runtime decisions and reasoning](plan.md#runtime-decisions-awaiting-review) remain
proposed. Open questions block only dependent implementation.

- Q-01 — Maintainer to review bounded MQTT startup/operations, subscription backoff
  and shutdown cutoff/defaults in NFR-04; preserve normal event timing.
- Q-02 — Implementer to demonstrate, maintainer to review fresh snapshot/delta and
  membership ordering, restart/session generations, bounded recovery and old-client
  compatibility. Cached hubstate plus blind replay is insufficient; queue reduction
  remains blocked. Additive transport/revisions require a contract addendum.
- Q-03 — Maintainer to review stable-ID fault-until-restart and cancellation of
  source-owned uncommitted actions; no replay, automatic reseed or false offline state.

- A-01 — Existing normal-load behavior is the baseline, not an incident to repair.
- A-02 — No additional persistence or broker service is needed for this scope.
- A-03 — Review findings need focused regression reproductions; a passing general
  race suite does not prove reconnect or logical ordering paths safe.

Open questions block only dependent implementation. No approval is recorded here.
