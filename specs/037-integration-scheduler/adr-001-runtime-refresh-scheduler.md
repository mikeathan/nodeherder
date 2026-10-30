# ADR 001: Runtime-owned persisted refresh scheduler

Status: Proposed | Date: 2026-10-29 | Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)

## Context

Integrations currently perform provider I/O only after an explicit Refresh or setup step. Weather
predictably becomes stale after one hour, and bin schedules need periodic retrieval. The household
should not depend on a browser being open, and multiple clients must not each schedule the same work.

The runtime already owns admission, four workers, a bounded queue, generation/credential/origin
fencing, persisted provider backoff and an ordered shutdown. `specs/032-external-integrations` defers
one due-work coordinator to M2 and requires persisted timestamps rather than sleeping goroutines.
`specs/034-weather-integration` and `specs/035-home-signal-integrations` exclude scheduling.

## Decision

Add one runtime-owned coordinator and persist a per-instance refresh schedule inside the existing
encrypted instance aggregate.

- Providers declare a default interval and an advisory minimum in the shareable definition policy.
- Each instance stores `enabled`, `interval_seconds`, `next_due_at`, last attempt/success/failure and
  consecutive failures.
- A single one-minute timer scans at most 50 instances and admits due work through the same
  `admitRefresh` path used by manual Refresh.
- Due timestamps and outcomes are written in the same repository transaction as operations and
  publications. A retryable operation is resumed by ID rather than re-created.
- A deterministic per-instance jitter of at most `min(300s, interval/10)` spreads work.
- Manual Refresh resets the next due time; enabling or editing a schedule performs no provider I/O.
- The schedule is exposed through the existing authenticated instance update surface and detail UI;
  no new storage bucket, schema version, dependency or transport is introduced.

## Alternatives considered

| Alternative | Why rejected |
| --- | --- |
| Reuse `internal/automations/scheduler.go` | In-memory jobs, wall-clock/Europe-London semantics, no persisted due state, no integration revision/generation fencing, and a second scheduler contradicts 032. |
| Per-instance goroutine/timer | Up to 50 timers, harder shutdown/join, more failure modes; a single scan is bounded and sufficient. |
| Cron expressions or calendar times | Unneeded parsing/timezone/DST complexity; the requested behavior is "every N hours/days". |
| New `schedules` bucket or generic job queue | Breaks atomicity with instance admission/deletion and adds migration/recovery work for no current need. |
| Browser-side timer | Stops when closed and duplicates across clients; provider work belongs to the backend. |
| Separate scheduler service/process | A second process would need the same encrypted DB, keyring and lifecycle ownership; the runtime is already the correct owner. |

## Consequences

- The coordinator must be included in `Close` cancellation and wait-group ownership; a callback must
  not re-arm after close. Queue insertion must not block shutdown: a durably admitted operation that
  cannot be queued is recovered on the next start.
- A fixed one-minute wake bounds lateness at 60 seconds. It deliberately trades exact timing for a
  simple, auditable scan.
- The schedule is additive JSON. An older binary ignores it but drops it on a later write, so rollback
  disables automatic refresh until the household re-enables it after re-upgrading.
- Editable YAML minimums are not a trusted provider maximum. The hard 60-second floor and persisted
  provider backoff protect the runtime; trusted provider retention/limits remain the 034 data-policy
  follow-up.
- Existing instances remain manual-only; new instances follow the provider default (product assumption
  A-01, subject to maintainer confirmation).
- No retention enforcement is delivered here and no retention-compliance claim is made.

## Compliance

NH-01 (spec), NH-03 (existing auth/origin/credential boundary), NH-04 (ADR and additive contracts),
NH-05 (fake-clock tests and recorded checks), NH-06 (bounded work, recovery, visible failures, ordered
shutdown) and BE-03 (timer/goroutine ownership) are the controlling rules. See the plan's Constitution
Check for evidence and pending checks.
