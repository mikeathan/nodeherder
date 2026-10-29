# Spec: Weather near-now fallback and unit formatting

ID: 036 | Status: Implemented; checks passed | Owner: NodeHerder contributors | Updated: 2026-10-29
Constitution: proposed 1.0.0 | Components: frontend

## Problem, scope, exclusions

A weather Home card can show `Unavailable · Unavailable` while its forecast is still present.
Reproduction: use Met Office weather, refresh at HH:59, wait until the next hour. The card keeps
its high/low, rain chance and checked time, but the headline becomes unavailable on every device.

Cause: `WeatherForecast.Project` deliberately drops `current` one hour after the observation, and
`current()` falls back to the latest hourly point. Met Office hourly observations are top-of-hour,
so a refresh at 17:59 yields a 17:00 observation that expires at 18:00. The Home card headline is
composed only from `current`, and both `weatherValue` and `weatherLabel` return `Unavailable` for
a missing value, so a fresh snapshot prints the sentinel twice.

Scope: a capability-owned near-now selector, its use in the Home card and results view, and one
shared unit formatter so quantities are never formatted with unit literals at call sites.
Exclusions: automatic provider refresh (deferred), backend projection/contract changes, and
changing the existing one-hour current validity rule.

## Scenarios and acceptance

### US-01 — A missing current observation still shows usable weather (P1)

Independent verification: `src/__tests__/integrations/weather/headline.test.ts` and
`presentation.test.ts`.

- AC-01: Given a projected resource with `current: null` and an hourly point at or before
  `evaluated_at`, the Home card shows that point's temperature and condition and labels it
  `Observed HH:MM`.
- AC-02: Given no elapsed hour, it uses the next hourly point and labels it `Next HH:MM`.
- AC-03: Given no current or hourly data, it uses today's daily summary and labels it `Today`.
- AC-04: Given no weather data at all, the card says `No saved forecast` once, never
  `Unavailable · Unavailable`.
- AC-05: A stale snapshot still shows a value and reports `Needs refreshing`.
- AC-06: The results view shows the same headline, hides the six current-only statistics when
  `current` is null, and states why ("No current observation. Showing observed 17:00.").

### US-02 — Units have one owner (P1)

- AC-07: Temperature, percent, speed, precipitation, distance and other known quantities render
  through `src/utils/units.ts`; no weather call site passes a unit string.
- AC-08: A missing quantity renders the formatter's explicit placeholder, never a fabricated zero.
- AC-09: Device exposes keep their existing rounding and spacing while using the same formatter.

## Requirements and success

- FR-01: `weatherHeadline` MUST select current → elapsed hour → next hour → today's daily → none
  using `evaluated_at`, not the browser clock (AC-01–05).
- FR-02: The Home presenter and results view MUST consume that one selector (AC-01–06).
- FR-03: `formatQuantity`/`formatWithUnit` MUST own unit symbols, rounding, spacing and the
  missing placeholder (AC-07–09).
- NFR-01: No backend or wire-contract change; old saved snapshots remain readable.
- NFR-02: The selector is pure and bounded: one pass over the hourly array, no I/O, no timers.
- NFR-03: The fix must work unchanged when automatic refresh lands; only the resource's
  `evaluated_at`/content moves forward, and no UI copy assumes a manual action.
- SC-01: Full frontend suite, lint, type-check and build pass; no new warnings or errors.

## Entities and external interfaces

No resource or API shape changes. `WeatherForecast` gains no field. The near-now value is a
presentation projection of existing `current`/`hourly`/`daily`. The unit formatter is a frontend
utility consumed by weather and device presentation.

## Compatibility and rollback

Presentation-only. Old snapshots without a valid current render the fallback. Reverting the change
restores the previous card behaviour; no migration or stored data is touched.
