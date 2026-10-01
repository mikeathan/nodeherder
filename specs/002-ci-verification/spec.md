# Spec: Pre-merge verification

ID: 002 | Status: Implemented; baseline failures block green CI | Updated: 2026-10-01
Parent: NH 1.0.0 (proposed) | Related: [stack manager](../001-stack-manager/spec.md)

## Scope and outcomes

CI lacks production compilation, lint, race detection, and image verification;
application tests share a sequential job. Add independent Ubuntu checks, with no
publishing/deployment, household devices, production credentials, branch-setting
changes, or runtime race fixes.

- FR-01 / AC-01: Verify Go tests/vet/Linux compilation, frontend lint/Jest/production
  type-check/build, and backend race detection. Failures stay visible, never green.
- FR-02 / AC-02: Check workflow syntax/expressions, Bash syntax, stack-manager
  ShellCheck, and existing mocked/real isolated Compose tests.
- FR-03 / AC-03: Build both production Dockerfiles with real contexts on Linux/amd64,
  without pushing. Offline smoke checks verify backend help, nginx configuration,
  and built frontend index; no Zigbee/broker startup.
- FR-04 / AC-04: Preserve `Run Tests` as aggregate: failed/skipped/cancelled checks
  prevent success/tagging. Run on main PRs/pushes and merge groups; document required
  status settings rather than changing them.
- FR-05 / AC-05: Bounded jobs, caches, read-only verification permissions, and
  cancellation of superseded PR runs. Only existing main-push tagging gets write access.
- SC-01: Validate workflow with actionlint; distinguish passed/failed/unrun checks.

## Baseline gaps

Local frontend lint passes with seven warnings; nine Jest suites/44 tests and its
production build pass. `go vet` flags discarded context cancellation in
`backend/utils/worker_test.go`. Race tests fail with existing backend races/concurrency
crashes on Go 1.26.1/macOS. Strict checks may block merging until separate fixes;
do not suppress failures. No arbitrary coverage threshold or blanket formatting.
