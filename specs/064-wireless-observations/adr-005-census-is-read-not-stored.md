# ADR-005: The observation census is resolved on read, not stored

Status: proposed | Date: 2026-11-18
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: none recorded |
Rules: BE-01, BE-04, FE-01, FE-03. Amends [ADR-004](adr-004-local-provider.md).

## Context

ADR-004 made wireless observation an ordinary integration backed by a local provider, and
gave it a stored resource: `publishLocal` normalises the hub's live snapshot into an
`observation_census/v1` payload and writes it to the instance record. `Refresh`,
`AddLocal`, re-enable and the scheduler are its only producers, and the scheduler
deliberately skips local providers because "its value is the hub's current state, so
refreshing it on a timer would republish the same thing".

That reasoning is wrong in one direction and incomplete in the other:

1. **The value is not static.** The hub's live state changes constantly while a radio is
   being read, so a timer would not republish the same thing.
2. **With no timer, nothing republishes at all.** The census is captured once and then
   frozen until the operator re-enables the integration.

Observed on real hardware (`specs/064-wireless-observations/plan.md` §11): the Results
page showed `0 devices heard` beside `571 heard · 571 decoded`, while the Home tile
simultaneously read `54 devices from 1 source · 39 new`. Both surfaces were correct about
different moments, and the stored one could never catch up.

Two mechanisms produced that specific pair of numbers:

- `observations.Service.ObserveAdvertisements` consults the `Active` gate and returns
  early when no enabled instance exists, but `RecordSourceActivity` — which owns the
  `571` counters — is called unconditionally from the ESPHome session. A sensor that was
  enrolled and streaming before the integration was added therefore accumulated counters
  while every batch was dropped before it could be aggregated.
- The census was captured during `AddLocal`, inside that window, so it recorded zero
  identities. Nothing refreshed it afterwards: `Refresh` is hidden in the UI for a
  continuous provider and the scheduler skips local instances.

Options:

1. **Give local providers a refresh schedule.** Rejected by ADR-004 on grounds that
   still hold: a cadence is an arbitrary answer to "how often is this current?" for a
   value that is current by construction.
2. **Republish on every observation snapshot.** Correct data, but `publishLocal` bumps
   `Instance.Revision`, and `Revision` is the ETag used for revision fencing on rename,
   enable and credentials. Publishing at the observation cadence (up to 4 Hz) would turn
   every settings write into a `revision_conflict` race and write a small record to disk
   several times a second for a view nobody may be looking at. Decoupling "config
   revision" from "snapshot freshness" is a model change with wider blast radius than the
   bug.
3. **Resolve the census when it is read, and store nothing.** Chosen.

## Decision and consequences

A local provider's resource is produced by `instanceDTO` at read time, through the same
`Normalize` call the stored path used. `publishLocal` stops writing a snapshot; adding,
enabling and refreshing a local instance now only clear or record the transient source
error.

Why this is the honest shape:

- **A local provider does not own its value.** For every authored provider the resource is
  the result of an operation the integration performed and can repeat; storing it is the
  point. For a local provider the "operation" is a read of state the hub already holds.
  Persisting it stores a copy whose only possible relationship to the truth is "older".
- **No staleness is possible, so no freshness policy is needed.** This removes the reason
  ADR-004 had to say "no schedule", and removes the `unsupported_schedule` rejection's
  awkward counterpart on the client.
- **Fix by construction, not by patching.** The reported defect was one instance of "two
  surfaces read different moments". With one source, the board row, the integration page
  and the Home tile cannot disagree, because they resolve the same snapshot.
- **Revision fencing is left alone.** `Revision` keeps meaning "what the operator owns",
  which is what the ETag was always for.

Costs and risks, stated rather than assumed:

- **The read path does work.** Listing instances resolves the census once per local
  instance (at most one — `rejectDuplicateLocal` refuses a second). `Snapshot()` is a
  bounded copy under the observation mutex, which is what the websocket path already does
  on every publish.
- **A lock-order constraint is introduced.** `Service.mu` (integrations) is now held while
  acquiring the observation mutex. The reverse must never happen: `Service.isActive`
  acquires `Service.mu` and is called from `ObserveAdvertisements`, which holds no
  observation lock, so the order is currently safe. That safety is an invariant, not an
  accident, and is recorded in the code beside both call sites.
- **A source failure is no longer durable.** It is reported on the DTO as a transient
  `last_error`, which is where an operator acts on it. It cannot survive a restart, which
  is correct: "the sensor is unreachable" is a live condition, not saved state.
- **The stored snapshot field becomes dead for local instances.** `Instance.Snapshot`
  stays for authored providers. An instance record written by an earlier build of this
  branch still contains one; `instanceDTO` ignores it for local providers, so no migration
  is needed.
