# Tasks: Home picker native selection under live reports

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: implemented; checks recorded

## Implementation

- [x] T001 — Own the pending native choice in `frontend/src/components/ui/UiSelect.vue`:
  commit on `input`, consume the paired `change`, and keep typed/disabled options unchanged.
  IDs: FR-01, AC-01, AC-04. Done: a pending `input` now survives reports and reports the chosen
  option; the affected entity and automation browser suites still pass.
- [x] T002 — Reuse the Home source and filter selectors while their entries are unchanged:
  `stableHomePickerOptions` / `stableHomePickerFilters` in `frontend/src/domain/home/picker.ts`,
  consumed by `AddTilesDialog.vue` through previous-value computeds. IDs: FR-02, FR-03,
  AC-01–AC-03. Done: identity tests pass; renamed, reordered and added devices still rebuild.

## Verification

- [x] T003 — `npx jest src/__tests__/domain/home/picker.test.ts --runInBand` in `frontend/`.
  Result: 6 passed, including the two selector-identity tests.
- [x] T004 — `npx playwright test e2e/home.spec.ts --grep "pending source choice"` with the
  real mock hub and live `deviceUpdated` frames. Result: desktop PASS (13.2 s), mobile PASS
  (13.3 s). The test dispatches the menu's `input`, waits for a new report, and expects the value
  and group to stay.
- [x] T005 — `npm test -- --runInBand`: 126 suites, 668 tests passed. `npm run lint`: 0 errors,
  28 pre-existing warnings. `npm run build` (includes `vue-tsc`): passed, built in 2.16 s.
- [x] T006 — Affected browser suites: `home` desktop 10/10; `DeviceMetrics` desktop + mobile;
  `EntityViewDialog` + `DeviceDashboard` desktop, 34 passed, 9 skipped; `automations` desktop,
  4 passed; `devices` desktop, 9 passed. Final diff reviewed against the Vue rules; no
  unresolved findings.
- [x] T007 — PR CI first failed `DeviceMetrics` on a stale count: the grouped History selector
  excludes the fixture's config `countdown` entity, so the test now expects 9 options and asserts
  `countdown` is absent. The rest of that spec passed unchanged; rerun passed on desktop and
  mobile. Lesson recorded: run the owning spec for each changed domain module before pushing
  (`historyEntityOptions` is owned by `DeviceMetrics` and `EntityViewDialog`), not only the specs
  reached from the changed component.

## Deferred

- The deployed service (build dated 2026-11-02) predates this change and needs a rebuild and
  redeploy to receive the fix. That host was not changed from this session.
