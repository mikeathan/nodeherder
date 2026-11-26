# Plan: Automation run history and duplication

Spec: [spec.md](spec.md) | Status: Draft | Author: engineering | Updated: 2026-11-26
Constitutions: parent 1.0.0; backend 1.0.0; frontend 1.0.0

## Context and current behaviour

- Execution boundary: `internal/automations/device.go` `Device.Evaluate` (device events) and
  `EvaluateTrigger` (manual runs). Failures are logged; successful executions leave no trace.
- Identity boundary: `AutomationEngine.HandleDevice` loads `device.Id`; automations are stored under
  that key (`configs/automations`). A second automation for one device cannot receive events.
- Frontend: `AutomationList.vue` exposes Run, enable and Delete; `useAutomations` wraps save/delete.
  `Automation.id` is also the device id in the editor fallback.
- Existing logs are a diagnostic stream, not an execution record: successes are absent and the
  buffer is shared/bounded independently.

## Design: run history

### Option A (recommended) — bounded in-memory recorder

- Add a `RunRecorder` owned by `AutomationEngine`, protected by a mutex/ring per automation and a
  global cap (100 per automation, 1,000 total).
- Record at the same points that execute triggers:
  - `Device.Evaluate` primary-trigger branch: success or joined error.
  - `EvaluateTrigger`: manual, success or error.
  - Do not record condition-only dependency evaluations that did not start actions.
- Record after execution returns; never delay or reorder execution. Truncate/sanitize error text
  (e.g. 500 chars) and never include device payloads or credentials.
- History is lost on restart; empty state and dialog copy say “since hub start”.

### Option B — persisted history

- Same recorder plus append-only per-automation file or bbolt bucket, with retention and secure
  deletion. Adds disk growth, privacy and migration concerns; not recommended for the first slice.

## Transport and UI

- Extend the existing WS automation channel rather than add a REST surface:
  - client request `loadAutomationHistory` with `{automationId}`;
  - server response `automationHistory` with bounded records;
  - server event `automationRun` after a record is added.
- Only send history to clients that requested it (or keep it bounded in the existing automation
  payload). The plan recommends request/response plus event, matching `loadAutomations`.
- Frontend `AutomationHistoryDialog`:
  - opened from a per-row History action (not the primary row click);
  - columns: time, trigger, origin, outcome, reason;
  - clear action with confirmation;
  - explicit empty/loading/error; no polling; unsubscribe on close/unmount.

## Design: duplicate (decision required)

### Why it does not work today

`HandleDevice` resolves the automation by `device.Id`; storages are keyed the same way. A copied
record with a new id would never evaluate. Duplicate is therefore an engine identity change, not a
button.

### Options

1. **Do not add duplicate.** Keep one automation per device; remove it from design expectations.
   Lowest risk, no migration.
2. **Add `deviceId` to the automation model and route by an index.** Existing records migrate with
   `deviceId = id`; new ids are generated. Requires updates to storage keys, scheduler/manual run
   lookups, editor draft identity, and spec 068’s feedback-loop guard review. Recommended only with
   an ADR and full migration/recovery plan.
3. **Duplicate only triggers inside one automation.** Not offered: trigger names are property keys
   and duplicates would be ambiguous.

No Duplicate UI ships until option 1 or 2 is explicitly approved.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 071 + this plan, decisions unresolved | PASS (artifacts) | PENDING |
| NH-02 | Recording is observational; execution order and payloads unchanged | PASS (design) | PENDING |
| NH-03 | History contains bounded error text, no payloads/secrets; WS auth reused | PASS (design) | PENDING |
| NH-04 | One recorder in engine; transport through existing WS channel | PASS (design) | PENDING |
| NH-05 | Recorder unit tests + browser history spec + failure/retention cases | PENDING | PENDING |
| NH-06 | Bounded memory, clear action, documented restart loss | PASS (design) | PENDING |
| BE-02 | No change to device update semantics or trigger selection | PASS (design) | PENDING |
| BE-03 | Mutex-owned recorder; no goroutines/timers | PASS (design) | PENDING |
| FE-01/03 | Dialog owns presentation; history state in composable/service | PASS (design) | PENDING |
| FE-04 | Keyboard dialog, focus, non-colour outcome labels, mobile sheet | PENDING | PENDING |

## Contracts/data/rollout

- WS add: `loadAutomationHistory` request; `automationHistory` and `automationRun` messages.
- No persistence change under Option A; restart clears history by design.
- Duplicate (if approved): new `deviceId` field plus migration of existing `id` values; storage
  filename compatibility; rollback must keep reading old records.
- No MQTT or device-command contract change.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01/AC-01–03 | `internal/automations/run_history.go`, `device.go`, `engine.go` | unit tests for success, failure, manual, truncation | NOT STARTED |
| FR-02/AC-05/06 | recorder + `automationHistory` clear | cap test, clear test | NOT STARTED |
| FR-03/AC-08 | `internal/ws/eventhub.go`, controller wiring | WS contract test | NOT STARTED |
| FR-04/AC-01–09 | `AutomationList.vue`, `AutomationHistoryDialog.vue`, `useAutomationHistory.ts` | browser spec desktop/mobile/open-close | NOT STARTED |
| FR-05/AC-10/11 | duplicate decision | ADR / no dead control | DECISION PENDING |
| NFR-01/02 | recorder | race + bounded-memory test | NOT STARTED |

## Debt, risks, exceptions

- In-memory history disappears on restart; this is deliberate and must be stated in UI. Persistence
  would require its own retention/privacy decision.
- Error text could contain provider detail; sanitize and cap before storage.
- The recorder must not become a second source of automation truth or alter spec 068 selection.
- Duplicate is blocked on architecture; implementing it without the identity change would create a
  silently dead automation, which is worse than no action.
- No exceptions requested; no maintainer approval claimed.

## Completion

This plan is intentionally planning-only. Implementation starts only after Q1/Q2 and (for duplicate)
Q3 are decided. Update this file with actual results when code begins.
