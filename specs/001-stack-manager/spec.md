# Spec: Stack manager

ID: 001 | Status: Implemented; deployment verification pending | Owner: repository contributor | Updated: 2026-10-01
Constitution: 1.0.0 (proposed) | Components: backend/frontend deployment

## Problem and scope

`scripts/manage-stack.sh` manages production backend/frontend Compose files, but
uses caller-relative paths and accepts invalid stacks. Rebuild tears down
services before version lookup/pull/build can fail. Separate `down` calls contend
over the shared project network. Docs barely expose the script and contain stale
build/port/toolchain guidance.

Improve the existing CLI and supporting deployment files/docs. Exclude live deployment,
credential generation, application behavior, auth hardening, and systemd installation.

## Scenarios and requirements

- FR-01 / AC-01: Existing start/stop/down/restart/pull/rebuild commands accept
  all/backend/frontend (default all), from any working directory; invalid input fails
  before Docker operations.
- FR-02 / AC-02: Rebuild pulls required external images and builds application images
  before `up`; a failed pull/build never stops/removes existing containers.
- FR-03 / AC-03: Single-stack operations affect only that stack; all-stack operations
  use one combined project. Partial down retains shared networks; no action deletes
  persistent data, volumes, images, or unrelated/orphan containers.
- FR-04 / AC-04: Add build/status/logs/validate, log tail/follow, optional no-cache,
  and dry-run. Dry-run prints commands without invoking Docker or requiring a daemon.
- FR-05 / AC-05: Help works without dependencies. Output is readable, with color only
  in terminals unless disabled. Invalid options/dependencies/config/daemon failures
  return nonzero and useful diagnostics. No rendered configuration/secrets are printed.
- FR-06 / AC-06: Frontend version reaches its build; explicit APP_VERSION works and
  missing host Node falls back to package version. Build contexts exclude dependencies
  and backend secrets/data. Docs describe actual files, commands, ports, prerequisites,
  and the separate native systemd examples. Required host mount/device variables reject
  missing/empty values rather than resolving unintended host paths.
- FR-07 / AC-07: Deployment targets Ubuntu/Linux with Bash and standard utilities;
  no macOS-only commands or host Node requirement. CI runs isolated CLI tests on Ubuntu.
- SC-01: Isolated CLI regressions demonstrate all acceptance cases with mocked Docker;
  syntax, links, and version-helper checks pass. Real Compose checks use disposable
  environment files and an unreachable daemon; verify all scopes, required variables,
  merged paths/network/ports, and frontend independence. Report builds separately.

## Interfaces and assumptions

CLI only; preserve Compose service/container names, ports, host bind mounts, and
existing nginx host-gateway proxy. `backend` scope includes mqtt/zigbee2mqtt.
`restart` restarts existing containers; `start` reconciles Compose configuration.
No healthchecks currently exist; successful `up` does not certify application readiness.
No automatic rollback is claimed after container replacement starts.
