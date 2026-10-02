# Plan: Stack manager

Spec: [spec.md](spec.md) | Status: Implemented; deployment verification pending | Updated: 2026-10-01
Constitutions: NH/BE/FE 1.0.0 (proposed)

## Design

Resolve root from the script, change to it, and use absolute Compose paths. Select
one/both files and explicit service arrays. Preserve derived project identity.
Preflight Docker/Compose, quiet configuration, and daemon access; validate needs no
daemon. Dry-run only prints shell-escaped commands. No environment files are sourced.
Build/rebuild embeds APP_VERSION; use existing Node version helper when available,
package fallback otherwise. Rebuild retains containers until pull/build succeed.
Partial down uses selected stop + rm, preserving the shared network; full down uses
combined files. Never use volume/image/orphan deletion flags.

Keep nginx/Compose networking unchanged. Fix Dockerfile path/port descriptions and
add context-specific ignores. Align setup, component READMEs, and Makefile startup.
Require nonempty Compose mount/device variables to prevent accidental root-level mounts.
Native systemd examples remain separate and host-specific.
Use Ubuntu-compatible Bash/standard utilities; add an independent Ubuntu CI job for
syntax/version/CLI tests, required alongside application tests before automatic tagging.
Follow-up verification: add real Compose configuration tests in disposable roots with
dummy environment files and an unreachable daemon. Run them locally via the installed
standalone Compose binary and in Ubuntu CI through `docker compose`; never use real
credentials or start containers. Expand mocked action/failure coverage as needed.

## Constitution Check

| Rule | Evidence before implementation | Final |
| --- | --- | --- |
| NH-01 | Spec ACs and scoped exclusions recorded | Spec/plan/tasks trace implemented changes |
| NH-02 | No direct device commands; existing service groups retained | Service selection tests pass; no live devices used |
| NH-03 | No new exposure; quiet config; backend context excludes secrets | Quiet-config tests pass; Docker context/ports reviewed |
| NH-04 | Existing CLI retained; new flags documented; project/network preserved | Selection/path tests pass; CLI and native/Compose distinction documented |
| NH-05 | Mocked Docker regression plan; no live devices | 56 mocked tests and 11 real Compose cases pass locally; Ubuntu CI configured but not run |
| NH-06 | Build before up; no volume deletion; failures stop subsequent steps | Failure/scoped-removal tests pass; real Compose rejects missing/empty host variables |
| BE-01/02/05 | N/A: domain, pipeline, routes unchanged | N/A |
| BE-03/04 | Lifecycle/data mounts preserved; no new Go concurrency/schema | Mounts/ports retained; no new application lifecycle/schema code |
| FE-01/02/03/04/05 | N/A: browser behavior/contracts unchanged; backend secrets excluded from build | N/A |

## Traceability and verification

| Requirements | Evidence | Result |
| --- | --- | --- |
| FR-01–05 | `node --test scripts/manage-stack.test.cjs` with isolated fixtures/mock Docker | 56 tests pass locally on macOS; mocked Docker only |
| FR-06 | Version helper from another cwd, fallback and APP_VERSION tests; Dockerfile/context review | Pass; real helper outputs version only from `/private/tmp`; real Compose confirms required-variable rejection and merged paths/ports/network |
| FR-07 | Bash/standard utilities review; `.github/workflows/ci.yml` independent Ubuntu job | No macOS-only runtime dependencies; Ubuntu execution pending CI |
| All | Bash/Node syntax; docs links; YAML parsing; `git diff --check` | Pass; 29 local links/anchors resolve; Compose and workflow YAML parse |
| Compose config | `NODEHERDER_COMPOSE_TESTS=1 NODEHERDER_COMPOSE_BINARY=/opt/homebrew/lib/docker/cli-plugins/docker-compose node --test scripts/manage-stack.compose.test.cjs` | 11 cases pass (12 reported tests including parent) using standalone Compose 5.1.0; isolated fixtures/unreachable daemon |
| Ubuntu/container builds | Independent Ubuntu CI job and deployment-host procedure below | Unrun: current host is macOS; no live daemon/builds used |

No application source changes: Go/Jest suites do not verify this shell/deployment change.
Use mocked CLI tests, real isolated Compose configuration checks, and static checks;
actual container builds and Linux Zigbee
startup remain unrun. No rollback, healthchecks, credential setup, or schema migrations
are added. Pulling unpinned external images may still change dependency versions;
application replacement may briefly interrupt service. No exceptions requested.

Deployment-host verification (not executed): after private host/env configuration,
run `./scripts/manage-stack.sh validate all`, preview with `rebuild all --dry-run`,
then in an authorized maintenance window run `rebuild all`, `status`, and scoped logs.
Confirm broker/Zigbee connection, backend on 4110, frontend on 80, and retained data.
Review logs privately. Real builds/startup need a working deployment-host daemon;
configuration tests validate interpolation/model only, not container behavior.
