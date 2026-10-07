# Tasks: Ordered device message processing

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: in progress (implementation authorised 2026-10-07)

Each task maps to requirement/acceptance IDs or a gate. Mark done only with evidence.
Backend commands run from `backend/`. No routine test may command household devices.

## Preparation

- [x] T001 — Resolve Q-02/Q-03/Q-04 in `spec.md`; IDs: all; done: recorded 2026-10-07.
- [x] T002 — Baseline: run `go test ./...`, `go vet ./...`, `go test -race ./internal/...` on the
  untouched tree and record results; gate: NH-05; done: output saved in plan. Evidence: plan "Recorded checks" (all exit 0).
- [x] T003 — Baseline benchmark of the current path (1,000-message mixed burst across 20 devices:
  p95 latency, allocs/op, idle goroutine count); IDs: NFR-01, NFR-02; done: numbers in plan. Evidence: plan T003 table (p95 4.04–4.56 ms, ~46.1–46.7k allocs/op, 25 idle goroutines).

## Coverage first (must fail on current code where a defect exists)

- [x] T004 — Replay harness with fake Zigbee2MQTT (answers `/set` with stale-then-new pair) driving
  the real MQTT callback → ingestion → processor → broadcast; 2,000 randomised trials, both
  directions; IDs: AC-01, AC-02, SC-01; done: fails on current code, output recorded. Evidence: FAIL before (1804/2000, 436/500), PASS after; plan.
- [x] T005 — Golden dial/button tests from the real automation JSON (`configs/automations/`):
  slow/fast left/right, min/max clamp, TOGGLE, preset cycling, identical repeated events; baseline
  first; IDs: AC-04–AC-07; done: baseline recorded, burst test (alternating directions) fails on
  current code. Evidence: golden PASS before/after; burst FAIL before (44–49 commands/40 events), PASS after. Traces are synthetic from the real JSON, not live captures.
- [x] T006 — Manual-trigger isolation test (unconditional trigger, concurrent device messages);
  IDs: AC-08, AC-09; done: fails on current code. Evidence: FAIL before (0 vs 1 manual run; 2 vs 1), PASS after.
- [x] T007 — Fake-clock delayed-action tests with the real presence JSON (cancel, no restart on
  repeat, illuminance gate); IDs: AC-11, AC-12; done: pass on current code (characterisation).
  Requires T012 clock injection. Evidence: 4 tests PASS with fake clock (`internal/automations/scenarios/presence_test.go`).

## Implementation

- [x] T008 — `internal/lanes` executor: lazy per-key lanes, `Submit`/`Do`/`Shutdown`, cap 1,000 with
  reject+log+count, per-task recover, depth/wait logging; IDs: FR-01, FR-02, FR-07, AC-03, AC-14,
  NFR-02, NFR-04; depends: T002; done: unit tests incl. order, isolation, overflow, panic, shutdown,
  `-race`. Evidence: `internal/lanes` tests, `-race -count=20` PASS.
- [x] T009 — MQTT: `SetOrderMatters(true)`, callback enqueues onto the ordered ingress lane only;
  IDs: FR-06, AC-16; depends: T008; done: blocking-handler test passes. Evidence: `TestBridgeSubscribeDoesNotStallOrderedDelivery` PASS; fails with synchronous ingestion (negative control).
- [x] T010 — `HubController.Ingest` single entry (MQTT + HTTP `Enqueue`), handler map owned by the
  ingress lane, route per device topic, `bridge/*` share one lane, replace `utils.WorkerPool` use
  (remove the pool only if unused elsewhere); IDs: FR-01, FR-06, AC-01, AC-15; depends: T008, T009;
  done: T004 passes. Evidence: T004 PASS; HTTP ordering test PASS; WorkerPool removed.
- [x] T011 — Per-run `runContext` decorator; remove `SetManualTrigger`/`SetDevicePayload` from
  `AutomationContext`; update conditions/actions/mocks/tests; IDs: FR-03, AC-08, AC-09; depends:
  T006; done: T006 passes, no behavioural diff in T005 baselines. Evidence: T006 PASS, golden unchanged.
- [x] T012 — Inject `utils.Clock` into delayed actions (behaviour-preserving); IDs: AC-11; done:
  existing tests green, T007 deterministic. Evidence: existing automation tests PASS; T007 deterministic.
- [x] T013 — Route `TriggerManual` through the device lane with request deadline; `ErrLaneBusy` → 503,
  existing 412 cases unchanged; IDs: FR-04, AC-10; depends: T008, T011; done: error-mapping table test. Evidence: hub busy-lane test and route mapping table PASS.

## Handoff

- [x] T014 — Full checks: `go test ./...`, `go build ./...`, `go vet ./...`, `go test -race` on
  `internal/lanes`, `internal/mqtt`, `internal/controllers`, `internal/services`,
  `internal/automations`, `internal/api`, `utils`; benchmark vs T003 (≤ +10%); gate: NH-05; done:
  actual results recorded in plan. Evidence: plan "T014 after change" (all exit 0; NFR-01 met).
- [ ] T015 — Live manual checks on approved targets only (Q-04): ten UI toggles each way on the
  Attic room Light; dial rotation, button 1/2 and presence on the Living Room Light, with a read-only
  Zigbee2MQTT feed capture; IDs: SC-02, Q-01 confirmation; done: results recorded.
- [ ] T016 — Docs and final rule check: update `backend/docs/architecture.md`,
  `docs/sdd/repository-map.md`; plan Constitution Check "before merge" column; gates NH-01/04/05;
  done: evidence recorded, nothing marked passed without a run. Progress: both docs updated and
  "before merge" column filled with evidence; left open until review and T015 (NH-05 live part).
