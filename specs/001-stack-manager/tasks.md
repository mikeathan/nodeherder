# Tasks: Stack manager

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)

- [x] T001 — Inspect script/Compose/Dockerfiles/version/nginx/setup/systemd; FR-01–06.
- [x] T002 — Isolated CLI regressions in `scripts/manage-stack.test.cjs`; FR-01–06.
- [x] T003 — Implement robust CLI in `scripts/manage-stack.sh`; depends T002; FR-01–06.
- [x] T004 — Align Docker build contexts/version helper/Makefile/docs/Ubuntu CI; FR-06/07.
- [x] T005 — Run regressions, syntax/link checks, final diff/rule review; NH-05/06.

All local checks passed, including real isolated Compose configuration. Ubuntu CI
and container/device startup remain unrun; see the plan's host procedure/limitations.

- [x] T006 — Real isolated Compose configuration checks and expanded regressions;
  FR-01–07/SC-01; run locally and configure Ubuntu CI; no live deployment.

- [x] T007 — Fix ShellCheck SC1007 reported by Ubuntu CI: explicit empty color
  assignments; FR-05/SC-01. Bash syntax and 56 mocked regressions pass locally;
  ShellCheck rerun remains in CI (not installed locally).
