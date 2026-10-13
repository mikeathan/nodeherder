# Tasks: Session reliability and local sign-in

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete (agent checks; maintainer review pending)

- [x] T001 — Read rules, inspect auth, resolve scope/lifetime; NH-01; evidence spec/plan.
- [x] T002 — Add cookie/mode/rejection and frontend failure regressions; FR-01/02/03.
- [x] T003 — Persist cookies, expose mode and gate local start; AC-01/05/07; depends T002.
- [x] T004 — Preserve auth on transient failure, bound/deduplicate checks and retries;
  AC-02/03/04/NFR-01; depends T002.
- [x] T005 — Local/Google/loading/error UI and mock endpoint; AC-05/06/NFR-02.
- [x] T006 — Browser acceptance and required Go/Jest/lint/build checks; NH-05.
- [x] T007 — Update setup docs and final constitutional/evidence review; NH-01/04/05.

Completion evidence: all task checks and limitations recorded in [plan.md](plan.md#completion).
