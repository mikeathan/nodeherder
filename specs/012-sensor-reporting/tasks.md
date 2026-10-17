# Tasks: Meaningful sensor reporting and automation routing

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Read rules and collect read-only Z2M examples; NH-01/03; observations.md.
- [x] T002 — Record scope/contracts/design checks; NH-01/04; spec.md and plan.md.
- [x] T003 — Source-owned policy/settings/automation regression tests and baseline CRAP;
  AC-01–08; depends T002; record failures before fixes.
- [x] T004 — Domain filter/settings validation and lifetime integration; FR-01/03; T003.
- [x] T005 — Derived automation input routing and lifecycle tests; FR-02; T003.
- [x] T006 — Shared threshold editor/contracts/mock UI verification; AC-09; T004.
- [x] T007 — Full tests/race/vet/build/lint/performance/CRAP; NH-05; T004–06.
- [x] T008 — Independent final diff/rules review, correct docs and actual evidence; NH-01–06.

Evidence: full Go/race/vet/build, 239 Jest tests, lint/build, mock UI checks, routing benchmark and CRAP comparison; see plan for exact scope, warnings, corrected intermediate failures and limitations. Final diff manually reviewed against NH/BE/FE rules. No exceptions, ratification, commits or deployment claimed.
