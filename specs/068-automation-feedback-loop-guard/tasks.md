# Tasks: Automation feedback-loop guard

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete

Each task maps to requirement/acceptance IDs or a constitutional gate.

## Preparation

- [x] T001 — Resolved owner options in `spec.md` Q-01/Q-02 (A now, B deferred);
  IDs: FR-01, FR-02.
- [x] T002 — Checked applicable rules in `plan.md`; depends: T001; done: constitution table.

## Coverage and implementation

- [x] T003 — Added `TestIlluminanceOnlyChangeDoesNotStartActions` in
  `backend/internal/automations/device_test.go`; IDs: AC-02, AC-06; done: failed pre-fix with
  `unexpected command "...state ON"`, passes post-fix; now also asserts that the
  dependency-only report was evaluated.
- [x] T004 — Implemented primary-vs-dependency execution in
  `backend/internal/automations/device.go` and `trigger.go`; IDs: FR-01, FR-02, AC-01,
  AC-03, AC-06; done: acceptance tests pass. `AffectedBy` reads the primary name through
  `GetName()`, matching `Evaluate`; recorded in spec/plan per the AGENTS.md guard rule.
- [x] T005 — Updated behavior-encoding tests: presence ordering, late-dark return,
  dependency/off-delay, condition evaluation ownership, action failure reporting,
  manual/device independence; IDs: AC-01..AC-07; done: suite passes. Review corrections
  added dependency-only darkening steps to the late-dark-return and off-delay tests so
  they fail pre-fix rather than merely restating new expectations.
- [x] T006 — Added invariant comments at the gating code, the authoritative architecture
  note and short agent-guidance pointers; IDs: NFR-01; done: reviewed.
- [x] T007 — Option B (manual-override latch) recorded as deferred in `spec.md` Q-02;
  forward plan removed at owner request; IDs: spec Q-02.

## Handoff

- [x] T008 — Ran the CI-equivalent matrix locally (details in `plan.md`).
  Backend: `go mod verify`, `go build ./...`, Linux/amd64 cross-build, `go vet ./...`,
  `go test -count=1 ./...` (53 packages ok), `go test -race -count=1 ./...` (exit 0,
  53 ok), `go test -run '^$' ./internal`. Frontend: `npm run lint` (0 errors),
  `npm test -- --ci --runInBand` (162 suites / 1097 tests), `npm run type-check:e2e`,
  `npm run build -- --mode production`. Browser: full Playwright desktop+mobile
  (421 passed, 9 skipped). Workflow/shell: actionlint, `bash -n`, shellcheck.
  Gate: NH-05/SC-01.
- [x] T009 — Independent `reviewer` pass completed. Verdict: no CI blockers; two
  majors and several minors, all addressed: documented the `Seed` full-snapshot
  limitation (spec + architecture), added superseded pointers to spec 010 plan/tasks,
  removed the unreachable manual clause, used `GetName()` consistently, restored the
  failure-no-latch assertion, made the discarded condition result explicit, corrected
  the spec evidence path, updated the architecture guide, renamed the selection test,
  and added a custom-trigger boundary test. Gate: NH-01..06, BE-01..05.
- [x] T010 — 2026-11-25 review corrections, recorded in this branch's same spec/plan/tasks
  set: corrected the spec's dial-recipe evidence to cite the committed recipe and its
  scenario test; removed the stale "manual events always execute" plan clause and the
  unresolved scenario hedge; completed the affected-paths and contracts sections for the
  CI guard; replaced `PASS (planned)` with design evidence; corrected FR-02's evaluation
  claim; consolidated the invariant into one authoritative home with short pointers;
  recorded the `AffectedBy` accessor change; added GoDoc to `Device.Evaluate` and
  `Device.EvaluateTrigger`; removed the unreachable dependency-only error path; made the
  custom-trigger test prove selection; strengthened
  `TestIlluminanceOnlyChangeDoesNotStartActions` to prove evaluation; made the CI guard
  derive its expected version from the image tag; corrected the false AC-03 claim in
  specs/027; refreshed dates and verification evidence. Gate: NH-01..06, BE-01..05,
  `.agents/rules/go-engineer.md` GoDoc rule.
- [x] T011 — Second independent review pass (2026-11-25; agent, not human approval):
  reviewed the final `device.go`/`trigger.go` diff and all execution/selection call sites;
  re-ran build, vet, full suite, automations/services race checks, the pre-fix negative
  controls and the CI-guard mismatch cases. No code finding required a change. Gate:
  NH-01..06, BE-01..05; depends T010.
