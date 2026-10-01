# Plan: Pre-merge verification

Spec: [spec.md](spec.md) | Status: Implemented; baseline failures block green CI | Updated: 2026-10-01
Constitutions: NH/BE/FE 1.0.0 (proposed)

## Design and rollout

Independent backend/frontend/race/tooling/container jobs retain repository toolchains;
actionlint uses separately supported Go. Images use Docker Buildx, actual component
contexts, isolated GHA caches, explicit CI version, load-only export, no registry
login. Smoke containers use `--network none`: backend `-h` exits before initialization;
nginx gets an isolated host mapping for config validation. Compose tests retain
disposable credentials/unreachable daemon.

Keep `test` / `Run Tests` as an always-evaluated aggregate. Main-push tagging depends
on it; document branch protection rather than mutating settings. No path-based skips.
Cancel superseded PR runs only, preserving main tagging. No runtime changes.

## Constitution Check

| Rule | Design evidence | Final |
| --- | --- | --- |
| NH-01 | Acceptance/scope recorded before workflow edits | Artifacts trace all workflow changes |
| NH-02 | Mocks/offline containers; no devices | Mocks/config suites pass; image commands statically reviewed, unrun |
| NH-03 | Read-only checks; no deployment secrets/publishing | Workflow review confirms read-only checks/load-only images |
| NH-04 | Existing status/tag interface retained; rollout documented | `Run Tests` retained; docs explain external required-check setting |
| NH-05 | Requirements map to checks; failures reported | Local checks recorded below; existing failures block green CI, not suppressed |
| NH-06 | Timeouts/caches/scoped cancellation; no production mutation | All jobs bounded; gate fails on failed/skipped/cancelled/empty results |
| BE-01–05 | N/A: no Go changes; verify existing domain/lifecycle tests | N/A |
| FE-01–05 | N/A: no browser changes; production checks added | N/A |

## Verification and debt

FR-01: local frontend lint/Jest/build and backend vet/build/race.
FR-02/04/05: actionlint/workflow review, stack-manager suites/syntax.
FR-03: Docker/smoke static review; real image builds run in GitHub, not household services.
Ubuntu/Buildx/ShellCheck checks remain unrun locally unless evidence says otherwise.

Backend vet/race failures need separate fixes before strict CI can be green. Existing
Go formatting debt stays outside this change. Frontend lint currently checks JS/config,
not comprehensive Vue/TypeScript rules; production build checks Vue/TypeScript compilation.
Branch protection is an external maintainer setting. SDD/constitutional review stays
manual. No exceptions/approval fabricated.

## Actual results (2026-10-01)

| IDs/check | Local result |
| --- | --- |
| FR-01 frontend | `npm run lint`: pass, 7 warnings; `npm test -- --ci --runInBand`: 9 suites/44 tests pass; `npm run build`: Vue/TS and Vite pass, bundle-size warning |
| FR-01 backend | `go mod verify`: pass; `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=readonly go build -o /private/tmp/nodeherder-ci-verification .`: pass; `go test -count=1 -timeout=5m ./...`: pass |
| FR-01 new strict checks | `go vet ./...`: fail, missing cancel in `utils/worker_test.go:38`; `go test -race -count=1 -timeout=5m ./...`: fail, existing races/concurrency crashes; Go 1.26.1/macOS, not the CI toolchain/OS |
| FR-02 | `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12 -shellcheck= .github/workflows/ci.yml`: pass; Bash syntax passes for all 4 tracked scripts; 56 mocked CLI tests and 11 real Compose config cases pass |
| FR-04/05 | YAML/needs/status/tag/timeout review passes; actual aggregate shell tested against success/failure/cancel/skip/empty inputs, all expected exit codes |
| FR-03/remaining FR-02 | Real image builds/offline smoke, Ubuntu runner execution, and ShellCheck (including inline actionlint integration) unrun locally; current environment has no Buildx/daemon or ShellCheck |

No backend runtime/test fixes are part of this change. Green CI/merge readiness is
not claimed. Follow-up remediation must address the vet/race findings with scoped
tests; maintainers still need to require the aggregate status in GitHub settings.

## CI failure follow-up

Ubuntu CI reported the expected vet/race failures; aggregate `Run Tests` correctly
failed with two failing prerequisites. Fixed the worker-test discarded cancellation,
joined worker cleanup, removed the shared completion flag, and replaced cancellation
test sleep/counter races with start/release signals and atomic counts. No production
worker or CI gate was changed. `go vet ./...` and 20 repeated race runs of both worker
tests pass locally on Go 1.26.1/macOS.
Follow-up verification also passes: `go test -race -count=1 -timeout=1m ./utils`,
`go build ./...`, and `go test -count=1 -timeout=2m ./...`.

Diagnostic command: `GORACE=halt_on_error=1 go test -race -count=1 -timeout=2m ./...`.
It still fails: automation scheduler writes `BaseAutomation.Enabled` while tests read
it; device mapper reads/writes its map without synchronization; legacy
`github.com/boltdb/bolt` triggers checkptr allocation crashes in storage-backed tests.
These are local findings, not an assertion that the user's truncated Ubuntu log
contains every same failure. Production synchronization and storage compatibility
changes need a separately scoped repair/verification plan; do not disable checks.
