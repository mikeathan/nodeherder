# Tasks: Responsive binary history chart

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Draft

Each task maps to spec acceptance or a constitutional gate. Test ownership follows the source: state-domain regressions go in `frontend/src/__tests__/domain/history/state.test.ts`; browser coverage extends existing `frontend/e2e/DeviceMetrics.spec.ts` and `EntityViewDialog.spec.ts`. Do not add a feature-named test file.

## Preparation

- [ ] T001 — Capture the current contact-history baseline and review the 032 spec/plan against the latest accepted Option A prototype; IDs: NH-01, AC-01…09; done: baseline screenshot/geometry and no-overflow behavior recorded at 320, 390, 768 and 1280px, both themes; confirm full-history scope and compact-dialog exclusion before editing.
- [ ] T002 — Add failing pure-domain expectations to `frontend/src/__tests__/domain/history/state.test.ts` for width bins; IDs: AC-05/07, NH-05/06; depends: T001; done: tests reproduce incorrect/lost active duration, unknown coverage, transition count, or input mutation in the proposed binner where applicable.

## Coverage and implementation

- [ ] T003 — Implement a pure, O(N + W) interval-to-pixel-bin helper in `frontend/src/domain/history/state.ts`; IDs: AC-05/07, NFR-01; depends: T002; done: exact total active/known durations are conserved, unknown-only bins remain `null`, transitions are correctly counted across unknown boundaries, repeated/same-time reports add no fabricated duration, and source segments are not mutated.
- [ ] T004 — Render Option A in the non-compact branch of `frontend/src/components/chart/history/StateHistory.vue`; IDs: AC-01/02/04, FR-01/02; depends: T003; done: active/inactive labels and colors derive from existing expose helpers; contact lower lane uses 70%/72% plot positions; unknown stays hatched; compact branch and exact `stateStats()` summary remain unchanged.
- [ ] T005 — Add width-aware geometry, adaptive ticks and sparse/dense render selection in `StateHistory.vue`; IDs: AC-03/05/07, NFR-01/02; depends: T004; done: `ResizeObserver` is disconnected on unmount, the plot does not overflow at required widths, sparse segments are time-proportional, dense output is capped at `min(plotWidth, RENDER_LIMIT)`, and bins encode active share only over known time.
- [ ] T006 — Add sparse/dense pointer, tap and keyboard readouts with accessible name/focus and bounded hit-testing in `StateHistory.vue`; IDs: AC-06, FE-04; depends: T005; done: sparse segment lookup is logarithmic, dense lookup is direct by column, tooltip contents identify state/time/duration or share/coverage/count, focus is visible, and touch inspection does not suppress page scrolling.
- [ ] T007 — Extend the existing `frontend/e2e/DeviceMetrics.spec.ts` fixture/regressions for Option A; IDs: AC-01…08, SC-01; depends: T004…T006; done: assert active/inactive text, token colors, raised lower lane, responsive ticks, unknown hatch, pointer/tap/keyboard readout, no horizontal overflow at 320/390/768/1280 in both themes, and a 10,000-sample aggregation with exact summaries and bounded bins; prove no extra history query or device command.
- [ ] T008 — Preserve the compact entity dialog and surrounding history behavior; IDs: AC-08/09, FR-05; depends: T004; done: existing `frontend/e2e/EntityViewDialog.spec.ts` and DeviceMetrics empty/partial/sample/route assertions still pass; compact chart screenshot remains unchanged.

## Handoff

- [ ] T009 — Run frontend checks; gate: NH-05, FE-04; depends: T003…T008; done: record exact results for `npm test -- --runInBand`, `npm run lint`, `npm run build`, `npm run type-check:e2e`, and isolated desktop/mobile Playwright runs for `DeviceMetrics.spec.ts` and `EntityViewDialog.spec.ts`; no real household devices used.
- [ ] T010 — Inspect screenshots and complete the final rule/acceptance review; IDs: AC-01…09, NH-01…06, FE-01…05; depends: T009; done: record results at required viewports/themes, verify spec/plan links and no backend/contract/dependency drift, list unrun manual browser checks, and leave no unapproved exception.

Planning only: all tasks are open; no implementation or test execution is claimed.
