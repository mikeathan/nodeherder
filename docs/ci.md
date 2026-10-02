# CI and pre-merge checks

[Workflow](../.github/workflows/ci.yml) · [Development policy](sdd/README.md)

CI runs on Ubuntu for pull requests to `main`, pushes to `main`, and merge groups.
Independent jobs provide feedback in parallel; container matrix failures do not
cancel the other component. Superseded PR runs are cancelled; main runs are not.

| Check | Errors caught |
| --- | --- |
| Workflow and Shell Checks | Invalid Actions syntax/expressions/inline shell; Bash syntax in tracked scripts; stack-manager ShellCheck findings |
| Stack Manager (Ubuntu) | CLI scope/failure/dry-run/version regressions; real Compose merge, required variables, paths/ports/network using dummy fixtures |
| Backend Tests and Build | Go module checksum changes, Linux compilation failures, uncached unit regressions, `go vet` findings |
| Backend Race Detection | Concurrent accesses exercised by backend tests; not proof that all races are absent |
| Frontend Tests and Build | Locked install failures, existing ESLint errors, Jest regressions, Vue/TypeScript compilation and production bundle failures |
| Container (backend/frontend) | Real Dockerfile/context/build failures on Linux/amd64; backend executable help; nginx config and built index smoke checks |

Images are built/loaded on disposable runners, never published. Smoke containers
use `--network none`; backend help exits before application initialization. nginx's
test-only host mapping resolves its proxy name without accessing a backend.
No production credentials, broker, Zigbee coordinator, or household devices are used.
Go/Node use repository toolchain files; Dockerfiles retain their builder versions.
Dependency and component-specific image caches reduce repeat cost. Jobs have explicit
timeouts; verification gets read-only repository permission. Only the existing
main-push tagging job can write, after all checks succeed.

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
