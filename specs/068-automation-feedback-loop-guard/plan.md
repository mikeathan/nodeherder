# Plan: Automation feedback-loop guard

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: mikeathan | Updated: 2026-11-25
Constitutions: parent 1.0.0; backend 1.0.0

## Context and design

Ownership stays in `backend/internal/automations`. `Device.Evaluate` is the single
device-selection path; `DeviceTrigger.Process` is the single execution path. Manual runs
never enter `Evaluate`: `AutomationEngine.HandleManual` calls `Device.EvaluateTrigger`,
which calls `Process` directly.

Affected paths:

- `backend/internal/automations/device.go` — choose execution vs dependency-only path;
  GoDoc for both exported evaluation methods.
- `backend/internal/automations/trigger.go` — add the unexported dependency-only path;
  document the invariant; `AffectedBy` and `conditionsPermitExecution` both read the
  primary name through `GetName()` (selection semantics unchanged).
- Tests: `backend/internal/automations/device_test.go`, `trigger_test.go`; existing
  `scenarios/dial_test.go` and `scenarios/presence_test.go` cover AC-07 unchanged.
- Docs/guidance: `backend/docs/architecture.md` (authoritative invariant), `AGENTS.md`,
  `.agents/rules/go-engineer.md` (short pointers).
- Historical pointers: `specs/010-presence-and-connection-recovery/{spec,plan,tasks}.md`
  and `specs/027-fast-e2e-jobs/{plan,tasks}.md` carry truthful superseded/correction
  notes; they are annotated, not reopened.
- `.github/workflows/ci.yml` — the branch's separate CI commit `b25266b2` repairs the
  Playwright image guard; review corrections derive the expected version from the job's
  own image tag. No application contract change.

Design (smallest suitable):

1. In `Device.Evaluate`, compute `primary` from the event payload
   (`payload[trigger.GetName()]`).
2. Primary: `trigger.Process(ctx)` — unchanged.
3. Dependency-only: if the trigger implements the unexported
   `processDependencyChange(ctx)` path, call it; otherwise do nothing (no device command).
4. `DeviceTrigger.processDependencyChange` runs `conditionsPermitExecution(ctx)` so a
   failing condition still calls `stopActions()` (cancellation), then returns without
   executing actions. It has no error return: nothing on that path can fail, so
   `Evaluate`'s failure flag only reflects `Process`.

Observed facts: selection already reaches the trigger for dependency changes, and
`conditionsPermitExecution` already owns cancellation. The change only skips the
action-start step, so there is no new state or ordering. `HasValueChanged` may end
condition evaluation early; that behaviour is retained and documented in spec FR-02.

Alternatives considered:

- Future manual-override latch (Option B): restores dusk auto-on but needs mutable
  provenance state and a spec-010 amendment; deferred to its own spec.
- Timed settle window: rejected — stateful and can re-fire after expiry.
- Full revert of `AffectedBy`: rejected — discards cancellation on dependency changes and
  broadens test churn.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | spec 068 scenarios/ACs | PASS (design) | PASS (spec + tasks) |
| NH-02 | primary vs dependency execution separated; repeated events and manual paths unchanged; loop prevention specified | PASS (design) | PASS (tests) |
| NH-03 | no trust/entry-point or data-exposure change | N/A (no change) | N/A |
| NH-04 | established automations boundaries; no external contract change | PASS (design) | PASS (no contract diff) |
| NH-05 | spec ACs map to named source-owned tests; design evidence only here | PASS (design) | PASS (tests below) |
| NH-06 | no persistence/resource/lifecycle change | N/A (no change) | N/A |
| BE-01 | domain logic only; no transport change | PASS (design) | PASS |
| BE-02 | dependency-only vs primary paths designed through the lifetime service | PASS (design) | PASS (tests) |
| BE-03 | no goroutines/timers/shared state added | PASS (design) | PASS (no new concurrency) |
| BE-04 | no storage/query change | N/A | N/A |
| BE-05 | no route/assistant change | N/A | N/A |

## Contracts, data, rollout

No HTTP/WS/MQTT/MCP/config/persistence change. `.github/workflows/ci.yml` changes only a
CI guard step: it now reads the Playwright version from the job's own container image tag,
so that tag is the single source of truth for the comparison. Workflow job names, artifacts
and required checks are unchanged. No migration or deployment ordering. Rollback is a code
revert; no data to restore.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, AC-02, AC-06 | `internal/automations/device_test.go` | `TestIlluminanceOnlyChangeDoesNotStartActions` (asserts both evaluation and no command), `TestPresenceConditionReportOrdering` | pass |
| FR-02, AC-03, AC-06 | `device_test.go` | `TestConditionDependencyCancelsButDoesNotStartDelayedAction`, `TestPresenceReturnAfterDelayedOffWithLateDarkReading`, `TestPresenceDependencyDoesNotRestartOrRetainOffDelay` | pass |
| FR-01, AC-04 | `device_test.go`, `trigger_test.go` | `TestManualAndDeviceRunsRemainIndependent`, `TestManualTriggerRunsDespiteConcurrentDeviceMessage` | pass |
| FR-03, AC-05 | `device_test.go`, `trigger_test.go` | `TestConditionChangesNeverReplayCachedPhysicalEvents`, `TestDeviceTriggerOwnsSingleConditionEvaluation`, `TestCustomTriggerExecutesOnlyOnPrimarySelection` | pass |
| AC-07 | `scenarios/dial_test.go`, `scenarios/presence_test.go` | existing golden/scenario suites | pass |
| NFR-01/02 | whole diff | no state/contracts added; review | pass |
| SC-01 | `backend/` | `go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./internal/automations/... ./internal/services/...` | pass |
| CI guard | `.github/workflows/ci.yml` | derive the tag from the image, compare to installed `@playwright/test` | pass (`sh -e` simulation, shellcheck) |

Before-fix evidence (reverted `device.go`/`trigger.go` hunks in a scratch copy):
`TestIlluminanceOnlyChangeDoesNotStartActions`, four `TestPresenceConditionReportOrdering`
subtests, `TestConditionDependencyCancelsButDoesNotStartDelayedAction`,
`TestDeviceTriggerOwnsSingleConditionEvaluation`,
`TestCustomTriggerExecutesOnlyOnPrimarySelection` and
`TestActionFailureReportingFollowsSelection` fail; after the fix they pass.
`TestPresenceReturnAfterDelayedOffWithLateDarkReading` and
`TestPresenceDependencyDoesNotRestartOrRetainOffDelay` passed pre-fix too; the 2026-11-25
review corrections added dependency-only darkening steps so they now fail pre-fix as well.

## Debt, risks, exceptions

- Accepted loss: auto-on when the room darkens while someone is already present; recovered
  only by Option B (future). Recorded in spec Q-01.
- Risk: radar flicker/re-acquisition produces a fresh `presence` change and may re-light.
  This is a genuine presence event and is left unchanged.
- `Seed` forwards a whole first snapshot; primary keys therefore read as changed on lane
  creation/restart. Pre-existing, documented in `backend/docs/architecture.md`.
- Historical specs 010 and 027 carry superseded/correction pointers. They are annotated,
  not reopened, consistent with `docs/sdd/README.md` branch-scope guidance.
- No exceptions requested.

## Completion

Acceptance evidence is recorded in `tasks.md`. The automation invariant has one
authoritative home in `backend/docs/architecture.md`, with short pointers in `AGENTS.md`
and `.agents/rules/go-engineer.md`. No approved exception required.

Checks run 2026-11-23 by the author (macOS; Node 26.0.0 available, repo pins 24); not
repeated during the 2026-11-25 review apart from the backend commands listed below:

Backend (`backend/`):

- `go mod verify` — all modules verified
- `go build ./...` — pass
- `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .` — pass
- `go vet ./...` — pass
- `go test -count=1 -timeout=10m ./...` — 53 packages ok, no failures
- `go test -race -count=1 -timeout=10m ./...` — exit 0, 53 packages ok, no data races
- `go test -run '^$' ./internal` — pass (CI fixture compile check)

Frontend (`frontend/`):

- `npm run lint` — 0 errors (113 pre-existing warnings)
- `npm test -- --ci --runInBand` — 162 suites / 1097 tests pass
- `npm run type-check:e2e` — pass
- `npm run build -- --mode production` — pass
- `npm run test:e2e` (desktop + mobile, all specs) — 421 passed, 9 skipped, exit 0

Repository (`/`):

- `actionlint` v1.7.12 — pass; `bash -n` all tracked scripts — pass;
  `shellcheck scripts/manage-stack.sh` — pass
- `node scripts/sync-reporting-defaults.mjs --check` + test — pass;
  `node --test scripts/dev.test.mjs` — 5 pass

Checks run 2026-11-25 during review (reported by the reviewer, then repeated locally):

- `gofmt -l` on the four changed Go files — clean
- `go build ./...`, `go vet ./...` — pass
- `go test -count=1 ./...` — 53 packages ok
- `go test -race -count=1 ./internal/automations/... ./internal/services/...` — pass
- `git diff --check` — clean
- `actionlint .github/workflows/ci.yml` (installed locally for this review) — clean
- CI guard extracted and exercised under `sh -e` for five tag/package combinations
  (match, package-only bump, image-only bump, coordinated bump, malformed tag);
  `shellcheck -s sh` on the guard script — clean
- Pre-fix regression: `main` copies of `device.go`/`trigger.go` with the current tests
  fail all eight targeted tests, including the two strengthened ones
- 2026-11-25 second independent review pass (agent, not human approval): repeated
  `go build ./...`, `go vet ./...`, `go test -count=1 ./...` and
  `go test -race ./internal/automations/... ./internal/services/...` (pass); a scratch
  copy with `origin/main` `device.go`/`trigger.go` again fails the eight targeted tests;
  both CI-guard mismatch cases re-exercised under `sh -e`. No code finding required a
  change.

Not run (environment-limited or unrelated to this diff):

- `scripts/manage-stack.test.cjs` fails locally because the working copy `.env` sets
  `TRUSTED_PROXY_CIDRS` (CI runs without it).
- `scripts/manage-stack.compose.test.cjs` and container image jobs need a running
  Docker daemon.
- `Tag Build` (the CD step) runs only on `main` pushes and creates a tag; per AGENTS.md
  it must not be run from a PR branch.
- Frontend unit/lint/build and Playwright suites were not re-run on 2026-11-25: no frontend
  source changed. The 2026-11-23 rows above are the author's evidence, not the reviewer's.
