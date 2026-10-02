# Spec: Scoped CI repair

ID: 003 | Updated: 2026-10-02 | Related: [CI](../002-ci-verification/spec.md)

User narrowed runtime changes to mapper synchronization and closing the key/value
DB when initialization fails. Automation/scheduler edits, driver replacement,
transaction-value copying and legacy fixture work were removed. Existing worker-test
repairs remain under spec 002. No deployment or household-data access.

- FR-01: Concurrent mapper configure/update/lookup preserves IEEE mapping,
  stale-entry removal and hashed fallback without map races.
- FR-02: Failed key/value initialization closes the opened DB, releasing its lock.
- AC-01: Concurrent mapper regression passes repeatedly under race detection.
- AC-02: Invalid bucket initialization followed by reopening the same temporary file succeeds.

Verify vet/build/ordinary backend tests and targeted races; do not claim full race
CI is repaired. Further runtime changes need separate user review/authorization.

## Authorized mock follow-up

FR-03: Make mock automation-cache operations, MQTT handler/response configuration,
and clock/timer state safe for concurrent access. Timer callbacks execute outside
state locks; resetting timers during callbacks must not lose or duplicate timers.
Provide explicit waiting for asynchronous mock MQTT responses. Preserve production
code/dependencies and existing response behavior. Verify concurrent mock regressions
and ordinary backend tests; report remaining runtime/Bolt failures separately.

## Authorized Bolt follow-up

FR-04: Replace only the legacy driver with a Go-1.24-compatible bbolt release.
Keep filenames, buckets, keys, JSON, locking options and automation behavior
unchanged. AC-04: a synthetic database created with Bolt v1.3.1 must retain root
values, nested metrics and device/bridge records after bbolt read/write/reopen.
Existing tests/assertions and CI checks remain intact; do not suppress checkptr
or race detection. No live database modification/deployment. Ordinary tests,
vet/build and storage race checks must pass; report unrelated full-race failures.

## Low-risk runtime follow-up

FR-05: Synchronize WebSocket client bookkeeping and the remote-log enable flag
without changing network payloads, automation state/timing, MQTT/Z2M handling or
storage. No deployment/live devices. AC-05: concurrent client lifecycle callbacks
and logger toggles/delivery pass race tests; disabled logging remains disabled.
No new locks across network calls, logging or user callbacks. Existing assertions
remain intact. Live workload safety is not inferred from test success.

FR-06 test-only follow-up: coordinate operation-test producers/responses and
permit-join test callbacks; await mock deliveries before reusing fixtures and
remove diagnostic reads of mutable actions. Retain arithmetic, bounds, command
counts and permit-join assertions. No production changes, race exclusions or
serialization of genuine production concurrency scenarios to conceal races.

FR-07 registry follow-up: protect device-service configuration lookups and
iteration with the existing registry mutex. Invoke configuration methods after
unlocking; preserve creation, lifecycle, automation and device-update behavior.
AC-07: concurrent registration/default/override lookup passes repeated race checks
with disabled devices to isolate registry ownership. Existing lifecycle race tests
remain unchanged. Authorized extension: Seed callbacks must execute outside the
registry lock and may synchronously update configuration without deadlock. Publish
only after Seed completes; no per-device update serialization/timing redesign.
Review refinement: defaults snapshot IDs, then use configureDeviceLifetime outside
the iteration lock, retaining the shared lookup/configuration flow for both paths.

FR-08 configuration follow-up: permit-join saves must not mutate previously returned
bridge values; cache reads return detached bridge snapshots. Metrics cleanup reads
latest configured sleep/expiry durations without racing history-pointer replacement.
Preserve duration units, persistence/errors, expiry callbacks and read-after-sleep
retention behavior. No pruning-rule, timer, automation or task-lifecycle redesign.
AC-08: concurrent bridge reads/saves and old-snapshot stability pass; existing permit
expiry and metrics-cleanup tests retain assertions and pass under race detection.

FR-09 cleanup lifecycle: Stop cancels an interruptible sleep, waits for any in-flight
Prune to finish, and is safe before Start or repeatedly. Start replaces/joins its
previous worker; only one worker per task, parent cancellation respected. Preserve
sleep-before-prune, retention units and pruning/error behavior. No pruning starts
after Stop returns. Repository Prune has no context: shutdown must await it, not
pretend to interrupt it. No unrelated shutdown/device/automation changes.

FR-10 LastSeen-only follow-up: Seed/Update use existing SetLastSeen; Update retains
the incoming timestamp locally for its emitted package. Preserve timestamp parsing/
fallback and callback ordering; no availability/entity/monitor/automation changes.
AC-10: concurrent LastSeenTime reads during updates pass race checks; emitted
timestamps retain payload values. Existing assertions remain unchanged. Direct
exported-field/JSON readers remain outside this narrow repair.

FR-11 monitor lifecycle: serialize start/stop/reset and disabled-state ownership;
each worker owns its ticker/context, and stale workers cannot clear or expire a
replacement monitor. Preserve real one-second polling, existing last-seen timeout
comparison, event/debounce/automation processing. Disable cancels and stops ticker;
callbacks execute outside lifecycle locks and may disable/re-enable synchronously.
Already committed/in-flight callbacks may finish after disable (no self-join).
Reset resumes stopped ticker or recreates a missing worker after exit. AC-11:
re-entrant disable/re-enable monitoring, repeated/concurrent config toggles, disabled
updates and existing semantics tests; no assertion removal or other field-race fixes.

FR-12 availability-access-only: Seed/Update use existing device availability
getters/setters instead of direct field access. Preserve online/timeout callbacks,
polling and automation; no atomic transition/order or entity/serialization redesign.
Existing assertions remain unchanged; remaining direct test/JSON reads stay visible.

FR-13 entity-value ownership: synchronize Value/SetValue/ValuesMatch/JSON access
inside EntityData only. Preserve primitive wire values, comparisons, mixed value
types, repeated events, unchanged-state dedup and fresh automation reads before
debounce. Decode failures leave prior state intact. No callbacks/custom encoders
under the lock; no event ordering, action timing or broad device locking change.
Pointer ownership: do not copy EntityData; all discovered repository uses are pointers.
Referenced composite values remain caller-owned and must not be mutated concurrently
after publication; no deep-copy or multi-entity transaction guarantee.
AC-13: concurrent value/compare/JSON regression; encoding re-entry; wire/error/type
compatibility; repeated-event and fresh-debounced automation regressions, existing
operation assertions retained. Remaining automation/availability races stay visible.

FR-14 test-only availability repair: use model accessors for concurrent device
assertions; await controller offline/online broadcasts rather than fixed sleeps.
Retain online→offline→online expectations and all existing payload/count checks.
Bound waits, stop service fixtures after assertions; no production edits, test
locks around processing, exclusions or changed timeout policies.

FR-15 enable-state only: synchronize IsEnabled/SetEnabled and JSON flag snapshots.
Preserve enabled wire field, metadata, initial/default values and scheduler/gating
semantics; no locks across evaluation, callbacks or encoding. Exported Enabled is
initialization-only; concurrent callers use accessors. No action/configuration fix.

FR-16 delayed-action lifecycle: reserve pending state before Execute returns;
duplicate requests while pending do not reset the delay. Stop cancels waiting work
and permits restart, repeatedly/concurrently without blocked sends or channel panic.
Each worker owns its cancellation channel/timer; old completion cannot clear a
replacement. Cancel before timer commitment prevents publication. Already committed
publication may finish; Stop is not a join and cannot retract MQTT commands.
Preserve payload/state updates, immediate-action locking, conditions and configuration.
Invalid/nonpositive duration returns an error rather than panicking in a timer worker.
AC-16: concurrent scheduling/Stop, duplicate suppression, cancellation/restart,
completion identity and callback re-entry tests; retain existing delay assertions.

FR-17 test-only rapid events: exercise JSON-decoded input → device processor/lifetime
→ real automation engine/conditions → immediate trigger/step → captured MQTT command.
Repeated identical button and rotary events retain exact command counts; alternating
directions use fresh action_time and confirmed target brightness despite UI debounce.
Characterize withheld feedback separately: do not invent confirmed brightness from
outgoing commands or claim every unconfirmed step accumulates. No production change,
live broker/device use, altered existing assertions or artificial event cooldown.

FR-18 approved configuration ownership: prepare/validate action configuration before
publishing one runtime snapshot (operation, target name, MQTT client and registrar).
Execution uses one snapshot throughout all publishes; failed reconfiguration keeps
the last working configuration. No configuration lock across registrar/MQTT/context
callbacks. Preserve arithmetic, presets, payload shape and delay/cancellation policy;
delayed actions select current configuration when execution starts, not when queued.
Unconfigured execution returns an error rather than dereferencing a nil operation.
Recipes are caller-owned before publication; concurrent direct exported-field edits
remain unsupported. Verify failed lookup/sanitization, simultaneous configure/execute,
multi-command consistency, callback reconfiguration and existing rotary/delay cases.
Trigger JSON snapshots normalized recipe fields under the short recipe lock and
encodes outside it; preserve wire fields/omitempty and decoding behavior.

FR-19 creation ownership: exactly one lifetime seeds each device ID. During Seed,
copy and queue incoming updates/config notifications; the creator drains them in
registry-arrival order before publishing. Re-entrant calls enqueue without waiting;
normal published updates remain synchronous. No callbacks/cache reads under registry
lock, extra workers, event coalescing or automation-policy changes. Defaults include
in-flight IDs; disable during Seed applies before later queued updates. Already
running Seed callbacks may finish. Queue retains events until callbacks return;
there is no timeout/drop policy. Test same-ID concurrency, disable and re-entry.

FR-20 approved ready-generation reload: execution, enabled checks and engine reads
use only successfully configured runtime generations, independently of storage's
recipe cache. Keep the previous generation on load/configuration/handler/save failure;
first-load failures remain unavailable. Publish per ID only after preparation and
successful persistence (Add); remove absent IDs only after a complete successful
disk scan. Invalid JSON/I/O aborts disk reload without clearing its previous cache.
Save stages/syncs JSON in the same directory and atomically replaces the recipe;
preserve existing permissions, remove temporary files on failure, and never expose
truncated JSON to concurrent readers. No schema change or multi-file transaction.
Directory-creation failures return errors, not panics; reads/deletes do not create
missing directories. Regression must work with root-run CI (no permission-only test).
Replacement schedules bind the replacement even when times are unchanged; compare
complete time/type lists, preserve current enabled state for identical schedules,
and retain old jobs until commit. Failed preparation/activation/save rolls back new
resources; replacement/deletion/removal cancel owned jobs and context watchers.
Old already-started evaluations/actions may finish, including previously accepted
delayed actions; no retry, event queue, speculative target state or action join.
Management updates are single-owner and reject concurrent/re-entrant writes with
an observable busy error; reads/triggers remain nonblocking during preparation.
Clone active recipes before editing/removing triggers; do not mutate live generations.
Immediate action errors reach manual callers; process remaining actions as before,
log physical-trigger failures, distinguish unknown trigger from a successful no-op.
Delayed execution errors remain asynchronous/logged; MQTT Publish has no ack/error
contract, so success means accepted execution, not confirmed physical device state.
AC-20: gated first/reload execution; successful/failed replacements; disk corruption/
missing root; failed Store/Delete; handler prepare/activate failures and rollback;
unchanged/changed/type-only/removed/invalid schedules; timer ownership/cancellation;
callback read/write re-entry; concurrent reads/updates; trigger errors and missing
names; existing rapid-event/delay/schedule semantics remain covered.

FR-21 authorized cleanup: preserve FR-01–20 semantics/ownership while reducing
duplication. One management gate per engine; one cleanup constructor and immutable
per-worker duration snapshot. Remove only unused private cooldown/timeout state,
not event safeguards, cancellation identities or public compatibility entry points.
Capture the processor's stable configuration-cache dependency during construction;
no store/interface getter is invoked under registry locks. Snapshot IDs before locking.
Preserve original/new behavioral assertions; add missing gates/config-snapshot
coverage before simplifying, then full native/Go1.24/race/static/Linux checks.
No API payload, arithmetic, order, cancellation or timing-policy redesign.
Contextual action errors wrap their causes rather than stringify them; remove the
single-error Join wrapper without changing displayed error details or execution.

FR-20 explicit-disable exception (user approved): an observed disabled Z2M source
overrides fallback. Remove its executable generation before retiring schedules,
even if handler cleanup fails; keep its persisted recipe for later re-enable.
Retry cleanup on subsequent reloads; only successful configuration can republish.
Ordinary lookup/configuration/persistence failures still retain a working generation.
Already accepted commands remain non-joining. Cover initial disable, Add/reload,
manual/physical rejection, schedules, cleanup failure/retry, re-enable failure/success,
and isolation of other sources. No new polling or device-state policy.

FR-22 review follow-up (user requested). Reproduction 1: a bridge/devices reload
arriving while a save/delete owns the management gate logged busy and was dropped;
the bridge hash still advanced, so automations kept stale bindings until the bridge
payload changed again. Expected: reload stays nonblocking but is deferred and runs
once the current owner releases the gate; repeated requests coalesce. User writes
keep the observable busy error. Reproduction 2: LoadAppConfig cloned only Bridge;
callers encoded the shared Hub (device defaults/overrides, debounce maps, dashboard
groups, sections) after unlock while writers mutated it. Expected: returned snapshots
own every Hub part writers replace or mutate; payload content is unchanged.
No change to device events, debounce, schedules, MQTT payloads or enabled semantics.
