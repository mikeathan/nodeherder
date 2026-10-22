# Plan: Automatic proxy trust for local sign-in

Spec: [spec.md](spec.md) | Status: Implemented | Author: NodeHerder contributors
Updated: 2026-10-22 | Constitutions: parent/backend proposed 1.0.0

## Context and design

Current behavior: `OFFLINE_STRICT_LOCAL=true` only selects local sign-in when
`utils.IsLocalRequest` can attribute the request to a private/loopback client through a
trusted proxy. `TRUSTED_PROXY_CIDRS` is empty by default, so behind the Compose Nginx
every request is treated as online and users are sent to Google, which fails with
`redirect_uri_mismatch` on a LAN-only deployment. Observed deployment evidence: with
`TRUSTED_PROXY_CIDRS=172.18.0.1/32` the same stack returns `mode=local`.

Ownership: deployment tooling owns peer discovery; the backend keeps enforcing the
allowlist. `scripts/manage-stack.sh` already owns Compose assembly, so it gains one
preparation step for `start`/`rebuild`:

1. If `TRUSTED_PROXY_CIDRS` is set in the shell or `backend/.env*`, use it unchanged.
2. Otherwise create the Compose network (`compose create`, no start) and read the
   network gateway from `IPAM.Config[0].Gateway`.
3. When the frontend service is in scope, read its container address via
   `compose ps -aq frontend` + `docker inspect`.
4. Write only those resolved addresses, as `/32` or `/128` single-host prefixes, into
   `.generated/compose.proxy-allowlist.yml` and append it as a Compose override file.
5. Any resolution failure prints a warning naming the manual setting and continues.

Alternatives rejected: trusting all Docker/private ranges (widens trust beyond the
dedicated proxy peers); computing inside the backend (the container cannot know which
host path the request used, and adding host discovery to the app expands its trust
surface); hard-coding Compose IPs in the checked-in env (host-specific and stale after
recreation); mTLS or authenticated proxying (larger design change, ADR-scale).

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec records reproduction, outcomes, boundaries | PASS | PASS |
| NH-02 | No device semantics or command paths touched | N/A | N/A |
| NH-03 | Trust stays an explicit, narrow allowlist; no wildcard ranges; explicit config wins | PASS | PASS |
| NH-04 | Existing script/config boundary; no new service or dependency | PASS | PASS |
| NH-05 | Script tests cover computed, configured, duplicate, unresolved and frontend-only paths | PASS | PASS |
| NH-06 | Generated file is per-run, gitignored, bounded; no persistence or migration | PASS | PASS |
| BE-03 | No goroutine, timer, or shared in-process state added | N/A | N/A |
| BE-05 | Auth routes, modes and payloads unchanged | PASS | PASS |
| FE-04 | No UI change; sign-in mode already shown from `/api/auth/mode` | N/A | N/A |

## Contracts, data, rollout

No HTTP/WS/MQTT/persistence change. Deployment change: `start`/`rebuild` now run
`compose create` before `up -d` and pass an extra Compose file. An explicit
`TRUSTED_PROXY_CIDRS` keeps today's behavior, so rollback is reverting the script; a
stale `.generated/` file is ignored by version control and unused without the flag.
Deployment order: recreate the backend through the stack manager; no image rebuild is
required for the env to apply.

## Traceability and verification

| IDs | Paths | Procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02 | scripts/manage-stack.sh | Computed-peers and duplicate-peer tests | PASS: gateway/32 only, gateway+frontend, no duplicates |
| FR-01, AC-03 | scripts/manage-stack.sh | Configured shell/env-file tests | PASS: no `create`, no override file, message names the configured value |
| FR-02, AC-04 | scripts/manage-stack.sh | Unresolved gateway and `<no value>` tests | PASS: warning, stack still starts, no trust written |
| AC-05 | scripts/manage-stack.sh | Existing dry-run/validate suites | PASS: no Docker calls, no mutations |
| NH-05 | scripts/manage-stack.test.cjs, Compose config | 62 script tests; real Compose merge inspection | PASS |

Execution (2026-10-22, Node 24.21.0, Docker Compose v5.1.0):

- `node --test scripts/manage-stack.test.cjs`: 62 passed (adds IPv6 and malformed-inspector coverage).
- Real Compose merge: `docker compose -f docker-compose.backend.yml -f
  .generated/compose.proxy-allowlist.yml config` resolved
  `TRUSTED_PROXY_CIDRS=172.18.0.1/32,172.18.0.2/32` with `APP_ENV=production`.
- Disposable probe project: `compose create` then
  `docker network inspect ... IPAM.Config[0].Gateway` returned `172.18.0.1`;
  `compose ps -aq` returned the created container id. `compose create` also builds
  a missing image, so a clean `start` is not blocked and keeps its existing
  version-arg behavior.
- `node --test scripts/sync-reporting-defaults.test.mjs scripts/dev.test.mjs` and the
  12 opt-in Compose config tests passed; actionlint, `bash -n`, shellcheck and
  `git diff --check` passed.
- Not run: the household stack itself (no containers, devices, or broker touched), and
  GitHub CI until the PR runs.

The frontend package version moves to 0.14.0 so the next release tag starts a new minor
line instead of a patch off the deployed 0.13.0.

## Debt, risks, exceptions

Trusting the network gateway means any client that reaches the published backend port
and supplies a private `X-Real-IP`/`X-Forwarded-For` can obtain a local session; that is
the same exposure the manual value already had, so the backend port must stay LAN-scoped.
A container-to-container proxy is covered by including the frontend address. Addresses
change when Docker recreates networks, which the per-run recomputation handles. Manual
values in `backend/.env` can still go stale; docs say so. No exceptions requested.

## Completion

Implemented on a follow-up branch with spec 026. Deployment docs updated for the
automatic path. No production auth, device, or data behavior changed.
