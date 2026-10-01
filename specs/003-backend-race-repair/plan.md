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

Passed after the targeted revert: `go vet ./...`, `go build ./...`,
`go test -count=1 -timeout=5m ./...` (includes failed-init reopen regression),
`go test -race -count=20 -timeout=2m ./repository -run TestMapperConcurrentConfigureUpdateResolve`,
`go test -race -count=20 -timeout=2m ./utils`, and `git diff --check`.
Local environment: macOS/arm64 Go 1.26.1; Ubuntu CI and repository-toolchain parity
are not established by local checks. Full race suite was not rerun after the revert;
known unresolved failures above remain. Acceptance evidence covers retained scope only.
