# ADR-001: Persistent evidence summaries for classification

Status: proposed | Date: 2026-11-12
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: none recorded |
Rules: NH-03, NH-06, BE-04. Amends
[064 ADR-002](../064-wireless-observations/adr-002-observation-storage.md).

## Context and options

Classification needs evidence spanning days: "present 24/7", "comes and goes", "never strong".
Spec 064 keeps a one-minute memory window, so every identity looks alike and no class is
possible.

064 ADR-002 rejected persisting raw observations for three reasons: volume, cardinality, and
privacy. Those reasons remain valid. The question is whether they also exclude *aggregates*.

Options:

1. **Keep memory-only.** No storage change, but classification degrades to a one-minute
   window, which cannot distinguish a fridge from a phone. Rejected: it fails the purpose.
2. **Persist raw observations.** Works, but at ~40 identities and tens of packets each per
   minute this is hundreds of thousands of rows a day, and it puts third-party identifiers
   and payload contents at rest. Rejected on all three of 064's grounds.
3. **Persist daily aggregates (chosen).** One row per identity per day, fields limited to
   counters and statistics. Keeps every property classification needs, at ~120 bytes per
   row.

## Decision and consequences

Choose **option 3**, gated by a write floor: a day is stored only when the identity was heard
in at least two distinct hours and with at least twenty packets. Noise earns no row, so the
maintainer's requirement — store data, provided it does not fill the database with noise — is
enforced at the write boundary rather than by review.

Consequences:

- 256 identities × 30 days ≈ 900 KB. Bounded and asserted by a test (NFR-01).
- No raw observations, packet contents or payloads at rest. A summary says how often, how
  strong and how mobile — not what was transmitted. This is what keeps retention defensible
  under NH-03.
- Storage follows the `internal/spatialhistory` shape exactly: own bbolt file, own `Prune`,
  worker-serialised writes, bounded queue that drops rather than blocking the pipeline.
- **The class is never stored.** Only evidence is. Changing the rules therefore needs no
  migration and no reclassification job, which is the main reason classification stays
  computed.
- This is a deliberate amendment to 064 ADR-002, recorded rather than diverged silently.

Verification: floor and retention tests with fixture data; a row-count ceiling test; a test
proving no file is created when there are no observations.

## Supersession

Amends 064 ADR-002's "no persistence" conclusion for aggregates only. That ADR's rejection of
raw-observation and metrics-database storage stands unchanged.
