# Tasks: Scoped CI repair

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)

- [x] Remove expanded runtime/dependency/legacy-fixture changes at user request.
- [x] Retain mapper mutex and failed-initialization DB close; FR-01/02.
- [x] Verify mapper concurrency, failed-init reopen, worker regressions, vet/build/ordinary suite.
- [x] Record actual results and deferred issues for user review; do not suppress CI checks.
