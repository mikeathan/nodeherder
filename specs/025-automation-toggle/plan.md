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

No production persistence schema, endpoints, environment variables, or MQTT payloads
change. Backend producer already provides `values.toggle`; UI stops discarding it.
Automation JSON normalizes an empty trigger list to `[]` (previously `null`); existing
consumers treat both as empty and no persistence migration is required. Deploy
UI independently for Toggle; deploy the backend for the callback and empty-list
normalization. No migration is needed. Rollback restores previous behavior without
changing saved automations. Unknown/unsupported values remain preserved until explicitly
edited. At save, a zero delay becomes absence, the backend's existing representation of
immediate execution; positive delays are unchanged. No engine semantics change.
Fixture-only health/control endpoints never compile into the application binary.
The frontend package minor version is bumped from 0.12.0 to 0.13.0 in
`frontend/package.json` and its lockfile, so the next release tag starts a new minor
line. `scripts/tag_build.sh` and `scripts/manage-stack.sh` read that package version;
no separate version file exists.

## Traceability and verification

| IDs | Paths | Procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02/04, NFR-01 | exposes.ts, automation.ts, editor inputs | Companion Jest regressions; bounded helper review | PASS: native/absent/invalid/duplicate values; commands excluded from readings/conditions; keyboard input at both widths |
| FR-02, AC-03, NFR-02 | hub_test.go, hubharness, automations.spec.ts | UI creates/saves/reloads; identical dial events; native commands; confirmed On/Off | PASS: both viewports, then three repeats each under Go race instrumentation; unauthorized fixture controls return 401 |
| AC-05 | Existing domain/scenario/browser suites | Round-trip, disabled/ignored/offline/reconnect/confirmation and leak-cycle checks | PASS: full suites below; unchanged valid automation payload retained |
| AC-06 | contracts/automations.ts, Vuex save | Companion contract/store tests and real-engine save acknowledgement | PASS: zero delay omitted once at shared save boundary; positive delay and draft preserved; validation copy explains zero |
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

Follow-up review corrections rerun on the final files (available Go 1.26.1,
Node 26.0.0; the repository-pinned toolchains were used for the original runs above):

- Frontend `npm test -- --runInBand`: 49 suites, 327 tests passed (new zero-wait
  validation-copy regression). `npm run lint`: passed, same seven pre-existing
  warnings. `npm run build`: passed after the TypeScript 6 config migration and
  dependency upgrades; Vite prints its default >500 kB chunk warning.
- Playwright/e2e TypeScript: `npm run type-check:e2e` passed after making node types
  explicit and typing the blocked-storage stub; `npm ci --dry-run` reports the lock
  in sync.
- TypeScript 6.0.3 config diagnostics (`rootDir`, `baseUrl`, `moduleResolution=node10`)
  are gone; `type-check:e2e` is clean under TS 6 and the TypeScript next build, and
  the project build uses `vue-tsc`, which passes.
- Backend `gofmt`, `go vet ./...`, `go test -count=1 ./...`, and
  `go test -race -count=1 ./...`: passed.
- `go test -run '^$' ./internal`: passed; this is the CI fixture warm-up and does
  not start the browser server.
- Playwright `e2e/automations.spec.ts` desktop and mobile: 8 passed. The dial case
  now leaves the light ON and verifies `/__test/reset` returns 204 and reports OFF.
- Pinned actionlint 1.7.12 on the updated workflow and `git diff --check`: passed.
- Dependency audit: `npm audit --omit=dev` reports zero vulnerabilities; the full
audit reports zero high and 20 moderate dev-only advisories (see residual risk).
- Post-upgrade browser subset: automations + panel desktop and mobile (28 passed),
  plus Home and DeviceDashboard desktop (8 passed).
- After the Aura subset and nginx gzip change: full `npx playwright test` passed
  (179 passed, nine existing skips, 4.1 minutes); main entry 501.7 -> 438.8 kB
  (149.4 -> 141.5 kB gzip) and the Vite chunk-size warning is gone.
- `frontend/config/nginx.conf` gzip was validated in the built frontend container:
  `nginx -t` passed and a JS response returned `Content-Encoding: gzip`.

CI-equivalent local run (Node 24.21.0, Go 1.24.8):

- Workflow lint: actionlint 1.7.12, `bash -n` on tracked scripts, shellcheck 0.11.0:
  passed.
- Stack tools: reporting-defaults check, 4 reporting tests, 61 deployment tests,
  and 12 real Compose config tests: passed. Compose used the standalone
  `docker-compose` adapter because this host's Docker CLI lacked plugin discovery;
  CI uses the Ubuntu plugin.
- Backend: `go mod verify`, Linux/amd64 build, tests, vet, and race detection:
  passed.
- Frontend: `npm ci`, lint, Jest 49 suites/327 tests, `type-check:e2e`, and production
  build: passed.
- Browser: `CI=true npx playwright test` passed 179 tests with nine existing skips
  in 4.3 minutes.
- Containers: backend/frontend images built; `-h`, `nginx -t`, index presence, and
  gzip response smoke checks passed. Images were built natively (arm64) because the
  local Docker CLI lacks buildx; GitHub's linux/amd64 buildx job remains unrun.
- Remote GitHub CI and GitHub-hosted runner specifics remain unrun; the warm-up plus
  300s startup timeout is the pre-merge CI verification step above.

Regression evidence: the original action summary said Unknown; native-option helper
regressions initially failed. Real-backend browser save then exposed zero-delay
rejection. The isolated race run reproduced callback registration versus markDirty;
its source-owned race regression failed before the lock/snapshot fix. The empty-list
unit test failed with [] encoded as null before the decoder fix. All now pass.
Temporary local execution logs are under `/private/tmp/nodeherder-toggle-*`; enduring
coverage is in the committed tests. No household device, coordinator, broker, or
production deployment was used. GitHub CI/container builds were not run for this
local branch; Dockerfiles and deployment configuration are unchanged. Frontend
dependencies were upgraded for security and TypeScript support (see follow-ups).

## Debt, risks, exceptions

Existing scenario test filenames predate companion ownership; reuse them without
adding new files there. Test server lifecycle and cookie/WS origin behavior must be
verified using real auth/routes (passed). Test-only fixture routes are absent from
application binaries; normal Go suites skip the long-running fixture. Existing
empty-recipe and zero-wait issues found by integration are fixed, not deferred.
No exceptions requested or approval invented.

Follow-up review corrections on the same branch: the browser E2E job warms the Go
fixture build before Playwright starts it and allows 300s for fixture startup, so a
cold build is not reported as a server failure. Fixture reset drains stale control
queues and reports the light OFF when a failed case left it ON. Round-trip comments
now state the intentional zero-delay exception, the delay validation copy explains
zero, and the plan contract wording covers empty-list normalization. Node types are
now explicit, `tsconfig.e2e.json` type-checks the Playwright config/tests, and the
frontend CI job runs `type-check:e2e`; the blocked-storage browser stub is typed
explicitly. The root TypeScript config now sets `rootDir`, uses relative `paths`,
and uses `moduleResolution: bundler` (the e2e config uses `rootDir: "."`), so
TypeScript 6 no longer reports migration errors in the editor; a vuex types path
covers its package-exports gap. The frontend package version is 0.13.0. No
device-command or automation semantics changed.

Dependency audit follow-up: the high advisories are resolved by Vue 3.5.44,
eslint-plugin-vue 10 + vue-eslint-parser 10, unplugin-vue-components 32,
Jest 30/ts-jest 29.4/@types/jest 30, and a `qs` 6.16.x override; `npm audit --omit=dev`
is clean. Twenty moderate advisories remain in the dev-only Jest coverage chain
(`@istanbuljs/load-nyc-config` -> js-yaml 3 -> argparse 1 -> sprintf-js); sprintf-js
has no patched version, and `npm audit fix --force` only suggests incompatible
downgrades. Mitigation: dev-only, not reachable from app traffic or untrusted format
strings; revisit when Jest/istanbul replace the chain or coverage tooling changes.

The Vue security patch initially lifted the main bundle to about 502 kB. Importing only
the Aura component token modules the app uses (rather than all 100) reduces the main
entry chunk to about 439 kB (141 kB gzip) and removes Vite's 500 kB warning. `nginx`
now gzips text assets and proxied JSON/JS, so the wire transfer for that chunk is
roughly 141–145 kB at gzip level 6. Add the matching Aura token module when adding a
new PrimeVue component.

## Completion

Implemented on `fix/restore-automation-toggle`, based on latest main `b60d2bdb`.
All acceptance scenarios have passing local evidence above. `docs/ci.md` documents
the real-engine browser fixture and its isolation/auth/toolchain requirements.
Production engine timing, retry policy, MQTT commands, state confirmation, and MCP
behavior are unchanged. Branch checkout/commit details are reported in the handoff.
