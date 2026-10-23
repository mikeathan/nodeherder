# Tasks: History preset, sidebar simplification and appearance palette

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete (automated + measured checks)

## Preparation

- [x] T001 — Record spec 031 scope and supersessions (030 preset count, drawer destinations);
  IDs: FR-01…FR-04; done: `spec.md` written and linked from the plan.
- [x] T002 — Confirm single sources (`HISTORY_PRESETS`, `NAV_ITEMS`, `THEMES`/`ACCENTS`); gate:
  FE-01/NH-04; done: design section in `plan.md` names paths and rejected alternatives.

## Coverage and implementation

- [x] T003 — Add `3d` (`72`) preset to `HISTORY_PRESETS`; IDs: FR-01/AC-01…02; done:
  `src/domain/history/range.ts` updated.
- [x] T004 — Extend `src/__tests__/domain/history/range.test.ts` with a 3d preset + 72h window case
  and run the history suite; IDs: AC-02/AC-03; done: `npx jest src/__tests__/domain/history`.
- [x] T005 — Remove the `Device list` entry from `NAV_ITEMS` and fold `DeviceList`/`DevicePage` into
  the `Devices` `includes` list; IDs: FR-02/AC-05; done: `src/components/shell/navigation.ts`.
- [x] T006 — Update `e2e/layout.spec.ts` (drop the link, add a `toHaveCount(0)` guard, keep 44px
  and focus checks); IDs: AC-04/AC-06; done: 6 passed.
- [x] T006a — Update the Period option count in `e2e/DeviceMetrics.spec.ts` (8 → 9) and assert the
  `3d` option after CI caught the stale count; IDs: AC-01; done: that spec passes on desktop and
  mobile projects (10 passed).
- [x] T007 — Replace `#0f8bd6` with `#c08a1e` and `#8a5a2b` in `ACCENTS`; IDs: FR-03/AC-07;
  done: `src/theme/theme-settings.ts`.
- [x] T008 — Add the `amber` descriptor (plus the now-rejected sunset/diner, later removed);
  IDs: FR-04/AC-08; done: `src/theme/theme-settings.ts` lists 8 themes.
- [x] T009 — Add/keep token blocks in `src/assets/styles/themes.css`; IDs: AC-09/NFR-02;
  done: amber light+dark kept, sunset/diner blocks removed.
- [x] T009a — Review round 2: remove `teak`, `avocado`, `plum`, `sunset`, `diner` from `THEMES` and
  `themes.css`, and drop accent `#8a5a2b`; IDs: AC-07/AC-08/AC-10; done: 8 themes / 8 accents,
  tests updated, `npm test` 345 passed, lint/build clean.
- [x] T009b — Add the approved replacements (Crimson, Emerald, Blueprint; accents `#c92a2a`,
  `#2b4fd8`, `#495057`) and remove the mustard accent; IDs: FR-03/FR-04/AC-07…09; done: 11 themes /
  10 accents, token-completeness test green, measured accent contrast 4.89–10.6, 0px overflow at
  320px, Jest 345 passed, lint/build clean, `e2e/layout.spec.ts` 6 passed.
- [x] T010 — Extend `src/__tests__/theme/theme-settings.test.ts` for the accent set (nine, no
  `#0f8bd6`) and new theme ids; IDs: AC-07/AC-10; done: Jest theme suite.

## Handoff

- [x] T011 — Run `npm test -- --runInBand`, `npm run lint`, `npm run build`; gate: NH-05/FE-04;
  done: Jest 53 suites / 345 tests passed; lint 0 errors / 7 pre-existing warnings; build success.
- [x] T012 — Run `npx playwright test e2e/layout.spec.ts` (desktop + phone projects); IDs:
  AC-04…06; done: 6 passed (desktop + mobile projects), 0 skipped.
- [x] T013 — Manual light/dark review of the three retro themes at 320/1280px (picker wrap, text
  and accent contrast); IDs: AC-08/AC-09/NFR-02; done: measured text 12.7–16.1, muted 6.0–7.6,
  accent 4.82–10.6 for the new themes; light accents darkened to reach ≥4.8; 0px overflow at 320px.
- [x] T014 — Update `plan.md` traceability/completion, verify every NH/FE rule row, and report
  unrun checks; gates: NH-01/NH-05; done: plan traceability updated with actual results; other
  browser specs left unrun and recorded as such.

Add discovered work; update spec/plan for scope changes.
