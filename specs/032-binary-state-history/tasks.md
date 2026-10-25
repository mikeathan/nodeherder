# Tasks: Responsive binary history chart

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Implemented; automated checks passed; manual Safari review pending

Each task maps to spec acceptance or a constitutional gate. Test ownership follows the source: state-domain regressions go in `frontend/src/__tests__/domain/history/state.test.ts`; browser coverage extends existing `frontend/e2e/DeviceMetrics.spec.ts` and `EntityViewDialog.spec.ts`. Do not add a feature-named test file.

## Preparation

- [x] T001 — Capture the current contact-history baseline and review the 032 spec/plan against the latest accepted Option A prototype; IDs: NH-01, AC-01…09; done: before-edit Playwright baseline passed 2/2 desktop/mobile projects with screenshots at 320/390/768/1280px in both themes and no page overflow; baseline showed the 2rem strip and three-label axis. Full-history scope and compact-dialog exclusion confirmed.
- [x] T002 — Add failing pure-domain expectations to `frontend/src/__tests__/domain/history/state.test.ts` for width bins; IDs: AC-05/07, NH-05/06; depends: T001; done: first red run failed because `stateHistoryBins` was not yet exported; after implementation the same conservation, unknown, transition and immutability assertions pass.

## Coverage and implementation

- [x] T003 — Reopen `frontend/src/domain/history/state.ts` for pure, O(N + W) visible-range bins plus `recentActivityWindow()` and `clipStateSegments()`; IDs: AC-05/07, AC-10/11, NFR-01/03; depends: T011; done: conserve active/known durations, keep unknown bins null, return latest transition cluster plus bounded focused range, clip without mutation, and cover same-time/no-transition/full-range cases in the existing `state.test.ts`.
- [x] T004 — Reopen the non-compact `StateHistory.vue` path for a one-chart activity-focus mode; IDs: AC-01/02/04, AC-10/12, FR-01/02/06; depends: T003; done: overview and focus reuse the same SVG, lane geometry and palette; coverage note is next to the chart; compact strip remains unchanged.
- [x] T005 — Reopen responsive geometry/render selection for local `viewRange`; IDs: AC-03/05/07/11, NFR-01/02/03; depends: T004; done: clip to overview or focus bounds, use pixel-bounded bins, keep unknown visible and retain responsive axes without changing source series.
- [x] T006 — Reopen interactions for the named Zoom-to-activity/Back-to-full-range control; IDs: AC-06/10/11, FE-04; depends: T005; done: focus action is keyboard/touch accessible, focused bounds are announced, reset restores the overview, and inspection remains request-free.
- [x] T007 — Extend the existing `frontend/e2e/DeviceMetrics.spec.ts` fixture/regressions for activity focus; IDs: AC-01…08, AC-10…12, SC-01/02; depends: T004…T006; done: 6h/24h sparse clusters show the callout, same chart zooms/resets, 1h desktop stays overview while phone adapts, `<1%` coverage is explicit, full stats/period/query count remain unchanged, and the 10,000-point regression still passes.
- [x] T008 — Preserve the compact entity dialog and surrounding history behavior; IDs: AC-08/09, FR-05; depends: T004; done: DeviceMetrics empty/partial/sample/route and EntityViewDialog regression cases pass in both Playwright projects; compact chart branch unchanged.

## Handoff

- [x] T009 — Re-run frontend checks; gate: NH-05, FE-04; depends: T003…T008, T011…T013; done: record final Jest/lint/build/e2e typecheck and desktop/mobile `DeviceMetrics.spec.ts` + `EntityViewDialog.spec.ts` results; use only configured isolated/mock hubs.
- [x] T010 — Reopen the final review; IDs: AC-01…12, NH-01…06, FE-01…05; depends: T009; done: inspect overview/focus/reset/coverage screenshots, review updated spec/plan and diff, record any unrun Safari check, and confirm no backend/contract/dependency drift or exception.

Implementation and automated verification are complete. Physical Safari/iOS review remains pending; see `plan.md`.

## Activity-focus follow-up

- [x] T011 — Add source-owned coverage in the existing `frontend/src/__tests__/domain/history/state.test.ts` and `frontend/e2e/DeviceMetrics.spec.ts` for the rare 6h/24h burst, focus-window bounds, coverage note, unchanged full-range stats and request count; IDs: AC-10…12, NH-05/06; depends: T002; done: helper boundary tests and sparse/zoom/reset browser assertions pass.
- [x] T012 — Implement the local focused viewport in `StateHistory.vue`; IDs: AC-10/11, NFR-03; depends: T003, T011; done: one chart changes x-range with 44px zoom/reset controls, latest ten transitions plus context/minimum window clamped to loaded bounds, visible-range rendering and exact full-range stats verified.
- [x] T013 — Add the adjacent coverage summary and accessible Zoom-to-activity/Back-to-full-range controls in `StateHistory.vue`; IDs: AC-10…12, FR-06, FE-04; depends: T012; done: one plot only, threshold <50 CSS px, `<1%` displayed honestly, full selected range remains explicit, and no network/device action.
- [x] T014 — Complete follow-up browser coverage and verify; IDs: AC-10…12, SC-02, NH-05, FE-04; depends: T012/T013; done: mocked eight-transition/four-minute burst at 6h/24h, 1h desktop overview and phone-adaptive policy, zoom/reset, exact summaries and full frontend checks recorded in `plan.md`.

## Review follow-up

- [x] T015 — Extract `domain/history/format.ts` with unit tests and replace every local duration/moment/coverage copy in the history chart family; IDs: AC-13, FR-07, FE-01; done: `src/__tests__/domain/history/format.test.ts` covers rounding, sub-second/sub-minute spans, `<1%`/`0%`/`Unknown` and axis labels; no duplicate formatters remain under `components/chart/history`.
- [x] T016 — Split the monolithic `StateHistory.vue` into `stateLayout.ts` (pure, tested), `StateHistorySvg.vue`, `StateHistoryChart.vue`, `StateHistoryCompact.vue` and an orchestrator; IDs: FE-01, FE-04, NH-04; done: `stateLayout.test.ts` covers lanes, gutters, dense/focus thresholds and tick collisions; orchestrator is ~220 lines and owns focus state only.
- [x] T017 — Remove duplicated transition scanning via `stateTransitions()`; IDs: NH-06, FE-01; done: `state.test.ts` asserts stats and transitions agree and that unknown time never invents a transition.
- [x] T018 — Harden interaction and testability: rAF-coalesced inspection, polite status announcement on zoom/reset, duplicate test-only `data-*` attributes removed, per-test browser fixtures with `route.fallback()`; IDs: AC-06, AC-10/11, FE-04, NH-05; done: `DeviceMetrics.spec.ts` asserts rendered summary text, announcements and the focus flow in both projects.
- [x] T019 — Fix `.vue` lint coverage and the pre-existing errors it exposed; IDs: NH-05, FE-04; done: ESLint parses TS SFCs; `npm run lint` reports 0 errors; `Range`, `Selection`, `MenuDropdown`, `EntityViewDialog` and `ExposeSelectionDialog` corrected with behaviour preserved (verified by the dialog/metrics browser suites).
