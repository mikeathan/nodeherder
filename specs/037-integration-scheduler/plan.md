# Plan: Integration refresh scheduler

Spec: [spec.md](spec.md) | Status: Implemented; automated checks passed; browser check unrun | Author: NodeHerder contributors | Updated: 2026-10-29
Constitution: parent/backend/frontend 1.0.0 (proposed)

## Context and design

### Observed facts

- `runtime.Service` owns the worker pool, queue, provider deadlines and janitor
  (`backend/internal/integrations/runtime/service.go`). `Refresh` is the only refresh admission path
  (`runtime/controls.go`), and `worker`/`execute` never sleep on provider state (`runtime/coordinator.go`).
- `runtime.janitor` re-arms a one-minute local timer with the `Close`/`wg` pattern already required for
  background ownership (`runtime/janitor.go`). It must stay network-free.
- `definitions.Compiler` rejects unknown YAML/JSON fields and applies policy defaults
  (`definitions/compiler.go`), so any grammar addition must change both decoders deliberately.
- `Instance` is one encrypted aggregate (`models/integrations/state.go`); publication, failure,
  renewal and deletion already update it atomically with its operation.
- `utils.Clock` exposes `Now` and `AfterFunc`; `mocks.MockClock.Advance` fires callbacks synchronously,
  so scheduler tests can be deterministic without sleeps.
- `internal/automations/scheduler.go` is an in-memory wall-clock scheduler with no persisted due state,
  revision fencing or integration lifecycle; 032's runtime contract excludes a second scheduler.

### Selected design

**1. Persist a schedule on the instance aggregate, not in a new bucket.**

Add `RefreshSchedule` to `models/integrations/state.go` and
`Instance.Schedule *RefreshSchedule` (`json:"schedule,omitempty"`). This keeps due-time updates,
operation admission and deletion in the same encrypted transaction. No bucket, AAD or schema-version
change is needed; unknown fields are already ignored by `repository.DecodeIntegrationRecord`.

```go
type RefreshSchedule struct {
    Enabled             bool       `json:"enabled"`
    IntervalSeconds     int64      `json:"interval_seconds"`
    NextDueAt           time.Time  `json:"next_due_at"`
    BackoffUntil        *time.Time `json:"backoff_until,omitempty"`
    LastAttemptAt       *time.Time `json:"last_attempt_at,omitempty"`
    LastSuccessAt       *time.Time `json:"last_success_at,omitempty"`
    LastFailureAt       *time.Time `json:"last_failure_at,omitempty"`
    ConsecutiveFailures int        `json:"consecutive_failures"`
}
```

`BackoffUntil` is the only retry-backoff source for an existing operation; the regular `NextDueAt` is
never used as a proxy. A new refresh admission, a pending continuation or a successful publication
clears it, so an interrupted later attempt resumes at its own `RetryAt` rather than waiting a regular
interval left by an earlier failure.

Add `Trigger string` to `Operation` and `RunSummary` (`manual|scheduled`; renewal keeps the parent
trigger) so the UI and tests can distinguish automatic work without a second history store. Older
records decode with an empty trigger and are treated as manual/unknown.

**2. Extend the provider policy with defaults and an advisory floor.**

`models/integrations/definition.go` gains:

```go
RefreshDefaultSeconds int `json:"refresh_default_seconds" yaml:"refresh_default_seconds"`
RefreshMinSeconds     int `json:"refresh_min_seconds" yaml:"refresh_min_seconds"`
```

`definitions/compiler.go` defaults `refresh_min_seconds` to 60, accepts 60..2,592,000, accepts
`refresh_default_seconds` 0 (manual-only) or a value within the same bounds, and requires
`default >= min` when the default is non-zero. The shipped weather and waste seeds set hourly/5-day
defaults and sensible floors. The frontend parser and policy form receive the same fields.

The YAML value is configuration, not a trusted provider maximum (see 034's data-policy debt). The
runtime floor is the hard 60-second minimum plus persisted provider backoff; the definition minimum
only prevents the UI from offering a faster value.

**3. One runtime coordinator reuses admission.**

New `backend/internal/integrations/runtime/scheduler.go`:

- `const scheduleWake = time.Minute`, `minimumRefreshInterval = time.Minute`,
  `maximumRefreshInterval = 30 * 24 * time.Hour`, `maximumJitter = 5 * time.Minute`.
- `scheduleRefresh()` mirrors `scheduleCleanup`: it registers `wg.Add(1)` under `s.mu`, calls an
  unexported `runDue()` and re-arms only while the service is open.
- `runDue()` reads the bounded instance set under the service lock and considers an enabled automatic
  instance when its regular due time has arrived **or** it still owns an unfinished operation. The
  second condition prevents a pending renewal from expiring while the five-day interval is still in
  the future. For each candidate it re-reads the instance in its own write transaction:
  - if an unfinished refresh/renewal operation exists, resume it when `RetryAt <= now` (same
    operation/idempotency key) instead of admitting a second one;
  - otherwise, when `NextDueAt <= now`, reload the compiled definition and reuse
    `admitRefresh(tx, &instance, compiled, "scheduled")`;
  - enqueue accepted operation IDs only after their transaction commits.
  Instances whose last error requires user action (`reconnect_required`, `restart_setup`,
  `reconfigure_needed`, `unsupported_setup`, `unsupported_capability`, `operation_expired`,
  `definition_unavailable`, `not_found`, `invalid_definition`, `reconciliation_required`) are skipped
  until the household repairs them. Blocking definition/origin admission failures are persisted as
  sanitized instance failures; repairs clear the definition-level failure before the overdue refresh
  is admitted again. Origin-review failures carry `retry_action: review_destination`; definition
  repair clears only that reconnect variant, never a provider's rejected-credential failure. An instance with no validated snapshot is skipped as well; automatic refresh only
  maintains established data, while manual Refresh remains available.
- Per-instance failures are logged safely and do not roll back other instances. A due instance whose
  definition is unhealthy or whose admission fails keeps its due time so a repaired definition resumes
  on the next wake; the one-minute wake bounds the retry rate, and logs use stable codes without
  private values.
- `enqueue` must not block shutdown. Make the send context-aware (or select on `s.ctx.Done()`); a
  durably admitted operation that cannot be queued is left for restart recovery rather than wedging a
  coordinator callback or `Close`.
- The wake is a fixed one minute. It is simple, bounds the scan to 50 records and satisfies the
  60-second due-time target without computing timer deltas from timestamps.

Extract the admission body from `Refresh` into:

```go
func (s *Service) admitRefresh(tx *repository.IntegrationTx, i *model.Instance,
    c *definitions.Compiled, trigger string) (model.Operation, error)
```

`Refresh` keeps its HTTP-facing checks (revision, enabled, coalescing, definition health, origin,
provider deadline) and calls the shared helper. The scheduled path performs the same checks minus the
external revision check. Both set `Operation.Trigger`, `Schedule.LastAttemptAt` and the next due time in
the same transaction.

Extract the continuation body of `CheckOperation` into an internal
`admitContinuation(tx, op, owner)` helper. `CheckOperation` retains its external revision/expiry/backoff
checks; the scheduler calls the same helper after confirming the operation is a due refresh/renewal.
This keeps ambiguous POST replay, idempotency-key reuse, expiry and provider-deadline checks in one
place.

**4. Schedule outcome updates stay at terminal boundaries.**

A small helper in `scheduler.go` owns the arithmetic:

- admission: `NextDueAt = now + interval + jitter(instanceID, interval)`; `OperationID` is set;
- success (`finishSnapshot`, `finishUnchanged`): `LastAttemptAt/LastSuccessAt = now`,
  `ConsecutiveFailures = 0`, `NextDueAt = now + interval + jitter`;
- retryable failure (`failOperation` with `retry_required`): `LastAttemptAt/LastFailureAt = now`,
  `ConsecutiveFailures++`, `BackoffUntil = NextDueAt = max(retryAt, now + backoff)` so the coordinator
  resumes the same operation;
- terminal failure: `LastAttemptAt/LastFailureAt = now`, `ConsecutiveFailures++`,
  `BackoffUntil = nil`, `NextDueAt = max(now + interval + jitter, failure.RetryAt)`.

Both successful publication paths use one `finishInstanceSuccess` helper to apply schedule success and
advance the instance revision/update time once, avoiding duplicated finalization.

Jitter is deterministic: FNV-1a over the instance ID modulo `min(300s, interval/10)` plus one. This
prevents synchronised hourly polls after a restart without persisting random state.

**5. Enable, disable and schedule edits never start I/O.**

`UpdateInstance` already forbids provider work on enable. When an instance becomes enabled and has an
automatic schedule, set `NextDueAt = now + interval + jitter`. `UpdateSchedule` writes only the
schedule and increments the instance revision; it does not cancel a manual operation or admit work.
Legacy instances have `Schedule == nil` and remain manual-only.

**6. API and UI.**

- `runtime/dto.go`: additive `ScheduleDTO` and `InstanceDTO.Schedule *ScheduleDTO`.
- `internal/api/integrations.go`: new protected route
  `PUT /api/integrations/instances/:id/schedule`, input fields
  `schedule_enabled`/`schedule_interval_seconds`, action `schedule`; reuse envelope, CSRF, `If-Match`,
  ETag and error mapping. Add `invalid_schedule` handling if a field error is not already represented.
- `frontend/src/types/integrations.type.ts`: `ConnectionSchedule` and `ConnectionInstance.schedule`.
- `frontend/src/contracts/integrations.ts`: strict parser for the schedule and the new policy fields.
- `frontend/src/services/integrations.service.ts`: `updateSchedule(id, etag, enabled, seconds)`.
- `frontend/src/domain/schedule.ts`: pure conversion/formatting helpers
  (`toIntervalSeconds`, `splitInterval`, `formatScheduleInterval`, `scheduleEditor`,
  `matchesScheduleEditor`) so unit handling and editor reconciliation are not duplicated.
- `frontend/src/composables/integrations/useConnectionDetail.ts`: schedule refs, initialisation from
  the loaded instance, reconciliation after every successful read (focus, minute timer, save and
  observer reloads) while the editor still matches the last confirmed schedule, draft base-ETag
  tracking so another session's schedule change conflicts instead of being overwritten, and
  `saveSchedule()` using the existing `perform`/reload flow.
- `frontend/src/components/integrations/ConnectionSettings.vue`: an accessible "Automatic refresh"
  fieldset with enabled control, number + unit input, next due/last success/last failure text, and
  save. `frontend/src/components/integrations/ProviderPolicyFields.vue`: the two policy fields.
- `frontend/src/components/integrations/ConnectionDetail.vue`: pass the new bindings/emit.

**7. Docs and operations.**

Update `docs/connections.md` to replace "No provider work runs on a recurring timer" with the actual
schedule contract, defaults, manual opt-out, restart behavior, provider-quota warning and rollback
note. Keep the WeatherAPI.com retention warning and the 034 data-policy follow-up explicit.

## Alternatives

| Option | Rejected because |
| --- | --- |
| Reuse `automations.Scheduler` | In-memory only, wall-clock `HH:MM`/Europe-London assumptions, no revision/generation fencing, no persisted due state, and 032 forbids a second scheduler boundary. |
| Cron expressions / calendar-time schedules | Adds parsing, DST and timezone ambiguity for no requested outcome; the examples are fixed intervals. |
| Per-instance goroutine or timer | Up to 50 timers, harder cancellation/join, unnecessary resource use; one coordinator plus persisted timestamps is the documented M2 design. |
| Scheduler in the browser | Stops when the tab closes and duplicates work across clients; server-side provider work must be owned by the backend. |
| New `schedules` bucket or job queue | Splits atomicity from the instance/operation and adds migration/recovery work; the repository has no job-queue abstraction. |
| New `NextDueAt` on `Operation` only | Operations are terminal/expiring; a schedule must outlive individual attempts and survive restarts. |
| Immediate refresh when enabling/saving | Violates the existing "enable never starts I/O" contract and creates surprise quota use. |
| Store cron-style provider maximums in editable YAML | YAML is host-editable configuration, not a trusted provider limit; see 034's recorded data-policy debt. |

## Constitution Check

Design and automated execution evidence are recorded below. No maintainer approval or real-device
verification is claimed; the browser settings check remains unrun.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | This spec, plan, tasks and ADR; unresolved product assumptions labelled. | PASS (design) | PASS (reviewed) |
| NH-02 | No device, event, MQTT or automation path changes. | N/A | N/A |
| NH-03 | Only enabled instances with approved origins/credentials schedule; existing destination policy, private storage and sanitized errors are reused. No new endpoint exposure beyond the authenticated schedule mutation. | PASS (design) | PASS (API auth/CSRF tests) |
| NH-04 | Additive instance field, policy grammar, DTO and endpoint; shared admission boundary; ADR records persistence/coordinator alternatives. | PASS (design) | PASS (diff review + ADR) |
| NH-05 | Fake-clock scheduler tests, compiler/API/parser tests, full backend/frontend suites; no household device or live provider. | PASS (design) | PASS (automated checks) |
| NH-06 | One timer, bounded scan/queue/workers, persisted due state, restart recovery, provider backoff, visible failures, ordered shutdown, no retention claim. | PASS (design) | PASS (race/lifecycle tests) |
| BE-01 | Runtime owns the coordinator/repository; no provider I/O in handlers or edit DTOs; no new package cycle. | PASS (design) | PASS (reviewed) |
| BE-02 | Device lifetime unchanged. | N/A | N/A |
| BE-03 | Coordinator/timer ownership, cancellation, `wg` join and `Close` ordering implemented; `Clock.AfterFunc` not `Sleep`. | PASS (design) | PASS (race + stop tests) |
| BE-04 | Additive configuration defaults documented; no schema/bucket change; legacy record behavior explicit. | PASS (design) | PASS (compiler/default tests) |
| BE-05 | Schedule route requires existing auth/CSRF/ETag; provider errors remain sanitized; no assistant/MCP change. | PASS (design) | PASS (API tests) |
| FE-01 | Existing Vuex module, service and composable ownership; schedule logic lives in a pure domain helper. | PASS (design) | PASS (reviewed) |
| FE-02 | Parser mirrors the additive backend fields; TypeScript is not treated as runtime validation. | PASS (design) | PASS (parser tests) |
| FE-03 | Save is a requested action; the UI shows confirmed next/last times after reload; failures and stale/manual states are distinct. | PASS (design) | PASS (code review; browser check unrun) |
| FE-04 | Labelled controls, keyboard operation, visible focus, shared theme controls. | PASS (design) | PASS (code review; viewport check unrun) |
| FE-05 | No secrets or household inputs added to browser storage/logs; schedule DTO contains only timing metadata. | PASS (design) | PASS (reviewed) |

## Contracts, data, rollout

- **Provider YAML/Form:** additive `policy.refresh_default_seconds` and `policy.refresh_min_seconds`;
  strict unknown-field rejection remains. Existing files without them stay valid.
- **Instance record:** additive `schedule` JSON inside the existing encrypted record. No schema-version,
  bucket, AAD or keyring change. A record written by this version is readable by an older binary, but a
  later older-binary write drops the field; document the rollback consequence.
- **HTTP:** new authenticated `PUT /api/integrations/instances/:id/schedule`; existing routes, envelopes,
  error codes and ETags are unchanged. No MQTT, WS, MCP or metrics contract change.
- **Frontend:** additive `ConnectionInstance.schedule`; existing clients ignore it.
- **Rollout:** backend first (it tolerates legacy records), then frontend. No data migration. Provider
  seeds are copied only when the host file is missing; existing host-edited definitions remain
  untouched and manual-only until the household adds the policy fields. Rolling back is safe for
  data integrity but disables automatic scheduling; re-enable after upgrading again.
- **Recovery:** restart recovery continues to mark interrupted operations retryable; the coordinator
  resumes them when due. A database that cannot open remains degraded and starts no coordinator.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01–03, AC-01–06 | `models/integrations/state.go`, `runtime/dto.go`, `runtime/controls.go`, `contracts/integrations.ts` | Model/DTO serialization, legacy nil schedule, schedule update tests; parser tests | Passed — runtime schedule/update tests and contract parser tests |
| FR-02, FR-11, AC-02, AC-04, AC-17, AC-18 | `models/integrations/definition.go`, `definitions/compiler.go`, `configs/integrations/**`, `domain/schedule.ts`, `ProviderPolicyFields.vue` | Compiler boundary/default tables; Form vs YAML parity; frontend parser and duration tests | Passed — compiler defaults/bounds and frontend schedule tests |
| FR-04–08, AC-07–13 | `runtime/scheduler.go`, `runtime/controls.go`, `runtime/coordinator.go` | Fake-clock due/not-due, interval boundary, coalescing, scheduled 304, 429/5xx backoff, existing-operation backoff, retry resume, pending renewal before expiry, jitter range, exactly-one admission | Passed — scheduler tests incl. scheduled 304, 5xx backoff, pending renewal resume and jitter |
| FR-13, AC-21 | `frontend/src/domain/schedule.ts`, `frontend/src/composables/integrations/useConnectionDetail.ts` | Pristine-editor reconciliation on focus/minute reads; dirty editor preserved | Passed — composable tests (`useConnectionDetail.test.ts`) |
| FR-14, AC-22 | `runtime/controls.go`, `runtime/scheduler.go` | Cadence edit during an outstanding retry/poll; retry deadline preserved and operation resumes before expiry | Passed — `TestSchedulerHonorsScheduleBackoffForExistingOperation` |
| FR-16, AC-24 | `runtime/controls.go`, `runtime/scheduler.go` | Instance and schedule disable/re-enable during an outstanding retry; deadline preserved and operation resumes before expiry | Passed — `TestInstanceReenablePreservesRetryDeadline`, `TestScheduleReenablePreservesRetryDeadline` |
| FR-17, AC-25 | `runtime/scheduler.go`, `runtime/controls.go`, `models/integrations/state.go` | Interrupted refresh after an earlier terminal failure; retry deadline distinct from regular cadence | Passed — `TestInterruptedRefreshUsesRetryNotRegularCadence`, `TestInterruptedRefreshSurvivesInstanceReenable` |
| FR-18, AC-26 | `runtime/scheduler.go`, `runtime/definitions.go` | Blocked definition/origin persistence, repeat-scan suppression and repair clearing | Passed — `TestSchedulerPersistsBlockedDefinitionAndClearsOnRepair` |
| FR-19, AC-27 | `runtime/scheduler.go`, `runtime/definitions.go` | Definition repair distinguishes destination-review and rejected-credential reconnect failures | Passed — `TestSchedulerDefinitionFailureDistinguishesCredentialAndOrigin`, `TestDefinitionRepairDoesNotClearProviderAuthenticationFailure`, `TestDefinitionRepairClearsSchedulerOriginFailure` |
| FR-20, AC-28 | `frontend/src/domain/schedule.ts`, `frontend/src/composables/integrations/useConnectionDetail.ts` | Draft base-revision rebase for unrelated revisions; conflict for a changed cadence | Passed — composable base-ETag tests |
| FR-15, AC-23 | `frontend/src/domain/schedule.ts`, `ConnectionSettings.vue` | Non-minute integer seconds display and resubmit exactly | Passed — schedule domain tests (`61`, `90`, `3601` seconds) |
| FR-09, AC-03, AC-05, AC-14 | `runtime/controls.go`, `runtime/service.go`, `runtime/scheduler.go` | Enable/save performs no HTTP; restart within 60s; disabled/manual skip | Passed — no-I/O, disabled/manual and timer tests; restart reuses existing `recoverOperations` suite |
| FR-10, AC-19 | `runtime/controls.go`, `internal/api/integrations.go`, `services/integrations.service.ts` | Auth/CSRF/If-Match/malformed/interval-bound API tests; service test | Passed — API auth/input tests and runtime revision/bounds tests |
| NFR-02, AC-10 | `runtime/scheduler.go`, `runtime/coordinator.go` | Queue-capacity admission test; existing four-worker construction unchanged; no goroutine-count assertion | Passed — queue-capacity test; worker count not separately asserted |
| NFR-03, AC-15 | `runtime/scheduler.go`, `runtime/janitor.go`, `main.go` | Shutdown-during-wake test; second Close; no timer re-arm; DB stays open until workers join | Passed — timer stop/no-work-after-close and existing shutdown/race tests |
| NFR-05, AC-16 | `runtime/scheduler.go`, `runtime/controls.go`, `runtime/publication.go` | Credential/origin/generation fence tests reuse existing fixtures; no secret in DTO/logs | Passed — existing fence suites plus scheduler skip/error tests |
| NFR-06, AC-01, AC-11 | `runtime/dto.go`, `ConnectionSettings.vue`, `ConnectionDetail.vue` | DTO field test; scheduled success/failure updates; browser settings check | Passed (automated) — browser settings check unrun |
| SC-01–03 | backend and frontend trees | `go test ./...`, `go build ./...`, `go vet ./...`, affected `go test -race`; `npm test -- --runInBand`, `npm run lint`, `npm run build` | Passed — see actual runs below |

Selected component checks, from `backend/CONSTITUTION.md` and `frontend/CONSTITUTION.md`:

```sh
# backend/
go test ./internal/integrations/... ./internal/api/... ./models/integrations/...
go test -race ./internal/integrations/runtime/...
go test ./...
go build ./...
go vet ./...

# frontend/
npm test -- --runInBand
npm run lint
npm run build
```

Actual runs (2026-10-29):

- Backend `go test ./...`: passed.
- Backend `go build ./...`: passed.
- Backend `go vet ./...`: passed.
- Backend `go test -race ./internal/integrations/runtime/...`: passed.
- Frontend `npm test -- --runInBand`: 92 suites / 522 tests passed.
- Frontend `npm run lint`: 0 errors, 28 pre-existing warnings.
- Frontend `npm run build`: passed (Vue type-check and Vite build).
- Browser settings/keyboard/viewport check: not run. No household device, live provider account or
  real scheduler wake was used as evidence.

## Debt, risks, exceptions

- **Recorded 034 debt:** provider retention and the shared freshness/data-policy service remain
  outstanding. This plan must not claim retention compliance; WeatherAPI.com stays operationally
  warned. A follow-up spec is required before asserting those limits.
- **Editable YAML is not a trusted maximum:** `refresh_min_seconds` is advisory. The runtime floor is
  60 seconds plus persisted provider backoff. Trusted provider constraints remain part of the data-policy
  follow-up.
- **New-instance default:** A-01 enables automatic refresh for new instances when the provider declares
  a default. This is a product decision for maintainers; it can be changed without redesign.
- **Existing instances:** A-02 leaves them manual-only after upgrade. No unsolicited provider traffic.
- **Rollback:** an older binary will drop the additive schedule field on its next write. Document and
  require re-enabling after re-upgrade.
- **Upgrade of in-flight retries:** records written before `backoff_until` existed have no separate
  deadline. A retryable operation from before the upgrade therefore resumes at its operation
  `RetryAt` rather than inferring a backoff from the ambiguous regular due time. New failures persist
  the distinct backoff.
- **Lateness:** fixed one-minute wake means up to 60 seconds of lateness; acceptable per NFR-01. No
  exact-time guarantee is offered.
- **Provider quota:** automatic polling can consume provider limits. Provider backoff, 429 handling,
  jitter, four-worker bound and visible last-failure state mitigate misuse; users can choose manual.
- **No automatic recovery from configuration errors:** a definition/credential/origin block skips
  scheduling and surfaces the existing error; it does not silently repair or borrow another binding.
- No exception is requested. Unapproved NH-03/NH-04/NH-06 violations would block merge.

## Completion

Implementation and automated verification are complete. The browser settings/keyboard/viewport
check remains unrun, and the product assumptions A-01/A-02 still need maintainer confirmation.
Live provider quota behavior remains an operational check. Documentation: [spec](spec.md),
[tasks](tasks.md), [ADR-001](adr-001-runtime-refresh-scheduler.md),
[schedule contract](contracts/schedule.md), `docs/connections.md`.
