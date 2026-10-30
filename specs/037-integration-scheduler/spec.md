# Spec: Integration refresh scheduler

ID: 037 | Status: Implemented; automated checks passed; browser check unrun | Owner: NodeHerder contributors | Updated: 2026-10-29
Constitution: proposed 1.0.0 | Components: backend and frontend

## Problem, scope, exclusions

**Problem.** Every provider request is admitted by a user action today. Weather becomes stale after
one hour and bin dates are only as current as the last manual Refresh. A household must remember to
open Integrations, and a browser timer cannot help because it stops when the tab closes and would
duplicate work across clients.

**Outcome.** Each saved integration can declare an automatic refresh cadence. Providers supply
sensible defaults (for example hourly weather and five-daily bin schedules), and the household can
choose another interval within server-enforced bounds. Due work runs from one bounded backend
coordinator, survives restart, and reuses the existing refresh execution and publication path.

**Evidence (observed in code, not assumptions).**

- `backend/internal/integrations/runtime/controls.go` `Refresh` is the only refresh admission path.
  It checks revision, enabled state, queue capacity, definition health, origin approval and the
  persisted provider deadline, then builds a durable operation through `makeOperation`.
- `backend/internal/integrations/runtime/service.go` owns four workers, a 64-slot queue, a 30-second
  attempt budget and a local-only janitor. Startup performs no provider request.
- `backend/internal/integrations/definitions/compiler.go` rejects unknown fields and applies bounded
  defaults to `Policy`; `backend/models/integrations/definition.go` currently exposes
  `poll_min_seconds`, `renewal_min_seconds` and `idempotency_window_seconds` only.
- `backend/models/integrations/state.go` has no per-instance schedule. `LastRenewal` controls POST
  renewal, not schedule refresh.
- `specs/032-external-integrations/contracts/runtime-lifecycle.md` defers one due-work coordinator to
  M2 and requires persisted timestamps, not sleeping goroutines. `specs/034-weather-integration/plan.md`
  and `specs/035-home-signal-integrations/spec.md` explicitly exclude scheduling from those deliveries.
- There is no cron/job-queue dependency in `backend/go.mod`; `internal/automations/scheduler.go` is an
  in-memory automation timer and is not a suitable integration scheduler.

**In scope.**

- A per-instance schedule aggregate (enabled flag, interval, due time, last attempt/success/failure)
  stored atomically with the existing encrypted instance record.
- Provider-definition defaults and advisory minimums, validated by the existing strict compiler and
  frontend parser.
- One runtime-owned due-work coordinator using the existing `utils.Clock.AfterFunc`/`utils.Timer`
  contract, worker pool and provider backoff.
- One shared admission path for manual and automatic refresh, including coalescing.
- An authenticated, revision-fenced schedule update endpoint and UI in the integration settings.
- Restart, shutdown, disable, credential/origin-change and failure behavior.
- Tests, operational documentation and rollback guidance.

**Exclusions.**

- Provider data-retention enforcement and the shared freshness/data-policy service recorded as debt in
  `specs/034-weather-integration/plan.md`. This spec makes no retention-compliance claim and leaves the
  existing WeatherAPI.com warning in place.
- Cron expressions, calendar-time ("every day at 08:00") schedules, time zones for schedules and
  per-user schedules. The requested examples are fixed intervals measured from the last attempt.
- Automatic discovery of new providers, new capabilities, MQTT/assistant integration, push
  notifications, or device/automation changes.
- Immediate refresh on enable, on schedule save or at startup. Enabling remains I/O-free per 034's
  "M1 enable does not start provider work" contract; the first automatic request occurs when due.

## Scenarios and acceptance

### US-01 — Configure a cadence per integration (P1)

Independent verification: `internal/api/integrations_test.go`, the schedule contract fixture and the
connection-settings browser check.

- AC-01: Given a saved instance, when the household opens Settings, then the schedule section shows
  whether automatic refresh is on, the chosen interval, the next due time, and the last
  attempt/success (or "never"), distinguishing automatic and manual triggers where available.
- AC-02: Given a provider definition declaring `refresh_default_seconds: 3600`, when a new instance is
  saved, then its schedule is enabled with a 3600-second interval and no additional provider request
  is sent during creation beyond the setup request already in progress.
- AC-03: Given the household selects manual-only or changes the interval, when the revision-fenced
  update succeeds, then the new values and a new next-due time are persisted and no provider request
  is sent by the update itself.
- AC-04: Given an interval below the effective minimum or above 30 days, when the update is submitted,
  then the server returns `invalid_schedule` with a field message and writes nothing.
- AC-05: Given an instance with automatic refresh off or the instance itself disabled, when time
  advances past any previously stored due time, then no provider request is admitted for it; a manual
  Refresh still works when the instance is enabled.
- AC-06: Given a legacy instance with no schedule record, when it is read, then it is reported as
  manual-only and remains readable; it is never silently enrolled in automatic refresh.

### US-02 — Due integrations refresh without a browser (P1)

Independent verification: `internal/integrations/runtime/scheduler_test.go` with `mocks.MockClock`.

- AC-07: Given an enabled instance with automatic refresh and `next_due_at` at or before the
  coordinator's current time, when the coordinator wakes, then exactly one refresh operation is
  admitted within 60 seconds and it uses the same request recipe, credential, origin and publication
  path as a manual Refresh.
- AC-08: Given a provider response of 304, when the scheduled refresh completes, then the stored
  content age and resource revision semantics are unchanged, the check time advances, and the next
  due time moves forward.
- AC-09: Given an operation is already queued or running for the instance (manual or automatic), when
  the coordinator wakes, then it does not admit a second operation and the existing work is joined.
- AC-10: Given more due instances than free workers, when the coordinator wakes, then admissions are
  bounded by the existing queue capacity and four-worker pool; no unbounded goroutine is created.
- AC-11: Given a successful scheduled refresh, when it publishes, then the instance snapshot is
  replaced atomically, a run summary is recorded, and `last_success_at` is updated.

### US-03 — Failure, backoff and restart (P1)

Independent verification: scheduler, coordinator and lifecycle tests with isolated storage and fake
HTTP providers.

- AC-12: Given a 429 or 5xx response carrying `Retry-After`, when the operation terminates, then the
  next automatic attempt is not earlier than `max(now + interval, retry_at, persisted provider
  deadline)` and the failure is visible in the instance status and last error.
- AC-13: Given a refresh or renewal operation is left `retry_required` or `pending` by a provider or
  by restart recovery, when its retry time arrives, then the coordinator resumes that same operation
  (same idempotency key/body where applicable) before its expiry rather than creating a second lookup.
- AC-14: Given the backend restarts with overdue automatic work, then the first coordinator wake
  admits or resumes enabled due work within 60 seconds and does not duplicate already-committed work.
- AC-15: Given shutdown begins, then the coordinator stops, cancels and joins in-flight callbacks,
  and the database is not closed under an active writer; no timer re-arms after close.
- AC-16: Given a credential change, origin review requirement, disabled definition or obsolete
  generation, when scheduled work reaches publication, then the existing fences reject it and the
  instance reports the safe failure instead of publishing stale or unauthorized data. A last error
  that requires user action (`reconnect_required`, `restart_setup`, `reconfigure_needed`,
  `unsupported_setup`, `unsupported_capability`, `operation_expired`, `definition_unavailable`)
  prevents further automatic admission until the household repairs or replaces the binding.

### US-04 — Provider configuration and compatibility (P2)

Independent verification: compiler tests, frontend parser tests and the existing Form/YAML editor
checks.

- AC-17: Given a provider definition with `refresh_default_seconds`/`refresh_min_seconds`, when it is
  validated through Form or YAML, then the same bounds are enforced in both editors and missing fields
  retain safe defaults.
- AC-18: Given the shipped weather and waste definitions are updated with defaults, when an existing
  host-edited definition lacks those fields, then it remains valid and manual-only unless the host
  explicitly adds them.
- AC-19: Given the schedule endpoint, when it is called without authentication, without the CSRF
  request header, with a stale `If-Match`, or with malformed fields, then it fails with the existing
  established codes and does not mutate the schedule.
- AC-20: Given an older binary reads a newer encrypted instance record, then it ignores the additive
  schedule fields; rollback guidance explains that a later write from the older binary drops the
  schedule and automatic refresh must be re-enabled after upgrading again.
- AC-21: Given the page remains open, when a later successful instance read returns a changed schedule
  and the editor has no unsaved schedule edits, then the editor reflects the confirmed values and the
  fresh ETag; with unsaved edits, the editor keeps them and the next save uses the fresh ETag.
- AC-22: Given an outstanding retry or pending continuation, when the household saves a new cadence,
  then the operation's existing retry/poll deadline is preserved until it completes, and the new
  cadence applies to subsequent refreshes; the edit cannot postpone the retry past its expiry.
- AC-23: Given a valid interval that is not a whole number of minutes, when the editor displays and
  resubmits it, then the exact integer seconds are preserved and the interval is not rounded.
- AC-24: Given an outstanding retry or pending continuation, when the instance or automatic schedule
  is disabled and re-enabled, then its retry/poll deadline is preserved and the operation resumes
  before expiry; a re-enable with no outstanding operation starts the new cadence in the future.
- AC-25: Given a later refresh is interrupted after an earlier terminal failure left the failure count
  positive, when the interrupted operation becomes retryable without a new provider failure, then the
  coordinator resumes it at its operation retry deadline rather than waiting for the regular interval.
- AC-26: Given automatic admission is blocked by a missing/invalid provider definition or a changed
  origin, when the coordinator examines the instance, then a sanitized failure is persisted and
  visible in settings; repeated scans do not admit provider work, and repairing the definition clears
  the failure so the overdue refresh can proceed.
- AC-27: Given the provider rejected the saved credential, when a same-origin definition is edited or
  its disk error is repaired, then the credential failure remains visible and automatic admission
  stays blocked; only a destination-review failure is cleared when the origin matches.
- AC-28: Given the household has unsaved schedule edits, when a later read changes the confirmed
  schedule, then saving uses the draft's base revision and returns a conflict instead of silently
  overwriting it; unrelated instance revisions may rebase the draft, and an explicit reload discards it.

## Requirements and success

- FR-01: A saved instance MUST support `enabled` and `interval_seconds` for automatic refresh;
  manual-only is represented by `enabled: false` (AC-01, AC-03, AC-05).
- FR-02: Provider definitions MUST support `refresh_default_seconds` (0 = manual-only) and
  `refresh_min_seconds` (0 = server floor). The strict compiler MUST default/validate them and the
  frontend parser MUST accept the same shapes (AC-02, AC-04, AC-17, AC-18).
- FR-03: `InstanceDTO` MUST expose the schedule and its due/last timestamps; legacy instances return
  `schedule: null` (AC-01, AC-06, AC-11).
- FR-04: One runtime-owned coordinator MUST scan persisted due times on a bounded wake interval and
  admit work only through the existing refresh admission/execution/publication path (AC-07, AC-10).
- FR-05: Automatic and manual refresh MUST coalesce per instance; no second operation may be created
  while one is queued or running (AC-09).
- FR-06: Admission MUST persist the next due time in the same transaction as the operation, and
  terminal publication MUST update last attempt/success/failure and the next due time atomically
  (AC-11, AC-12, AC-14).
- FR-07: The coordinator MUST honor the existing generation, credential, origin, definition-health,
  snapshot-presence and persisted provider-deadline checks; it MUST NOT bypass backoff. Automatic
  refresh only maintains an instance that has a validated snapshot; an instance that has never
  completed setup remains manual until it does (AC-12, AC-16).
- FR-08: A manual Refresh MUST reset the next due time so a manual action is not immediately repeated
  automatically (AC-03, AC-09).
- FR-09: Enabling a schedule, changing it, enabling an instance, saving a definition or restarting
  MUST NOT itself initiate provider I/O (AC-02, AC-03, AC-05, AC-14).
- FR-10: The schedule update MUST be an authenticated, CSRF-protected, `If-Match`-fenced mutation
  returning the updated `InstanceDTO` (AC-19).
- FR-11: The effective interval MUST be at least `max(60 seconds, definition refresh_min_seconds)` and
  at most 30 days; requested values outside that range are rejected, never silently clamped
  (AC-04).
- FR-12: Existing encrypted records, endpoints, resource contracts and manual refresh behavior MUST
  remain compatible; no repository bucket or schema-version change is introduced (AC-06, AC-20).
- FR-13: Every successful instance read MUST reconcile a pristine schedule editor with the fresh
  confirmed schedule and ETag while preserving unsaved schedule edits (AC-21).
- FR-14: Saving a cadence MUST preserve an outstanding operation's retry/poll deadline; the new
  interval governs only subsequent refreshes (AC-22).
- FR-15: The editor MUST represent an interval in exact integer seconds when it is not evenly
  divisible by a larger unit, and resubmit that exact value (AC-23).
- FR-16: Disabling and re-enabling an instance or automatic schedule MUST preserve an outstanding
  operation's retry/poll deadline; only a re-enable with no outstanding operation starts the new
  cadence in the future (AC-24).
- FR-17: Retry backoff MUST be persisted separately from the regular next-due time, cleared when a
  new attempt is admitted or a pending/successful attempt replaces it, and used instead of the
  regular cadence for continuation admission (AC-25).
- FR-18: Blocking admission failures MUST be persisted as sanitized instance failures, MUST suppress
  repeated automatic admission, and MUST be cleared when the blocking definition or origin condition
  is repaired (AC-26).
- FR-19: Definition repair MUST NOT clear a provider authentication failure; only a
  destination-review reconnect failure is cleared when the origin matches (AC-27).
- FR-20: An unsaved schedule draft MUST retain the base revision of its confirmed schedule; background
  reads may advance it only while the confirmed cadence is unchanged, and a changed cadence MUST
  produce a save conflict rather than a silent overwrite (AC-28).

- NFR-01 (lateness): Enabled due work MUST be admitted within 60 seconds of its persisted due time
  under normal operation; the coordinator wake is bounded to one minute.
- NFR-02 (bounds): The coordinator MUST own exactly one timer and no per-instance goroutine; at most
  50 instances are scanned per wake, and provider I/O remains bounded by the existing four workers,
  64-slot queue, 30-second attempt budget and per-instance single-operation rule.
- NFR-03 (restart): In-memory state MUST be reconstructible from persisted records; the worst-case
  lost work on abrupt shutdown is one in-flight attempt, which recovery marks for bounded retry.
- NFR-04 (jitter): Next-due times MUST include deterministic per-instance jitter of at most
  `min(300 seconds, interval/10)` so instances with the same interval do not stampede.
- NFR-05 (security): Scheduled requests MUST use only the already-approved origin, declared credential
  slot and existing destination/HTTP policy; schedules MUST NOT expose tokens, raw provider bodies or
  household inputs in logs, DTOs or run summaries.
- NFR-06 (observability): Each instance MUST report automatic state, effective next-due time, last
  attempt, last success, last failure, consecutive failures and whether the attempt was manual or
  scheduled, without requiring a provider request.
- NFR-07 (no read/UI work): GETs, startup loading, Home/board reads, the cleanup janitor and all
  frontend polling MUST continue to perform zero provider requests.

- SC-01: Scheduler, API, compiler and frontend parser suites pass, including boundary, restart,
  disable, 304, 429/backoff and shutdown cases.
- SC-02: A controlled-clock test advances past an hourly and a five-day due time and observes exactly
  one admission per instance, within the stated lateness bound.
- SC-03: Backend `go test ./...`, `go build ./...`, `go vet ./...` and affected race tests pass;
  frontend `npm test -- --runInBand`, `npm run lint` and `npm run build` pass.

## Entities and external interfaces

- **Refresh schedule (private aggregate):** nested in `models/integrations.Instance`; fields
  `enabled`, `interval_seconds`, `next_due_at`, `last_attempt_at`, `last_success_at`,
  `last_failure_at`, `consecutive_failures`. It is encrypted with the instance and deleted with it.
- **Provider policy:** additive `refresh_default_seconds` and `refresh_min_seconds` in
  `models/integrations.Policy` and the shareable YAML/Form grammar. `renewal_min_seconds` continues to
  govern POST renewal and is not overloaded as a refresh interval.
- **Schedule DTO:** additive `schedule` object on `InstanceDTO`; UTC RFC 3339 timestamps.
- **Coordinator:** runtime-owned background task; no HTTP/MQTT/MCP surface and no new dependency.
- **API:** `PUT /api/integrations/instances/{id}/schedule` with body
  `{"enabled": bool, "interval_seconds": int}` and `If-Match`; returns the existing instance envelope
  and ETag.
- **Consumers:** the integration detail settings update their presentation after a successful save;
  Home/board continue to consume the existing resource freshness fields. No resource contract changes.

## Compatibility and rollback

Adding JSON fields to the encrypted instance record is additive for this schema. An older binary
ignores unknown fields when reading, but will omit them when it rewrites the record, so rolling back
drops the schedule and automatic refresh must be re-enabled after upgrading again. No bucket, AAD or
schema-version change is made. Existing definitions without the new policy fields remain valid and
manual-only. The schedule endpoint and DTO field are additive; clients that do not understand them
continue to work with manual refresh.

## Questions and assumptions

- **Q-01 (product, owner maintainers):** Should new instances start with automatic refresh enabled
  when the provider declares a default? Assumption A-01: yes; existing instances remain manual-only
  until edited. If rejected, only the creation default changes, not the scheduler design.
- **Q-02 (product, owner maintainers):** Should existing instances be auto-enrolled? Assumption A-02:
  no, to avoid unsolicited provider traffic and quota use after upgrade.
- **Q-03 (constitutional, owner maintainers):** Retention/data-policy work recorded in 034 is split
  from this spec. Assumption A-03: acceptable because 037 makes no retention claim and keeps the
  WeatherAPI.com warning; a follow-up spec remains required before retention compliance is asserted.
- **Q-04 (security, owner maintainers):** Editable YAML `refresh_min_seconds` is advisory UX, not a
  trusted provider maximum. Assumption A-04: the runtime hard floor is 60 seconds plus persisted
  provider backoff; trusted provider constraints remain part of the 034 data-policy follow-up.
- **Q-05 (scope, owner maintainers):** All signed-in users share schedule controls, matching the
  existing shared integration controls. Assumption A-05: no per-user schedules or admin role is added.
- **Q-06 (scope, owner maintainers):** Fixed intervals are sufficient; no cron/calendar-time schedule.
  Assumption A-06: yes, based on the requested "every N days/hours" examples.
