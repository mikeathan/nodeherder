# CI and pre-merge checks

[Workflow](../.github/workflows/ci.yml) · [Development policy](sdd/README.md)

CI runs on Ubuntu for pull requests to `main`, pushes to `main`, and merge groups.
Independent jobs provide feedback in parallel; container matrix failures do not
cancel the other component. Superseded PR runs are cancelled; main runs are not.

| Check | Errors caught |
| --- | --- |
| Workflow and Shell Checks | Invalid Actions syntax/expressions/inline shell; Bash syntax in tracked scripts; stack-manager ShellCheck findings |
| Deployment Tooling (Ubuntu) | Local launcher startup/shutdown/descendant cleanup; generated reporting-defaults drift and generator regressions; CLI scope/failure/dry-run/version regressions; real Compose merge, required variables, paths/ports/network using dummy fixtures |
| Backend Tests and Build | Go module checksum changes, Linux compilation failures, uncached unit regressions, `go vet` findings |
| Backend Race Detection | Concurrent accesses exercised by backend tests; not proof that all races are absent |
| Frontend Tests, Lint and Build | Locked install failures, existing ESLint errors, Jest regressions, Vue/browser-test TypeScript compilation and production bundle failures |
| Frontend E2E (desktop/mobile, shard 1/4) | Broken user flows in a real Chromium at desktop and phone sizes: OAuth sign-in/out, dashboards and device commands, automation save payloads, device pages; UI-created dial toggles also execute through the real Go engine with in-memory MQTT (never real devices) |
| Container (backend/frontend) | Real Dockerfile/context/build failures on Linux/amd64; backend executable help; nginx config and built index smoke checks |

Images are built/loaded on disposable runners, never published. Smoke containers
use `--network none`; backend help exits before application initialization. nginx's
test-only host mapping resolves its proxy name without accessing a backend.
No production credentials, broker, Zigbee coordinator, or household devices are used.
Go/Node use repository toolchain files; Dockerfiles retain their builder versions.
Dependency and component-specific image caches reduce repeat cost. Jobs have explicit
timeouts; verification gets read-only repository permission. Only the existing
main-push tagging job can write, after all checks succeed.

## E2E elapsed time and isolation

E2E uses eight jobs: desktop/mobile × file-level shards 1/4. Each job starts a
fresh mock hub, isolated Go automation hub, and Vite server, with one Playwright worker. Keep `workers: 1` and
`fullyParallel: false`: the hub has mutable shared state, and files retain their
internal execution order. Matrix failures do not cancel other partitions. All four
must pass through the existing `Run Tests` aggregate, on PRs, main and merge groups.
The same tests, assertions, leak-cycle counts, retries and deadlines remain enabled.

E2E jobs run inside `mcr.microsoft.com/playwright:v<version>-noble`, which ships
Chromium and every system library it needs. Nothing is installed from apt at run
time: doing that per job previously ranged from 20 seconds to 11 minutes depending on
hosted mirror load, and was the dominant cost of the whole workflow. The image tag
must equal the `@playwright/test` version in `frontend/package.json`; a dedicated step
fails with the fix if they drift. Container jobs use `--ipc=host`, as Playwright
recommends for Chromium.

Playwright manages ports 4100 (UI), 4110 (Node mock), and 4120 (Go fixture).
The Go toolchain comes from `backend/go.mod`; CI installs it alongside Node and
compiles the fixture test binary before Playwright starts it. Fixture startup allows
300s, so a slow cold build is not reported as a server failure.
`backend/internal/hub_test.go` skips its long-running fixture unless
`NODEHERDER_BROWSER_HUB=1`. It binds loopback, uses production auth/WS/routes and
`testing/hubharness` with synthetic dial/light metadata and in-memory MQTT. Its
control routes exist only in the test binary and require the normal session cookie.
Only the relevant page's API/WS traffic is redirected to this fixture; all other
browser tests retain the Node mock. Two identical dial events must publish native
TOGGLE commands and receive separate MQTT On/Off confirmations in the UI. These
fixtures are reset between cases; they never load household automation configs.

For an externally managed run (`E2E_EXTERNAL_SERVERS=1`), start all three isolated
servers; start the Go fixture from `backend/` with
`NODEHERDER_BROWSER_HUB=1 go test -count=1 -timeout=20m -run '^TestAutomationBrowserHub$' ./internal`.
The fixture uses localhost ports 4100/4120 for auth/origin checks, so keep those
ports for its automation cases. Stop it with SIGINT/SIGTERM when finished.

Each completed partition uploads `playwright-timings-<project>-<shard>` containing
`timings.json` with per-test durations, including successful runs. Failed partitions
also upload `playwright-report-<project>-<shard>` with HTML, traces and screenshots.
Retention is seven days. This lets a future slowdown be traced to a test rather
than inferred from the total job time.

Measured before this change in [PR run 39298242659](https://github.com/mikeathan/nodeherder/actions/runs/39298242659)
(2026-10-19):

| Work | Elapsed time |
| --- | --- |
| Whole PR workflow | 6m57s |
| E2E job, desktop then mobile on one worker | 6m45s |
| Browser test step, including server startup/reporting | 6m08s |
| Chromium/system dependency install | 21s |
| Backend ordinary/race jobs, already in parallel | 1m28s / 1m50s |
| Frontend unit/lint/build job | 37s |
| Backend/frontend container jobs | 51s / 1m22s |

After the Go browser fixture landed, [main run 39781252241](https://github.com/mikeathan/nodeherder/actions/runs/39781252241)
showed where the time actually went: E2E jobs finished their tests in about two
minutes, but the `playwright install --with-deps` step took 11m24s and 11m10s on two
of the four jobs while the other two took 18–21s. The run therefore ended at 14m15s
because of apt mirror variance, not test cost. The Playwright container image removes
that step entirely, so a shard is now compile, install dependencies, and about two
minutes of tests.

The E2E log contains 155 passed cases and nine existing desktop skips for phone-only
checks. Entity-dialog tests cost about 145s across both viewports, including about
50s for the two 50-cycle leak checks. Confirmation-timeout tests cost about 31s;
the home light tests wait for periodic mock updates (~20s combined). Layout checks
visit eleven routes and use `networkidle` (~20s combined). These checks retain their
current behavior. Ordinary backend package execution is heaviest in automations
(21.7s) and controllers (16.3s); some tests use real-time waits. Compilation and
package scheduling also contribute to job duration.

The matching [main run 39299251928](https://github.com/mikeathan/nodeherder/actions/runs/39299251928)
took 7m15s, with E2E at 6m53s. The earlier main run 39053480876 took 4m22s,
with E2E at 4m02s; the intervening change expanded entity-dialog/history coverage.
Main verifies the merged commit with the full suite again, then tags. The current
adjustment speeds both PR and main runs. Based on the measured file durations, the
partitions should take roughly 1.5–2 minutes each including setup once the
Playwright image removes the install step. Four shards per project are used because
the remaining floor is fixed overhead (about 33s container start plus Go/Node setup)
and ~30 tests per shard, so more shards buy little while the repository is public and
run minutes are not billed. Four E2E
runners add installation/runner overhead and may queue under account concurrency
limits. File-level sharding can become uneven as suites grow; inspect timing reports
before increasing partition count. Tests execute afresh; existing dependency/build
caches remain enabled.

Reproduce one partition from `frontend/`, using Node from `.nvmrc` and only isolated
mock servers (do not run these commands concurrently on the same default ports):

```bash
CI=true PLAYWRIGHT_JSON_OUTPUT_FILE=test-results/timings.json \
  npm run test:e2e -- --project=desktop --shard=1/4 --reporter=list,html,json
```

Repeat for desktop shards 2–4 and mobile shards 1–4. See
[spec 023](../specs/023-ci-and-device-controls/plan.md) for discovery, local execution and
remaining rollout verification. Sharding follows [Playwright's documented file-level behavior](https://playwright.dev/docs/test-sharding).

Local validation (macOS/Node 24): the unchanged suite passed in 3.7 minutes;
individual fresh-server partitions passed in 54.4s, 51.4s, 66.7s and 54.2s.
The executed reports cover the same 164 cases exactly once (155 passes, nine
existing skips). GitHub installation and queue time are excluded from these results.

## Enforce before merge

In GitHub branch protection/rulesets, require **Run Tests**, the retained aggregate
status. It fails if any verification job fails, is cancelled, or is skipped; the
tag job depends on it. Workflow configuration alone does not enforce merge blocking.
Maintainers must configure required checks; this change does not alter GitHub settings.
Do not require `Tag Build` for PRs—it runs only on main pushes.

## Local reproduction and current gaps

```bash
# backend/
go mod verify
go build ./...
go test -count=1 -timeout=5m ./...
go vet ./...
go test -race -count=1 -timeout=5m ./...

# frontend/ (Node from .nvmrc)
npm ci
npm run lint
npm test -- --ci --runInBand
npm run build -- --mode production
```

Jest's global setup sets `TZ=Europe/London` before workers start, including on UTC CI runners.
History's spring/autumn DST and 33-date assertions run unconditionally; launch with `TZ=UTC`
to verify the setup is effective.

From the root: `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`,
`shellcheck scripts/manage-stack.sh`, and the [stack-manager checks](setup.md#stack-manager).
Actionlint requires Go 1.25+; CI installs it separately from the application toolchain.

Baseline verification on 2026-10-01: frontend lint passes with seven warnings,
9 Jest suites/44 tests pass, and production type-check/build passes. Backend module
verification, Linux/amd64 cross-compilation, and uncached ordinary tests pass. Backend vet
initially reported missing context cancellation in `utils/worker_test.go` (now fixed
along with test-local worker synchronization); race tests expose
existing backend concurrency failures on Go 1.26.1/macOS. Strict checks intentionally
report these failures rather than suppress them; scoped fixes are needed for green CI.
Full Ubuntu/GitHub image execution has not yet been observed locally.

2026-10-02 follow-up: bbolt v1.4.3 replaces legacy Bolt; synthetic old-file
read/write/reopen tests and storage race suites pass on macOS Go 1.26.1/1.24.8.
All ordinary backend tests, vet and Linux/amd64 cross-build pass. Full race checks
still report runtime concurrency failures, but no longer hit the Bolt pointer-check
crash. Existing test assertions/CI checks remain enabled; see
[scoped repair](../specs/003-backend-race-repair/plan.md) for evidence/rollout limits.

Limits: frontend ESLint currently covers JS/config, not comprehensive Vue/TypeScript
rules; the build checks compilation. Image smoke checks do not establish end-to-end
API/WS/auth/device behavior or ARM64 support. Coverage thresholds, vulnerability/secret
scanning, and browser contract tests are useful follow-ups, not claimed checks.
SDD acceptance and constitutional review remain manual. See the
[verification spec](../specs/002-ci-verification/spec.md) for scope/results.

References: [GitHub required checks](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches),
[Go race detector](https://go.dev/doc/articles/race_detector),
[Docker CI](https://docs.docker.com/build/ci/github-actions/).
