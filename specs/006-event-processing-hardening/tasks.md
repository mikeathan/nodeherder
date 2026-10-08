# Tasks: Event processing hardening

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: T002 and scoped T006/T007 complete; broader work pending review
Updated: 2026-10-08. [Runtime decisions](plan.md#runtime-decisions-awaiting-review) | [Review](plan.md#consolidated-design-review).
Evidence: [T002 evidence](evidence.md#t002-characterization-and-faster-tests), [T006/T007 evidence](evidence.md#t006t007-validation-and-output-errors).

The user authorized T002/test improvements and subsequently the scoped T006/T007
validation/error-reporting work on 2026-10-08.
Other tasks remain reviewable proposals. Maintain acceptance IDs and record actual commands/results
when executed. Changes with different failure semantics should be separate commits.

## Preparation and decisions

- [ ] T001 — Confirm scope and review Q-01 (deadlines/backoff/shutdown) and Q-03
  (initialization fault recovery) in `spec.md` and the runtime decisions section of `plan.md`; review ADR-001. IDs: FR-02/03/07,
  NFR-04, NH-01/02/06. Depends: none. Done: actual decisions, reviewer and date are
  recorded; no inferred approval. Unresolved decisions block their dependent tasks.
- [x] T002 — Characterize existing semantics and refresh performance baseline in
  `backend/internal/controllers/hub-controller_test.go`, `internal/services/`,
  `internal/automations/scenarios/` and `testing/hubharness/`. IDs: FR-01,
  AC-01–03, NFR-01/03. Depends: none. Done: reusable golden trace inventory,
  baseline five-run output/environment and gaps recorded; no live broker use.
  Evidence: [T002 evidence](evidence.md#t002-characterization-and-faster-tests), benchmark/check tables, new
  seed/update sequence test, full backend tests/vet, focused races and 20 repeated
  changed-test runs passed. Only tests/documentation changed; pending policy reviews
  and complete harness lifecycle cleanup remain outside this slice.
- [ ] T003 — Design and demonstrate safe browser resynchronization; inspect
  `backend/internal/api/routes.go`, WS event production, frontend WS module,
  `services/hubstate.service.ts`, Vuex device state and persisted initialization.
  IDs: FR-04, AC-10, FE-01/02/03, NH-04. Depends: none. Done: resolve Q-02 with
  reviewed `contracts/reconnect.md` (create only once a concrete design is selected),
  required fields/auth/errors/ordering/removal/old-client handling, rollout and a
  deterministic prototype/test proving no stale snapshot/delta rollback or membership
  loss/resurrection, with copied snapshot values and backend-restart fencing. If existing
  contracts cannot provide the fence, propose additive revisions in the ADR/spec.
  Do not lower queues or implement an unreviewed wire change during this spike.
- [ ] T004 — Recheck constitution matrix against resolved decisions and implementation
  scope; inspect any additional applicable AGENTS/rules. IDs: NH-01–06, BE-01–05,
  FE-01–05. Depends: T001; T003 for WS work. Done: supported design entries updated;
  remaining blockers explicitly scoped; no fabricated exceptions.

## Diagnostics and failure boundaries

- [ ] T005 — Add bounded diagnostics observer/snapshots and lane accounting in
  `backend/internal/lanes/`, controller ingress/routing and focused new tests.
  IDs: FR-08, AC-15, NFR-02. Depends: T002, T004. Done: queue-depth/oldest-wait/
  stage-duration/outcome counters; deterministic ingress-full, lane-full, panic,
  discard and unfinished accounting; bounded labels/no raw payloads; no callback
  blocking or remote-log recursion. Existing overflow capacities remain unchanged.
- [x] T006 — Add fault-injection coverage for swallowed store/broadcast errors and
  malformed bridge/device input in controller tests and `backend/mocks/`.
  IDs: FR-05, AC-11/12. Depends: T002. Done: reproduce discarded errors/unchecked
  assertions where feasible; fixtures assert no suppressed independent outputs and
  preserve accepted unknown-expose behavior. Enabled-automation/no-duplicate-command
  failure coverage is deferred to T026; completion here is scoped accordingly.
  Evidence: `backend/internal/controllers/handlers_test.go` and `hub-controller_internal_test.go`; pre-fix regressions
  reproduced lost errors, rename partial effects/panics and invalid JSON suppression.
- [x] T007 — Report output errors and replace duplicated bridge parsing with typed
  decoding/validation in `controllers/handlers.go`, callback wiring and, if needed,
  `models/devices/requests.go`. IDs: FR-05/09, AC-11/12/16. Depends: scoped design
  check in `evidence.md` validation section and T006; T005 counters are a separate follow-up.
  Done: T006 passes; successful callback order remains unchanged; errors include
  stage/cause; no event replay; golden traces equal baseline.
  Evidence: [T006/T007 evidence](evidence.md#t006t007-validation-and-output-errors). Uses existing error logging;
  AC-11 counters/publication completion remain pending T005/T008. T004's broader
  Q-01/02/03 decisions are not prerequisites for this compatible failure slice.
- [ ] T008 — Add bounded publication completion tracking in `internal/mqtt/` and
  narrow publisher adaptation in `internal/automations/action.go`, controller command
  wiring and fakes. IDs: FR-05, AC-03/11, BE-03/05. Depends: T001, T004, T005.
  Done: immediate error/completion error/timeout/late completion/tracker saturation/
  saturated Paho submission/shutdown tests; local Stop/Wait tests; no command retries or unbounded token goroutines; normal submission
  order and timing preserved; timeout reported as indeterminate; no false device ack.

## MQTT subscriptions

- [ ] T009 — Reproduce subscription registry races, wrong unsubscribe prefix,
  identical-announcement recovery failure and unbounded waits with fake Paho in
  `internal/mqtt/` and controller tests. IDs: FR-02, AC-04–06. Depends: T002.
  Done: tests exercise actual MQTT adapter callbacks with controlled token completion,
  concurrent reconnect/removal and late success; do not serialize away the race.
- [ ] T010 — Implement desired-topic reconciler and bounded operation/retry scheduling
  in `internal/mqtt/mqtt.go`, new `subscriptions.go`, and test fakes. IDs: FR-02,
  AC-04–06, NFR-04. Depends: T001, T004, T005, T009. Done: instance-owned normalized
  set, generation checks, short locks, coalesced worker wake, cancellable backoff,
  bounded startup (connection plus six mandatory bridge acknowledgements), late-result
  reconciliation, local Stop/Wait and race tests pass; no QoS changes.
- [ ] T011 — Integrate desired membership into bridge registration/rename/removal in
  `controllers/handlers.go` and `services/device_registrar.go`. IDs: FR-02/05,
  AC-04/05/12. Depends: T007, T010. Done: errors propagate, hash success does not
  suppress unfinished work, new-topic-before-old-removal behavior tested, fakes and
  every internal MqttClient call site updated; no new wire acknowledgement promise.

## Lifecycle and device consistency

- [ ] T012 — Add application lifecycle failure tests in `internal/`, `internal/api/`,
  `store/`, controller tests and `testing/hubharness/`. IDs: FR-03, AC-07/08,
  NFR-02/04. Depends: T001, T002, T004. Done: observable resource-close order,
  repeated shutdown, bind/start failure, partial construction, queued/running work,
  expired deadline and delayed-action commitment fixtures with barriers on both
  sides of the actual commitment gate; temporary storage only.
- [ ] T013 — Introduce runtime stop/join ownership in `main.go`, `internal/hub.go`,
  `internal/api/server.go`, controller lanes and store construction/cleanup. IDs:
  FR-03, AC-07/08. Depends: T005, T012. Done: main waits for shutdown; reliable
  server errors; stop-intake distinct from execution cancellation; ingress drains
  before device lanes; repositories remain open until users join; deadline failure
  explicitly reports remaining work and avoids concurrent close.
- [ ] T014 — Complete owned-worker retirement in services, automation engine/handlers/
  actions, MQTT trackers/reconciler, store tasks and configuration listeners. IDs:
  FR-03/06, AC-07/08/13, NFR-02. Depends: T008, T010, T013. Done: non-joining Stop
  plus outer Wait where needed; no self-join; no delayed rearm after cutoff; correct
  default scheduler context; remove/partial-startup/shutdown release resources;
  harness cancels and joins; rebaseline after harness changes; 100 lifecycle cycles
  return owned counts to baseline. New workers already require local ownership tests
  in T008/T010; this task integrates them rather than postponing cleanup.
- [ ] T015 — Add controlled timeout-vs-update, disable, remove, replacement and rename
  tests in lifetime/controller tests. IDs: FR-01/06, AC-02/03/13. Depends: T002.
  Done: reproduce stale offline decision; document effective config and alias lane
  behavior; detect separate-lane or metadata races before changing routing.
- [ ] T016 — Extract clock-driven availability monitor and submit coalesced generation
  checks through the device ordering boundary. Paths: `services/device_lifetime.go`,
  proposed `availability_monitor.go`, processor/registrar/controller wiring. IDs:
  FR-06/09, AC-13/16. Depends: T004, T005, T014, T015. Done: no stale offline
  transition, no timer callback blocking/full-queue growth, one-second check semantics,
  config/remove/rename cleanup and race tests pass. Amend routing contract first if
  stable-ID serialization requires more than compatible alias handling.
- [ ] T017 — Add seed/drain panic and pending-overflow fixtures; implement explicit
  ready/faulted/retired initialization states in `services/device_processor.go`.
  IDs: FR-07, AC-14, NFR-02. Depends: T001, T004, T005, T014. Done: deferred cleanup
  releases bounded pending work/monitor resources, counts failure, rejects further
  faulted-device work without replay, permits other devices to progress; reentrant
  config tests still pass; stable-ID fault latch survives enable/rename/remove-readd;
  source-owned uncommitted delayed actions/schedules are cancelled; healthy target-
  sharing automations continue; fault is not offline; reviewed policy documented.

## Slow browsers and recovery

- [ ] T018 — Implement the reviewed fresh-snapshot/reconciliation contract in existing
  backend transport and frontend services/Vuex modules. Add backend contract fixtures
  and Jest WS/store tests, adapting development mocks where used. IDs: FR-04,
  AC-10, FE-01/02/03/05. Depends: T003, T004, T005. Done: concurrent updates/removal,
  stale generations, failed refresh and bounded buffering pass; no command retries;
  authenticated helpers used; unready state visibly distinguished; existing clients
  remain compatible according to the addendum.
- [ ] T019 — Measure session allocation and slow-client behavior using local WS
  clients in `internal/ws/`; select queue cap and frame buffer settings. IDs: FR-04,
  AC-09, NFR-05. Depends: T002, T003. Done: before/after retained-memory methodology,
  normal burst/blocked reader matrix and selected capacity recorded; message counts
  distinguished from bytes. Keep production default unchanged until T018 passes.
- [ ] T020 — Apply bounded WS queue and idempotent slow-session disconnect in
  `internal/ws/websocket.go`; delete retired client-map entries and integrate local
  diagnostics. IDs: FR-04, AC-09/10, NFR-02/05. Depends: T005, T018, T019. Done:
  ≥90% empty-session retained-memory reduction, no normal-burst disconnect, healthy
  clients unaffected, no recursive logger overflow, 100 reconnect cleanup cycles,
  proven resync after forced overflow; compatible rollout documented.

## Cleanup and performance assessment

- [ ] T021 — Extract cohesive request wiring and small consumer interfaces; remove
  only touched dead code/TODOs. Paths: controller wiring, service constructor,
  bridge decoding and optional Update result helper. IDs: FR-09, AC-16, BE-01/02.
  Depends: T007, T011, T016, T017. Done: Seed/Update remain distinct, state-before-
  debounce and callback order explicit, no package cycles, unchanged golden outputs.
  Avoid combining mechanical moves with new timing behavior in one commit.
- [ ] T022 — Run benchmark/profile matrix and assess conditional improvements.
  Paths: controller benchmark, lanes, metrics, WS fixture. IDs: NFR-01/02/05,
  FR-01/08. Depends: T020, T021. Done: five-run comparison within +10%, slow-storage/
  slow-client/reconnect bursts with p50/p95/p99 and queue/rejection/memory evidence;
  hotspot conclusions recorded. Do not implement asynchronous sinks, batching, ring
  buffers or new timing policies without measured need and the required amendment.

## Handoff and merge evidence

- [ ] T023 — Run backend test/build/vet and affected race commands from `plan.md`;
  run frontend Jest/lint/build for the recovery slice. IDs: SC-01, NFR-03, NH-05.
  Depends: T022, T026. Done: exact commands, toolchain, passed/failed/unrun checks and causes
  recorded; required failures resolved; no household devices involved.
- [ ] T024 — Verify recovery and deployment compatibility using local mocks: late
  broker ACK, broker unavailable at startup, dirty/failed snapshot, overflowing WS,
  shutdown deadline and restart of faulted device. IDs: AC-04–14, FE-03/04/05.
  Depends: T023. Done: outcomes match contracts; client-before-queue rollout and
  rollback tested; keyboard/text/offline states verified at 390/1280 px if UI touched.
- [ ] T025 — Update `backend/docs/architecture.md`, configuration/readme references
  and this plan's actual-results matrix; recheck every constitutional rule and
  acceptance ID. IDs: NH-01/04/05/06, SC-01. Depends: T024. Done: no open blocking
  questions for shipped behavior, approved scope decisions recorded, review links,
  limitations and remaining optional performance work stated. Record only actual
  maintainer review; a completed test suite does not grant approval.

## Deferred behavioral coverage

- [ ] T026 — Complete controller save-without-execution characterization and enabled-
  automation output-failure coverage in controller tests/harness and automation
  scenarios. IDs: FR-01/05, AC-03/11, NH-02/05. Depends: T002, T007. Done: a real
  controller configuration/save operation produces no physical command; subsequent
  intended events still execute. Inject independent output failures with automation
  enabled and assert exactly the baseline command count/order, no replay and remaining
  outputs attempted. Use fake publishers/storage and explicit barriers, no household
  devices. Reuse existing scenarios where possible. This is a test-only follow-up
  candidate; passing disabled-automation failure tests does not establish these claims.

## Suggested review slices

1. Evidence and error boundaries: T002, scoped T006/T007 complete; T026 is the
   smallest recommended independent follow-up. T005 diagnostics follows scoped review.
2. MQTT recovery and publication observation: T008–T011.
3. Application/device lifecycle and initialization: T012–T017.
4. Browser recovery, then queue reduction: T003, T018–T020.
5. Mechanical cleanup and final measurement/evidence: T021–T025.

T001/T004 gate the dependent slices throughout. Independent design/test work can
proceed without waiting for every decision; dependencies above remain authoritative.
No agent delegation or parallel implementation is requested by this task list.
