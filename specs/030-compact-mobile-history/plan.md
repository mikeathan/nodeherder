# Plan: Compact mobile history and navigation

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

## Navigation follow-up — 2026-10-23

User requests removing the bottom strip; retain the hamburger drawer as the phone navigation.
Before production edits: remove BottomNav and its import/render/declaration, primary-only metadata
and phone link hiding, and the unused height token. AppShell phone content keeps only ordinary
spacing and `--nh-safe-bottom`, with safe padding on NavSidebar too. Keep the inset token for
UiDialog and the manifest/viewport contract. No route, panel dock, session or device action change.
The focused test exposed the drawer's existing 40px link height; set a 44px minimum on drawer
layouts (≤960px), retaining larger density tokens and the desktop geometry.
Give the now-primary header hamburger the same 44px minimum within its existing header height.
This overrides the bottom-bar presentation in completed spec 029 rather than rewriting that history.

Reopen NH-01/04/05 and FE-01/03/04 final checks; other design evidence still applies. Replace the
existing bottom-tab test in `e2e/layout.spec.ts` with drawer/geometry/navigation assertions before
editing production. Verify all destinations/current state, keyboard/route dismissal, safe padding,
metadata/failure boot and wider sidebar. Run the full isolated desktop/mobile suite plus frontend
Jest/lint/build/e2e types because the shell is shared. Update docs and PR #75 around the final scope.

Navigation verification: baseline failed with 84px reserved phone padding (expected ≤32px).
The first focused run also exposed existing 40px drawer targets; scoped minimums now protect the
hamburger and drawer at 44px. The corrected focused layout run passed 6/6. Final full isolated
browser run passed **203 tests, nine existing desktop phone-only skips, zero failures (212 cases)**
in 5.0 minutes. Includes restored core links, secondary/actions presence, route/current state,
Escape dismissal, safe inset, manifest failure, desktop sidebar and panel dock regression checks.

Jest passed 53 suites/341 tests; lint passed with seven existing JS warnings; build and e2e types
passed. Entry is 116.18kB gzip; Chart.js unchanged. Updated phone screenshot inspected: no bottom
strip and chart range/readout visible. All final NH/FE checks rechecked; no backend, API, persistence,
unit or dependency change. `git diff --check` and the unused-reference scan passed. No physical
Safari/cutout/frame-timing verification is claimed; prior manual acceptance remains separate.

## Chart-first redesign — 2026-10-23

User rejects the remaining control-heavy phone layout. Reopen NH-01/04/05/06 and FE-01/03/04;
prior runs remain historical. Replace the two-row toolbar with entity/period/History options on one
44px row. HistoryToolbar owns sheet visibility only; DeviceMetrics still owns selections/query/export
and supplies the selected zone/range text. Reuse UiDialog for focus restoration, Escape and scroll
locking; its phone Close button also gets a 44px minimum. The sheet contains a native View picker,
refresh/export and their help. Mobile default hides the redundant zone/unit/keyboard-help rows;
required failure/partial messages remain. Current mode is described on the options opener.

HistoryChart keeps range ownership. Desktop buttons remain; phone buttons move into native details
below the canvas, with Escape-to-close and full visible bounds. No new shared state, global event
listener or dependency. Full mobile canvas height is clamp(280px,48dvh,420px); compact dialog charts
keep 105px. Daily Statistics reorders only on phones to show legend/plot before summary/table and
uses the same height policy. Avoid floating controls covering measurements. No fullscreen or new
network requests are needed to give the plot a substantially bigger surface.

Update the existing DeviceMetrics browser regression first: single-row budget, canvas gap/height,
options keyboard/focus/mode/disabled state, Statistics ordering and tools. Update responsive page
objects and existing mode/export/gesture assertions without dropping coverage. Verify desktop and
phone, long labels, short viewport, both themes, keyboard/gestures and exports; inspect screenshots.
Run full frontend checks and isolated browser suite; update docs and PR #75 with actual evidence.

Local verification uses mock/Vite ports 4810/4800 because the user's dev hub owns 4110/4100.
Temporary loader/config overrides stay uncommitted. Existing EntityViewDialog socket intercepts
must derive their URL from the already-configurable HUB_API instead of assuming port 4110;
confirmation assertions remain unchanged. Local affected-spec coverage uses these separate mocks;
the standard complete browser suite runs in CI, where fixture ports are dedicated.

Chart-first verification: baseline failed at 124.17px against the new 48px toolbar budget. Final
phone header is 44px; canvas starts directly below it and is ~410px at 854px viewport height
(previously 240px). Both themes and 320/384/390 widths, 568px short height, Statistics-first ordering,
44px controls, native details keyboard/Escape and options focus restoration are browser-asserted.
Default, Statistics, options sheet and expanded tools screenshots were inspected.

Final alternate-port affected suite (`DeviceMetrics.spec.ts EntityViewDialog.spec.ts layout.spec.ts`)
passed **93 tests, nine existing desktop phone-only skips, zero failures (102 cases)**, 2.7 minutes.
The final visual/keyboard/touch-target follow-up passed 2/2. Initial runs exposed old phone export
selectors and two hardcoded socket intercepts; intermediate runs were interrupted while migrating
those test assumptions. The final run retains both CSV schemas, gesture/no-extra-request assertions,
confirmation semantics, deep links, layout, and dialog cleanup. The initial standard-port attempt
ran no tests because the user's dev hub occupied 4110; no live hub/device was used or stopped.
The temporary local config was removed; standard full-suite verification is performed by CI.
The options-sheet assertion must use the browser's actual named zone rather than assuming London;
verify the final assertion with a New York browser context as well as the runner's zone.

Jest passed 53 suites/341 tests with a UTC launch. Lint passed (seven existing JS warnings), build
and e2e types passed. Entry is 116.25kB gzip, Chart.js unchanged at 71.35kB; no dependency/lockfile,
backend or API contract diff. `git diff --check` passed. Final NH/FE review confirms original data,
range and command ownership, honest failure/partial feedback, accessible deferred controls and
listener/focus cleanup. Physical Safari/cutout/frame timing remain unrun as already documented.

CI on `4feeb4f4` passed every check except the new phone test on both shard-1 projects: it expected
Europe/London while the runner's browser uses UTC. This was a test assumption, not an app timezone
conversion error. The assertion now compares with the browser's resolved zone. A separate
America/New_York browser follow-up passed 2/2 with the same options, selected-day and gesture UI;
production code is unchanged. Temporary config removed; e2e types/lint/diff checks rerun before push.
