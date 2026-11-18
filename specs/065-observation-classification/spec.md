# Spec: Behavioural classification for wireless observations

ID: 065 | Status: Draft | Owner: agent draft for maintainer review | Updated: 2026-11-12
Constitution: 1.0.0 (Proposed) | Components: both (backend + frontend)
Supersedes: nothing. Amends [spec 064](../064-wireless-observations/spec.md) scope and
[ADR-002](../064-wireless-observations/adr-002-observation-storage.md) persistence.

This spec states outcomes. It does not claim ratification, reviewer approval, or
implementation authorization.

## Problem, scope, exclusions

### Problem

The viewer works and is accurate, and it is not useful. Used against real data it shows
roughly forty unlabelled identities: the household's own televisions and bulbs, phones whose
addresses rotate every fifteen minutes, and neighbours' devices. The operator cannot tell
which is which, and there is no pattern to find because nothing distinguishes a fridge from
a person.

The measured facts behind that:

- A phone's address rotates by design (Resolvable Private Address). Without pairing there is
  no way to name it. Confirmed against the [ESPresense enrolment
  documentation](https://espresense.com/guides/enrolling-devices/), which states that
  devices with rotating identifiers and no extraction path are hard, and that Find-My tags
  cannot be distinguished from one another.
- Random identifiers inflate the list: one phone appears as dozens of unrelated keys.
- Most identities never move. They carry presence information only in aggregate, which the
  current one-minute window cannot express.

### Outcome

Identities are grouped into four behaviour classes — **infrastructure**, **personal**,
**neighbour**, **transient** — with a stated confidence and the evidence behind it. The
default view hides infrastructure and probable neighbours, leaving the small set of things
that move. Classes are recomputed continuously from recent evidence, so a device that
changes behaviour is reclassified rather than mislabelled forever.

Identification by name remains out of reach without device cooperation. This spec does not
attempt it; it makes the data legible without it.

### Affected users and data

Operator on an authenticated browser session. Data: per-identity daily summaries (aggregates
only) and class transitions. No raw observations are persisted.

### Exclusions

- No IRK enrolment, no pairing, no app on the phone. Stated constraint from the maintainer.
- No multilateration or floor-plan position; that needs three or more proxies and a separate
  spec.
- No machine-learned model. There is no labelled training data, and obtaining it would
  require the device registration this spec excludes.
- No change to spec 064's identity resolution, triage semantics, or transport seam.

### Findings that shaped the design

- Presence evidence must span days, but raw observations cannot be stored (ADR-002). Daily
  summaries satisfy both: classification needs aggregates, not packets.
- A fixed feature such as "seen at 04:00" is evidence, not a verdict. The maintainer
  explicitly rejected treating it as a label, because a device that later moves must be able
  to change class.
- Rotating addresses mean classification must also operate at the level of an advertisement
  family, or phones can never be recognised as personal at all.

## Scenarios and acceptance

### US-01 — See what matters, not everything (priority P1)

Independent verification: classifier unit tests with synthetic evidence; a browser check with
fixture data.

- AC-01: Given an identity present in 22+ of the last 24 hours with no band transitions, when
  the viewer loads, then it is classed infrastructure and hidden by default.
- AC-02: Given an identity appearing intermittently with multiple band transitions, then it
  is classed personal and shown by default.
- AC-03: Given an identity that is only ever heard weakly relative to the strongest seen,
  then it is classed neighbour and hidden by default.
- AC-04: Given an identity seen for under an hour in total, then it is classed transient and
  hidden by default.
- AC-05: Given fewer than 24 hours of evidence for any identity, then no class is asserted;
  the row shows a learning state with hours observed.

### US-02 — Classes follow behaviour (priority P1)

- AC-06: Given an identity classed infrastructure that begins moving, when its recent
  evidence outweighs its older evidence, then the class changes to personal.
- AC-07: Given a class change, then a transition is recorded with time and both classes, and
  the row states the previous class.
- AC-08: Given a single day of contrary evidence against a long history, then the class does
  not change; recency is weighted but a change must hold across evaluations.
- AC-17: Given a sustained behaviour change, then the class changes in roughly two days of
  consistent new evidence, not sooner. This is the measured consequence of the 2-day
  half-life and is stated so the delay is a decision, not a surprise.
- AC-09: Given an operator triage decision, then that decision overrides the class for
  display permanently, and the class is still shown as advisory.

### US-03 — The evidence is inspectable (priority P2)

- AC-10: Given any classified row, then the operator can see which features contributed,
  beyond a bare class name.
- AC-11: Given a rotating address family, then the family carries the class even though its
  individual addresses are short-lived.

### US-04 — Bounded, prunable storage (priority P1)

Independent verification: store and recorder tests with fixture data.

- AC-12: Given an identity seen in only one hour or with fewer than the packet floor, then no
  daily summary is written.
- AC-13: Given summaries older than the retention window, then they are pruned.
- AC-14: Given no wireless observations at all, then no database file is created and no
  goroutine runs beyond the existing service.

### US-05 — Day one is already quieter (priority P1)

The default filter hides infrastructure, but during learning there is no class to filter on,
so the first day would still be the full census. Two signals require no history and are
available on the first packet, so they carry the filter until the classifier has evidence.

Independent verification: unit tests over synthetic first-session evidence; a browser check
with only learning-state fixture data.

- AC-18: Given an identity with a rotating address (`stable: false`), when less than 24 hours
  of evidence exists, then it is classed `mobile-unidentifiable` and shown, because a rotating
  address is a person's device rather than household equipment.
- AC-19: Given an identity whose strongest reading in the session never exceeded the weak
  threshold, when less than 24 hours of evidence exists, then it is classed `likely-neighbour`
  and hidden by default.
- AC-20: Given an identity with a stable address that has been heard strongly, when less than
  24 hours of evidence exists, then it is shown, because it may be household equipment the
  operator cares about.
- AC-21: Given the provisional class is superseded by a real class, then the row's class
  changes and the transition records that the provisional class preceded it.
- AC-22: Given the pre-filter is active, then the row states that the class is provisional and
  gives the hours observed, so a withheld class is never mistaken for a missing one.

Rationale for the two signals: a rotating address means the device is someone's, not the
building's, so it can never be household infrastructure. A device never heard strongly is
outside the room, which with one anchor is the closest honest approximation of "neighbour".
Both are weaker claims than the full classifier and are labelled as provisional for that
reason.

### Navigation and presentation

- AC-15: Given the default view, then infrastructure and neighbour classes are hidden; the
  operator can reveal them without a reload.
- AC-16: Given any view, then no horizontal overflow at 1280 px and 390 px.

## Requirements and success

- FR-01: MUST accumulate per-identity daily evidence: duty cycle, packet count, RSSI mean and
  variance, hour-of-day presence, dB threshold crossings, strongest smoothed reading, and
  advertisement family. Acceptance: AC-01…AC-05.
- FR-02: MUST classify from a rolling window with recency weighting, and MUST NOT persist the
  class itself. Acceptance: AC-06, AC-08.
- FR-03: MUST apply hysteresis so a class change requires a margin across consecutive
  evaluations. Acceptance: AC-08.
- FR-04: MUST record class transitions with timestamp and both classes. Acceptance: AC-07.
- FR-05: MUST let operator triage override the class. Acceptance: AC-09.
- FR-06: MUST classify advertisement families as well as identities. Acceptance: AC-11.
- FR-07: MUST write a daily summary only when the identity clears a presence floor.
  Acceptance: AC-12.
- FR-08: MUST prune summaries beyond retention. Acceptance: AC-13.
- FR-09: MUST create no storage and start no worker when there are no observations.
  Acceptance: AC-14.
- FR-10: MUST show a learning state instead of a class below the evidence threshold.
  Acceptance: AC-05.
- FR-11: MUST apply a provisional class before the evidence threshold, using only signals
  available on the first packet: rotating address, and strongest-reading-so-far.
  Acceptance: AC-18, AC-19, AC-20.
- FR-12: MUST mark a provisional class as provisional in the payload and the UI, and MUST
  record a transition when a real class supersedes it. Acceptance: AC-21, AC-22.
- NFR-01: Summary storage MUST stay bounded: at most one row per identity per day. Measured
  ceiling for 256 identities × 30 days plus 1,000 transitions is **≈ 1.8 MB** (223 B per
  summary, 159 B per transition), asserted by a test so growth is caught rather than assumed.
- NFR-02: Classification MUST be computed on the aggregation worker, never per render.
- NFR-03: No new runtime dependency.
- NFR-04: Stored summaries MUST contain no raw observations, packet contents, or payloads.

Units and retention: distances in metres; timestamps UTC; daily summaries retained 30 days,
pruned daily. Class transitions retained **30 days, matching the summaries**: a transition
older than the evidence that justified it could be displayed but not explained, so retaining
it longer would be inconsistent rather than useful. Raw observations remain memory-only, as
ADR-002 requires.

## Entities and external interfaces

| Entity | Owner | Meaning |
| --- | --- | --- |
| Daily summary | new store | Aggregates for one identity on one UTC day |
| Behaviour class | computed | infrastructure, personal, neighbour, transient |
| Class transition | new store | One recorded change with time and both classes |
| Classification | computed, broadcast | Class, confidence, and contributing evidence |

External interfaces: additive fields on the existing `observations` snapshot; one new
read-only route for summary history; no new public surface.

## Questions and assumptions

| ID | Question | Consequence | Owner | Resolution |
| --- | --- | --- | --- | --- |
| Q1 | Is "neighbour" distinguishable from "far corner" with one anchor? | Determines whether the class is labelled as a guess | Maintainer | **Agreed: advisory only, worded as likely.** With one source it is a heuristic. |
| Q2 | Should daily summaries be per source or per identity? | Affects memory and multi-anchor future | Maintainer | **Agreed: per identity, with per-source readings nested**, so a second proxy adds evidence without a schema change. |
| Q3 | Retention for transitions: 90 days proposed | Affects storage growth | Maintainer | **Resolved: 30 days, matching summaries.** A transition outliving its evidence could not be explained. |

### Assumptions

- A1: One observing source cannot separate "outside the property" from "far inside it", so
  neighbour is a heuristic. It must never be presented as fact.
- A2: The household is a single deployer; no role model, so triage remains global.
- A3: Summary storage is acceptable to the maintainer provided noise is filtered at the
  write boundary, as stated.

Unresolved safety, security, contract, or acceptance questions block dependent work; Q1
blocks the neighbour wording, not the rest.
