# Spec: Runtime security and resource lifecycle

ID: 024 | Status: Review fixes verified locally; deployment validation pending | Owner: Codex | Updated: 2026-10-22
Constitution: proposed parent/backend/frontend v1.0.0 | Components: both

## Problem, scope, exclusions

The 2026-10-21 audit reproduced a storage deadlock, excessive WebSocket allocation,
forwarded-IP authentication bypass, cross-client MCP responses, stale stream auth,
and frontend timer/conversation/session isolation defects. The user authorized fixes
on a new branch from updated main while preserving working device and integration
behavior. Baseline commit: 682ebf07. Tests use isolated storage/fake MQTT, never devices.

Scope: audited runtime defects and the affected Node serving dependency. Preserve
HTTP/WS payloads, MQTT command/event semantics, metric units/UTC/inclusive ranges,
persistence formats, existing Vuex ownership and acknowledged-state presentation.
Exclude unrelated rewrites, broad dependency upgrades, ratification, deployment,
and historical test-file reorganization (record existing ownership debt).

## Scenarios and acceptance

### US-01 — Reliable storage and bounded work (P1)

- AC-01: Concurrent pruning, writing and reading finish without deadlock; errors
  roll back transactions; pruning neither skips eligible keys nor changes retained data.
- AC-02: Get returns independently owned bytes, preserves missing versus empty data,
  and existing legacy databases read/write/reopen without migration.
- AC-03: A socket uses at most 256 queued outbound messages and 4 KiB read/write
  buffers, retains the 1 MiB inbound message limit, and removes disconnected resources.
  Slow clients are disconnected on queue overflow rather than silently losing state.
  Reconnect refreshes the shared HTTP snapshot without overwriting newer live reports.
  Logout immediately invalidates pending snapshots; failed refreshes can be retried
  even when the registry was initialized. Live configuration edits preserve unrelated
  fields received in the recovery snapshot. Disconnect emits no non-contract text frames.
  Recovery retains at most 256 configuration mutations; overflow fails observably
  without replacing the existing registry or retrying device commands.
- AC-04: Oversized HTTP bodies return 413 before domain mutation; slow headers and
  idle connections are bounded. Normal requests and long-lived SSE remain functional.
- AC-05: Assistant/OAuth calls honor caller cancellation and finite deadlines;
  oversized upstream responses fail observably; existing successful payloads persist.
- AC-06: Legacy metric queries preserve results/order/limits and error compatibility,
  perform one successful scan per device for all selected exposes (per-expose error
  fallback preserves legacy adapters), canonicalize repeated expose names, and abort canceled/budget-exceeded
  work. Missing values stay missing. No change to automation evaluation or reporting.

### US-02 — Explicit trust and session isolation (P1)

- AC-07: Untrusted forwarding headers cannot select local sign-in. Direct local
  non-proxy requests without forwarding headers and explicitly trusted proxy requests retain
  the applicable sign-in mode. TRUSTED_PROXY_CIDRS is an explicit comma-separated
  IP/CIDR allowlist, empty by default. Forwarded requests from an unlisted peer,
  invalid configuration, malformed client addresses and trusted proxies lacking a
  client address cannot select local sign-in. The shipped Nginx overwrites client-IP
  headers; X-Real-IP is authoritative, with right-to-left X-Forwarded-For fallback.
- AC-08: Expired/revoked authenticated streams cannot dispatch or receive household
  data; disconnect cancels auth timers/subscriptions. Allowed browser origins work,
  unlisted origins are rejected; clients without Origin retain bearer-token access.
- AC-09: MCP POST replies reach only the matching session/HTTP caller; resource-change
  notifications may be shared. Legacy SSE clients use the advertised unique endpoint.
- AC-10: MCP/context authentication rollout is explicit and matches the user's client
  capabilities; no silent breaking migration. The user deferred this migration on
  2026-10-21: retain existing public access and revisit separately.
- AC-11: Cookies are Secure on direct/trusted-proxy HTTPS; intentional local HTTP
  remains usable. State, PKCE, session and expiry cookies share one policy, and the
  generated local sign-in URL uses that same explicit HTTPS setting without trusting
  client-supplied scheme headers.

### US-03 — Browser lifecycle and dependency hygiene (P2)

- AC-12: Permit-join countdown follows confirmed bridge state across panel navigation,
  unmount/remount, logout and subsequent sessions; timers stop with their owner.
- AC-13: Replies/history are applied only to their initiating conversation/session;
  cancellation, errors, switching chats and unmount cannot repopulate another chat.
  Deleting a conversation invalidates an in-flight selection of that conversation.
- AC-14: Logout clears assistant, console and activity data and invalidates late work;
  theme/activity preferences remain preferences, not household records.
- AC-15: npm start resolves patched compression >=1.8.2 without serving-path/API
  changes; the locked install, build and existing tests work. No speculative SSR exploit claim.

## Requirements and success

FR-01: AC-01–03 MUST hold without schema migrations or new runtime libraries.
FR-02: AC-04–06 MUST bound work with documented limits and existing errors/contracts.
FR-03: AC-07–11 MUST retain one auth policy and explicit integration migration.
FR-04: AC-12–14 MUST retain Vuex as shared truth and dispose per-view work.
FR-05: AC-15 MUST patch the affected dependency with a narrow lockfile change.
NFR-01: All affected tests plus full backend tests/race/vet/build and frontend
unit/lint/build/E2E pass; new regressions exercise failures, ownership and compatibility.
NFR-02: No automatic device retries, sampling changes, optimistic confirmations,
new household persistence or secret logging. Frontend layout/controls remain unchanged.
SC-01: Each AC has recorded regression evidence and exact check results in plan.md.

## Entities and external interfaces

Existing routes/events/domain types remain authoritative. TRUSTED_PROXY_CIDRS adds
explicit proxy trust without changing Docker routing or published ports. MCP session routing is transport
metadata, not a device/domain identifier. Authentication changes require service clients
to use the established service-token endpoint when enabled. Storage byte ownership changes
no disk bytes or key format. Retry of physical commands remains manual.

## Questions and assumptions

Q1 (resolved scope): The user requested leaving MCP authentication for a later review.
Public MCP/context access remains unchanged; this is deferred security work, not a claim
that those routes are protected.
Q2 (resolved topology): The user directed inspection of the Dockerfile. The frontend
image runs Nginx using frontend/config/nginx.conf; Compose supplies host.docker.internal
and Nginx proxies to its published backend port 4110. Keep this path. Docker allocates
addresses per deployment: do not guess a subnet or automatically trust private peers.
Document how to identify the immediate proxy address and set TRUSTED_PROXY_CIDRS before
upgrading. No deployment is authorized or performed; unconfigured forwarded requests
fail closed for local sign-in while Google sign-in remains available.
Assumption: ordinary compatibility means preserve valid inputs/results; reject malformed,
oversized, expired or unauthorized work with explicit failures rather than silently execute.
