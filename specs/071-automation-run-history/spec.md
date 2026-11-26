# Spec: Automation run history and duplication

ID: 071 | Status: Draft | Owner: engineering | Updated: 2026-11-26
Constitution: 1.0.0 (proposed) | Components: both

## Problem, scope, exclusions

Automations currently fail silently unless the failure happens to reach the log buffer. The UI can
start a manual run but cannot answer “what ran, when, and did it work?”. Separately, “Duplicate
automation” is offered in design discussions but the engine has no identity that permits a second
automation for the same device.

Evidence: `internal/automations/device.go` `Evaluate`/`EvaluateTrigger` execute triggers and only log
errors; no run record is retained. `internal/automations/engine.go` `HandleDevice` loads the
automation by `device.Id`, so one device maps to one automation. The frontend `Automation.id` is used
both as the automation key and the device key (`useAutomationDraft` falls back to `findDevice(id)`).

**In scope (future):**

- Bounded execution history per automation: timestamp, trigger, outcome, failure reason, manual vs
  device origin.
- Read and clear history from the automations UI, including live updates for an open history view.
- A decision and, only if approved, an engine identity change that enables duplicate automation.

**Excluded:** automation-editor redesign, long-term audit retention, cross-hub aggregation, and
duplicating a device itself. Duplicate is **not** implementable under the current identity model;
the plan records options and gates it behind an ADR.

## Scenarios and acceptance

### US-01 — See what an automation did (P1)

Independent verification: Go recorder tests; browser history dialog.

- AC-01: Given an automation runs successfully, when the user opens History, then one record shows
  time, trigger name, `device` origin and success.
- AC-02: Given a trigger fails, then the record shows failure and the bounded error text, without
  exposing secrets or raw payloads.
- AC-03: Given a manual run, then the record is marked `manual` and remains distinguishable from a
  device-triggered run.
- AC-04: Given no runs, then an explicit empty state explains that history is in-memory and when it
  resets.

### US-02 — History stays bounded and clearable (P1)

- AC-05: Given more than the retention limit, then only the newest N records remain per automation.
- AC-06: Given the user clears history, then records disappear and the action is confirmed; clearing
  does not disable the automation.
- AC-07: Given a restart, then history is gone (unless the persistence option is selected in the
  plan) and the UI says so rather than implying durability.

### US-03 — Live history for an open view (P2)

- AC-08: Given the history dialog is open and the automation runs, then a new record appears without
  a page reload and without stealing focus.
- AC-09: Given the dialog is closed, then no per-row subscription or timer remains.

### US-04 — Duplicate automation (decision-gated)

- AC-10: Given an approved identity model, when the user duplicates an automation, then a new
  disabled copy exists with a distinct id, the same triggers/schedules and a name suffix; the
  original is unchanged.
- AC-11: Given the unapproved current model, then no Duplicate action is shown (no dead control).

## Requirements and success

- FR-01: Record every trigger execution at the engine boundary with outcome and origin.
- FR-02: Bound retention per automation (proposed 100 records) and expose clear.
- FR-03: Expose history through an authenticated read/clear contract and live event.
- FR-04: UI history dialog with empty, loading, error and live states.
- FR-05: Duplicate remains absent until identity/routing is approved.
- NFR-01: No unbounded memory growth; no per-automation goroutines/timers.
- NFR-02: Recording must not change execution ordering or swallow errors (NH-02/spec 068).
- NFR-03: History read must not block device event processing.
- SC-01: A failed run is visible from the automation UI within 1 s of completion; memory stays
  bounded under sustained runs.

## Entities and external interfaces

`AutomationRun` (proposed): `id`, `automationId`, `triggerName`, `at` (RFC3339), `origin`
(`device`/`manual`), `outcome` (`success`/`failure`), `error?` (bounded, sanitized).

Transport options are decided in the plan: WS request/event vs authenticated HTTP snapshot. The
record is not device state and must never feed automation evaluation.

## Questions and assumptions

- Q1: In-memory vs persisted history. Consequence: durability vs disk growth/privacy. Recommended:
  in-memory first, documented reset; persistence needs its own retention/privacy decision.
- Q2: Transport. Consequence: live updates and auth path. Recommended: WS request + event, matching
  existing automation load/save patterns.
- Q3: Duplicate identity. Consequence: engine routing, storage filenames, scheduler and frontend
  drafts. Requires ADR; do not implement without it.
- Q4: Retention 100/automation vs global 1,000. Consequence: memory bounds and UX. Recommended:
  100 per automation with a global cap.
