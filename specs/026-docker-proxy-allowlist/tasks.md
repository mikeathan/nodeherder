# Tasks: Automatic proxy trust for local sign-in

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Record the deployment failure and outcomes in `spec.md`; NH-01,
  NH-03; done: reproduction and acceptance IDs above.
- [x] T002 — Add script regressions for computed peers, configured values,
  duplicates, unresolved peers and frontend-only scope in
  `scripts/manage-stack.test.cjs`; FR-01/02, AC-01–05; depends T001; done:
  failures observed before the script change.
- [x] T003 — Resolve and inject proxy peers in `scripts/manage-stack.sh` for
  `start`/`rebuild`; FR-01/02, AC-01–04; depends T002; done: 61 script tests pass.
- [x] T004 — Ignore the generated override and document the automatic path in
  `docs/setup.md` and `docs/runtime-reliability.md`; NFR-01, NH-03/06; depends
  T003; done: docs describe precedence and the unresolved warning.
- [x] T005 — Verify against real Compose merge semantics and a disposable probe
  project; NH-05; depends T003; done: env resolves in merged config, gateway and
  container id discovered; household stack never started.
- [x] T006 — Mention the automatic trust step in `--help` and bump the frontend
  package to 0.14.0 so the next tag starts a new minor line; NH-01; depends T004;
  done: help test and lockfile in sync.

Completion evidence: [plan verification](plan.md#traceability-and-verification).
