# Tasks: Reliable proxy peer discovery for local sign-in

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: Complete

- [x] T001 — Reproduce the miss with a directory Compose normalizes differently; NH-01;
  done: `compose create` produced `<normalized>_nodeherder` while the script looked for
  the raw directory name and warned.
- [x] T002 — Read the project name from `docker compose config` and list networks by
  the Compose project label; FR-01, AC-01/02; depends T001; done: 64 script tests pass.
- [x] T003 — Keep the fail-soft warning, dedupe single-host prefixes, and report the
  generated file path; FR-02/03, AC-03/04; depends T002; done: warning and message
  tests pass, docs show the `docker inspect` command.
- [x] T004 — Verify against a real container; NH-05; depends T002; done: container
  `Config.Env` showed `TRUSTED_PROXY_CIDRS=172.19.0.1/32` from a mixed-case checkout.

Completion evidence: [plan verification](plan.md#traceability-and-verification).
