# Plan: Meaningful sensor reporting and automation routing

Spec: [spec.md](spec.md) | Status: Complete | Author: Codex | Updated: 2026-10-16
Constitutions: parent/backend/frontend 1.0.0 proposed

## Context and design

`DeviceLifetimeService.collectUpdate` already separates raw delta, live output and
sampled metrics. Add a pure `models/devices/observation_filter.go` last-published baseline
policy, owned by one lifetime service on its device lane. Filter read-only numerics
only; use generic configured threshold and history maps resolved in settings cache.
Real Z2M cached payloads include unchanged writable settings. Track one protected,
nonserialized confirmation deadline per entity value, registered by the existing UI
command handler; broadcast unchanged writable echoes only for an active request.
No timer is needed: the next report consumes or expires the deadline. The existing command window moves to `reporting-defaults.json` compatibility data; saved defaults/device overrides are resolved by the browser and server. Validate command
id/name/write access before marking interest. Seed primes baselines; do not discard raw values or change automation/metrics deltas.
No timer/buffering: meaningful changes publish immediately; below-threshold values
remain available through raw state. UI sees a significance-filtered projection.

`Device` automation compiles conservative input names from primary triggers and
condition DependsOn declarations against configured source schema. The derived set is
immutable for a published generation, never persisted. Engine selection shares it;
service uses an optional input-aware query so existing querier adapters remain valid.
Unknown/unconfigured implementations fall back to current enabled behavior. Actual
trigger.AffectedBy remains authoritative for event replay and condition semantics.

Optional reportingThresholds map is copied in configuration snapshots, validated before
cache mutation/persistence, and edited through one shared settings component using the
existing hub configuration transport. New read-only distance schemas follow generic measurement classification; explicit `metricOverrides` can enable history for any diagnostic field, including existing distance entities.
No metric records are rewritten. No additional household data committed.

Alternatives: whole-device Z2M throttle loses presence events; UI-only filtering retains
server/network overhead; auto-learned smoothing obscures thresholds and adds state.
An explicit absolute deadband is sufficient for observed jitter. No architectural
replacement/ADR needed: existing raw/observation separation is retained.

## Constitution Check

| Rule | Evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec, observations, tasks before source edits | PASS | PASS |
| NH-02 | Raw state/events/confirmations preserved; routing includes conditions | PASS design | PASS |
| NH-03 | Validated finite numeric overrides; no new endpoint/auth changes | PASS design | PASS |
| NH-04 | Existing layers/contracts; derived index not second recipe truth | PASS design | PASS |
| NH-05 | Replay/companion/race/performance/UI verification | PASS design | PASS |
| NH-06 | Native units, unchanged retention, zero/remove recovery, bounded state | PASS design | PASS |
| BE-01 | Pure domain policy, services and settings ownership | PASS design | PASS |
| BE-02 | Existing repeated events/seed/disabled/echo/raw automation tests retained | PASS design | PASS |
| BE-03 | No new workers/timers; immutable published routing, lane-owned baselines | PASS design | PASS |
| BE-04 | Existing history plus distance, metrics opt-in remains | PASS design | PASS |
| BE-05 | Existing authenticated settings routes, no assistant changes | PASS design | PASS |
| FE-01 | Shared editor and Vuex config ownership | PASS design | PASS |
| FE-02 | Optional map shared Go/TS/mock contract | PASS design | PASS |
| FE-03 | Actual report remains required; configurable shared deadline; reset clears pending work | PASS design | PASS |
| FE-04 | Named numeric inputs/keyboard/mobile/desktop | PASS design | PASS |
| FE-05 | No household data persisted in browser | PASS design | PASS |

## Contracts, data, rollout

Optional `DeviceConfig.reportingThresholds` map returned by settings/hub snapshots and
accepted by existing saveDeviceConfigOverride/Defaults commands. Empty/missing inherits;
zero publishes every change. Old clients/records remain valid. Rollback ignores map;
remove entries to restore defaults. History inclusion only from configuration onward with metrics enabled. Schema units unchanged; previous diagnostic-distance metadata supported.

## Traceability and verification

| IDs | Paths/tests | Procedure | Result |
| --- | --- | --- | --- |
| FR-01, AC-01–05 | observation_filter_test.go; device_lifetime_test.go; device_test.go | Jitter/drift/units/events/echo/raw/metrics replays | PASS |
| FR-02, AC-06–08 | automations/device_test.go, engine_test.go; device_lifetime_test.go | Dependency routing, replacement, mixed state/event messages; benchmark | PASS |
| FR-03, AC-04/09 | settings/cache_test.go, UI editor companion tests, mock browser | Precedence/copy/persistence/invalid writes; keyboard/viewport | PASS |
| SC-01/NFR-01 | all touched code | Full Go/Jest, race relevant packages, vet/build/lint; CRAP comparison | PASS |

## Debt, risks, exceptions

Existing debounce is a leading-edge metrics sampler and may miss the final suppressed
metric followed by duplicates. Its semantics are documented, not silently replaced.
Per-device category/global expose debounce precedence is separate legacy debt; keep
this change's threshold precedence explicit. Meaningful rapid changes remain immediate;
this is not a hard packet-rate guarantee. No distance tolerance is preselected: configure it in native units after observing changing distances. No exception or approval invented.

## Completion

Implementation and final review complete. All changes remain uncommitted, including
pin wording. No household commands/settings were changed. Only read-only Z2M UI
observation used the live installation; all verification used mocks or pure code.

## Configuration amendment and consolidation

User requested no field-specific operational policy in code. Remove the builtin switch
and name-based metric exception. `models/settings/reporting_policy.go` now owns shared typed
resolution/validation, with optional maps and timeout. Compatibility JSON is one source
of the initial timeout used by backend, frontend and mock fixture hydration; persisted
configuration is authoritative. No observation filtering occurs until configured.
`domain/reporting-settings.ts` owns frontend policy input validation and timeout resolution.
The shared editor owns add/remove/edit interactions for thresholds, history and timeout.
Source-owned activity/commands companion files replace the old combined test file without
dropping assertions. Command pending timers clear on hub reset, important for longer
configured windows. Legacy payloads retain a compatibility default, and removing device
entries restores global inheritance. Invalid settings never publish to active cache.

## Verification and final review (2026-10-16)

- Full backend: `go vet ./...`, `go build ./...`, and
  `go test -count=1 -timeout=3m ./... -coverprofile=/private/tmp/nodeherder-sensor-after.cover`
  passed. Full-suite regression assertions retained. Relevant race run passed:
  `go test -race -count=1 -timeout=3m ./internal/services ./internal/automations/... ./internal/controllers ./models/devices ./models/settings`.
- Frontend: 32 Jest suites / 239 tests passed. Shared settings adapter has 100%
  statement/branch coverage; reporting validation/resolution has 100% statements and
  92.85% branches. `npm run lint` passed with seven pre-existing script warnings;
  `npm run build` passed with pre-existing duplicate DeviceCard/large chunk warnings.
- Targeted mock browser checks cover sensor settings, arbitrary field names, history
  inclusion, timeout edits, removal/inheritance, single-save patches, command feedback,
  pin wording, keyboard and overflow at desktop 1280×800 and mobile 390×844. Screenshots
  reviewed. Final run: 16/16 passed in 54.7s.
- Routing benchmark with 100 configured triggers: 30.01–30.14 ns/op,
  0 B/op and 0 allocs/op (Apple M4). This measures dependency lookup only, not MQTT,
  database or end-to-end latency. The 10,000-message noisy mixed-payload replay emits
  zero updates/automation callbacks after baseline; relevant mixed presence still passes.
- Installed CRAP tool with actual Go coverage: collectUpdate complexity/CRAP 7 → 6;
  Device.Evaluate stays 5 / 5.01. New observation filter Accept CRAP 8 (100% coverage),
  routing helpers ≤5, policy resolver 2 and settings validation 12 (100% coverage).
  All new functions stay below 30. Baseline coverage ran the four affected packages;
  final coverage ran the full backend, so no whole-repository aggregate reduction claimed.
- Review corrected an early over-broad writable-echo change: actuator regressions remain
  unchanged; only cached configuration echoes require explicit pending interest. The
  old live-versus-metrics controller regression now explicitly disables significance
  filtering and marks requested config echoes; assertions were not weakened. A browser
  test initially raced lazy loading; it now waits for the editor/Create Override branch.
  Intermediate compile errors and sandbox cache/server binding failures were corrected;
  they are not reported as passing runs. No test expects a known new production bug.
- Final readability pass renamed new files to observation_filter, trigger_dependencies,
  reporting_policy and SensorReportingSettings; added comments on ownership, raw versus
  published values, configuration precedence, and safe fallback for unknown extensions.
- Remaining limits: no hard live rate cap, no automatic noise learning, no changing-distance
  trace observed; the legacy metrics sampler still has no trailing flush. Numeric deltas
  affect UI/activity only; raw values needed by configured conditions remain immediate.
  Observation thresholds start unset and must be configured through the UI. Old clients
  omitting optional maps remain accepted. Unrelated complete Playwright suite was not run.


Follow-up: [spec 015](../015-reporting-build-boundary/plan.md) corrects the isolated
frontend build boundary and notification snapshot ownership found in independent
review. That reviewer reported 72/72 full browser tests passing; the results above
remain the checks actually run during this implementation. Current verification is
recorded separately in spec 015.
