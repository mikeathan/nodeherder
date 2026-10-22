# ADR-024: Explicit trust and stream ownership

Status: proposed | Date: 2026-10-21
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner: Codex; maintainer review pending
Rules: NH-03/04/06, BE-03/05, FE-05

## Context and options

Current sign-in trusts forwarded headers from every peer; MCP responses fan out across
clients; authentication is checked only at stream upgrade. Global broadcast and implicit
proxy trust preserve apparent connectivity at the cost of crossing trust boundaries.
Replacing the entire HTTP/WS/MCP stack would expand risk and operating cost unnecessarily.

## Decision and consequences

Retain existing protocols/services and one auth policy, with explicit proxy trust and
per-session stream lifetime/reply routing. Legacy SSE uses its advertised session URL;
direct JSON POST remains supported. Shared notifications carry only existing resource
invalidation messages. No new dependency or device command semantics.

The user deferred public MCP/context authentication migration on 2026-10-21; existing
access remains unchanged. The user confirmed the repository Docker setup by directing
inspection of its Dockerfile: Nginx proxies via host.docker.internal:4110. Retain this
routing and accept replacement addresses only from an explicit TRUSTED_PROXY_CIDRS
allowlist. The empty default permits direct local requests without forwarding headers;
forwarded requests require a configured peer and valid client IP. Never fall back to a
trusted proxy's private address as the client address. Invalid trust configuration
fails closed. X-Real-IP must be a single valid address; otherwise use X-Forwarded-For
from right to left through configured proxies, stopping at the first untrusted hop.
Nginx overwrites both headers with its observed peer. No DNS lookups, workers, fixed
Docker subnets or blanket private-network trust are introduced.

Alternative: ignore headers and authorize the immediate private peer. Rejected because
public visitors behind Nginx would then appear local. Trusting every Docker/private
address is also rejected because direct callers could forge headers. A shared proxy
secret or dedicated fixed-address network would add deployment machinery unnecessarily.

Before rollout, identify the immediate Nginx peer and set the allowlist in backend/.env.
Docker restarts can change addresses; verify and update the allowlist on recreation.
Missing configuration routes proxy users to Google rather than silently issuing a
local session. This coordinated configuration step is required to preserve local login.
Test all parts with isolated peers and no real devices. Document required
client configuration and rollback before enabling a breaking authentication boundary.
An ADR does not grant a constitutional exception or claim maintainer acceptance.
