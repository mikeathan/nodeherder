# Plan: Restore automation toggle commands

Spec: [spec.md](spec.md) | Status: Implemented | Author: NodeHerder contributors
Updated: 2026-10-22 | Constitutions: parent/backend/frontend proposed 1.0.0

## Context and design

Keep capability and presentation logic in `frontend/src/domain/exposes.ts`: binary
options default to reported on/off; explicit action mode adds a valid distinct
declared toggle. Command formatting is separate from `formatExposeValue`.
`ExposeValueInput.vue` opts in via a prop only from `actions/SetValuesAction.vue`.
`domain/automation.ts` uses command formatting for action summaries only.

Real-backend E2E found that `createActionFromType` supplies a zero delay rejected
by `MqttTriggerAction.Configure`. Normalize zero delays to absence in one pure
`contracts/automations.ts` helper used by the Vuex automation save action; keep
positive delays, other fields and the caller's draft unchanged. Verify in its
source-owned `src/__tests__/contracts/automations.test.ts` and real-engine E2E.

Race-instrumented startup reproduced `store.RegisterIsDirtyCallback` appending while
`markDirty` reads the slice after MQTT device creation. Protect publication with
an RWMutex and copy the callback slice only on registration. Notification snapshots
the immutable slice under a read lock, then invokes callbacks unlocked. This
preserves callback order, permits reentrancy and allocates nothing per notification.
Companion `backend/store/store_test.go` covers concurrency and reentrancy under race
detection. No new goroutine or lock is held across an external callback.

The initially empty saved fixture also exposed `TriggerList.UnmarshalJSON` returning
nil for an empty array. Decode into a preallocated local list and publish only on
success, preserving [] and standard replacement semantics. Existing empty recipes
then reload in the editor without a migration. Companion `trigger_test.go` covers
empty/replacement/invalid input; browser E2E saves/reloads an empty dial recipe and
adds its trigger before creating the toggle automation.

Regressions belong in existing companion `src/__tests__/domain/exposes.test.ts`
and `automation.test.ts`. Extend `e2e/automations.spec.ts` with a UI-created dial
automation, saved/reloaded before execution. Reuse shared UI/auth/WS support.

The existing Node browser hub cannot execute Go automations. Reuse
`backend/testing/hubharness` with a real WS transport and in-memory Paho client.
Extract route assembly from `backend/internal/hub.go` without changing routes;
its companion `hub_test.go` hosts an opt-in, loopback-only isolated browser hub
with synthetic dial/light metadata, authenticated fixture controls, and simulated
light confirmations. Context, server, WS hub, and MQTT cleanup are owned by the
test; waits and queues are bounded. Playwright owns the fixture process, and CI
installs the module's Go toolchain. Normal suites skip the long-running fixture.

Alternatives: a JS automation emulator would duplicate production behavior and
cannot establish AC-03; a live broker would add operational cost and device risk.
No new library, architecture, storage, or trust decision warrants an ADR.

## Constitution Check

Before implementation checks reference the supported design, including documented
scope updates before their edits. Final local self-review on 2026-10-22 checked the
diff, source ownership, callback locks, native command values, lifecycle and evidence
below. PASS is a local assessment, not maintainer approval or a GitHub CI result.

| Rule | Design/verification evidence | Before implementation | Final local review |
| --- | --- | --- | --- |
| NH-01 | Spec includes reproduction, outcomes, failures | PASS | PASS |
| NH-02 | Native command; separate MQTT confirmation; repeated event E2E | PASS | PASS |
| NH-03 | Existing auth; test controls only in opt-in test; synthetic data | PASS | PASS |
| NH-04 | Existing domain/controller/store/transport boundaries | PASS | PASS |
| NH-05 | Companion tests and real-engine E2E below | PASS | PASS |
| NH-06 | No persistence changes; bounded fixture resources and cleanup | PASS | PASS |
| BE-01 | Existing harness/fake MQTT, production router reuse | PASS | PASS |
| BE-02 | Ingestion unchanged; existing dial scenario and repeated-event E2E | PASS | PASS |
| BE-03 | Context-owned fixture, synchronized hooks, bounded waits; race checks | PASS | PASS |
| BE-04 | No storage/query/default changes | N/A | N/A |
| BE-05 | Existing production auth/routes preserved; test-only controls protected | PASS | PASS |
| FE-01 | Pure domain helpers; Vuex/WS owns device truth | PASS | PASS |
| FE-02 | Native typed values and unchanged payload tests | PASS | PASS |
| FE-03 | Only confirmed light reports; existing offline feedback/reconnect | PASS | PASS |
| FE-04 | Existing named UiSelect; keyboard/desktop/mobile E2E | PASS | PASS |
| FE-05 | Existing auth, text summaries, no secrets/household fixtures | PASS | PASS |

## Contracts, data, rollout

No production schema, endpoints, environment variables, persistence, or MQTT changes.
Backend producer already provides `values.toggle`; UI stops discarding it. Deploy
UI independently for Toggle; deploy the backend for the callback and empty-list
fixes. No migration is needed. Rollback restores previous behavior without changing saved
automations. Unknown/unsupported values remain preserved until explicitly edited.
At save, a zero delay becomes absence, the backend's existing representation of
immediate execution; positive delays are unchanged. No engine semantics change.
Fixture-only health/control endpoints never compile into the application binary.

## Traceability and verification

| IDs | Paths | Procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02/04, NFR-01 | exposes.ts, automation.ts, editor inputs | Companion Jest regressions; bounded helper review | PASS: native/absent/invalid/duplicate values; commands excluded from readings/conditions; keyboard input at both widths |
| FR-02, AC-03, NFR-02 | hub_test.go, hubharness, automations.spec.ts | UI creates/saves/reloads; identical dial events; native commands; confirmed On/Off | PASS: both viewports, then three repeats each under Go race instrumentation; unauthorized fixture controls return 401 |
| AC-05 | Existing domain/scenario/browser suites | Round-trip, disabled/ignored/offline/reconnect/confirmation and leak-cycle checks | PASS: full suites below; unchanged valid automation payload retained |
| AC-06 | contracts/automations.ts, Vuex save | Companion contract/store tests and real-engine save acknowledgement | PASS: zero delay omitted once at shared save boundary; positive delay and draft preserved |
| AC-07 | store/store.go | Concurrent registration/write and reentrant callback regressions | PASS: race reproduced before fix; full race suite and race-instrumented browser startup/shutdown clean afterward |
| AC-08 | internal/automations/trigger.go | Empty/replacement/invalid list regression; UI empty-save/reload/add-trigger | PASS: [] decoded as null before fix; companion and final browser cases pass afterward |
| NH-05, SC-01 | frontend/backend/workflow | Checks below | PASS locally; remote CI and real devices unrun |

Actual execution (2026-10-22, macOS, Node 24.21.0, Go application toolchain 1.24.8):

- `npm test -- --runInBand --no-cache`: 49 suites, 326 tests passed.
- `npm run lint`: passed, seven pre-existing script warnings; no new warnings.
- `npm run build`: passed, including Vue/TypeScript checks; existing DeviceCard
  component-name warning remains.
- `GOTOOLCHAIN=go1.24.8 go test -count=1 ./...`: passed on final backend.
- `GOTOOLCHAIN=go1.24.8 go test -race -count=1 ./...`: passed on final backend.
- `GOTOOLCHAIN=go1.24.8 go vet ./...` and `go build ./...`: passed on final backend.
- `CI=true npx playwright test`: 179 passed, nine existing skips, 4.1 minutes.
  After the final Go fixes, reran all eight desktop/mobile automation cases: passed
  in 13.6 seconds. Other browser cases use the unchanged Node mock.
- Built the final hub fixture with `go test -c -race`; ran the two Go integration
  browser cases with `E2E_EXTERNAL_SERVERS=1`, `--repeat-each=3`: 12 passed in
  18.4 seconds. Direct test binary received SIGINT, returned PASS/exit 0, and
  produced no race reports. Node/Vite/process groups were also stopped.
- Pinned actionlint 1.7.12 (built with the installed Go 1.26.1, matching the
  checker's Go >=1.25 requirement): passed. `git diff --check`: passed.

Regression evidence: the original action summary said Unknown; native-option helper
regressions initially failed. Real-backend browser save then exposed zero-delay
rejection. The isolated race run reproduced callback registration versus markDirty;
its source-owned race regression failed before the lock/snapshot fix. The empty-list
unit test failed with [] encoded as null before the decoder fix. All now pass.
Temporary local execution logs are under `/private/tmp/nodeherder-toggle-*`; enduring
coverage is in the committed tests. No household device, coordinator, broker, or
production deployment was used. GitHub CI/container builds were not run for this
local branch; Dockerfiles, dependencies, and deployment configuration are unchanged.

## Debt, risks, exceptions

Existing scenario test filenames predate companion ownership; reuse them without
adding new files there. Test server lifecycle and cookie/WS origin behavior must be
verified using real auth/routes (passed). Test-only fixture routes are absent from
application binaries; normal Go suites skip the long-running fixture. Existing
empty-recipe and zero-wait issues found by integration are fixed, not deferred.
No exceptions requested or approval invented.

## Completion

Implemented on `fix/restore-automation-toggle`, based on latest main `b60d2bdb`.
All acceptance scenarios have passing local evidence above. `docs/ci.md` documents
the real-engine browser fixture and its isolation/auth/toolchain requirements.
Production engine timing, retry policy, MQTT commands, state confirmation, and MCP
behavior are unchanged. Branch checkout/commit details are reported in the handoff.
