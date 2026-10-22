# Spec: Automatic proxy trust for local sign-in

ID: 026 | Status: Implemented | Owner: NodeHerder contributors | Updated: 2026-10-22
Constitution: proposed 1.0.0 | Components: backend deployment tooling

## Problem, scope, exclusions

`OFFLINE_STRICT_LOCAL=true` only offers local sign-in when the backend recognizes the
request as local. Behind the Compose Nginx proxy that requires `TRUSTED_PROXY_CIDRS`;
containers reach the backend through the host, so the observed peer is the Compose
network gateway rather than the LAN client. The allowlist is empty by default, so a
clean install sends users to Google OAuth, which fails with `redirect_uri_mismatch` on
a LAN-only deployment. Recovery required manually inspecting Docker networks and
editing `backend/.env`.

Fix the Compose stack lifecycle so local sign-in works on a clean install without
host-specific manual configuration. Do not change authentication rules, trust any
private subnet, or make trust implicit in the backend.

## Scenarios and acceptance

### US-01 — Clean install signs in locally (P1)

- AC-01: Given a clean checkout with no `TRUSTED_PROXY_CIDRS` anywhere, when an
  operator runs `scripts/manage-stack.sh start` (or `rebuild`), then the backend
  container is created with the Compose network gateway as its trusted proxy peer and
  `GET /api/auth/mode` returns `local` for a LAN client.
- AC-02: The computed value is exactly the dedicated proxy peers for the current
  stack — the network gateway and, when present, the frontend container address — not
  a whole Docker subnet, all private ranges, or `0.0.0.0/0`.

### US-02 — Explicit configuration still wins (P1)

- AC-03: Given `TRUSTED_PROXY_CIDRS` in the shell or `backend/.env*`, the script does
  not override it, and reports that the configured value is in use.
- AC-04: When the proxy peers cannot be determined, the stack still starts, with a
  warning naming the manual setting, instead of failing or trusting a guess.
- AC-05: `--dry-run` and `validate` still need no Docker daemon and change nothing.

## Requirements and success

- FR-01: `scripts/manage-stack.sh` computes the proxy allowlist for `start`/`rebuild`
  and injects it into the backend service only when none is configured; acceptance:
  AC-01, AC-02, AC-03.
- FR-02: A missing or unreadable network degrades to a warning, never a wider trust
  set; acceptance: AC-02, AC-04.
- NFR-01: No new runtime dependency, daemon, or stored state; the generated file is
  ignored by version control and rewritten per run.
- SC-01: Deployment script tests cover computed, configured, and unavailable cases,
  plus the existing dry-run/validate guarantees.

## Entities and external interfaces

Trust membership remains `TRUSTED_PROXY_CIDRS` consumed by `utils.IsLocalRequest`;
`/api/auth/mode`, `/api/auth/offline/start`, and session cookies are unchanged. The
generated Compose override exists only on the deployment host inside the repository
working copy and never ships in an image.

## Questions and assumptions

No blocking questions. Assumption: the supported path is the repository Compose stack
started through `scripts/manage-stack.sh`; hand-written `docker compose up` remains a
manual configuration as documented. Local sign-in stays restricted to private,
loopback, or link-local client addresses reported by the trusted proxy.
