# Event processing: implementation and verification evidence

Date: 2026-10-08 | Baseline source: `26f48d08` | [Plan](plan.md) | [Tasks](tasks.md)

Historical execution record for authorized T002 (test characterization/reliability)
and scoped T006/T007 (controller validation/error reporting). No pending Q-01/02/03
policy is accepted by these results. Consolidates the former baseline/failure reports
and four selected-output text files without dropping commands, numeric measurements,
coverage, compatibility details or limitations. No raw household payload logs retained.

## T002: characterization and faster tests

### Changes and reasoning

- Removed 3.3 seconds of fixed sleeps from seven service tests. Their Update and
  configuration callbacks are synchronous, and debounce tests already advance a
  mock clock. Waiting in wall time added no verification.
- Replaced disable/enable wait groups with exact callback counts and state checks
  before/after the transition. Missing callbacks now fail immediately instead of
  waiting indefinitely; disabled updates must leave the expose unchanged. Explicit
  zero brightness debounce isolates configuration behavior from sampling policy.
- Kept debounce/metrics-only test devices online so those cases do not accidentally
  create real availability-monitor workers. Offline-to-online coverage remains in
  `TestDeviceLifetimeService_UpdateWithNewData` and controller availability tests.
- Made Seed test only its synchronous contract, avoiding a duplicate one-second
  offline wait. Added cleanup through the existing disable API for tests that start
  monitors. This cancels workers but does not implement or prove production joining;
  full lifecycle ownership remains T014.
- Corrected the separate offline test's timeout from 10 nanoseconds to 10 seconds
  and supplied old last_seen in Seed's payload. Previously Seed overwrote the field
  assigned by the test, so its comment did not describe its actual setup. The test
  retains the production one-second tick and uses a bounded five-second receive.
- Added `TestDeviceLifetimeService_SeedAndUpdateOutputSequence`: full seed payload
  vs changed-only update delta, automation-before-sampled-output order, fresh state
  under debounce, identical-state suppression, and measurement-only metric output.
  Callback data is copied when observed. No scheduler sleeps or network are needed.

Only `_test.go` files and SDD evidence changed. No queue capacity, retry, timeout,
shutdown, reconnect or initialization recovery policy was implemented.

### Existing coverage reused

Paths below are relative to `backend/`. These tests passed in this slice's suites;
they are characterization coverage, not proof of all hardening acceptance criteria.

| Behavior | Existing test evidence |
| --- | --- |
| Identical physical events; fresh state before debounce | `internal/services/device_lifetime_test.go`: `TestEntityValueAutomationReadSemantics` |
| Rapid physical actions and creation reentry | `internal/services/device_processor_test.go`: `TestDeviceProcessorRapidPhysicalActions`, `TestDeviceProcessorCreationOwnership`, `TestDeviceProcessorSeedCallbackCanConfigureRegistry` |
| Stale-then-new and identical-state ordering | `internal/controllers/hub-controller_test.go`: `TestReplayStaleThenNewPairEndsOnNewState` (2,000 trials; randomized seed logged), `TestReplayIdenticalPairBroadcastsOnce` (500 trials) |
| HTTP/manual ordering | Same controller file: `TestHTTPIngestionIsOrderedPerDevice`, `TestTriggerManualRunsOnDeviceLane`, `TestTriggerManualKeepsEngineErrors` |
| Per-key isolation, capacity and FIFO | `internal/lanes/lanes_test.go`: `TestTasksForOneKeyRunInSubmissionOrder`, `TestSlowKeyDoesNotBlockOtherKeys`, `TestFullLaneRejectsAndCountsWithoutDropping` |
| Dial arithmetic, repeated buttons, presets/clamping | `internal/automations/scenarios/dial_test.go`: `TestDialGoldenSequence`, `TestDialBurstEachEventUsesItsOwnValues` |
| Presence delay, cancellation, repeated false and illuminance gate | `internal/automations/scenarios/presence_test.go`: all four presence scenarios |
| Manual/event context isolation | `internal/automations/device_test.go`: existing concurrent manual/device evaluation tests |
| Schedule generation/window and disabled source | `internal/automations/engine_test.go`: `TestSchedulerPreservesEnabledWindowOnReorder`, `TestEngineExplicitSourceDisable`, `TestEngineInitiallyDisabledSource`, schedule replacement/rollback tests |
| Seed/disabled state, state dedupe, measurement filtering | `internal/services/device_lifetime_test.go`: Seed, config-disable, unchanged-state, noisy/debounce and metrics tests; new sequence characterization complements them |

Coverage gaps retained for subsequent slices: reconnect/late-token fault injection,
application join semantics, availability logical interleavings, initialization panic,
slow-browser recovery, and an explicit controller save-without-execution regression.
Engine save/update coverage alone does not prove that final controller invariant.
Existing real-time scheduler tests and presence negative-observation windows were
not shortened: they need proper completion seams before waits can safely be removed.

### Performance baseline

Host: Apple M4, Darwin arm64. Actual Go: **go1.26.1** (installed toolchain selected by
`go`; go.mod requests Go 1.24.0/toolchain go1.24.8). Default Info console logging,
GOMAXPROCS 10 as reported by the benchmark suffix. Ordinary build, five runs, three
1,000-message iterations/run across 20 devices. No concurrent test/benchmark job
was started during these measurements. Harness and benchmark source were unchanged.

Run from `backend/`:

```sh
go test ./internal/controllers -run '^$' -bench '^BenchmarkMixedBurst$' -benchtime=3x -count=5
```

Exact measurements: [benchmark table](#benchmark-measurements). Median T002 p95
3.253167 ms; 2,325,245 bytes/burst; 47,279 allocations/burst.

All runs reported zero burst goroutine growth and 21 idle hub goroutines. These
figures do not prove teardown cleanup or production capacity; fake Paho, mock WS
observation and test storage exclude important real-I/O costs. The existing harness
does not yet join all owned resources. Future same-environment comparisons should
use the unchanged fixture or record a new baseline if harness lifecycle changes.

There is no production performance before/after claim: only tests changed. The
benchmark was captured before edits and did not require rerunning afterwards.

### Test speed and checks

Seven directly comparable tests previously reported 1.00, 0.31, 0.31, 0.40, 0.40,
0.56 and 0.36 seconds (3.34 seconds total). Each reports 0.00 seconds after removing
unnecessary waits. This is reporting precision, not literally zero work. The old
targeted package run took 3.58 seconds; the new broader targeted run took 1.257 seconds,
including the retained one-second offline test and added characterization. Those
package totals cover different selections and are not a strict speedup ratio.

All commands below ran from `backend/` and passed:

```sh
go test -count=1 -timeout=5m ./...
go vet ./...
go test -race -count=1 -timeout=5m ./internal/lanes ./internal/controllers ./internal/services ./internal/automations/...
go test -race -count=20 -timeout=2m ./internal/services -run 'TestDeviceLifetimeService_(Seed|SeedAndUpdateOutputSequence|UpdateWithSameData|UpdateWithDebouncer|MetricsAvailabilityWithMetricsEnabled|UnchangedValuesNeverConsumeDebounce|NoisyValuesDebouncedCorrectly)$|TestOnConfigUpdated_ShouldDisableDevice'
```

Before timing command:

```sh
go test -count=1 -json -timeout=2m ./internal/services -run 'TestDeviceLifetimeService_(UpdateWithSameData|UpdateWithDebouncer|MetricsAvailabilityWithMetricsEnabled|UnchangedValuesNeverConsumeDebounce|NoisyValuesDebouncedCorrectly)$|TestOnConfigUpdated_ShouldDisableDevice'
```

After timing command:

```sh
go test -count=1 -json -timeout=2m ./internal/services -run 'TestDeviceLifetimeService_(Seed|SeedAndUpdateOutputSequence|ShouldChangeAvailability_ToOffline|UpdateWithSameData|UpdateWithDebouncer|MetricsAvailabilityWithMetricsEnabled|UnchangedValuesNeverConsumeDebounce|NoisyValuesDebouncedCorrectly)$|TestOnConfigUpdated_ShouldDisableDevice'
```

Package results and targeted timings are retained in [historical check outputs](#historical-check-outputs).
Full and race suites ran concurrently after isolated timing/benchmark measurements;
their durations should not be used as a package-speed comparison. Go cache access
used approved elevated commands. No household devices were commanded.

Unrun: separate `go build ./...` (test-only source changes; packages compiled by tests),
frontend tests/lint/build (no frontend changes), full-repository race suite (focused
event packages and 20 repetitions selected), live-device checks (not required).
`gofmt` and `git diff --check` passed. Document links and requirement references were
checked. Q-01/02/03 remain open; T002 completion grants no acceptance of those policies.

## T006/T007: validation and output errors

Authorization: user accepted this scoped slice with “ok go” on 2026-10-08.
Status: complete for FR-05/09, AC-11 store/broadcast subset and AC-12/16.

Scope: return/log controller callback errors; attempt broadcast and device-store
effects independently in their existing order; decode bridge envelopes with typed
success data and validate required identifiers before effects; return invalid-device
JSON errors. Preserve MQTT commands, successful notification payloads/order and
permit-join transaction behavior. Valid bridge response status values are `ok` and
`error`; missing/unknown status and missing/empty/wrong-type required success fields
are malformed. Error responses do not require success data. Unknown fields remain
accepted. Public BridgeResponse fixture constructor remains compatible.

No whole-event retry, async output stage, counters infrastructure, publication-token
handling, subscription reconciliation, HTTP status change or pending Q-01/02/03
policy is introduced. T007's T005 dependency is narrowed for this slice to existing
logging; stage counters remain T005, and publication outcomes remain T008. These
portions of AC-11 are not claimed complete. No external schema changes are required.

Implementation sequence: demonstrate failure tests; add shared internal typed bridge
decoder and operation-specific validation; retain contextual errors using wrapping/
joining; report device callback failures at controller wiring; test independent
outputs, successful response compatibility and unchanged event/automation traces.

Before implementation rule check: NH-01 spec/plan/tasks plus this scoped amendment;
NH-02 command/state/ordering logic retained; NH-03 malformed input rejected before
effects, raw bridge-response debug payloads removed; NH-04 established boundaries
and small decoding helper; NH-05 fake dependencies/regression-before-fix; NH-06 no
new resources/retries or storage meaning. BE-01 transport validation stays in
controllers; BE-02 lifetime semantics unchanged and existing scenarios rerun; BE-03
no goroutines/timeouts added; BE-04 write failures surfaced without migration;
BE-05 no network/auth expansion. These were supported pre-implementation design checks; actual execution follows.
Frontend rules are N/A for this backend-only compatible-output slice. The preceding
checks describe the pre-implementation assessment; actual results follow below.

Planned checks (backend): full `go test -count=1 -timeout=5m ./...`, `go build ./...`,
`go vet ./...`, affected controller/service/automation race suites, repeated focused
failure regressions, five-run existing burst benchmark against T002. Tests use
in-memory dependencies only; household commands are prohibited. Rollback is reverting
the slice, with no data migration. Pending broader policies remain pending.

### Implementation and regression evidence

Changed backend files: `internal/controllers/handlers.go`, `hub-controller.go`, new
`bridge_response.go`, and source-matching `handlers_test.go` /
`hub-controller_internal_test.go` (originally collected in `failures_test.go`). The prior T002 test changes remain
in the worktree and are independent of this slice.

- Shared internal envelope decoding validates status and unmarshals typed success
  fields. Rename validates both names before unsubscribing; remove/interview require
  a nonempty ID. Failure envelopes accept missing or unrelated success data; unknown
  fields are ignored. The exported BridgeResponse test helper remains unchanged.
- Broadcast/store errors are wrapped and joined, preserving causes for errors.Is.
  Both effects run once in the original broadcast-then-store order. Seed, update,
  measurement and availability callbacks report failures through existing logging.
  Independent output failures do not interrupt the remaining sampled effects.
- Bridge notification and renamed-device emission failures now return to the routing
  error boundary. Remove-store failures still notify the browser and retain both
  error causes if notification also fails. Permit-join remains transaction-driven;
  a response without a transaction does not create or execute a request.
- Invalid device JSON and null return errors rather than reporting success or
  reaching a nil payload processor. Empty objects/unknown exposes retain their
  existing handling. Null bridge logging is a harmless no-op instead of a panic.
- Removed raw response-payload debug logging in touched operation handlers. Existing
  successful/negative WS notification payloads are checked, including the existing
  rename-error text `error`. No command publication or retry path was added.

Pre-fix focused regressions failed as expected: lost broadcast/store causes;
rename unsubscribed then panicked on missing/numeric `to`; null interview panicked;
missing identifiers silently succeeded; invalid-device JSON returned nil; and
rename/interview/remove output errors were discarded. The first regression run
used the original production code and exited 1. Its selected reproduction output
is summarized here rather than retaining raw payload logs in the repository.

### Verification results

All following commands passed, run from `backend/` on Go go1.26.1 Darwin arm64:

```sh
go test -count=1 -timeout=5m ./...
go build ./...
go vet ./...
go test -race -count=1 -timeout=5m ./internal/controllers ./internal/services ./internal/automations/...
go test -race -count=20 -timeout=2m ./internal/controllers -run 'TestDeviceOutputFailures|TestBridgeResponsesReject|TestDeviceHandlerReturns|TestBridgeResponseEffects|TestDeviceCallbackFailures|TestBridgeResponseNotification|TestBridgeLogging'
go test ./internal/controllers -run '^$' -bench '^BenchmarkMixedBurst$' -benchtime=3x -count=5
```

The full/focused race suites include existing stale/new replay, identical events,
manual ordering, dial/preset and presence-delay scenarios. The new callback test
injects failures on seed and update, verifies five contextual error reports and
the exact metrics/broadcast/store sequence, and leaves command publication unwired
so any attempted command execution would fail the test. New output tests also check
success, individual failures, combined errors and retained successful effect order.
The injected callback test disables automation: it proves output continuation and
error reporting, not absence of duplicate physical commands when enabled automation
and output failures coincide. T026 closes that gap and the deferred controller
save-without-execution regression. Existing happy-path automation scenarios remain
useful evidence but do not replace those combined failure cases. No live producer
capture was used to validate bridge status vocabulary; compatibility evidence is
repository contracts/fixtures, not a household broker experiment.

Package results: [historical check outputs](#historical-check-outputs).
Full and race suites ran concurrently; timing is not a package-speed comparison.
`gofmt`, whitespace and local document-link checks passed.

Five-run unchanged burst benchmark on the same Apple M4/Go toolchain, with default
Info logging and no competing test jobs: median p95 **2.853125 ms** versus T002's
3.253167 ms; median **47,229 allocations/burst** versus 47,279; median 2,322,456
bytes/burst versus 2,325,245. The +10% no-regression thresholds were met. All runs
reported zero burst goroutine growth and 21 idle hub goroutines. These small local
mocked measurements do not establish a production speed improvement or validate
shutdown cleanup. Exact measurements: [benchmark table](#benchmark-measurements).

No household devices or live brokers were contacted. Frontend suites were unrun
(backend-only changes, compatible event payloads); full backend race detection
beyond the affected event packages was unrun. Stage counters, publication-token
failures, reconnect/timeout/recovery policies and fresh-browser synchronization
remain outside this implementation; full FR-05/AC-11 is not marked complete.

After-design/final rule check: the supported NH-01–06/BE-01–05 design checks above
remain satisfied for this slice, backed by regression, build/vet/race and unchanged
scenario evidence. No new external entry point, dependency, goroutine, timer or
persistence format was introduced. No constitutional exception or approval of
broader Q-01/02/03 policies is asserted. Actual maintainer code review remains pending.

## Benchmark measurements

Both stages: `BenchmarkMixedBurst`, five runs × three iterations; each iteration is
1,000 messages/20 devices. Units below are nanoseconds, bytes and allocation counts
as emitted by Go. `ns/op`, bytes/op and allocations/op are per burst; p50/p95 are
reported per-message latency. Every run: zero burst goroutine growth, 21 idle hub
goroutines. Setup logs were omitted; both commands passed (T002 package 3.539 s,
T007 3.515 s). Mocked I/O and incomplete harness join preclude production-capacity
or leak-free claims.

| Stage | Run | ns/op | p50-ns | p95-ns | B/op | allocs/op |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| T002 | 1 | 3193722 | 1365917 | 2715666 | 2323714 | 47287 |
| T002 | 2 | 4110972 | 1735250 | 3359375 | 2336197 | 47279 |
| T002 | 3 | 4048708 | 1871500 | 3253167 | 2311349 | 47242 |
| T002 | 4 | 4386319 | 2072708 | 3662291 | 2325245 | 47279 |
| T002 | 5 | 3688667 | 2099375 | 2957125 | 2330210 | 47360 |
| T007 | 1 | 3039319 | 1305667 | 2389750 | 2308672 | 47209 |
| T007 | 2 | 3579375 | 1638209 | 2991708 | 2315840 | 47222 |
| T007 | 3 | 4047778 | 1759333 | 3459750 | 2330250 | 47264 |
| T007 | 4 | 2413250 | 1071500 | 2392333 | 2322456 | 47229 |
| T007 | 5 | 3417847 | 1806625 | 2853125 | 2330402 | 47269 |

## Historical check outputs

All listed results passed. Seconds are execution records, not comparable performance
measurements: full and race suites ran concurrently after isolated benchmark/timing
runs. Package paths omit `node-herder/`. `—` means not selected in that stage.
T002 vet passed; separate build unrun (test-only changes). T007 build and vet passed.

| Package | T002 full | T007 full | T002 race | T007 race | T002 repeat-race | T007 repeat-race |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| internal/api | 1.256 | 1.549 | — | — | — | — |
| internal/auth | 0.179 | 0.235 | — | — | — | — |
| internal/automations | 22.708 | 23.771 | 24.984 | 24.041 | — | — |
| internal/automations/scenarios | 0.784 | 0.799 | 1.931 | 1.918 | — | — |
| internal/controllers | 18.151 | 18.130 | 19.216 | 19.058 | — | 1.295 |
| internal/fs | 0.316 | 0.310 | — | — | — | — |
| internal/lanes | 0.218 | 0.214 | 1.285 | — | — | — |
| internal/mcp/intent | 0.200 | 0.154 | — | — | — | — |
| internal/mcp/resolver | 0.162 | 0.144 | — | — | — | — |
| internal/mcp/resources | 0.175 | 0.153 | — | — | — | — |
| internal/mcp/server | 0.162 | 0.161 | — | — | — | — |
| internal/mcp/timescope | 0.163 | 0.142 | — | — | — | — |
| internal/mcp/tools | 0.154 | 0.155 | — | — | — | — |
| internal/mcp/transport/http | 0.458 | 0.462 | — | — | — | — |
| internal/metrics/query | 0.158 | 0.161 | — | — | — | — |
| internal/metrics/services | 0.154 | 0.151 | — | — | — | — |
| internal/metrics/storage | 11.256 | 10.873 | — | — | — | — |
| internal/mqtt | 0.156 | 0.160 | — | — | — | — |
| internal/ratelimiter | 0.174 | 0.183 | — | — | — | — |
| internal/services | 5.574 | 5.474 | 6.813 | 6.820 | 1.343 | — |
| internal/ws | 0.683 | 0.656 | — | — | — | — |
| mocks | 0.211 | 0.214 | — | — | — | — |
| models/assistant | 0.150 | 0.147 | — | — | — | — |
| models/devices | 0.177 | 0.195 | — | — | — | — |
| models/settings | 0.256 | 0.327 | — | — | — | — |
| repository | 2.597 | 2.585 | — | — | — | — |
| store | 10.730 | 10.699 | — | — | — | — |
| utils | 0.605 | 0.589 | — | — | — | — |
| utils/storage | 4.574 | 4.351 | — | — | — | — |

No test files in either stage: `node-herder`, `internal`, `internal/mcp/protocol`, `internal/metrics/domain`, `models/automations`, `models/bridge`, `models/hub`, `models/logging`, `testing`, `testing/hubharness`.
Compiled but no matching tests in either stage: `internal/mqtt`.

### Targeted service test timings

Seconds at Go test reporting precision; `0` means below that precision. Different
before/after selections make package totals unsuitable for a strict speedup ratio.
Test names below omit `TestDeviceLifetimeService_` unless starting `TestOnConfig`.

| Test | Before | After |
| --- | ---: | ---: |
| UpdateWithSameData | 1 | 0 |
| UpdateWithDebouncer | 0.31 | 0 |
| MetricsAvailabilityWithMetricsEnabled | 0.31 | 0 |
| TestOnConfigUpdated_ShouldDisableDevice | 0.4 | 0 |
| TestOnConfigUpdated_ShouldDisableDevice_OnStartUp | 0.4 | 0 |
| UnchangedValuesNeverConsumeDebounce | 0.56 | 0 |
| NoisyValuesDebouncedCorrectly | 0.36 | 0 |
| PACKAGE | 3.58 | 1.2570000000000001 |
| SeedAndUpdateOutputSequence | — | 0 |
| Seed | — | 0 |
| ShouldChangeAvailability_ToOffline | — | 1 |

## Documentation consolidation and test organization

2026-10-08: merged baseline/failure reports and their four text outputs here; merged
runtime recommendations and review into plan.md. Retained spec/plan/tasks/ADR roles;
removed eight redundant files. All ten benchmark rows, package/race results, targeted
timings, reproduction details, commands and known gaps remain above.

Moved `TestDeviceLifetimeService_SeedAndUpdateOutputSequence` into
`backend/internal/services/device_lifetime_test.go`. Split the seven controller
failure tests into source-matching `handlers_test.go` (five) and
`hub-controller_internal_test.go` (two). The internal suffix preserves access to
private callbacks; existing `hub-controller_test.go` uses the external package.
Shared failure fakes remain package-local. No assertions or production code changed.

Initial split compilation identified missing/unused imports; corrected before final
checks. All final commands passed from `backend/`:

```sh
go test -count=1 -timeout=5m ./...
go vet ./...
go test -race -count=20 -timeout=2m ./internal/controllers ./internal/services -run 'TestDeviceOutputFailures|TestBridgeResponsesReject|TestDeviceHandlerReturns|TestBridgeResponseEffects|TestDeviceCallbackFailures|TestBridgeResponseNotification|TestBridgeLogging|TestDeviceLifetimeService_SeedAndUpdateOutputSequence'
```

Repeated race suites: controllers 1.415 s, services 1.308 s. Full suite includes
controllers 17.902 s and services 5.485 s; timings are records, not speedup claims.
No new benchmark, frontend suite, separate build or full-repository race run: only
test organization/docs changed; prior T007 build/benchmark evidence remains above.
Document links/anchors, task uniqueness/dependencies, requirement references,
benchmark row preservation, gofmt and whitespace checked during consolidation.

Final document checks passed: 35 links/anchors, 26 unique tasks with acyclic
dependencies, all requirement references, ten benchmark rows and eight relocated
tests retained. `gofmt -l` returned no files; `git diff --check` passed.

## Review follow-up fixes

2026-10-08, after review of the T006/T007 slice:

- `EmitDevice` (`backend/internal/ws/eventhub.go`) now returns its broadcast error.
  Previously only device-lookup failures reached the rename handler, so the earlier
  claim that renamed-device emission failures return was only partly true. Covered by
  `TestEmitDeviceReturnsBroadcastError` (new `eventhub_internal_test.go`, deterministic
  fake server; it failed before the fix).
- `TestDeviceCallbackFailuresAreReportedWithoutRetry` counted every error-level log
  from the process-wide logger hook, so unrelated background errors from other tests
  could break its exact-count assertion. It now counts only this device's
  `device <id> output failed` reports. Assertions are otherwise unchanged.

Not changed: possible error logs after rename if `bridge/devices` arrived after the
rename response (both share the ordered bridge lane, so not expected), and one error
log per message under persistent store/broadcast failure (shutdown ordering is T013).

All passed from `backend/`: `go build ./...`, `go vet ./...`,
`go test -count=1 -timeout=5m ./...`,
`go test -race -count=1 ./internal/ws ./internal/controllers ./internal/services`,
`go test -race -count=20 ./internal/controllers -run TestDeviceCallbackFailuresAreReportedWithoutRetry`.
Changed Go files are gofmt-clean. No household devices or brokers were contacted.

## T026 and MQTT topic mapping

2026-10-08. Authorization: after review, the user asked for T026 and the confirmed
unsubscribe-prefix bug to be done in this branch, following the repository rules.

### T026: enabled-automation output failures and save-without-execution

Test-only. `backend/internal/automations/scenarios/output_failures_test.go`:

- `TestDialCommandsUnchangedWhenOutputsFail` runs the real dial automation through
  the hub with every broadcast and every device/metrics write failing. It asserts the
  same golden commands as `TestDialGoldenSequence`, from the same `dialGoldenCases`
  table (moved out of that test so both use one baseline). No extra, missing or
  reordered command; broadcasts are still attempted (the rig waits on them); writes
  are still attempted. Automation reads in-memory state, so failed persistence does
  not change the next step's brightness basis.
- `TestSavingAutomationDoesNotExecuteIt` saves the dial automation through the
  controller's UI save handler, then delivers one dial event: exactly that event's
  command is published. Scope: it catches a command published by the save itself
  or routed through the hub's lanes, not a timer the save might start.

`testing/hubharness` gained `EventHub.FailBroadcasts`, `EventHub.SaveAutomation`, a
`Store` wrapper with `FailWrites`/`FailedWrites`, and `AutomationConfigPayload`
(sharing the config reader with `LoadAutomationConfig`). Existing harness users only
call `Store` methods, so the wrapper is compatible.

These pass on unchanged production code, so each was mutation-checked. Making a failed
broadcast skip the store write failed the first test ("no device or metrics write was
attempted"). Publishing a command from the save handler failed the second (published
`[map[state:ON] map[brightness:110]]`). Both mutations were reverted.

### MQTT topic mapping (FR-02 subset, from T009/T010)

Confirmed bugs in `backend/internal/mqtt/mqtt.go`, each reproduced by a failing test
in `mqtt_test.go` before the fix:

- `RemoveTopic` unsubscribed `<name>` while `subscribe` used `zigbee2mqtt/<name>`, so
  renamed devices stayed subscribed at the broker. `TestRemoveTopicUnsubscribesTheSubscribedTopic`.
- Inbound topics used `ReplaceAll`, so a friendly name containing `zigbee2mqtt/` was
  mangled (`garage/zigbee2mqtt/sensor` became `garage/sensor`). `TestInboundTopicStripsOnlyTheBasePrefix`.
- Every client's topic list aliased the package `bridgeTopics` array, which
  `RemoveTopic` edits in place; a later client then subscribed `bridge/logging` twice
  and never `bridge/devices`. Production creates one client (`internal/hub.go`), so
  this affected only multi-hub tests. `TestRemoveTopicLeavesOtherClientsBridgeTopics`.

Fix: `fullTopic`/`sanitizeTopic` are the single mapping used by subscribe, unsubscribe,
publish and receive (`TrimPrefix`), and each client clones the bridge topics, as the
plan's MQTT design specifies. `mocks.FakePahoClient.Unsubscribe` now removes the exact
filter and `Subscribed()` reports current filters; previously unsubscribe was a no-op,
so no test could observe this bug. Registry locking, reconciliation, bounded waits and
retry remain T010 (Q-01). Topics, QoS, retain flags and payloads are unchanged.

### Pre-existing WS test flake

The first full run failed `TestLoadMCPStatusMessage` ("melody instance is closed",
bad handshake); it passed in 90 isolated runs. melody 1.2.1 sets its hub open from a
goroutine started by `New`, so a dial made immediately after construction can be
refused under load. It predates this branch and does not affect production (clients
connect long after startup). The shared `NewTestWsServer` helper now retries only
`websocket.ErrBadHandshake`, for at most two seconds, because melody exposes no
readiness signal to wait on.

### Verification

All passed from `backend/` (Go go1.26.1, Darwin arm64):

```sh
go build ./...
go vet ./...
go test -count=1 -timeout=5m ./...   # three consecutive runs
go test -race -count=1 ./internal/mqtt ./internal/automations/... ./internal/controllers ./internal/services ./internal/ws
go test -race -count=20 ./internal/mqtt
go test -race -count=20 ./internal/mqtt ./internal/automations/scenarios -run 'TestRemoveTopic|TestDialCommandsUnchanged|TestSavingAutomation'
go test -race -count=3 ./internal/ws
```

Changed Go files are gofmt-clean. Unrun: frontend checks (no frontend change) and
live-broker or household checks (none permitted or required).

### Benchmark

`BenchmarkMixedBurst` uses `hubharness`, so the harness wrapper changed the NFR-01
fixture, and topic mapping runs on every receive and publish. A single five-run batch
gave median p95 4.198 ms, but comparing against hours-old T007 numbers mixes in machine
drift, so a same-conditions A/B was run instead. Base was branch head `b142dcc7` in a
temporary worktree, alternated with this change, three rounds of the recorded
five-run command each, with no other jobs.

| Side | Median p95 (15 runs) | Round medians p95 | allocs/burst | B/burst |
| --- | ---: | --- | ---: | ---: |
| Base `b142dcc7` | 3.395209 ms | 3.617959 / 3.316125 / 3.369834 ms | ~47,270 | ~2,324,000 |
| This change | 3.550667 ms | 3.497667 / 3.486417 / 3.593084 ms | ~46,240 | ~2,307,000 |

p95 +4.6%, within the +10% threshold; single runs ranged 2.28–5.65 ms on the base
side alone. Allocations fell about 1,030 per 1,000-message burst (−2.2%), because
`TrimPrefix` returns a substring where `ReplaceAll` allocated a string per received
message. Every run: zero burst goroutine growth, 21 idle hub goroutines. Mocked I/O;
no production throughput claim.
