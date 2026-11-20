# Tasks: Wireless viewer legibility

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

Each task maps to requirement/acceptance IDs or a constitutional gate. Mark done only with
evidence.

## Preparation

- [x] T001 — **DONE**: ADR-001 recorded; Q1–Q4 resolved with the maintainer (2026-11-20), Q5
  recorded in the plan. IDs: FR-01…FR-08.
- [x] T002 — **DONE**: no contract or backend change; `classification` is already on every
  identity and absent values are handled as Learning. Recorded in `plan.md` §4.

## Coverage and implementation

- [x] T003 — **DONE**: red tests in `classes.test.ts` (new) and `presentation.test.ts`; IDs:
  FR-01, FR-03, FR-04, FR-05.
- [x] T004 — **DONE**: `frontend/src/domain/observations/classes.ts` plus `classLabel`,
  `matchesIdentityQuery` and `seenLabel` in `presentation.ts`. 13 class tests, 7 new
  presentation tests. IDs: FR-01, FR-03, FR-04, FR-05, NFR-01.
- [x] T005 — **DONE**: composable tests for presets, counts, default-hiding override, search,
  composition and volatile-only stability. IDs: FR-02, FR-07, AC-06…AC-11, AC-18…AC-20.
- [x] T006 — **DONE**: `useObservationViewer.ts` gained `classFilter`, `search`,
  `classOptions`, `now` and a single ordered path. IDs: FR-02, FR-07, NFR-02.
- [x] T007 — **DONE**: new `ObservationRow.vue` + `ObservationList.vue`; `ObservationGroupList`
  reuses the row; `ObservationCards.vue` deleted; generated `components.d.ts` updated. IDs:
  FR-05, FR-06.
- [x] T008 — **DONE**: `ObservationFilters.vue` gained a labelled search field and ordered
  Class presets with counts; `WirelessViewer.vue` wires them and the empty message. The
  misleading empty-facets line was removed. IDs: FR-01, FR-07, FR-08.
- [x] T009 — **DONE**: `e2e/wireless.spec.ts` locators updated to the unified row, the section
  test asserts class sections, and four new scenarios cover presets, demoted reveal, search
  and the no-match state. The triage test now targets its row by name instead of position,
  because ordering is deliberately independent of arrival. IDs: AC-01…AC-23.

## Handoff

- [x] T010 — **DONE**: `npm test -- --runInBand` 154 suites / 1030 tests passed; `npm run lint`
  0 errors (23 pre-existing warnings); `npm run build` passed (vue-tsc + vite); `npm run
  type-check:e2e` passed. Gate: NH-05.
- [x] T011 — **DONE**: `npx playwright test e2e/wireless.spec.ts` 74/74 passed across the
  desktop and mobile projects (37 each), including the no-overflow checks. IDs: FR-08,
  AC-22, AC-23.
- [x] T012 — **DONE**: self-review found and fixed one a11y defect (Clear button nested in the
  search `<label>`); re-ran lint, build and the full browser suite. No template-side sorting,
  no duplicated class-to-label mapping (labels come from `CLASS_PRESENTATION`), no stale
  comments. No commit or PR created.

## Follow-up: filter layout and stability (2026-11-20)

- [x] T013 — **DONE**: one labelled-field anatomy in `ObservationFilters.vue`; fieldset
  legends carry an explicit margin because they are not flex items. IDs: FR-09, AC-24.
- [x] T014 — **DONE**: phone fields collapse behind `Filters (N)`; search stays visible; the
  Radar/List switch is a full-width row on phones; distance bands wrap as pills on mobile.
  IDs: FR-10, AC-25, AC-26, AC-29.
- [x] T015 — **DONE**: every class group is always rendered in fixed order with a
  `min-width: 3ch` count slot; zero counts disable a chip. IDs: FR-11, AC-27.
- [x] T016 — **DONE**: mapped wireless components from undefined legacy token names to
  `--nh-*` tokens and removed light fallbacks; `tokens.test.ts` now guards the folder.
  IDs: FR-12, NFR-06, AC-28.
- [x] T017 — **DONE**: unit suites 154 / 1032 passed; lint 0 errors (23 pre-existing
  warnings); build and e2e type-check clean; wireless browser suite 80/80 on desktop and
  mobile, including new geometry assertions for alignment, chip stability and no mobile
  clipping. Gate: NH-05.

## Follow-up: reading space (2026-11-20)

- [x] T018 — **DONE**: List is the default view; Radar is a separate view with no truncated
  side rail. The flat `ObservationList.vue` was deleted. IDs: FR-13, AC-30, AC-31, AC-34.
- [x] T019 — **DONE**: clearing a selection restores the view it was opened from
  (`lastNonFocusView`). IDs: FR-14, AC-32.
- [x] T020 — **DONE**: wide rows show `via <source>` and a signal sparkline;
  `sparklinePoints()` is pure and tested; `SignalSparkline.vue` draws nothing without a trend.
  IDs: FR-15, NFR-07, AC-33.
- [x] T021 — **DONE**: full gate re-run — 154 suites / 1038 unit tests, lint 0 errors, build
  and e2e type-check clean, wireless browser suite 80/80 (desktop + mobile). The radar-specific
  tests now open the Radar view explicitly. Gate: NH-05.

Add discovered work here; update spec/plan when scope changes.
