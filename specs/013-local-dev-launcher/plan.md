# Plan: Local development launcher

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-16
Constitutions: parent/backend/frontend v1.0.0 (proposed).

## Context and design

`scripts/dev.mjs` uses Node 24's built-in child_process APIs, with no dependencies.
Build `backend/nodeherder` in backend/ (existing native setup convention), then run
that executable and `npm run dev -- --strictPort` in their component directories.
The launcher owns POSIX process groups, including npm/Vite descendants. Shutdown
sends SIGTERM, then SIGKILL after five seconds, including surviving descendants.
Environment is inherited with APP_ENV defaulting to development. No protocol,
authentication, configuration-file, persistence, or application source changes.
Alternative shell job control is less explicit about bounded descendant cleanup.

## Constitution Check

Before implementation and design review: NH-01 PASS (spec and tasks); NH-02 PASS
(existing native startup only); NH-03 PASS (existing configuration/exposure unchanged);
NH-04 PASS (standalone tooling, no dependencies); NH-05 PASS (isolated process tests
planned); NH-06 PASS (owned groups and bounded cleanup designed).
BE-03 applies to launcher lifecycle; other BE/FE implementation rules are N/A because
application code/contracts/UI are unchanged. Final review: all applicable NH checks and BE-03 PASS with the process tests below.

## Traceability and verification

FR-1–3: `scripts/dev.test.mjs` companion tests using fake executables only.
Passed `node --test scripts/dev.test.mjs`: 5/5 tests (6.4s), including
SIGINT/SIGTERM cleanup of descendants, startup/build failure, frontend failure,
and five-second forced shutdown. Node syntax checks and git diff --check passed.
The first test run exposed macOS temporary-directory symlink normalization in the
fixture; canonicalized its path and reran successfully. No production assertion
was weakened. Final review confirms no application or configuration changes.
Application suites/builds are unrun: no application changes; real service startup
is intentionally excluded from tests to avoid household access.

## Debt, risks, exceptions

POSIX-only (macOS/Linux); no Windows process-tree support. Backend rebuild is on
launch only. Existing native data/configuration and port requirements still apply.
No exceptions or migrations. Rollback removes launcher/docs only.
