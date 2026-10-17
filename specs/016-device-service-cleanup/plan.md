# Plan: Device service retirement and launcher CI

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-16
Constitutions: parent/backend v1.0.0 proposed; frontend N/A (no UI changes).

## Context and design

Bridge remove responses currently call registrar.RemoveDevice directly on the bridge
lane. Inject a removal callback into that handler. HubController routes removal onto
the device's existing friendly-name lane using bounded lanes.Do, preserving ordering
with accepted updates and avoiding callbacks under the processor registry lock.

DeviceProcessor.RemoveDevice finds the current entry without holding registry locks
across I/O. DeviceLifetimeService retires it using the publisher's output mutex as a
transaction fence: execute storage deletion only after an active output finishes;
close pending output only on success. Failed deletion leaves the generation live.
On success, cancel an owned child context, stop/wait availability workers and evict
that exact entry and creation queue. Old references see the cancelled context and
cannot update/re-enable. Rediscovery still follows existing accepted-report behavior
and gets a fresh entry; no tombstones are retained.

ObservationPublisher retains/unregisters its context.AfterFunc stop callback on Close.
Registration/closure are mutex-protected even when the context is already cancelled.
The lifetime child context also parents availability monitoring, so hub cancellation
cannot leave a worker behind. A wait group tracks workers even across disable/re-enable;
start/close coordinate through monitorMu to prevent Add/Wait races. Close must run
outside output/availability callbacks; production removal runs on its owning lane.
See [ADR](adr-001-retirement.md) for the fence and failure alternatives.

CI adds `node --test scripts/dev.test.mjs` to stack-tools alongside the generator
checks. No runtime/frontend settings changes or new contracts/dependencies.

## Constitution Check

| Rule | Design evidence | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | AC-01–05 and tasks before implementation | PASS | PASS |
| NH-02 | Retired generation only; no debounce/action changes | PASS | PASS |
| NH-03 | Same validated bridge responses; no new access | PASS | PASS |
| NH-04 | Controller lane, processor lifecycle, publisher fence | PASS | PASS |
| NH-05 | Fake clocks/channels, failure and churn companions | PASS | PASS |
| NH-06 | Wait/unregister/evict; failed-delete preservation | PASS | PASS |
| BE-01 | Handler delegates removal; store remains registrar-owned | PASS | PASS |
| BE-02 | Seed/update/disabled/availability regressions retained | PASS | PASS |
| BE-03 | Context ownership, worker join, guarded closure | PASS | PASS |
| BE-04 | Storage failure never reported as success | PASS | PASS |
| BE-05 | Existing bridge messages/access unchanged | PASS | PASS |
| FE-01–05 | No frontend implementation changes | N/A | N/A |

## Traceability and verification

AC-01: Node 24 launcher tests plus actionlint; verify aggregate needs stack-tools.
AC-02–05: companion tests observation_publisher_test.go, device_lifetime_test.go,
device_processor_test.go, handlers_test.go and hub-controller_internal_test.go:
callback unregistration, pending timer cancellation, no late output/re-enable,
failed removal retention, repeated retirement, fresh rediscovery, churn eviction,
owned lane ordering and shutdown. All isolated stores/contexts, no household commands.
Final verification (2026-10-16):

- `go test -count=1 -timeout=3m ./... -coverprofile=...`: PASS, all packages.
- `go test -race -count=1 -timeout=3m ./internal/services ./internal/controllers`:
  PASS after the final changes. The broader affected automation/device/settings race
  suites also passed before the final cancellation regression and deferred-unlock edit.
- `go vet ./...` and `go build ./...`: PASS after final changes.
- Node 24 `node --test scripts/manage-stack.test.cjs scripts/dev.test.mjs
  scripts/sync-reporting-defaults.test.mjs`: 65/65 PASS, including all five launcher
  lifecycle tests. Workflow actionlint PASS; aggregate `test` requires `stack-tools`.
- `git diff --check`: PASS. Reviewed output fencing, callback ownership, monitor
  Add/Wait ordering, failed deletion and lane independence after implementation.
- CRAP analysis with the final coverage: new retirement methods score 1–6, all below
  30. `Close`, `Retire`, `closeAfter` and processor `RemoveDevice` have 100% package
  coverage. Controller removal is also exercised through its real lane/handler path;
  `DeviceName` is exercised by that cross-package regression, outside service-package
  coverage instrumentation.

Initial new processor/controller fixtures used an unknown raw property and failed
schema inference before exercising retirement. Corrected only those fixture payloads
to the supported `temperature` property; assertions were retained. All final checks
above passed. Added the unpersisted-first-snapshot lane regression, failure retention,
repeated churn/rediscovery, pending-output fence and hub-cancellation coverage.

Browser/Jest/frontend build checks were not rerun: this change touches backend,
CI and documentation only. Previous spec 015 records their historical results.
No household devices, commits or deployments were used.

## Contracts, risks and recovery

No persisted schema, metrics retention or command changes. Removal now waits for
existing output before deleting storage. A failed deletion keeps the generation live;
existing failure broadcasts remain. Repeated deletion follows repository idempotency.
A later accepted packet may rediscover the device, as before; suppressing packets
until explicit rejoin would be a separate product policy. Lane names follow the
existing TriggerManual convention. Avoid same-bridge-key lanes.Do self-deadlock.
No new cleanup timeout can safely abandon already-started synchronous store I/O.
No exceptions claimed. Work stays uncommitted on the existing branch.
