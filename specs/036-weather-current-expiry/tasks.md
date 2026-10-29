# Tasks: Weather near-now fallback and unit formatting

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: implemented

## Implementation

- [x] T001 — Add a pure near-now selector in `frontend/src/integrations/weather/headline.ts`
  (current → elapsed hour → next hour → today's daily → none, using `evaluated_at`).
  IDs: AC-01–05, FR-01, NFR-01/02. Done: 8 table tests pass, including the Met Office
  capture-17:59/current-17:00 case and the no-sentinel property.
- [x] T002 — Consume the selector in `integrations/weather/presentation.ts` and set the icon
  condition from the headline. IDs: AC-01–05, FR-02. Done: 5 presenter tests pass.
- [x] T003 — Update `components/integrations/WeatherResults.vue` to show the headline, hide
  current-only statistics when `current` is null, and explain which forecast is shown. Remove
  manual-refresh copy so automatic refresh later needs no change. IDs: AC-06, NFR-03.
  Done: type-check/build pass; presenter state covered.
- [x] T004 — Add `frontend/src/utils/units.ts` with `formatQuantity`, `formatWithUnit` and
  `formatDistance`; migrate weather and device expose formatting. IDs: AC-07–09, FR-03.
  Done: unit tests (7 cases) and the existing expose suite pass; no unit literals remain in the
  weather path.
- [x] T005 — Record the formatting rule in `.agents/rules/vue-engineer.md`: unit-bearing values go
  through `src/utils/units.ts`; no unit strings or manual concatenation at call sites.
  Done: rule added.

## Verification

- [x] T006 — `npm test -- --runInBand` in `frontend/`. Result: 90 suites, 513 tests passed.
- [x] T007 — `npm run lint`. Result: 0 errors, 28 pre-existing warnings (unchanged).
- [x] T008 — `npm run build` (includes type-check). Result: passed, built in 2.13s.

## Deferred

- Automatic provider refresh (spec 034 scheduling work). The selector already derives from the
  server's `evaluated_at`, so it needs no change when refresh becomes automatic.
- Migrating remaining legacy formatters outside weather/device exposes; the new rule applies to
  new and changed code.
