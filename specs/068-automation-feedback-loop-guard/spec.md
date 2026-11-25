# Spec: Automation feedback-loop guard

ID: 068 | Status: Implemented (unreviewed) | Owner: mikeathan | Updated: 2026-11-25
Constitution: 1.0.0 | Components: backend

## Problem, scope, exclusions

Turning a light off while someone is present and the room is dark makes the presence
recipe turn it back on within about one second. Evidence:

- `backend/internal/automations/trigger.go` (`BaseTrigger.AffectedBy`) selects a trigger
  when either its primary property or any declared condition dependency changes.
- `backend/configs/automations/0xa4c13894070052fc.json` defines trigger `presence` with
  conditions `presence == true AND illuminance <= 45` and action `state: ON`.
- The committed dial recipe `backend/configs/automations/0x001788010d7d9d3f.json` maps
  `button_1_press_release` to `state: TOGGLE` on the same Living Room Light;
  `backend/internal/automations/scenarios/dial_test.go` covers it. Turning the light off
  darkens the room; the next illuminance report from the presence sensor changes only
  `illuminance`, which selects the presence trigger and publishes `state: ON`. Reproduced
  with the presence recipe through `DeviceLifetimeService.Update` in
  `backend/internal/automations/device_test.go`.

The existing "feedback protection" (`ExposeCondition.HasValueChanged`) cannot stop this:
it reads `DeviceContext.currentData`, which is only written from action outputs keyed by
the action target, so cross-device condition inputs are always reported as "changed".

In scope: separate *selection for evaluation/cancellation* from *selection for execution*.
A condition-dependency-only change may evaluate conditions and stop pending actions, but
must not start actions. Out of scope: a manual-override latch (future Option B, deferred);
recipe thresholds; UI; persistence; MQTT/WS contracts.

## Scenarios and acceptance

### US-01 — A manual light-off is not undone by its own darkening (P1)

Independent verification: table-driven tests with the real presence recipe driven through
the production acceptance/deduplication path.

- AC-01: Given presence is false, when presence becomes true while illuminance <= 45,
  then exactly one `state: ON` is published.
- AC-02: Given presence is true and a light was turned off, when an update changes only
  `illuminance`, then no `state: ON` is published.
- AC-03: Given a pending delayed action, when a condition dependency changes such that a
  condition fails, then the pending action is cancelled and no new action starts.
- AC-04: Given a manual trigger, when it is evaluated, then its actions still execute
  regardless of primary/dependency selection.
- AC-05: Given repeated physical events, when identical events arrive, then each still
  executes; a cached event source is never replayed because a dependency changed.
- AC-06: Given an update that changes only a dependency, when conditions still pass, then
  no action starts (evaluation only).
- AC-07: Given the dial, front-door and other recipes, when their primary property or
  events change, then their behaviour is unchanged.

## Requirements and success

- FR-01: Execution MUST require the trigger's primary property in the changed set, or a
  manual trigger; acceptance: AC-01, AC-02, AC-04, AC-07.
- FR-02: Dependency-only selection MUST still run condition evaluation (through the first
  condition that fails or reports no change) and MUST retain cancellation of pending
  actions on a failing condition; acceptance: AC-03, AC-06. `conditionsPermitExecution`
  returns early on `HasValueChanged == false`, so later conditions are not evaluated on
  that path; this is pre-existing behaviour, not redesigned here.
- FR-03: Event sources MUST never be replayed because another property changed;
  acceptance: AC-05.
- NFR-01: MUST NOT add mutable selection state, caches, timers, goroutines or persistence;
  selection stays a bounded per-report, O(1) check.
- NFR-02: MUST NOT change external contracts (MQTT topics/payloads, WS, HTTP, config,
  storage) or require migration.
- SC-01: Before-fix regression fails and after-fix acceptance passes; `go test ./...`,
  `go build ./...`, `go vet ./...` pass; no new race in `internal/automations`.

## Entities and external interfaces

`Trigger.AffectedBy` remains the selection authority; a new internal (unexported) execution
path distinguishes dependency-only updates. `AffectedBy` reads the primary name through
`Trigger.GetName()`, matching `Device.Evaluate`, and `conditionsPermitExecution` uses the
same accessor; selection semantics are unchanged. `DeviceTrigger` is the only production
trigger type. No public interface, route, topic or persisted schema changes. Unknown trigger
extensions are not executed on dependency-only updates (safe default).

## Questions and assumptions

- Q-01 (resolved by owner): Option A is chosen for now — auto-on when the room darkens while
  someone is already present is intentionally dropped; entering a dark room still works.
- Q-02 (resolved by owner): Option B (manual-override latch) is deferred; it would require
  its own spec and a spec-010 amendment.
- Assumption: household rules from spec 010 AC-01/AC-02 are superseded for dependency-only
  execution by this spec; spec 010's cancellation and dedup obligations remain.
- Limitation (accepted): `DeviceLifetimeService.Seed` forwards a device's whole first
  snapshot, so on lane creation/restart its primary keys read as changed and primary
  triggers may run once. That is pre-existing restart behaviour and not the live loop
  this spec fixes; it is documented in `backend/docs/architecture.md` rather than changed
  here to avoid altering seed/output ordering.
