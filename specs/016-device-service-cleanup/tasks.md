# Tasks: Device service retirement

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Inspect removal/output/worker ownership, define AC-01–05 and gates.
- [x] T002 — Add launcher suite to CI, verify Node tests/actionlint; AC-01.
- [x] T003 — Publisher callback unregistration/fenced retirement with companions; AC-02/04/05.
- [x] T004 — Owned lifetime context and joined monitoring, processor eviction and
  fresh-generation regressions; AC-02–05; depends T003.
- [x] T005 — Wire validated bridge removal through device lane, preserve failure
  messages, test ordering; AC-02/04; depends T004.
- [x] T006 — Full backend checks/race review, final complexity/docs and gates; all AC.
