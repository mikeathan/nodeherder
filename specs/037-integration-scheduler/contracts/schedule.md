# Contract: integration refresh schedule

Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md) | Status: Draft

## Provider grammar

Additive fields in the existing `policy` object of `nodeherder.integration/v1`:

```yaml
policy:
  poll_min_seconds: 5
  renewal_min_seconds: 86400
  idempotency_window_seconds: 0
  refresh_default_seconds: 3600   # 0 = manual-only
  refresh_min_seconds: 60         # advisory UI/API floor; 0 = server default
```

| Field | Meaning | Validation |
| --- | --- | --- |
| `refresh_default_seconds` | Initial interval for a new instance. `0` keeps the instance manual-only. | Integer; `0` or 60..2,592,000; when non-zero it must be at least the effective minimum. |
| `refresh_min_seconds` | Minimum interval offered by the editor/API for this provider. | Integer; `0` becomes 60; otherwise 60..2,592,000. |

`renewal_min_seconds` continues to govern bounded POST renewal. It is not a refresh interval and the
scheduler must not reinterpret it.

The existing strict decoders reject unknown fields. Missing new fields are valid and decode to zero;
the compiler then applies the defaults above. Form and YAML editing produce the same validation.
Shipped seeds may declare defaults, but an existing host-edited file is never overwritten.

The YAML minimum is configuration, not an immutable provider maximum. The runtime always enforces a
hard 60-second floor and persisted provider backoff. Trusted provider constraints belong to the
separate data-policy work recorded in `specs/034-weather-integration/plan.md`.

## Private instance record

Additive field on the encrypted `Instance` aggregate:

```json
{
  "schedule": {
    "enabled": true,
    "interval_seconds": 432000,
    "next_due_at": "2026-11-03T13:00:00Z",
    "backoff_until": null,
    "last_attempt_at": "2026-10-29T13:00:00Z",
    "last_success_at": "2026-10-29T13:00:02Z",
    "last_failure_at": null,
    "consecutive_failures": 0
  }
}
```

- `enabled` controls automatic admission; the instance's existing `enabled` field still controls the
  whole integration and manual refresh.
- `interval_seconds` is the household value; it is always within the effective bounds when written.
- `next_due_at` is persisted in UTC and is the coordinator's regular scheduling input.
- `backoff_until` is a retryable failure's retry deadline, persisted separately from the regular
  cadence. It is cleared when a new attempt is admitted, a pending continuation is saved, or an
  attempt succeeds; terminal failures also clear it. It is private and is not exposed in the public
  DTO.
- `last_*` timestamps and `consecutive_failures` are observability, not authorization.
- A legacy record has no `schedule` and is treated as manual-only. It is never auto-enrolled.
- The schedule is encrypted with the instance and deleted with it.
- `Operation.Trigger` and `RunSummary.Trigger` are additive values `manual` or `scheduled`; renewal
  work inherits the triggering operation's value. Missing values are treated as manual/unknown.

## Public DTO

`InstanceDTO` gains:

```json
{
  "schedule": {
    "enabled": true,
    "interval_seconds": 432000,
    "next_due_at": "2026-11-03T13:00:00Z",
    "last_attempt_at": "2026-10-29T13:00:00Z",
    "last_success_at": "2026-10-29T13:00:02Z",
    "last_failure_at": null,
    "consecutive_failures": 0
  }
}
```

`schedule` is `null` for a legacy/manual-only record with no stored schedule. When automatic refresh is
off, `next_due_at` is `null` and the stored interval is retained for a later re-enable. All timestamps
are RFC 3339 UTC. No provider address, credential, body or household input is included.

## HTTP

`PUT /api/integrations/instances/{id}/schedule`

Request (the existing envelope fields are unchanged):

```json
{
  "schedule_enabled": true,
  "schedule_interval_seconds": 432000
}
```

- `If-Match` is required and must equal the current instance ETag.
- Existing authentication, cookie-origin and `X-NodeHerder-Request: 1` checks apply.
- Enabling (`schedule_enabled: true`) requires a compilable, healthy provider definition so the
  interval can be checked. Disabling is allowed while the definition is missing, invalid or disk-
  unhealthy, so the household can always turn automatic work off.
- Response: the existing instance envelope with the updated `InstanceDTO` and `ETag`.
- The request performs no provider I/O.

Errors use the existing envelope and codes:

| Condition | Code | HTTP |
| --- | --- | --- |
| Missing/stale `If-Match` | `precondition_required` / `revision_conflict` | 428 / 409 |
| Unknown field or malformed body | `invalid_input` | 400 |
| Enabled with interval outside bounds, or interval 0 while enabled | `invalid_schedule` | 400 |
| Instance not found | `not_found` | 404 |
| Storage degraded/unavailable | existing 503 codes | 503 |

## Scheduling semantics

1. **Admission.** The coordinator wakes at most once per minute and scans at most 50 instances.
   For each enabled automatic instance, it re-reads the record in its own write transaction and:
   - re-examines any outstanding refresh/renewal operation even when the regular next-due time is
     still in the future, resuming it when `retry_at <= now` (same operation ID, idempotency key and
     pinned definition), or
   - admits a new refresh when no unfinished operation exists and `next_due_at <= now`.
   A pending continuation brings `next_due_at` forward to its retry time so it is checked before the
   operation's 24-hour expiry. For `retry_required`, `backoff_until` is the only retry-backoff source;
   a regular `next_due_at` is never treated as a provider backoff.
   Instances whose `last_error.code` requires user action (`reconnect_required`, `restart_setup`,
   `reconfigure_needed`, `unsupported_setup`, `unsupported_capability`, `operation_expired`,
   `definition_unavailable`, `not_found`, `invalid_definition`, `reconciliation_required`) are
   skipped until the household repairs or replaces the binding. Blocking definition/origin admission
   failures are persisted as sanitized instance failures; a successful definition activation clears
   definition-level failures before the overdue refresh is admitted again. A destination-review
   `reconnect_required` failure carries `retry_action: review_destination`; definition repair clears
   only that variant, never a provider's rejected-credential failure. An instance with no validated
   snapshot is also skipped; automatic refresh only maintains established data, while manual Refresh
   remains available.
2. **Coalescing.** A queued or running operation, manual or scheduled, prevents a second admission for
   that instance. A manual Refresh shares the same path and resets the next due time.
3. **Due calculation.** Admission and terminal publication set
   `next_due_at = now + interval_seconds + jitter`. Jitter is deterministic from the instance ID and is
   at most `min(300s, interval_seconds / 10)`.
4. **Success.** `last_attempt_at`, `last_success_at` are set, `consecutive_failures` becomes 0, and the
   next due time advances. A 304 keeps the stored content age/revision semantics and only advances the
   check time, exactly as manual Refresh does.
5. **Retryable failure.** `last_attempt_at`/`last_failure_at` are set, the counter increments, and both
   `backoff_until` and `next_due_at` are set to the later of the exponential backoff and any provider
   retry deadline. The next wake resumes the same operation at `max(retry_at, backoff_until)`. A
   pending lookup response clears `backoff_until` and uses its poll deadline.
6. **Terminal failure.** The counter increments, `backoff_until` is cleared, and the next attempt is not
   before `max(now + interval + jitter, failure.retry_at)`. The failure remains visible on the instance.
7. **Enable/edit/startup.** Enabling an instance or saving a schedule sets a future due time and starts
   no provider request. While an outstanding refresh/renewal operation exists, a cadence edit keeps
   that operation's existing retry/poll deadline; the new interval applies to subsequent refreshes.
   Disabling and re-enabling an instance or the automatic schedule also preserves that deadline; only
   a re-enable with no outstanding operation starts the new cadence in the future.
   Startup starts the coordinator but does not refresh immediately. Restart recovery marks interrupted
   operations retryable; the coordinator resumes them when due.
8. **Disable/delete.** Disabling the instance or the schedule prevents admission. Deleting the instance
   removes the schedule with it. In-flight work is cancelled/fenced by the existing lifecycle.
9. **Bounds.** Four runtime workers, a 64-slot queue, a 30-second attempt budget, one operation per
   instance, at most 50 instances and no per-instance goroutine/timer.
10. **No read/UI work.** GETs, catalog/board/Home reads, the cleanup janitor and all browser polling
    perform zero provider requests.

## Security and compatibility

- Scheduled requests reuse the instance's approved origin, declared credential slot, destination
  policy, HTTP client limits, generation/credential fences and sanitized failure mapping.
- Schedules are shared installation state, matching existing shared integration controls; no per-user
  schedule or admin role is introduced.
- The additive JSON fields require no bucket, AAD, keyring or schema-version change. An older binary
  ignores them when reading but omits them when rewriting the record, so rollback disables automatic
  refresh until re-enabled after re-upgrading.
- The endpoint and DTO field are additive; clients that do not understand them continue to use manual
  Refresh.
