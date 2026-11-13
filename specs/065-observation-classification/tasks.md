# Tasks: Behavioural classification and a specialised viewer

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft

Each task maps to requirement/acceptance IDs or a constitutional gate and names its evidence.
Mark done only with evidence. No task is started.

## Preparation

- [x] T001 — **DONE**: ADR-001, ADR-002 and ADR-003 recorded. Q1 agreed (neighbour is
  advisory, worded as likely); Q2 agreed (per identity, per-source readings nested); Q3
  corrected to 30 days so a transition never outlives its evidence. IDs: FR-01, FR-06.
- [ ] T002 — Confirm the summary schema against a fixture day and the storage ceiling in
  NFR-01; gate: BE-04; done: arithmetic recorded in the plan.

## Backend

- [ ] T003 — Red tests for evidence accumulation: hour mask, Welford mean and variance, band
  transitions, write floor. Paths: `backend/internal/observations/evidence_test.go`.
  IDs: FR-01, AC-01, AC-12; done: failing for the right reasons.
- [ ] T004 — Implement `evidence.go`; IDs: FR-01; depends: T003; done: green.
- [ ] T005 — Red tests for classification: the four classes from synthetic evidence, the
  learning threshold, recency weighting, hysteresis needing two evaluations, family rollup,
  **and the measured reclassification delay** (a change needs about five days of sustained
  contrary evidence; one day must not flip). Paths: `classify_test.go`. IDs: FR-02, FR-03,
  FR-06, FR-10, AC-01…AC-08, AC-11, AC-17.
- [ ] T006 — Implement `classify.go` as a pure function; IDs: FR-02, FR-03; depends: T005;
  done: green, table-driven, no clock dependency outside an injected `now`.
- [ ] T006a — Red tests for the provisional pre-filter: rotating address is mobile and shown;
  never-strong is likely-neighbour and hidden; stable-plus-strong stays learning and shown;
  a real class supersedes a provisional one and records the transition.
  Paths: `provisional_test.go`. IDs: FR-11, FR-12, AC-18…AC-22; depends: T006.
- [ ] T006b — Implement the provisional pre-filter, used only below the evidence threshold.
  IDs: FR-11, FR-12; depends: T006a; done: green, and a test proving the full classifier
  always wins once evidence exists.
- [ ] T007 — Red tests for the summaries store: write floor, retention, `Prune`, absence of a
  file when there are no observations, and that **transitions are pruned on the same 30-day
  boundary as summaries**. Paths: `internal/observations/summaries/*_test.go`.
  IDs: FR-07, FR-08, FR-09, NFR-01.
- [ ] T008 — Implement `internal/observations/summaries`, mirroring `spatialhistory`:
  repository interface, bbolt store, worker-serialised recorder, bounded queue that drops.
  IDs: FR-07, FR-08, NFR-01; depends: T007; done: green, `-race` clean.
- [ ] T009 — Wire the summariser into the service and add the additive snapshot fields.
  IDs: FR-01, FR-02, FR-04; depends: T004, T006, T008; done: snapshot carries class and
  evidence; transitions recorded on change only.
- [ ] T010 — Add the summary history route, protected, bounded. IDs: AC-10; depends: T008;
  done: route test with a bounded response.
- [ ] T011 — Triage override and class advisory precedence. IDs: FR-05, AC-09; done: snapshot
  shows both.

## Frontend

- [ ] T012 — Contracts and types for the new optional fields; mock fixtures covering each
  class plus learning. IDs: FE-02; done: malformed and unknown-class cases rejected.
- [ ] T013 — Tracked/Census split with the default filter; "N hidden" footer. IDs: FR-09,
  AC-15; done: default view hides infrastructure and neighbour.
- [ ] T014 — Class presentation: word, confidence, "why" phrase, transition note, and a
  distinct **provisional** treatment that states the class is not settled. IDs: AC-07, AC-10,
  AC-22; done: no bare badges anywhere and a provisional class is never shown as settled.
- [ ] T015 — Timeline pane, desktop grid and mobile horizontal scroll with a sticky identity
  column. IDs: AC-15; done: renders a day per identity.
- [ ] T016 — Radar: class colours, worded legend, labels for the tracked set. IDs: FE-04;
  done: colour is never the only cue.

## Handoff

- [ ] T017 — Browser verification at 1280 px and 390 px, keyboard traversal recorded.
  IDs: AC-16; gate: FE-04.
- [ ] T018 — Run all planned checks and record actual results. Gate: NH-05.
- [ ] T019 — Final rule audit against the full diff; report unresolved findings.
  Gates: NH-01, NH-04, NH-05.

Add discovered work here and update the spec/plan when scope changes.
