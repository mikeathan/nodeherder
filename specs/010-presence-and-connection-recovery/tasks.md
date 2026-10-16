# Tasks: Presence recovery and truthful connection notices

Spec: [spec.md](spec.md) | Plan/evidence: [plan.md](plan.md) | Status: Complete

- [x] T001 — Read AGENTS.md, constitutions, Go/Vue/integration rules, SDD skill/templates;
  create fix branch and specify AC-01…09 before edits. Gates: NH-01/04.
- [x] T002 — Add before-fix source-owned regressions for presence and startup/recovery;
  AC-01…09. Failed before fixes, passed after; evidence recorded in plan.
- [x] T003 — Implement generic stateless Condition.DependsOn / Trigger.AffectedBy and
  single processing path; FR-01/AC-01…04. Remove rejected eligibility cache; verify custom
  conditions, errors, cancellation, events, manual runs and historical changed-reading behavior.
- [x] T004 — Repair socket/snapshot ownership, lifecycle cleanup and fixed notices;
  FR-02/AC-05…09. Verify startup DOM, viewport geometry, keyboard and recovery.
- [x] T005 — Run full component checks, races, minimum Go toolchain and scoped quality;
  NH-05/NFR-01/SC-01. Record actual results, warnings and limits in plan.
- [x] T006 — Review final diff/rules; consolidate required docs and remove temporary
  investigation/test outputs. Keep production fixes and regression tests; no deployment.

- [x] T007 — Extract shared frontend types into src/types; retain store-only interfaces in
  type-only state files and private implementation types locally. Update imports; run
  frontend unit/lint/build checks and remove generated outputs. Mechanical ownership cleanup.
