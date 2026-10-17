# Plan: Category-based live reporting

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-16
Constitutions: parent/backend/frontend v1.0.0, proposed.

## Context and design

Existing `settings/cache.go` contains a history-only leading-edge DeviceDebouncer.
`DeviceSettings.GetDebounceForEntity` duplicates cache lookup; unused public helpers
and UI `debounce` mappings have no production consumers. `DebounceEditor.vue` retains
and mutates its initial props object. Settings containers render raw config keys.
`ObservationFilter` owns numeric publication baselines; reuse it for delta semantics.

Add `liveIntervalByCategory` and `liveIntervalOverrides` to DeviceConfig. Retain old
history/threshold fields, avoiding a migration to a second competing policy schema.
One ordered resolver handles device field → global field → category → fallback.
History uses the same device-field → global-field → global-category ordering. Existing
category and device-field values retain meaning; global history field rules become
editable through the common editor.
Hydrate missing category live defaults from shared reporting-defaults.json on load.
All new intervals validated before save, maps cloned, explicit zero preserved.

[ADR 001](adr-001-live-publication.md) records the lifecycle decision.
An ObservationPublisher in services composes the numeric filter and interval gate.
It owns a mutex, latest pending value per known field, and at most one AfterFunc timer
for the earliest deadline. Input updates replace pending values without resetting the
deadline. Emission is serialized by that owner; timer callbacks never mutate raw
entity values or run automation. Events invoke existing output/repository boundaries.
Context cancellation and device disable/offline clear pending state. Configuration
changes re-evaluate retained values under new policy without requiring another report.
No separate mutable sensor model and no whole-device throttle. Both filter and
publisher remain independently testable with existing utils.Clock/MockClock.

Frontend keeps Vuex/useDeviceSettings ownership and one shared editor. Explicit
containers expose device enabled/history controls and use consistent section cards.
Show global category live/history settings and per-field exceptions, inherited effective
values and units. Separate command timeout into Advanced. Replace and remove obsolete
three debounce editors plus their settings registry when references are eliminated.
Keep no parallel visible editor for the same policy. Existing field editor browser
assertions are updated only to reflect the new accessible UI structure, retaining
save/reload/zero/removal coverage.

## Constitution Check

| Rule | Design evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec AC-01–08, tasks before edits | PASS | PASS |
| NH-02 | Independent output; event/confirmation bypass | PASS | PASS |
| NH-03 | Existing authenticated WS/config paths only | PASS | PASS |
| NH-04 | Resolver/domain, publisher/services, Vuex/UI boundaries | PASS | PASS |
| NH-05 | Fake clocks, companions, browser/race checks | PASS | PASS |
| NH-06 | Bounded pending maps/timer; no storage rewrite | PASS | PASS |
| BE-01 | No transport business logic | PASS | PASS |
| BE-02 | Raw/seed/events/availability regressions retained | PASS | PASS |
| BE-03 | Publisher mutex/context cancellation, fake-clock races | PASS | PASS |
| BE-04 | Legacy history behavior/units preserved | PASS | PASS |
| BE-05 | No changed access/routes/assistant contracts | N/A | N/A |
| FE-01 | Shared editor, Vuex sole saved state | PASS | PASS |
| FE-02 | Optional fields synchronized with fixtures/types | PASS | PASS |
| FE-03 | Validation and connection feedback; no false saved claim | PASS | PASS |
| FE-04 | Shared theme/buttons, 390/1280 viewport checks | PASS | PASS |
| FE-05 | Existing transport, no persistence/secrets added | PASS | PASS |

## Contracts, rollout, recovery

Additive optional config maps; Go JSON, TS types, fixture hydration and copy/save paths
updated together. Existing files remain readable and keep history meaning. New UI
requires updated backend; previous backend ignores live fields and rollback loses
live filtering until upgraded. No automatic device commands. Changes uncommitted.
Snapshot live defaults are authoritative for UI, with shared bootstrap for old fixtures.

## Traceability and verification

AC-01/02/08: reporting_policy/config/cache companion tests for precedence, explicit
zero, invalid intervals, legacy load, snapshot isolation and failed saves.
AC-03–05: observation_filter, observation_publisher and device_lifetime companions;
latest trailing flush, cumulative threshold, return-to-baseline, multiple deadlines,
no starvation, disable/offline/config/cancellation, repeated events and automation.
AC-06/07: reporting-settings/useDeviceSettings unit tests and source-owned browser
spec at 390x844 and 1280x800, keyboard/errors/inheritance/reload/no-overflow/screenshots.
Commands planned: full Go test/vet/build; affected race tests; Jest/lint/build;
Playwright reporting/panel/layout/control; complexity comparison with coverage.
Actual checks and limits are recorded below. No routine tests ran against household devices.

## Risks and debt

Output callback ownership must remain safe against existing availability callbacks;
new timer state is mutex owned and cancellation checked inside that lock. Existing
history has no trailing flush; preserve and explain its separate behavior here.
Existing classification whitelist remains outside this change. Category filters must
never apply to discrete alarm/battery-low values merely because they are diagnostic.
No new architectural dependencies or exceptions.

## Verification results and final review

- PASS: `go test -count=1 -timeout=3m ./... -coverprofile=/private/tmp/nodeherder-category-after.cover`.
- PASS: `go test -race -count=1 -timeout=3m ./internal/services ./internal/controllers ./models/settings ./models/devices`.
- PASS: `go vet ./...` and `go build ./...` on final backend changes.
- PASS: `npm test -- --runInBand`: 32 suites, 242 tests.
- PASS: `npm run lint` (seven existing script warnings) and `npm run build`
  (existing duplicate DeviceCard/large bundle warnings).
- PASS: final `npm run test:e2e -- e2e/SensorReportingSettings.spec.ts`: ten
  desktop/mobile reporting checks, 10.9 seconds. Save/reload, category/field precedence,
  explicit zero, removal, non-saving draft rows, keyboard focus, invalid-input feedback,
  disconnection, independent history and no overflow all verified against mock services.
- PASS: twelve existing desktop/mobile controls/layout/panel checks in the broader
  browser run. That run initially had two reporting-selector failures; both resolved
  in the final ten-test reporting run. The complete unrelated browser suite was unrun.
- PASS: visual review of desktop category defaults and mobile field exceptions from
  final browser screenshots under frontend/test-results/. Native-unit labels, readable
  inheritance hints, theme colors, 44px controls and stacked mobile fields verified.
- PASS: `node --check frontend/tools/server/state.mjs` and `git diff --check`.

Regression evidence includes an initially failing derived-config ownership test:
legacy interval pointers were shallow copied. The shared clone now copies those
values as well; the original assertions pass unchanged. An older distance replay
explicitly disables live intervals to retain its original threshold/history focus;
its assertions are unchanged, with new fake-clock diagnostic interval integration
coverage added separately. Browser failures were scoped locator ambiguity (two
similar inheritance labels, then an ancestor-qualified has locator), corrected by
scoping to the actual live input. No tests were weakened to conceal production bugs.
Final lifecycle review also preserves pending readings across ordinary configuration
edits: they are re-evaluated against new policy instead of being silently discarded.

Complexity: `/Users/mikeathan/go/bin/crap` with full Go coverage, baseline from spec
012's final report. NewDeviceConfigFrom complexity/CRAP 3/4.13 → 1/1; duplicate unused
GetDebounceForEntity (complexity 4, CRAP 20) removed. ObservationFilter.Accept 8/8 →
3/3 with reusable Allows 5/5. Collector remains 6/6. New publication methods max
CRAP 9.05; interval/threshold validation CRAP 16 with 100% statement coverage. All
new reporting functions are below 30. Reports: /private/tmp/nodeherder-category-crap.json
and /private/tmp/nodeherder-sensor-crap-after.json. No claim of a whole-repository
score reduction or household CPU/throughput benchmark. The thousand-reading fake-clock
regression verifies one timer allocation for an unchanged pending deadline.

Final gates: all applicable rules PASS with the evidence above. Application protocols,
authentication and raw automation semantics are preserved; additive policy values
and mock defaults match. Old history configuration remains intact. No deployment,
household configuration writes, commit or push performed. Restart the native launcher
to rebuild the backend before trying the new behavior. Architecture and setup docs
explain category defaults, override priority, latest-value flush and separate history.


Follow-up: [spec 015](../015-reporting-build-boundary/plan.md) corrects the isolated
frontend build boundary and notification snapshot ownership found in independent
review. That reviewer reported 72/72 full browser tests passing; the results above
remain the checks actually run during this implementation. Current verification is
recorded separately in spec 015.
