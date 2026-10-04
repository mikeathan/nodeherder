# Tasks: Pre-merge verification

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)

- [x] T001 — Inspect CI/tooling/safe tests and baseline gaps; FR-01–05.
- [x] T002 — Independent strict jobs, image smoke checks, aggregate gate;
  `.github/workflows/ci.yml`; FR-01–05; depends T001.
- [x] T003 — Document checks/failures/required-status rollout/limits;
  `docs/ci.md`, README/SDD links; FR-04/SC-01; depends T002.
- [x] T004 — Run actionlint/regressions/local checks; record actual outcomes and
  unrun GitHub/container verification; SC-01/NH-05; depends T002/T003.

Separate follow-up: remediate backend vet and race/concurrency failures before
expecting green strict CI.

- [x] T005 — Fix the reported worker-test context leak using explicit cancellation;
  replace test-local unsynchronized flags/counters and timing assumptions with
  completion signals/atomic counts; verify vet and repeated worker race tests.
  Diagnose application/storage failures separately before runtime changes;
  preserve strict CI gates.
  Evidence: vet/build/full ordinary backend suite pass; full `utils` race suite and
  20 repeated runs of both worker tests pass locally. Full backend race detection
  remains failing on separate automation/mapper/Bolt issues recorded in the plan.
