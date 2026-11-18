# ADR-002: Observation storage and retention

Status: accepted (maintainer decision 2026-11-11) | Date: 2026-11-11
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: maintainer decision recorded |
Rules: NH-03, NH-06, BE-04, FE-05

## Context and options

Observations are high-cardinality and high-volume: a single BLE proxy reports tens to
hundreds of advertisements per second, and consumer phones re-key their address roughly
every fifteen minutes. Three storage models were considered.

1. **`metrics.db` per identity.** Create an entity per observed identity and let the
   existing time-series machinery store it. Rejected: `AggregateExposeValue` assumes
   known expose names, `DefaultMetricsCleanupTask` prunes as entity history, and every
   rotating address becomes a new series. It would also put third-party identifiers into
   the device registry and out through `/api/context/devices` to MCP — an NH-03 problem.
   This mirrors the rejection of add-on device projection in
   `docs/external-integrations-investigation.md` §1.3.
2. **Settings only, no live buffer.** Persist every observation into `settings.db`.
   Rejected: write amplification with no user value; the viewer needs the last few
   seconds, not a durable log.
3. **Three-tier split (chosen).** Operator decisions in `settings.db`; live observations
   in a bounded in-memory ring; optional capture to a dedicated BoltDB file when an
   operator explicitly wants offline replay.

## Decision and consequences

Choose **option 3**.

| Data | Store | Retention |
| --- | --- | --- |
| Identity decisions (unknown/known/ignored, label) | `settings.db` via `AppConfigCache` | indefinite; user intent |
| Live observations, aggregation, render buffer | in-memory bounded ring, drop-oldest | process lifetime |
| Optional capture/replay | dedicated BoltDB under `GetDataDir()` | 7 days with `Prune` |

Retention for the optional capture file is **7 days**, matching
`spatialhistory.Retention`, pruned by the same `Prune` discipline. Unbounded capture was
rejected: an operator-forgotten capture would grow without limit and would accumulate
third-party identifiers, which NH-03 and NH-06 both discourage.

Consequences:

- No history of individual sightings unless capture is enabled. Accepted: the feature is
  a live instrument, not an archive.
- Triage survives restart; the live view does not, which matches operator expectation.
- A settings read failure fails closed: all identities present as unknown rather than
  inventing decisions.
- Storage follows existing precedents — `DeviceConfigCache.HasOverride`
  (`store/store.go:157`) for decisions, `spatialhistory.NewStoreAt`/`Prune` for capture.
- Bound enforcement is mandatory (NH-06): bounded queue, observable drop counter,
  defined retention, shutdown cancels workers.

Verification: settings round-trip with existing-data fixtures; bounded-growth test under
sustained load; `Prune` test with a fake clock.

## Supersession

None.
