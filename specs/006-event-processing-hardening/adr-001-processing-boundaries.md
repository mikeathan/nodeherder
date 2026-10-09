# ADR-001: Preserve ordered processing and make resource ownership explicit

Status: Proposed | Date: 2026-10-08
Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)
Author: Codex | Maintainer review: pending | Rules: NH-02/04/06, BE-01/02/03, FE-02/03

## Context and options

The existing synchronous device lane guarantees that an automation evaluates its
own event against freshly applied device state. Decoupling every stage with queues
would create new ordering, durability, retry and snapshot contracts. The service is
currently functioning and no measured production bottleneck justifies that change.

1. Keep the implementation unchanged: preserves behavior but retains observed
   reconnect, shutdown, error-reporting and buffer-size gaps.
2. Keep FIFO processing and add small ownership boundaries: targeted fixes, direct
   failure tests, no new broker or persistence dependency. Selected proposal.
3. General event bus/actor framework or durable queue: can provide broader delivery
   guarantees but adds operations and changes timing/replay behavior. Rejected here.
4. Async metrics/UI stages now: may isolate slow I/O, but changes sampled timestamps,
   ordering and recovery. Deferred until measured evidence and a separate decision.

## Decision and consequences

Retain ingress → keyed device processing → synchronous automation → sampled metrics
and UI effects. Availability decisions join the same device ordering boundary.
The stable device identity must not accidentally create a second lane during rename;
test existing friendly-name routing and resolve aliases consistently before changing
key selection. Preserve source event order while metadata is reconciled.

Introduce only focused boundaries:

- MQTT subscription reconciler owns desired membership, retry scheduling and
  connection generations. No mutex is held across a broker operation or callback.
- Application runtime owns startup/stop/join ordering; lower-level Stop remains
  non-joining where callbacks may reenter, with Wait invoked by the outer owner.
- Device monitor owns its timer and submits generation-tagged timeout candidates.
- Processing diagnostics records stage outcomes, with bounded labels and no payloads.
- Consumer-defined interfaces distinguish command publication, sampled metrics and
  UI delivery without putting transport types into domain rules.

There is no retry of an entire event or uncertain physical command. Subscription
retry is safe reconciliation of desired membership. Initialization panic is proposed
to fault the device rather than reseed it and repeat side effects; Q-03 gates this.

Shutdown first freezes intake and independent timer commitment, then drains accepted
work while dependencies remain usable. Work drained after cutoff may perform its
existing immediate effects, but cannot arm delayed work that would outlive shutdown.
Join active work before closing its resources; a deadline reports unfinished work
and lets the process exit without closing resources concurrently with those tasks.
Go goroutines cannot be forcibly killed safely. A successful graceful shutdown is
distinct from a deadline exit.

Slow-browser overflow is explicit disconnection plus safe state refresh, not silent
coalescing of event payloads. Queue reduction depends on a proven reconnect contract
(Q-02); no claim is made that current browser reconnection already provides it.

No storage migration, new dependency or configuration service is proposed. No new
endpoint is selected: Q-02 may require an authenticated recovery read/request or
additive metadata, which must be reviewed in a contract addendum before implementation.
Detailed proposed policies are in [runtime decisions](plan.md#runtime-decisions-awaiting-review). Costs are a reconciler worker, bounded diagnostics, explicit lifecycle
methods and additional fault tests. Reconsider async sinks only if slow-I/O fixtures
and production-safe measurements show a material bottleneck.

## Supersession

Not superseded. Maintainer acceptance has not been recorded. This ADR does not grant
an exception to proposed constitutions or authorize implementation.
