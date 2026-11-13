# Tasks: Behavioural classification and a specialised viewer

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft

Each task maps to requirement/acceptance IDs or a constitutional gate and names its evidence.
Mark done only with evidence. No task is started.

## Preparation

- [x] T001 — **DONE**: ADR-001, ADR-002 and ADR-003 recorded. Q1 agreed (neighbour is
- [x] T002 — **DONE**: schema confirmed and the ceiling measured rather than estimated. A

## Backend

- [x] T003 — **DONE**: `evidence_test.go`, 11 tests. Hour mask, Welford over 10k readings, mobility thresholds, write floor, day reset, unusable readings.
- [x] T004 — **DONE**: `evidence.go`. RSSI 0 counts as activity but not as signal evidence, because 0 dBm means unmeasured, not zero distance.
- [x] T005 — **DONE**: `classify_test.go`, 12 tests. Hysteresis is measured against real scores rather than asserted.
- [x] T006 — **DONE**: `classify.go`. Per-day scoring then recency weighting. Mobility counted in dB, not distance bands.
- [x] T006a — **DONE**: `provisional_test.go`, 5 tests.
- [x] T006b — **DONE**: `provisional.go`. `Provisional.asClassified` is the single bridge to the classifier's type.
- [x] T007 — **DONE**: `summaries/store_test.go`, 11 tests. Includes the prefix-collision guard, which caught a real key-layout bug.
- [x] T008 — **DONE**: `summaries/store.go` + `observations/summarise.go`. The recorder lives in `observations` because it needs the accumulator; the dependency runs one way only.
- [x] T009 — **DONE**: class reaches the snapshot; the store is read on a slow interval outside the aggregation lock, and a read failure keeps the last good class.
- [x] T010 — **DONE**: `GET /api/observations/history?key=…`, protected, bounded to 30 rows, returns summaries and transitions together.
- [x] T011 — **DONE**: `Advisory: true` in the payload, and a test proving an ignored identity never appears regardless of its class.

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
