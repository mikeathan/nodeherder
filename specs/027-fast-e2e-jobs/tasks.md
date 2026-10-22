# Tasks: Fast, deterministic browser E2E jobs

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Measure where browser-job time goes on main and record it in `spec.md`;
  NH-01, AC-01; done: two jobs spent ~11m in `--with-deps`, tests took ~2m.
- [x] T002 — Run browser jobs in the version-matched Playwright image and remove the
  per-job browser install; FR-01, AC-01/02; depends T001; done: local containerized
  desktop shard 1/2 passed 52 tests (9 skipped) in 1.4 minutes.
- [x] T003 — Add the image/version guard and tighten the job timeout; FR-02, AC-03;
  depends T002; done: guard logic verified, image v1.56.1 matches package 1.56.1.
- [x] T004 — Rename jobs and steps to describe what runs, keeping `Run Tests`;
  AC-04; depends T002; done: workflow and `docs/ci.md` agree.
- [x] T005 — Validate the workflow and record evidence; NH-05, SC-01; depends T002–T004;
  done: actionlint and YAML parse passed; GitHub timing evidence comes from this PR.

Completion evidence: [plan verification](plan.md#traceability-and-verification).
