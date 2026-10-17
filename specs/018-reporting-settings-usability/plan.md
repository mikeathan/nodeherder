# Plan: Clear reporting settings and device category overrides

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-17
Constitutions: parent/backend/frontend v1.0.0 proposed.

## Context and design

Retain DeviceConfigurationSettings/useDeviceSettings ownership and Vuex save feedback.
Reuse liveIntervalByCategory and defaultDebounceByCategory on device configs; these
already serialize/validate/clone but were ignored for device resolution. Extend the
shared Go interval resolver and framework-free TypeScript resolver with category scope.
No schema/protocol changes. Specific field rules outrank categories at either scope.

A reusable ReportingRateControl owns only inherit/every/custom presentation, with
ReportingIntervalInput providing validated seconds input. Category rows use it directly;
sensor timing is under Advanced update timing. Main sensor rows contain threshold and
Reset adjustment. Sensor history adjustments have their own collapsed history section,
using separate row lists/reset patches so editing/resetting live never deletes history.
No new component framework, listeners, timers or transport.

Prefer progressive disclosure over extra explanatory paragraphs. Show current source
and effect beside relevant controls; inherited values are text, not editable-looking
placeholders. Existing configured overrides remain visible via summaries. Global page
uses the same controls with all-device scope. No fake default values are saved by merely
opening controls or adding rows. Custom values save through existing validated patches.

## Constitution Check

| Rule | Design evidence | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | AC-01–06 and tasks before edits | PASS | PASS |
| NH-02 | Raw automation untouched, zero defaults retained | PASS | PASS |
| NH-03 | Existing authenticated saves, validation unchanged | PASS | PASS |
| NH-04 | Shared interval resolvers and existing config maps | PASS | PASS |
| NH-05 | Companion unit/Go tests and browser checks | PASS | PASS |
| NH-06 | Reset only selected map entries; no resources | PASS | PASS |
| BE-01 | Policy owned in settings domain | PASS | PASS |
| BE-02 | Existing event/seed/state regressions plus category replay | PASS | PASS |
| BE-03 | No workers/locks added; affected race checks | PASS | PASS |
| BE-04 | History independent; old data/settings retained | PASS | PASS |
| BE-05 | No new route/access/assistant behavior | PASS | PASS |
| FE-01 | Shared Vuex/composable persistence | PASS | PASS |
| FE-02 | Same wire maps; frontend/backend precedence fixtures | PASS | PASS |
| FE-03 | Existing save/errors/offline state; no commands | PASS | PASS |
| FE-04 | Native selects/details, labels, 390/1440 layouts | PASS | PASS |
| FE-05 | No new secrets, auth paths or HTML | PASS | PASS |

## Contracts, data, rollout

Consumers: settings backend/cache, frontend domain/editor, mock server stores maps.
Device category entries now take effect rather than being ignored. Global field rules
retain priority. Missing means inherit; zero is explicit immediate. No migration needed.
Older versions ignore device category maps; upgrading both components enables controls.
Rollback preserves settings/readings, but stops applying new device category overrides.

## Traceability and verification

AC-01/02/04: reporting_policy_test.go, cache_test.go and reporting-settings.test.ts.
AC-03/05: SensorReportingSettings.spec.ts with mobile/desktop, reload, keyboard, empty,
invalid/offline and live/history isolation. Adapt existing assertions to relocated
controls without removing behavior coverage. AC-06: lifetime companion replay and full
backend suite/race. Run Go full tests/build/vet, settings/services races, Jest/lint/build,
affected browser suite and visual screenshots. No household devices. Final results (2026-10-17):

- `go test -count=1 -timeout=3m ./... -coverprofile=...`: PASS all packages.
- `go test -race -count=1 -timeout=3m ./models/settings ./internal/services`: PASS.
- `go vet ./...`, `go build ./...`: PASS.
- `npm test -- --runInBand --silent`: PASS 32 suites / 244 tests after final changes.
- `npm run build`: PASS after final changes; existing duplicate DeviceCard name and
  large-chunk warnings remain. `npm run lint`: PASS, seven existing script warnings.
- Full `npm run test:e2e`: 74/74 PASS. After the final history-inheritance wording and
  example classification correction, reran `npm run test:e2e --
  e2e/SensorReportingSettings.spec.ts`: 12/12 PASS, covering both viewport projects.
- Visual review: 1280px and 1440px desktop, 390px mobile; light and dark sensor views.
  Native keyboard focus/Tab, collapsed controls, explicit zero/reset, reload, invalid
  values and offline disabling verified. No horizontal overflow in browser assertions.
- Shared Go resolver / GetReportingPolicy / GetDebounce have 100% package coverage,
  CRAP scores 2 / 2 / 5 respectively. Added no workers, listeners, timers or dependencies.
- `git diff --check`: PASS. Reviewed final precedence, scope keys, save patches,
  reset isolation, category reconfiguration and unchanged raw automation path.

Visual review identified stale `docs/hub_state.json` target-distance diagnostic metadata
in the mock preview; corrected the example to measurement, matching spec 017/backend.
Kept legacy decoding regression fixtures in source-owned tests. The final browser test
asserts distance has immediate default timing and captures the dark mobile view.
Updated docs/setup.md and backend/docs/architecture.md. No tests were weakened to hide
failures; existing history/timeout/persistence checks were relocated to the separate
controls, retaining assertions and adding live/history reset isolation.
No household-device tests, commits, deployment or approval steps.

## Risks and completion

Avoid auto-copying inherited values into device config. Category rules can affect many
readings; scope text explicitly states this device/all devices. Config/actuator/event
semantics remain owned by existing publisher. No significant new boundary/dependency
requiring ADR. No exceptions claimed; work remains uncommitted.
