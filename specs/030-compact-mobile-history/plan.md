# Plan: Compact mobile history toolbar

Spec: [spec.md](spec.md) | Status: Implemented; automated checks passed | Updated: 2026-10-23
Constitutions: parent/frontend proposed 1.0.0. Branch: `fix/mobile-history-toolbar`.

## Context and design

Merged PR #74 is the baseline; leave spec 029 as history. `HistoryToolbar.vue` will retain one set
of selection props/events and render responsive native selectors below 720px, chips above it. Two
pickers share a grid row; mode buttons form a compact segmented control alongside refresh/export
icons, with an accessible Export CSV name and existing unit help. Use scoped local dimensions and
44px targets. Tighten phone date/zone margins in `DeviceMetrics.vue` without removing information.
No breakpoint JS, new dependency or duplicate selection state. The alternative of merely shrinking
pills retains horizontal scrolling and several rows, which does not address the supplied screenshot.

## Constitution Check

| Rule | Design evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec acceptance precedes production edits | PASS | PASS |
| NH-02 | Read-only presentation; no device/event changes | PASS | PASS |
| NH-03 | No network/auth or household-data changes | PASS | PASS |
| NH-04 | Existing component props/events; no dependency or boundary decision needing ADR | PASS | PASS |
| NH-05 | Existing DeviceMetrics e2e plus geometry/keyboard; isolated mocks | PASS | PASS |
| NH-06 | Units, gaps, ranges and storage unchanged | PASS | PASS |
| FE-01 | Component presentation; parent retains state | PASS | PASS |
| FE-02 | No API/WS changes | N/A | N/A |
| FE-03 | Existing loading/error/export-disabled feedback preserved | PASS | PASS |
| FE-04 | Named pickers, keyboard/focus, 44px and both-theme geometry checks | PASS | PASS |
| FE-05 | No sensitive operations | N/A | N/A |

## Contracts, data, rollout

No interface/data migration. Existing native/display CSV disclosure remains visible. Frontend-only
rollback. No backend checks needed for this layout-only change; inspect the diff for contract drift.

## Traceability and verification

AC-01/04: migrate existing `e2e/DeviceMetrics.spec.ts` selection/restore and `EntityViewDialog.spec.ts`
deep-link checks through responsive
page-object helpers in `e2e/support/ui.ts`; keep desktop chip assertions and mobile picker assertions.
AC-02/03: add a phone geometry, keyboard, export and screenshot case in the same source-owned file;
measure 320/384/390 widths, long colour-temperature label and dark/light appearance. Demonstrate
baseline failure, then inspect the resulting screenshots. Run all DeviceMetrics and EntityViewDialog
checks on both browser projects, plus frontend Jest, lint, build, and e2e TypeScript.

## Debt, risks, exceptions

Long labels must stay inside the picker and all options remain readable when opened. Native popup
appearance is browser-owned. Physical Safari/safe-area/frame-timing acceptance from 029 remains
unrun and is not established by this correction. No exceptions or maintainer approval implied.

## Completion

Checks and visual evidence pending; implementation scope and rules reviewed before edits.

Initial evidence: the phone regression failed before the fix (243.78px toolbar versus ≤144px).
Phone geometry passed after the fix. The first full affected-spec run passed 84 with nine existing
desktop skips and three test failures: the keyboard test needed to wait for numeric mode controls
after reload, and the sheet deep-link assertion still expected a phone chip. Migrate that assertion
without dropping its expose precedence check; use native picker selection plus Tab/focus and Space
mode activation rather than OS-specific End-key popup behavior. Rerun the affected specs.

## Final verification — 2026-10-23

- `TZ=UTC npm test -- --runInBand`: 53 suites / 341 tests passed.
- `npm run lint`: passed, zero errors and seven existing JS warnings.
- `npm run build`: passed; entry 116.20kB gzip and Chart.js 71.35kB unchanged. Existing DeviceCard
  auto-import naming warning remains. No dependencies added.
- `npm run type-check:e2e`: passed.
- `CI=true npm run test:e2e -- DeviceMetrics.spec.ts EntityViewDialog.spec.ts`: final run **87 passed,
  nine existing desktop phone-only skips, zero failures (96 cases)**, 2.3 minutes.
- Phone toolbar ≤144px at 320/384/390px (baseline 243.78px at 320), 44×44 targets, long colour
  temperature label, actions on one row, both themes and no overflow asserted. Keyboard focus/Tab
  order and Space mode activation, Yesterday restore, Custom feedback and export are covered.
- Both-theme screenshots inspected; the chart fits in the 384×854 screenshot. Artifacts are generated
  under `frontend/test-results/DeviceMetrics-mobile-histo-b9354--labels-and-usable-controls-*/`.
- Wider 768/1280 layouts and responsive selected values asserted; sheet deep-link precedence kept.
- `git diff --check` passed; backend/contracts/package/lockfile diff empty. Backend checks unrun
  for this scoped presentation change. No household hub/device used.

All NH/FE checks were rechecked against the final diff with no exception. User docs updated in
`docs/history.md`. No physical Safari popup, real cutout or phone frame timing result is claimed;
029's remaining manual acceptance is unaffected.
