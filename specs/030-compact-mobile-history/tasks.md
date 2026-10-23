# Tasks: Compact mobile history and navigation

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete (automated scope)

- [x] T001 — Inspect screenshot/current layout and specify AC-01…04 before code edits; NH-01.
- [x] T002 — Add geometry/keyboard regression to existing `e2e/DeviceMetrics.spec.ts`; AC-02/03;
  depends T001; done: baseline failure recorded and responsive flow assertions retained.
- [x] T003 — Compact scoped toolbar and date margins in HistoryToolbar/DeviceMetrics; AC-01…04;
  depends T002; done: all choices/actions accessible and phone height ≤144px.
- [x] T004 — Migrate selections through existing `e2e/support/ui.ts`; AC-01/04; depends T003;
  done: phone picker and desktop button selection/restore/export checks pass.
- [x] T005 — Run Jest/lint/build/e2e types and both-project DeviceMetrics/EntityViewDialog checks;
  NH-05; depends T004; done: record actual checks, inspect light/dark screenshots and final rule diff.
- [x] T006 — Update user docs and record evidence/limits; NH-01/04/05; depends T005.

Evidence: baseline toolbar 243.78px failed the 144px budget. Final affected browser run 87 passed /
nine existing skips; Jest 341 passed, lint/build/e2e types and diff checks passed. See plan.md.

- [x] T007 — Update existing `e2e/layout.spec.ts` for AC-05/06 before production edits; demonstrate
  bottom-space regression, then remove BottomNav/metadata/height reservation and restore drawer links
  in AppShell/NavSidebar/navigation/tokens; preserve safe inset and panel dock.
- [x] T008 — Run full isolated browser suite and frontend component checks; AC-01…06/NH-05;
  depends T007; done: actual checks, docs and rule review recorded, PR #75 updated.

Navigation follow-up: full browser 203 passed / nine existing skips; Jest 341 passed; lint/build/types
and diff checks passed. BottomNav and its reservations/metadata removed; drawer retains all routes.
