# Tasks: Default view on this browser

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Record behavior, assumptions and NH/FE design gates (spec/plan).
- [x] T002 — Add source-owned domain policy tests for AC-01–04/06; reproduce missing implementation.
- [x] T003 — Implement policy, composable and router integration; depends T002; FR-01.
- [x] T004 — Add shared responsive pin button to title and Panel dock; depends T003; FR-02.
- [x] T005 — Add PanelMode companion browser coverage and run mock UI flows; AC-01–06.
- [x] T006 — Run frontend tests/lint/build, review diff/rules, record actual evidence; NH-05, SC-01.

Evidence: 219 Jest tests, 100% new domain/composable branch coverage, 30 mock browser tests, lint/build/diff checks passed; see plan for warnings, corrected intermediate failures and unrun checks.
