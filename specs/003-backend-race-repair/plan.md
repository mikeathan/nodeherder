# Plan: Scoped CI repair

Spec: [spec.md](spec.md) | Updated: 2026-10-02

Keep mapper RWMutex around map mutation/lookups; repository loading remains outside
the lock. Keep DB Close only on initialization failure. No automation, scheduler,
wire-format, schema or dependency changes. Regression tests use immutable mock bridge
metadata and temporary DB files. No performance claims or deployment.

NH-01: narrowed scope recorded. NH-02/BE-02: no automation/event changes.
NH-03/BE-05: no network/auth changes or household data.
NH-04/BE-01: existing repository boundaries/dependencies preserved.
NH-05: targeted regressions, vet/build and ordinary suite; results below.
NH-06/BE-03: map ownership synchronized; failed initialization releases DB handle.
BE-04: unchanged storage format/driver. No approvals/ratification invented.

## Deferred findings for review

Full detector previously reproduced automation enable-state reads/writes; timer
start/stop/callback ownership; entity-data reads/writes in automation operations;
device availability monitoring versus initialization; mock automation-cache map
access. Legacy Bolt v1.3.1 also crashes under checkptr. These are investigation
findings, not authorization to change runtime behavior. Full race CI remains unresolved.

## Verification

Test-only follow-up: operation cases model successive confirmed device responses,
so execute each step on the test goroutine and wait for mock response completion
before the next calculation. Keep numeric/bounds assertions and verify exact
delivery counts, including no extra command at limits. Trigger tests await response
completion and stop printing live mutable action objects. Permit-join tests use
explicit callback signals and synchronized test counters; do not change runtime
state access to manufacture passing race results. Defer production enable/entity/
availability/configuration races. Repeated focused and ordinary/full-race evidence.

Low-risk runtime follow-up: a plain mutex around only the two WebSocket map writes;
no session/network/logging operations under it. Use atomic.Bool for the logger's
enable flag: no new lock/callback boundary, disabled check remains at Fire entry,
already accepted deliveries may finish after disabling as before. Emitter setup
remains outside this scoped change. Verify existing WS/utils suites and repeated
focused concurrent regressions, vet/build/ordinary/full race. Defer service/state/
action changes. No timing or performance guarantees; live Z2M validation not run.

Bolt follow-up: pin bbolt v1.4.3 (Go 1.23 minimum), change the two driver imports
only and resolve required transitive dependencies. Add legacy compatibility cases
to existing storage/repository test files, preserving all existing assertions.
Generate the fixture with Bolt v1.3.1 before replacement; use only temporary copies.
See [storage ADR](adr-storage.md). Verify local/repository-toolchain tests, Linux
cross-build and Ubuntu execution if available; never equate cross-build with tests.
NH-04/06 and BE-04: preserve storage layout, test old-file read/write/reopen, require
stopped-service database backup before rollout and restore backup for rollback.
Other constitutional checks above apply; no authentication/device/timing changes.

Mock follow-up: use local mutexes for cache/clock/MQTT state, snapshot handlers at
publish time, and track async responses for explicit test cleanup. Clock callbacks
run after releasing locks; timer registration/reset is deduplicated. Add one mock
test file (no existing mock test file) covering concurrency, callback reentrancy,
reset/stop, cache ordering and response completion. No runtime/dependency edits.

Passed after the targeted revert: `go vet ./...`, `go build ./...`,
`go test -count=1 -timeout=5m ./...` (includes failed-init reopen regression),
`go test -race -count=20 -timeout=2m ./repository -run TestMapperConcurrentConfigureUpdateResolve`,
`go test -race -count=20 -timeout=2m ./utils`, and `git diff --check`.
Local environment: macOS/arm64 Go 1.26.1; Ubuntu CI and repository-toolchain parity
are not established by local checks. Full race suite was not rerun after the revert;
known unresolved failures above remain. Acceptance evidence covers retained scope only.

Mock follow-up results: `go test -race -count=20 -timeout=2m ./mocks`,
`go vet ./...`, `go build ./...`, `go test -count=1 -timeout=5m ./...`, and
`git diff --check` passed (macOS/arm64 Go 1.26.1).
Affected domain rerun (`go test -race -count=1 -timeout=3m ./internal/automations
./internal/controllers ./internal/services`) still failed: enabled state, delayed
actions, configuration/execution overlap, entity values, availability/last-seen,
and test coordination; controllers/services also hit the legacy Bolt pointer crash.
No mock cache, clock, or handler-state race reports appeared in this rerun;
job-construction/mock-clock reports disappeared too. This is observed coverage,
not a claim all runtime races are fixed. Bolt remains v1.3.1; no production changes.

Bolt follow-up completed after the above mock-only snapshot: imports now use
bbolt v1.4.3; old Bolt dependency removed. Required testify 1.10.0 and x/sys 0.29.0
versions resolve bbolt's module requirements; repository Go baseline unchanged.
Legacy fixture tests passed against old Bolt before replacement and then under
bbolt (including race detection). Existing tests/assertions/CI flags unchanged;
only new cases and required imports added to existing repository/storage test files.

Passed: module verification, vet, native build, Linux/amd64 CGO-disabled cross-build,
all uncached ordinary tests, five race runs of `./utils/storage ./repository
./internal/metrics/storage` (Go 1.26.1/macOS), and those storage race packages with
`GOTOOLCHAIN=go1.24.8`. Full uncached race suite completed without the Bolt checkptr
crash but still reports runtime races; exact failing packages are recorded below.
Ubuntu execution unverified: local Docker daemon is unavailable. No production DB
access or deployment. Backup/rollback requirements remain in the ADR.

Remaining full-race failing packages: `internal/api`, `internal/automations`,
`internal/controllers`, `internal/services`, `internal/ws`, `store`. Besides
previously reported state/action/test issues, execution now exposes DeviceProcessor
service-map configuration versus creation, RemoteHook enable versus log delivery,
bridge-permit-join configuration versus test reads, and history-config writes versus
MetricsCleanupTask reads. These are deferred findings, not repaired in this scope.
Full log: `/private/tmp/nodeherder-bbolt-full-race.log` (local diagnostic artifact).

At user request, removed the standalone `legacy-bolt.db.gz.base64` and its README.
The same synthetic bytes are embedded in the existing test helpers, preserving
legacy read/write/reopen coverage and assertions; no tests disabled or skipped.
Reran `go test -race -count=1 -timeout=5m ./...`: storage/repository/metrics-storage
and mocks pass; the same six runtime packages above still fail, with no Bolt crash.
Log: `/private/tmp/nodeherder-race-after-fixture-removal.log`.

Low-risk runtime follow-up results: 20 focused race runs (concurrent 16-client
connect/broadcast/disconnect, logger toggling/delivery and callback disabling),
vet/build and all ordinary tests passed on macOS/arm64 Go 1.26.1. Existing assertions
remain intact. New test cleanup initially called Melody.Close twice and timed out;
removed only that redundant test shutdown, then repeated checks passed. Production
lock scope reviewed: one map write, no nested locks/network/callbacks; atomic logger
flag needs no lock. Emitter configuration and automation/service races remain out
of scope. Full race run now passes internal/ws and utils; API, automations,
controllers, services and store still fail. RemoteHook enable/Fire report disappeared.
No live Z2M/performance/Ubuntu guarantees; no deployment or physical-device access.
Also passed: five complete affected-package race runs (`./internal/ws ./utils`)
with `GOTOOLCHAIN=go1.24.8`; `git diff --check` clean. Production edits remain
limited to websocket.go and logger.go; pre-existing staged changes preserved.

Test-only follow-up results: operation/trigger cases pass five race repetitions
on both Go 1.26.1 and 1.24.8; permit-join rejection passes ten race repetitions.
Vet/build and all ordinary backend tests passed; final affected ordinary suites
also passed after command-limit guard additions. Arithmetic/bounds assertions
retained; explicit counts now catch extra commands. Producer goroutines/sleeps
were removed only from sequential confirmed-response calculation tests. Test
callback counters use atomics, errors are local, and completion is registered
before requests. No production edits in this follow-up; genuine concurrent
service/automation tests remain unchanged.

Permit-join expiry race check still fails at AppConfigCache.SaveBridgePermitJoin
versus the test's configuration read; no test lock/skip was added to conceal it.
Full race run still fails API, automations, controllers, services and store;
operation index/value fixture races and mutable-action diagnostic races no longer
appear in that run. Log: `/private/tmp/nodeherder-test-sync-full-race.log`.
Remaining production/configuration/timing issues need separately reviewed changes;
passing isolated calculations is not proof of safe live concurrent Z2M operation.

FR-07 design (review refinement): snapshot IDs under the existing registry RLock
for defaults, then call configureDeviceLifetime for each ID after unlocking.
Both paths use its protected current-service lookup; release before OnConfigUpdated.
This supersedes the initial pointer snapshot, preserving the original shared flow.
Registry read locks are owned by private getDeviceIDs/getDeviceLifetime helpers;
configuration and update callers use them, with callbacks outside their locks.
Safety finding: creation holds the write lock while Seed invokes callbacks.
Adding configuration read locks can deadlock if a Seed callback synchronously
updates configuration. This is a potential re-entry path, not a reproduced live
incident. User authorized creation/callback ownership follow-up: construct and Seed
outside the registry lock, then lock only publication. Preserve publish-after-Seed
ordering. Configuration snapshots exclude in-flight seeds; concurrent same-device
creation/update and lifecycle fields remain separate unresolved ownership issues.
Regression uses disabled services (no timers/state transitions), concurrent creation
and immutable default/override configurations. Run repeated focused races, vet/build,
ordinary tests and the full race suite; report unrelated failures. NH-01/05 and
BE-03 apply; no transport/storage/timing/device semantics change (other rules N/A).

FR-07 results: new registry/configuration and Seed callback re-entry regressions
pass ten race repetitions each on Go 1.26.1 and 1.24.8 (macOS/arm64). Initial new
callback fixture had no exposed entries and was rejected; corrected its payload,
not production validation. Vet/build/all ordinary backend tests pass. Full race
suite still fails API, automations, controllers, services and store; registry map
reports are absent in this run, while lifecycle/configuration/action races remain.
Logs: `/private/tmp/nodeherder-registry-{focused,go124-final,ordinary-final,full-race-final}.log`.
Only device_processor.go changed in production for this follow-up; existing tests
and assertions remain unchanged, with new cases appended in the existing file.
No deployment/live devices/Ubuntu execution; no claim of complete live safety.
Also passed: five race repetitions of all processor/configuration tests
(`-run 'TestDeviceProcessor|TestOnDeviceConfigUpdated'`); `git diff --check` clean.

ID-snapshot refinement verified: ten repetitions of all processor/configuration
race tests pass on Go 1.26.1 and 1.24.8; vet/build/full ordinary suite pass after
rerunning cache-access sandbox failures with approved access. Existing tests and
Seed implementation unchanged in this revision. Full race suite not rerun for
this refinement; previous unrelated failures remain outstanding. Logs:
`/private/tmp/nodeherder-registry-ids-{race,go124,ordinary}.log`.
Private read-helper refactor: ten processor/configuration race repetitions,
vet/build and diff checks pass; tests unchanged. Full suites not rerun for this
mechanical extraction (prior results above). Log: registry-helpers-race.log in
`/private/tmp/`.

FR-08: copy bridge plus its interval on save/read, under the cache mutex; copy the
AppConfig shell in LoadAppConfig so bridge replacement does not escape that lock.
Other nested app sections remain shared, not a promise of general immutable config.
Cleanup owns an atomic duration-pair snapshot updated synchronously in Start under
the existing cache caller lock; each loop loads sleep before sleeping and current
expiry afterward. No goroutine reads mutable app history. Existing Stop does not
cancel/join workers; leave lifecycle redesign outstanding, and ensure old workers
still observe new settings rather than retaining obsolete retention values.
NH-01/02/04/05/06 and BE-01/03/04: existing boundaries, storage shape, pruning rules,
units/timing retained; focused race regressions and ordinary suites. NH-03/BE-02/05
unchanged (no new access/API/device processing). No live databases/devices.

FR-08 results: five focused race repetitions (snapshot ownership, both permit-join
cases, existing cleanup/retention assertions) pass on Go 1.26.1; one repetition
passes on Go 1.24.8. Vet/build/full ordinary backend suite and diff checks pass.
Full race suite now passes store/settings; API, automations, controllers and
services still fail with existing domain races. No bridge-permit/history-cleanup
race stacks in this run. Logs: `/private/tmp/nodeherder-config-{focused-race,
go124,ordinary,full-race}.log`. Existing assertions/tests unchanged; added snapshot
regression in cache_test.go. Production edits limited to cache.go and store/tasks.go.
Readers must reload to observe later permit-join changes; Hub remains shared.
Cleanup settings publish on successful save/task restart; a failed history save
still changes the cache before returning its existing error, but no new durations
publish to workers. Existing lifecycle Stop/duplicate-worker issues remain deferred.
No live-device/data/deployment or Ubuntu execution; passing tests do not prove
complete live safety.

FR-09: serialize Start/Stop with a lifecycle mutex; each worker owns child context
and done channel. Cancel/join before replacement; worker never takes lifecycle lock.
Clock.AfterFunc plus a buffered wake channel replaces uninterruptible Sleep without
changing Clock interface. Stop timer on either completion/cancellation; check context
before Prune. Wait for existing Prune (non-context-aware repository); no forced
timeout/abandoned prune, no mutex held by worker across repository calls. Retain
duration publication and sleep-before-prune; old workers now exit before restart.
Tests in existing store_test.go: idle/repeated stop, sleeping stop, restart/concurrent
lifecycle, parent cancellation, in-flight prune joining; retain metrics assertions.
NH-01/04/05/06 and BE-01/03/04 apply; no new transport/storage/pruning/device semantics
(NH-02/03, BE-02/05 unchanged). Full/focused races, vet/build/ordinary checks.

FR-09 results: ten lifecycle race repetitions pass on Go 1.26.1; complete store
race suite passes on Go 1.24.8, including unchanged retention assertions. Vet/build,
all ordinary backend tests and diff checks pass. Full races pass store; API,
automations, controllers and services still fail on separate domain races. Logs:
`/private/tmp/nodeherder-cleanup-{lifecycle-race,go124,ordinary,full-race}.log`.
Only tasks.go changed in production; lifecycle regressions added to existing
store_test.go. No live household storage/device/deployment/Ubuntu execution.
Review: actual Prune implementation only accesses metrics storage/clock and does
not re-enter config/lifecycle locks. Stop/reload waits on an in-flight DB prune;
non-context-aware Prune can delay shutdown and configuration save completion.
Generic repository implementations must not re-enter Start/Stop from Prune.
This supersedes FR-08's deferred duplicate-worker/Stop finding, not broader
application/repository shutdown debt or test-helper resource ownership.

FR-10: replace two direct timestamp writes with existing SetLastSeen. Retain
getLastSeen result locally in Update and use it in the package, avoiding a shared
field reread after callbacks. Locks cover assignment only, never callbacks; no
timestamp-ordering policy or broad processing lock. New regression in existing
lifetime test file uses one updater and concurrent LastSeenTime reader (no Seed
timer), retains exact payload timestamp assertions. Existing seed/update/disabled/
debounce/event tests remain intact. NH-01/02/04/05, BE-01/02/03 apply; remaining
rules unchanged/N/A (no transports/storage/privacy/pruning change). Run focused
races, full races, vet/build/ordinary suites; other known races remain visible.

FR-10 results: concurrent LastSeenTime/update regression passes twenty race
repetitions on Go 1.26.1 and ten on Go 1.24.8. Exact emitted timestamps/count and
callback reads verified. Vet/build/all ordinary backend tests and diff checks pass;
existing device/event/debounce/disabled assertions unchanged. Full races still
fail API, automations, controllers and services; previous LastSeenTime-versus-
Seed/Update race stacks absent in this run. Direct exported-field/JSON and invalid-
timestamp logging reads remain outside the guarantee; no broad device safety claim.
Production diff only two setter calls plus Update's local timestamp/payload use.
Logs: `/private/tmp/nodeherder-lastseen-{focused,go124,ordinary,full-race}.log`.
No deployment/live devices/data or Ubuntu execution.

FR-11: lifecycle mutex protects disabled state and current monitor pointer. Private
locked helpers start/stop; worker captures local ticker/context and identity-checks
under mutex before an offline transition or clearing ownership on exit. Stop detaches,
cancels and stops under mutex, without joining callbacks; callbacks/logging outside
lock. Reset uses same ownership lock and creates missing monitor if necessary.
Retain real ticker (existing tests inject clocks for debounce, not monitor polling).
No lock across Seed/Update/automation; Update only snapshots disabled status. Device
availability/entity/JSON races remain deferred; no claim of atomic timeout-vs-message
ordering. NH-01/02/04/05/06, BE-01/02/03 apply; other contracts/access/storage unchanged.
Tests in existing lifetime file; repeated targeted races, Go1.24, ordinary/full races.

FR-11 results: five repetitions of all three new monitor regressions pass under
race detection on Go 1.26.1; final Go1.24.8 run passes (initial two regressions also
passed three repetitions). Eight concurrent togglers, synchronous callback disable/
re-enable, and offline→online reset tested. Existing assertions unchanged. Full
ordinary backend suite, final affected services/automations/controllers suites,
vet/build/diff checks pass. Full races still fail API, automations, controllers,
services on remaining domain/availability/entity races; no suppression or broader
field synchronization added. Logs: `/private/tmp/nodeherder-monitor-{focused-final,
go124-final,ordinary,ordinary-final,full-race}.log`.
Disabled updates snapshot stopped state; previously accepted/in-flight updates are
not drained. Committed callbacks can finish after cancellation, intentionally
avoiding callback self-join deadlock. Offline monitor retains its worker with a
stopped ticker for reset; a cancelled worker cannot clear replacement ownership.
Device read/expiry comparison is not atomic against incoming messages; this remains
separate transition work. No live deployment/devices/performance/Ubuntu guarantee.

FR-12: replace Seed's online assignment and Update's offline check/online assignment
with existing SetAvailability/GetAvailability. Individual device locks release
before lifecycle/reset/logging/callbacks: no nested lock added. Check-and-set remains
non-atomic, explicitly outside scope. NH-01/02/05 and BE-02/03: preserve all existing
assertions, verify seed/update/disabled/debounce/events, accessor concurrency and
full/targeted races; remaining contracts/access/storage unchanged. No device/entity
or timer redesign. Direct test/JSON reads are not protected by these accessor edits.

FR-12 results: production diff is exactly three accessor substitutions; no tests
edited. Vet/build/full ordinary suite and diff checks pass. Initial broader focused
race checks fail on both Go1.26.1 and Go1.24.8: UpdateWithNewData directly reads
exported Availability while monitor writes it. Failure retained, not suppressed.
Separately, seed/same-data/disabled/monitor paths pass three race repetitions on
Go1.26.1 and one on Go1.24.8. Full races pass API in this run, but automations,
controllers and services fail (including direct availability assertion/callback
reads and entity/automation races). One API pass is not exhaustive safety evidence.
Logs: `/private/tmp/nodeherder-availability-{focused,go124,safe-paths,safe-go124,
ordinary,full-race}.log`. No deployment/live data/devices or performance guarantee.

FR-13: private EntityData RWMutex; pointer Value/MarshalJSON receivers avoid copying
locks. Audit found no value-copy/value-JSON callers; Entity.Data is *EntityData.
Snapshot via Value for comparison/JSON; encode/decode outside locks, SetValue after
successful decode. Zero value works. name is constructor-only. Preserve existing
conversion/comparison rules (including existing unsupported composite comparison).
JSON remains primitive data, no schema migration; Go value method set changes,
unsupported value-copy callers would need pointers. No model lock surrounds service
or automation callbacks. Add tests before source fix; capture failing race regression,
then repeated model/service/operation races, full ordinary/vet/build/full races.
NH-01/02/04/05, BE-01/02/03/04 apply; trust/storage shapes unchanged, no deployment.
RWMutex adds per-value access cost; no benchmark/live latency guarantee.

FR-13 results: pre-fix numeric concurrent regression fails with Value/SetValue/
UnmarshalJSON race evidence (`/private/tmp/nodeherder-entity-before.log`). Final
mixed-type model/JSON and service event/debounce regressions pass twenty race
repetitions on Go1.26.1 and five on Go1.24.8. Existing increase/decrease/multi-step
operation race assertions pass five Go1.24.8 repetitions. Wire/null/mixed types,
zero value, malformed decode retention, encode errors, and custom encoder/callback
re-entry verified. Initial new test used a nonexistent type alias; corrected to
existing ExposeDataType (no production validation or existing assertion change).
Final vet/build/full ordinary suite and diff checks pass. Full races still fail
automations/controllers/services; EntityData race stacks absent in that run.
Direct availability reads and automation action/state races remain unmodified.
Production diff limited to device.go's EntityData access; new tests in existing
device_test.go/lifetime_test.go, no removed assertions or new test files. Logs:
`/private/tmp/nodeherder-entity-{focused-mixed,go124,go124-mixed,ordinary-final,full-race}.log`.
No deployment/live household data/devices; pointer method-set and referenced-value
ownership boundaries remain as documented, not a full-update ordering guarantee.

FR-14: replace shared Availability assertion reads with GetAvailability; initialization
and immutable UpdatePackage reads stay unchanged. Controller mock broadcasts signal
offline/online events; initial repository registration has a bounded readiness poll
(broadcast precedes store publication). Keep exact expected state sequence; service
fixtures disable monitoring only in cleanup after assertions. Existing failures
outside this scope remain visible. NH-05/BE-03 test ownership, no production change;
other contracts/rules unchanged. Repeated focused, affected/full race and ordinary
checks; reuse spec003 and existing test files.

FR-14 results: final four repaired cases pass five native and three Go1.24.8 race
repetitions. Full services race suite passes Go1.24.8; complete controllers still
fails manual-trigger action/configuration races (unchanged). Full backend race run
now passes services/API, fails automations/controllers only. Getter changes do not
serialize production processing; real broadcasts drive bounded state waits.
Original online/offline/re-online and payload/count assertions preserved; corrected
one diagnostic's wrong expected-state text, not its assertion. Service fixture
cleanup disables monitoring after assertions. No new test files/production edits.
Vet/full ordinary suite pass; final affected ordinary suites also pass after service
cleanup additions. Diff checks clean. Logs: `/private/tmp/nodeherder-availability-tests-
{final-focused,go124,go124-focused,ordinary,final-ordinary,full-race}.log`.
No deployment/live devices, no blanket availability transition ordering guarantee.

FR-15: private enable mutex; explicit base/device JSON snapshot avoids copying a
used mutex or promoted marshaler dropping fields. Decode retains existing metadata
handling, publishes flag through setter. Preserve scheduler assertions via getter;
add concurrent toggle/read/encode and JSON compatibility tests in engine_test.go.
Run focused races, ordinary/vet/build and full races; report remaining failures.

FR-15 results: new regression reproduces accessor and reflection/JSON races before
fix. Flag/JSON plus existing scheduler tests pass three native and two Go1.24.8
race repetitions; flag/JSON + scheduler-config test also passes five native runs.
Vet/build/full ordinary suite pass. Full races fail only automations/controllers:
delayed action cancellation/completion and action configuration/execution remain.
No enabled race reported. Existing scheduler expectations preserved via accessors;
no exclusions, new test files, action/scheduler changes or live-device testing.
Only flag ownership is synchronized: metadata/configuration decoding still requires
exclusive ownership, exported direct flag access is initialization-only, and base
objects must not be copied after use. No live performance/safety guarantee.
Logs: `/private/tmp/nodeherder-enabled-{focused,go124,ordinary,full-race}.log`.

FR-16: use existing action mutex for pending/channel publication, Stop and identity-
checked completion. Capture worker-local cancellation; Stop closes it once under
lock and detaches it. One-shot timer stopped on every worker exit; verify ownership
under lock before processing, then release before MQTT/context callbacks. No joins,
configuration locking or synchronous Execute redesign. Cancellation/commit serialized
by the mutex; after commit, Stop cannot retract execution. Brief locks only, no live
latency guarantee. NH-01/02/04/05/06, BE-01/03 apply; storage/auth/wire unchanged.
Add regressions in existing triggers_test.go, demonstrate baseline races, repeat
affected cases on native/Go1.24, run vet/build/ordinary/full races. Real delay expiry
uses bounded channel waits; hour-long stress delays isolate cancellation from expiry.
Use the interval's full duration; remove the old wall-clock millisecond truncation
(less than one millisecond of rounding), not a new cooldown or rescheduling policy.
Immediate Execute still holds its existing mutex across callbacks; its re-entry and
configuration ownership require separate review. Rapid Stop/restart can overlap an
already committed old publish; this fix does not serialize/reorder those commands.

FR-16 results: baseline stress reproduces seven race warnings and close-of-closed-
channel panic (`/private/tmp/nodeherder-delay-before.log`). Final four lifecycle
regressions pass twenty native repetitions; original delayed-trigger cases plus new
regressions pass five Go1.24.8 repetitions. Full automations races pass three native
and one Go1.24.8 runs. Vet/build/full ordinary tests and diff checks pass. Full backend
races now fail controllers only: processAction/CreatePayload overlap Configure/
newTriggerOperation in TestManualTriggerTurnsOnLightAutomation; left untouched.
Source diff limited to action.go delayed lifecycle/Stop, invalid duration guard;
existing assertions unchanged, tests appended in triggers_test.go. No new test files,
exclusions, deployment or live-device checks. No proven live latency/safety guarantee.
Other logs: `/private/tmp/nodeherder-delay-{focused,final-focused,automations,go124,
go124-all,ordinary,full-race}.log`. Acceptance covers cancellation, duplicate/restart,
old completion, callback re-entry and invalid duration; no joins under action lock.

FR-17: append integration regressions to device_processor_test.go using memory store,
mock automation storage, real engine/conditions/processor and synchronous captured
MQTT client. Seed outside tested sequence; long debounce window, no inter-event sleeps.
Confirmed-feedback case feeds commands back through target processing before next
input; withheld case preserves reported target state and checks resulting commands.
Verify fresh source values, exact counts/topics/brightness and sampled UI behavior.
No network/controller transport-routing coverage or real-latency claim. NH-02/05,
BE-01/02/03; other rules unchanged. Run focused native/Go1.24 races, services ordinary/
race and vet/build/full ordinary; do not hide remaining configuration races.

FR-17 results: three scenarios × 120 events pass twenty native and twenty Go1.24.8
race repetitions. Full services/automations race command passes (automations cached);
vet/build/full ordinary and diff checks pass. All commands/topics, source state/delta,
UI event counts and reported target state checked. Confirmed rotations finish at
brightness 60 from 200; withheld feedback leaves reported brightness 200 and commands
calculate from that value. No production changes or original assertion modifications.
Initial new fixture run correctly hit the brightness minimum (suppressed commands)
and read an unregistered fixture pointer. Corrected setup retrieves registered
devices and starts at 200 so the 120-event accumulation case stays within bounds;
existing min/max tests unchanged. No test failures were fixed through source changes.
Logs: `/private/tmp/nodeherder-rapid-{first,focused,go124,ordinary,affected}.log`.
Full backend races not rerun for test-only work; prior controller configuration races
remain unresolved. Ordered input/simulated confirmation only: no broker transport,
real feedback latency, concurrent MQTT delivery or live-device safety guarantee.

FR-18: short recipe mutex for internal copies/normalization; immutable runtime binding
published by atomic pointer after successful preparation. No registry/network calls
under recipe lock. processAction loads one binding; emit uses that same client/name.
Step operation captures registrar and copied recipe instead of reading mutable base
configuration; existing execution mutex still owns step/preset mutable operation state.
In-flight old execution may finish while new configuration is published. A delayed
run loads current binding at execution start, preserving prior reconfiguration timing.
No global engine lock, command join/retry, scheduling redesign or dependency change.
NH-01/02/04/05/06, BE-01/03; auth/storage unchanged. Add regressions before source,
repeat focused native/Go1.24 races, full ordinary/vet/build/full races; no live devices.
Trigger MarshalJSON must snapshot its recipe because Configure normalizes exposed
Data; do not add a base marshaler that could omit embedded derived fields. Decode/
external recipe mutation remains exclusive-owner work. No encoding under recipe lock.

FR-18 results: pre-fix tests reproduce lookup-failure recipe/payload mutation,
configuration races and single-publish actions split across clients. Final focused
binding/failure/unconfigured/JSON tests pass ten native and ten Go1.24.8 repetitions;
JSON compatibility/re-entry plus rapid-event/delay cases pass ten native repetitions.
Complete native backend race suite passes, including controllers; complete Go1.24.8
automations/controllers race suites pass. Final vet/build/full ordinary and diff checks
pass. Original assertions unchanged; new tests appended to triggers_test.go. During
development, corrected a missing source import and new-test mistakes (trigger wire
assertion initially attached to step fixture; enum Values requires []string), then
reran checks; no source changes to conceal a failed existing assertion.
Failure behavior intentionally changes to retain last successful recipe/binding;
unconfigured execution/missing step/preset expose now report errors instead of nil
dereference. Trigger JSON fields/null/empty/omitempty and encoder re-entry verified.
Production limited to action.go and operations.go; no engine/scheduler/controller,
arithmetic, command-confirmation, cancellation or preset ordering rewrite. Per-command
binding load is atomic; no configuration lock around external callbacks, no benchmark
or live safety guarantee. Immediate action mutex/re-entry behavior remains unchanged.
Only internally owned recipe updates synchronized; direct concurrent recipe/Client
edits, mutable referenced composite values and concurrent decoding are unsupported.
No deployment/live broker/devices. Logs: `/private/tmp/nodeherder-config-{before,
focused-final,compatibility,go124-final,go124-focused,handoff-ordinary,handoff-race}.log`.

FR-19: reserve one in-flight lifetime per ID before Seed; queue copied updates and
config changes without waiting, drain outside registry lock, then publish atomically.
Retain existing configure helper/default routing. Add existing-file regressions
before source edits; repeat focused races and full backend checks. No lifetime,
monitor or automation rewrite. Queued calls return before delivery during creation;
blocked callbacks retain their queue (no new worker or event loss/drop policy).

User-requested cleanup: one registry entry owns service/readiness/queue; one routing
helper owns lookup/reservation/enqueue locking. Event snapshot/apply and queue drain
are private helpers; no duplicated update/config routing. Published-event lookup
now uses a short exclusive registry lock; callbacks remain unlocked. This deliberately
avoids a second fast-path/readiness check and its duplicate source of truth.

FR-19 results: pre-fix regressions fail with duplicate Seed callbacks. Final tests
cover registered/new devices, concurrent/re-entrant first updates, defaults/override
disable during Seed, ordered delivery, input-map copying and config/update re-entry
while draining. Original assertions unchanged; appended cases in existing file.
Full native uncached race/ordinary suites, vet/build, twenty Go1.24.8 focused race
runs (including rapid physical actions), Linux/amd64 cross-build and diff checks pass.
Production changes limited to device_processor.go. Normal delivery allocates no
event queue/snapshot; only creation-window events queue. Maps copy shallowly: nested
payload values remain caller-owned/immutable. Blocking callbacks can retain pending
events; no timeout, join, worker or live-device guarantee. Logs:
`/private/tmp/nodeherder-creation-{before,focused,full-race,ordinary,go124}.log`.

FR-20: runtime registry is the execution/read source of truth; storage owns persisted
recipes only. Short registry locks, no callbacks under them; TryLock management
ownership rejects competing writes rather than blocking re-entry. Prepare handler
changes before activation, roll them back on activation/save failure, publish then
retire previous resources. Scheduler changes own child contexts; validate complete
recipes, preserve enabled state for unchanged schedules, cancel discarded workers.
Disk Initialize stages a complete cache and reports I/O/JSON failures. Trigger Process
and manual evaluation propagate joined action errors without skipping later actions.
DeleteTrigger clones via existing serializer; error bounds include negative indices.
Add regressions in existing engine/trigger/scheduler/storage test files before source;
full ordinary/race/vet/build, repeated Go1.24 focus and Linux cross-build after repair.
NH-01/02/04/05/06 and BE-01/03/04 apply; auth, MQTT format, physical-event arithmetic,
timezone and repeat cadence unchanged. No live-device checks. Design/alternatives:
[runtime generation ADR](adr-runtime-generation.md). Constitutions remain proposed.

Persistence refinement: direct overwrite can corrupt a recipe on failed writes.
Stage/sync/close then rename within its directory under the existing storage lock;
retain file permissions and clean discarded temp files. Add external concurrent-
reader JSON and permission/failure regressions; no filesystem layout/schema change.
Final failure-path review reproduces a mkdir panic with a dangling directory link.
Store now creates its directory and returns errors; pure path construction avoids
read/delete side effects. Existing directory mode retained; cache/runtime unchanged
on failure. Directory-link regression works independently of effective UID.

FR-20 evidence: pre-fix regressions demonstrate unready publication, hidden save
errors, stale/type-incomplete schedule matching, discarded disk cache and partial
JSON writes. Final new matrix passes ten native race repetitions; prior complete
native ordinary/race/vet/build and affected Go1.24 suites pass; ten Go1.24 expanded
repetitions and Linux build pass. Existing assertions preserved. Source also validates
nil nested recipe entries, cancels partially started schedulers, and guards old timer
completion by run identity. No background-context watcher is created; prepared/retired
schedule contexts are canceled. New lifecycle tests stay in existing test files.
Core engine read/delivery/handler-transaction functions reached 100% statement
coverage; coverage is not proof of all interleavings. Final verification follows FR-21.
Internal Trigger.Process/manual-evaluation contracts now return errors and handlers
implement prepared changes; no transport schema change. Already committed callback/
action completion is non-joining; overlapping whole-event context transactions remain
pre-existing debt. Atomic save requires a writable containing directory; valid symlinks
preserved, dangling symlink saves fail explicitly. No directory fsync/power-loss or
multi-file durability claim. No live broker/Ubuntu runtime test. Logs:
`/private/tmp/nodeherder-ready-{before,schedule-before,disk-before,atomic-before,
final-focused,atomic-final,ordinary,final-race,go124,go124-focused}.log`.

FR-21 cleanup design: central engine update helper owns TryLock/release for all
management paths. Cleanup duration atomics are redundant once Start joins its old
worker; capture the immutable durations locally instead. Default cleanup constructor
delegates to the existing constructor. Unused lifetime cooldown fields/constants
misrepresent the actual unthrottled event path; remove without adding any throttling.
Keep the registry, immutable action binding and monitor/run identities; superficial
similar lock accessors do not warrant generic wrappers. Add missing configuration
snapshot/busy-gate regressions, rerun the complete checks after cleanup. No new ADR:
ownership and policies are unchanged; existing-file tests/FR-20 guard this refactor.

FR-21 results: shared management gate, worker-local immutable durations, delegated
default constructor and unused private cooldown/timeout removal implemented. Processor
stores only its stable configuration-cache dependency; no Store getter inside its
registry lock. GetId is evaluated outside registry locks. Single-error Join removed;
action contextual errors preserve causes with unchanged displayed details. New error-
cause regressions failed before this simplification and pass afterwards. Both cleanup
constructors/config snapshots and all re-entry gates tested before/after refactoring.
Complete native uncached ordinary/race suites, vet/build, ten focused Go1.24.8 race
repetitions and Linux/amd64 cross-build pass on final source. Original behavioral
assertions retained. No live devices, frontend changes or Ubuntu runtime execution.
Logs: `/private/tmp/nodeherder-cleanup-{before,errors-before,focused,ordinary,
full-race,go124-focused}.log`; coverage `/private/tmp/nodeherder-cleanup-coverage.out`.

Final FR-20 audit (regression found; follow-up authorized): explicit Z2M source Disabled makes
Device.Configure fail; generic fallback retains the old enabled generation. An
isolated overlay regression demonstrates HandleManual returning nil and publishing
one command after source disable. It is not part of the normal passing suite and
does not justify a completion/safety claim. Recommended policy: explicit disable
stops/removes its executable generation and schedule, while ordinary configuration/
lookup/persistence failures retain the working generation. User approved this exception:
classify disabled-source errors, remove runtime before handler retirement, retain
recipes and retry failed cleanup on reload. Commit the
regression in engine_test.go, test physical/manual/scheduled/re-enable paths, rerun
checks. Diagnostic: `/private/tmp/nodeherder-disabled-audit.TKZJ4W/overlay.json`,
`/private/tmp/nodeherder-disabled-audit.log`. Preset-cycle and whole-event-context
coverage/ownership debt remain separately recorded; do not hide or suppress them.

FR-20 explicit-disable follow-up: four Add/reload × cleanup-failure regressions
failed before repair (old source remained executable). Classified source-disable
error now removes the ready entry before ordinary handler retirement; failures keep
execution unavailable and report both error causes. No persistence mutation, new
worker/polling, callback lock, command join or normal-failure fallback change.
Existing engine_test.go covers initial disable, manual/physical rejection, another
source, retained recipe values, lookup fallback, failed/successful re-enable,
schedule retirement/prepare+activate failure/cleanup retry/restored windows, and
non-joining accepted commands. Ten Go1.24 race repetitions pass. Full native race
and ordinary tests, vet/build, Linux/amd64 CGO-disabled build and diff checks pass.
No live-device/Ubuntu-runtime test or claim of complete coverage; preset-cycle and
whole-event-context debt remain outside this follow-up.

FR-22 design/results: an atomic reload-pending flag is set by Initialize; whoever
releases the engine gate (withUpdate or a reload) drains it with TryLock, so a request
is either run by its caller or observed by the owner after Unlock. A reload re-requested
from inside reload reconfiguration repeats by design; the existing re-entry test probe
now runs once (assertions unchanged) because its callback re-requested on every lookup.
LoadAppConfig clones Hub sections, device defaults/overrides with debounce maps and
dashboard groups; intervals/expose lists are replaced, never mutated, and stay shared.
Red-before-green: TestEngineReloadDuringUpdateRunsAfterRelease ("deferred reload was
dropped"), TestAppConfigSnapshotOwnsHub ("snapshot changed after writes"). Checks
(go1.26.1 darwin/arm64): build, vet, full `go test -race -count=1 ./...`, ordinary
automations/controllers, 20 race repetitions of affected engine/settings tests pass.
gofmt applied to touched key-value.go. Go1.24.8 and live-device checks not run.
