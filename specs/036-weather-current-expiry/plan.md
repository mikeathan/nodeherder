# Plan: Weather near-now fallback and unit formatting

Spec: [spec.md](spec.md) | Status: Implemented; checks passed | Updated: 2026-10-29
Constitution: parent/frontend 1.0.0 (proposed)

## Context and design

`WeatherForecast.Project` intentionally removes `current` one hour after the observation. The
Home card and the results view treated that as "no weather", even though `hourly` and `daily` were
intact. The fix keeps projection semantics and adds one presentation selector:

- `src/integrations/weather/headline.ts` — `weatherHeadline(forecast)` chooses current → latest
  elapsed hour → next hour → today's daily → none, using the server's `evaluated_at`. It returns
  the value, condition, source and a label (`Now`, `Observed 17:00`, `Next 18:00`, `Today`).
  `weatherHeadlineValue` composes one string and never joins two sentinels.
- `src/integrations/weather/presentation.ts` — the Home card uses the headline and its condition.
- `src/components/integrations/WeatherResults.vue` — the summary panel uses the same headline;
  current-only statistics are hidden when `current` is null, with one explanatory line.
- `src/utils/units.ts` — one owner for quantity formatting. `formatQuantity(value, kind)` covers
  temperature, percent, speed, precipitation, pressure, power, energy, illuminance, UV, AQI,
  pollutant and distance. `formatWithUnit(value, unit)` handles data-provided units. Both own
  rounding, spacing, missing placeholders and non-finite values.
- `.agents/rules/vue-engineer.md` — new rule requiring unit-bearing values to use that utility.

No backend, repository, API or resource-contract change. Automatic refresh later only updates the
stored resource; the selector already derives from `evaluated_at`, and the "Manual refresh" copy
was removed so the UI stays truthful either way.

## Alternatives

| Option | Rejected because |
| --- | --- |
| Keep `Current` past its validity window | Presents stale observations as current (NH-06) |
| Add a `headline` field to the wire contract | More surface for a presentation-only bug; not needed for automatic refresh |
| Fix only the Home card | Duplicates the fallback in the results view and panel |
| Widen the one-hour window | The observation can already be nearly an hour old at capture; still fails |
| Format units per component | The original bug; unit literals drift and missing values diverge |

## Constitution Check

| Rule | Evidence | Result |
| --- | --- | --- |
| NH-01 | Spec 036 with reproduction, acceptance and plan | PASS |
| NH-02 | No device pipeline, command or automation change | N/A |
| NH-03 | No new network entry, secret or household-data exposure | N/A |
| NH-04 | Capability-owned selector; generic cards keep `value`/`detail`/`status`; no API change | PASS |
| NH-05 | Selector, presenter and formatter tests; full suite/build recorded | PASS |
| NH-06 | Missing data stays missing, no zero substitution; stale labelled; no mutation of stored data | PASS |
| FE-01 | Pure module + presenter; no competing state | PASS |
| FE-02 | No contract change; existing resource fields only | PASS |
| FE-03 | Observed/next/today/stale/no-data states are distinct | PASS |
| FE-04 | Existing labels and status semantics retained; no new control | PASS |
| FE-05 | No secret or household data added to browser storage | N/A |

## Traceability and verification

| Requirement | Paths | Test/procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01–05 | `frontend/src/integrations/weather/headline.ts` | Table tests: observation, elapsed hour, next hour, daily, empty, no sentinel | PASS (8 cases) |
| FR-02, AC-01–06 | `presentation.ts`, `WeatherResults.vue` | Presenter tests: expired current, stale, empty, icon condition, observation | PASS (5 cases) |
| FR-03, AC-07–09 | `frontend/src/utils/units.ts`, `domain/exposes.ts` | Unit tests for temperature, percent, speed, precipitation, distance, missing, dynamic units; existing expose suite | PASS (49 tests across 5 suites) |
| SC-01 | `frontend/` | `npm test -- --runInBand`; `npm run lint`; `npm run build` | 90 suites / 513 tests passed; lint 0 errors (28 pre-existing warnings); build passed |

No live provider, household device or browser run was used. The Met Office top-of-hour case is
covered synthetically (capture 17:59, current 17:00, evaluated 18:00).

## Risks and completion

Risk: the selector could hide a genuinely empty provider response. It only falls forward when a
point exists; with no data it returns `No saved forecast`. Risk: unit refactoring could change
device output; `formatNumber` keeps the previous `toFixed` rounding and the expose suite passes.

Documentation: the frontend rule records the formatting convention. No ADR is needed for a
presentation-only fix.
