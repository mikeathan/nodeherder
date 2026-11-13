# Plan: Behavioural classification and a specialised viewer

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft for maintainer review | Updated: 2026-11-12
Constitutions: parent 1.0.0 (Proposed); backend 1.0.0, frontend 1.0.0 (Proposed)
ADRs: [adr-001](adr-001-evidence-summaries.md), [adr-002](adr-002-classification-rules.md),
[adr-003](adr-003-classification-viewer.md)

Repository baseline: `feat/wireless-observations` at `d0e55112`. This is new work in the
sense of the branch-scope rule only if 064 is merged first; otherwise it extends the same
branch's set. Recorded as a separate directory because the scope is a new subsystem.

No ratification or reviewer approval is claimed. "Pending" results are unrun.

## 1. The design problem

Forty identities, four of which matter. Classification is the filter; the viewer must then
make those four legible. The current UI is organised for a census — one flat table, one
radar — which is the wrong shape for a small set of tracked things.

Two things change it:

- **A class is a claim about behaviour over days.** The UI must show the claim, its evidence,
  and when it last changed. A bare badge would be untrustworthy and unarguable.
- **Most identities should never be seen.** The default view is the tracked set. The census
  becomes a secondary, deliberate view — "show everything the hub can hear".

## 2. Backend

### 2.1 Evidence accumulation

Extend `aggregate` (`internal/observations/service.go`) with a summariser that folds each
observation into hourly buckets, then promotes completed days to storage.

```go
// evidence accumulates the aggregates classification needs for one identity. It holds no
// packets: every field is a bounded counter or statistic, so memory grows with identities,
// not with traffic.
type evidence struct {
    firstSeen, lastSeen time.Time
    packets             uint64
    hours               [24]bool     // seen during this hour of day, ever in the window
    hoursToday          [24]bool     // this UTC day, promoted to storage at rollover
    rssi                welfordState // count, mean, M2
    strongest           int          // strongest smoothed median seen
    weakest             int
    bandTransitions     uint32
    family              string
}

// welfordState is a numerically stable running mean and variance.
type welfordState struct{ Count uint64; Mean, M2 float64 }
```

`bandTransitions` increments when the smoothed median crosses a `DISTANCE_RANGES` boundary
since the previous observation. This is the mobility signal: variance conflates "moved" with
"someone walked past", whereas a band crossing is an actual change of estimated range.

### 2.2 Storage

New package `internal/observations/summaries`, following `internal/spatialhistory` exactly —
own bbolt file, own `Prune`, worker-serialised writes, bounded queue that drops.

```
data/observation-summaries.db
  bucket "summaries":  key "<identityKey>|<YYYY-MM-DD>"  value DailySummary
  bucket "transitions": key "<unixNano>|<identityKey>"   value Transition
```

```go
// DailySummary is one identity's behaviour on one UTC day. It holds aggregates only: no raw
// observations, packet contents or payloads, so retention carries no privacy blast radius.
type DailySummary struct {
    Key             string
    Day             string  // YYYY-MM-DD, UTC
    Packets         uint64
    ActiveHours     uint8   // hours with at least one observation, 0-24
    HourMask        uint32  // bit N set when heard during hour N
    RSSIMean        float64
    RSSIM2          float64
    Strongest       int
    Weakest         int
    BandTransitions uint32
    Family          string
}
```

Write floor (AC-12): a day is stored only when `ActiveHours >= 2 && Packets >= 20`. A device
that appears once never earns a row, so noise costs nothing.

Retention: summaries 30 days, transitions **30 days to match**, pruned daily. 256 identities ×
30 days × ~120 B ≈ 900 KB — stated in NFR-01 and asserted by a test. A transition must not
outlive the evidence that justified it, or the UI could show a change it cannot explain.

### 2.3 Reclassification speed, measured

Recency weighting uses a 5-day half-life. The arithmetic, recorded so the delay is a decision
rather than a surprise:

| Evidence | Weighted day-equivalents |
| --- | --- |
| 21 days as infrastructure | 7.30 |
| 2 hours of new behaviour, today | 0.08 |

So a single unusual day cannot move a long-standing class. How long sustained contrary
behaviour needs before it wins:

| New behaviour held for | New | Old | Flips |
| --- | --- | --- | --- |
| 1 day | 1.87 | 5.43 | no |
| 3 days | 3.29 | 4.02 | no |
| 5 days | 4.36 | 2.94 | yes |

**A genuine reclassification therefore takes about five days of consistent evidence.** That
is the intended trade: slow enough that a fridge does not flap, fast enough that a device
which really moved is corrected within a week. Reducing the half-life to 2 days would flip in
about two days at the cost of more flapping; that single constant is the tuning lever.

### 2.4 Classification

`internal/observations/classify.go`, a pure function over the rolling window so it is
directly testable and never persisted.

```go
// Class is a behaviour category, not an identity. It answers "what kind of thing is this"
// from observed behaviour, and is always advisory: operator triage overrides it.
type Class string

const (
    ClassInfrastructure Class = "infrastructure"
    ClassPersonal       Class = "personal"
    ClassNeighbour      Class = "neighbour"
    ClassTransient      Class = "transient"
    ClassLearning       Class = "learning"
)
```

Scoring, per class, over summaries weighted by age (half-life 5 days):

| Feature | Infrastructure | Personal | Neighbour | Transient |
| --- | --- | --- | --- | --- |
| Duty cycle | strong + | mid | strong + | very low |
| Band transitions | ~0 | many | ~0 | ~0 |
| Strongest vs weakest span | small | large | small | n/a |
| Night presence (00–05) | corroborates | anti | corroborates | n/a |
| Family continuity | high | low per address | high | low |

Each feature contributes a bounded score; the highest total wins, confidence is the margin
over the runner-up. Two safeguards the spec requires:

- **Recency weighting** so a device that changes behaviour can change class (AC-06), with
  hysteresis so one unusual day does not flip it (AC-08): a class change needs the candidate
  to win on two consecutive evaluations.
- **Learning state** below 24 h of evidence (AC-05, FR-10): no class, hours observed instead.

The class is **never stored**. Only evidence and transitions are stored, so changing the rules
needs no migration and no reclassification job.

Family-level classification (FR-06, AC-11): summaries for short-lived addresses roll up by
`Family` (vendor + AD signature from spec 064's decoder). A family with a personal profile is
treated as personal even though no individual address survives.

### 2.5 Broadcast

Additive fields on `IdentityView`:

```go
Class      string   `json:"class"`       // learning until enough evidence
Confidence float64  `json:"confidence"`
Evidence   Evidence `json:"evidence"`    // dutyCycle, hoursObserved, bandTransitions, nightPresence
ChangedAt  *time.Time `json:"changedAt,omitempty"`
PreviousClass string `json:"previousClass,omitempty"`
```

Optional at browser ingress, so an older payload still parses. `Evidence` is what makes
AC-10 possible.

## 3. Frontend redesign

The viewer becomes two surfaces with different jobs, plus one new pane.

### 3.1 Structure

```
Wireless
├── Tracked          ← default. Only personal + anything the operator marked known.
│   ├── Radar        (spatial, for the tracked set)
│   ├── Focus        (one identity: distance, trend, class, evidence)
│   └── Timeline     (NEW: when each tracked identity is present)
└── Census           ← deliberate. Everything, classified, filterable.
    ├── List         (grouped by class, not by vendor)
    └── Band filter, class filter, distance bands
```

The current tab strip becomes a **two-level** control: a small Tracked/Census switch, and the
view tabs within it. One level cannot express "these are different questions".

### 3.2 Radar changes

The radar is only meaningful for a small set and currently wastes that fact. Three changes:

1. **Class-coloured blips with a legend**, not vendor tinting. Colour is a fast read of *what
   kind of thing* this is; the legend states it in words so colour is never the only cue
   (FE-04).
2. **Labels for the tracked set.** With a handful of identities, a name beside each blip is
   affordable and removes the need to hover.
3. **Ring labels that follow the span**, already implemented, plus a "N hidden" footer stating
   what the filter removed — so the radar never silently lies about the household.

Deliberately unchanged: no bearing claim, no movement animation beyond position drift, and
the range-only caption.

### 3.3 Timeline pane (new)

The thing a census cannot show and a radar cannot either: **presence over time**. One row per
tracked identity, 24 columns per hour, filled where the identity was heard.

```
               00  03  06  09  12  15  18  21
  Front door   ▓▓▓ ▓▓▓ ░░░ ░░░ ▓▓▓ ▓▓▓ ▓▓▓ ▓▓▓
  Mike's phone ░░░ ░░░ ░░░ ▓▓▓ ▓▓▓ ▓▓▓ ▓▓▓ ░░░
  Garden tag   ░░░ ░░░ ▓▓▓ ▓▓▓ ▓▓▓ ░░░ ░░░ ░░░
```

This is where patterns actually appear, and it is the honest answer to "I can't find
patterns": a week of these rows shows routine. On mobile it becomes a horizontally scrollable
single-day view with hour ticks, not a squeezed grid.

### 3.4 Classification presentation

A class is never a bare badge. Each row carries:

- the class word,
- a confidence indicator,
- **why** — one short phrase from the dominant features, e.g. *"present 96% of the day, never
  moved"*,
- and, if it changed recently, *"was infrastructure, 2 days ago"*.

The operator's triage stays a separate, stronger control: marking a device known pins it into
Tracked regardless of class, and the UI states that the class is advisory.

### 3.5 Mobile

- Tracked is the mobile default; Census is behind the switch.
- The radar keeps the full column, already implemented.
- Timeline scrolls horizontally with a sticky identity column; never a grid squeeze.
- Focus stays a single column.
- Every control remains keyboard reachable, and status is never colour-only.

### 3.6 Reuse

| Need | Reuse | Not |
| --- | --- | --- |
| Persistence shape | `spatialhistory` store, recorder, `Prune` | new storage idiom |
| Recency weighting | none exists; new, pure and tested | ML dependency |
| Radar projection | `radialRatio`, `ringsForSpan` from 064 | new geometry |
| Distance bands | `DISTANCE_RANGES` | duplicate bands |
| Filter chrome | `ObservationFilters` segmented groups | new control style |
| View ids | `ObservationViews` union | string literals |

`SpatialCanvas.vue` remains excluded per 064 ADR-003.

## 4. Contracts, data, rollout

`contracts/classification.md` documents the additive snapshot fields and the summary route:
producer, cadence, field meanings, and the rule that evidence is advisory.

Compatibility: all additions optional. A hub with no summaries broadcasts `class: "learning"`,
which the UI renders as a learning state rather than a false class. An older browser ignores
the fields.

Rollout: summaries (inert) → classifier behind a computed field → default filter → Tracked/
Census split → Timeline pane.

Rollback: the summaries file is additive and prunable; deleting it returns the viewer to spec
064 behaviour with every class reading `learning`. No migration is required in either
direction because classes are never stored.

## 5. Constitution Check

| Rule | Evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | AC-01…AC-16 precede code | PASS (design) | PENDING |
| NH-02 | No device state or command; classes are advisory display data | PASS (design) | PENDING |
| NH-03 | No new public route; summaries store no payloads or raw observations; identity keys stay private | PASS (design) | PENDING |
| NH-04 | New package with a documented boundary; ADRs 001–003 | PASS (design) | PENDING |
| NH-05 | Traceability §6; synthetic evidence fixtures, no live devices | PASS (design) | PENDING |
| NH-06 | Bounded queue, bounded rows, defined retention, shutdown joins the worker | PASS (design) | PENDING |
| BE-01 | Classifier is pure; storage behind a repository interface | PASS (design) | PENDING |
| BE-03 | One worker per store, context-bound, `-race` on affected packages | PASS (design) | PENDING |
| BE-04 | Retention and Prune specified; existing-data fixtures required | PASS (design) | PENDING |
| FE-01 | Classification reaches state through the existing observations module | PASS (design) | PENDING |
| FE-02 | Runtime validation for the new optional fields; mocks updated | PASS (design) | PENDING |
| FE-03 | Learning state and "N hidden" are explicit; no invented class | PASS (design) | PENDING |
| FE-04 | Legend states classes in words; keyboard and two viewports verified | PASS (design) | PENDING |
| FE-05 | No secrets; summaries stay server-side; nothing added to browser storage | PASS (design) | PENDING |

## 6. Traceability and verification

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-01, AC-01…04 | `observations/evidence_test.go` | synthetic evidence per class, table-driven | Pending |
| FR-02, FR-03, AC-06/08 | `observations/classify_test.go` | recency weighting; hysteresis needs two evaluations | Pending |
| FR-04, AC-07 | `observations/summaries`, recorder tests | transition recorded on change only | Pending |
| FR-05, AC-09 | `observations/service_test.go` | triage overrides class in the snapshot | Pending |
| FR-06, AC-11 | `classify_test.go` | family rollup classifies a rotating set | Pending |
| FR-07, AC-12 | recorder tests | below the floor writes nothing | Pending |
| FR-08, AC-13 | store tests | `Prune` removes expired rows | Pending |
| FR-09, AC-14 | service tests | no file and no goroutine without observations | Pending |
| FR-10, AC-05 | `classify_test.go` | under 24 h yields learning | Pending |
| NFR-01 | recorder tests | row-count ceiling asserted | Pending |
| FE, AC-10/15/16 | `e2e/wireless.spec.ts` | classification, filter default, timeline, two viewports | Pending |

Commands: `backend/` — `go build ./...`, `go vet ./...`, `go test ./...`,
`go test -race ./internal/observations/...`; `frontend/` — `npm test -- --runInBand`,
`npm run lint`, `npm run build`, `npx playwright test e2e/wireless.spec.ts`.
No test may depend on a live device.

## 7. Risks and limits

- **Cold start.** First 24 h shows learning only. Stated in the UI, not hidden.
- **A phone on a charger looks like infrastructure.** Real ambiguity; the transitions log and
  the "why" phrase are the mitigation, and triage is the override.
- **Neighbour is a heuristic.** One anchor cannot separate "outside" from "far corner". The
  wording must be advisory (A1).
- **A TV switched off at night** sits around 60% duty cycle, not 100%. Thresholds are tuned
  against fixtures, not guessed, and are a single documented constant block.
- **Rule tuning is a judgement call.** Keeping the class unpersisted means rules can be
  changed without migration, which is the main reason for that decision.
- **Storage is a scope change** relative to ADR-002; recorded as an amendment in ADR-001
  rather than a silent divergence.

## 8. Task sequence

- T001 Record ADR-001 (evidence summaries), ADR-002 (rules not model), ADR-003 (viewer split).
- T002 Red tests: evidence accumulation and the classifier table, from fixtures.
- T003 Implement `evidence` and `classify.go`; green.
- T004 Red tests for the summaries store: floor, retention, prune.
- T005 Implement `internal/observations/summaries` mirroring `spatialhistory`.
- T006 Wire the summariser into the service; add the additive snapshot fields.
- T007 Extend contracts/types and mock fixtures; unit tests.
- T008 Frontend: Tracked/Census split and the default filter.
- T009 Frontend: class presentation with the "why" phrase and transition note.
- T010 Frontend: Timeline pane, desktop and mobile.
- T011 Frontend: radar class colours, legend, tracked labels.
- T012 Browser verification at 1280 px and 390 px; record results.
- T013 Run all checks, update artifacts, final rule audit.

## 9. Completion

Acceptance evidence per §6, contracts published, browser checks recorded, reviewer approval
noted. Draft until reviewed.
