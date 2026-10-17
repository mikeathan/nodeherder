# Tasks: Category-based live reporting

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Inspect lookup/UI/lifecycle and record AC-01–08 and gates.
- [x] T002 — Shared policy/validation/default hydration and companion regressions;
  paths: backend/models/settings; AC-01/02/08; depends T001.
- [x] T003 — Bounded publisher and filter/state integration with fake-clock lifecycle
  tests; paths: backend/models/devices, internal/services/controllers; AC-03–05; T002.
- [x] T004 — Shared responsive editor, explicit containers, inherited summaries,
  cleanup obsolete editors, synchronized contracts/fixtures; AC-06/07/08; T002.
- [x] T005 — Source-owned Jest/browser regressions, screenshot review at both widths,
  failures/offline/keyboard; AC-06/07; T004.
- [x] T006 — Full applicable checks, race/complexity review, update architecture/setup
  and final gates, record actual evidence; NH-05, all AC; T003/T005.

Completion evidence: full Go suite with coverage, affected race suite, vet/build,
242 Jest tests, frontend lint/build, ten final reporting browser checks plus twelve
existing desktop/mobile control/layout/panel checks. Screenshots reviewed at both
widths; complexity evidence and intermediate failures recorded in plan.md. No commits.
