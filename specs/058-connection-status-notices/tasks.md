# Tasks: Compact connection status instead of stacked notices

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete (unreviewed)

- [x] T001 — Define the outcome and exclusions in `spec.md` (AC-01…AC-05). IDs: NH-01.
- [x] T002 — Turn `ConnectionStatus.vue` into a status button with an attention badge, an
  accessible label and a `UiDialog` listing notices with their retry actions. IDs: AC-01,
  AC-02, AC-03, FR-02, FR-03.
- [x] T003 — Remove the fixed connection/session/snapshot panel from `StatusBanners.vue`
  and keep the permit-join banner. IDs: AC-01, FR-04.
- [x] T004 — Update `e2e/StatusBanners.spec.ts` to open the status dialog, assert
  loss/recovery stability and remove dismiss-based checks. IDs: AC-02, AC-04, SC-01.
- [x] T005 — Record unit/build/lint/type-check and full desktop/mobile results in
  `plan.md`. Gate: NH-05.
