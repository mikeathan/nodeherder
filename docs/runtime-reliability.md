# Runtime resource and session behavior

Implementation scope/evidence: [spec 024](../specs/024-runtime-security-and-resource-lifecycle/plan.md).
These changes preserve device commands, confirmed-state presentation, automation inputs,
metric units/UTC ordering and database formats. They introduce no automatic command retry.

## Resource limits

| Boundary | Behavior |
| --- | --- |
| Incoming HTTP body | Maximum 1 MiB, including unknown-length/chunked bodies; 413 before handler mutation |
| HTTP connection | Header 5s, request read 15s, idle 60s; no fixed stream write timeout |
| WebSocket | 1 MiB incoming message, 4 KiB socket buffers, 256 queued messages |
| Slow WebSocket client | Close on outbound overflow; frontend reconnect refreshes the shared snapshot and retains newer live reports/config |
| Snapshot recovery | One pending read, 8s timeout, at most 256 captured config mutations; superseded/logout requests aborted; failed refreshes can be retried |
| Assistant proxy | Caller cancellation, 60s upstream timeout, maximum 4 MiB upstream reply; failed persistence is an error |
| Google OAuth | Caller cancellation, 15s overall exchange/userinfo deadline; access token in Authorization header |
| Legacy metrics query | At most 64 requested devices/exposes, limit 0–10001, ordered range at most 366 days, 250000 scanned rows and 8s deadline |
| SSE | At most 64 connections, 10 messages/client, 10s refreshed write deadline, 15s heartbeat |

Legacy metrics use one successful scan/device for all exposes; adapter errors retain
per-expose fallback compatibility. Repeated expose names produce one result. Query
budgets fail explicitly rather than returning partial success. The existing history
endpoint keeps its stricter request constraints and shared scan/time limits.

Recovery replays captured config changes through the same Vuex mutations used for
live updates, retaining unrelated server settings. Replay overflow fails the refresh
without replacing the existing registry; retry reads state again without resending
device commands. WebSocket disconnects emit no debug text into the event protocol.

Storage reads return owned bytes. Callback-based reads borrow bytes only during their
transaction. Pruning uses one atomic write transaction and propagates failures; it
preserves the on-disk format and retained data. Shutdown releases owned stream resources.

## Sessions and browser state

Authenticated WebSockets and SSE stop on token expiry/revocation. Listener/timer cleanup
belongs to the transport request, including failed upgrades. WebSocket browser origins
must match the request host or FRONTEND_ALLOWED_ORIGINS; non-browser clients may omit
Origin and continue using the existing authentication mechanism.

Cookies use Secure for direct HTTPS. A TLS terminator can set AUTH_COOKIE_SECURE=true
explicitly; the generated local sign-in URL uses this same HTTPS policy. Client-supplied
protocol headers do not control it. Leave it unset
for intentional local HTTP. Apply the setting with the deployment's actual HTTPS
configuration, and roll it back together with that configuration.

Assistant requests carry cancellation signals and conversation/session ownership.
Switching chats, logout and unmount invalidate late responses; the last history selection
wins. Logout clears assistant, console and activity records. Appearance and activity
preferences remain preferences. The application shell owns the shared permit-join timer,
including when panel mode hides sidebar/banner consumers.

## MCP reply routing and compatibility

Direct POST /api/mcp still returns its JSON-RPC reply in the HTTP body. A legacy SSE
client connects to GET /api/mcp or GET /api/mcp/events and posts to the advertised
endpoint, now /api/mcp?sessionId=<opaque-id>. That session alone receives the SSE reply;
the session is also bound to its authenticated principal when authentication applies.
Resource-change notifications retain shared delivery. Unknown/disconnected session IDs
return 404. Stream overflow closes that session rather than silently losing a reply.

The existing MCP SDK SSE client is covered by an integration regression for initialization
and resource listing. A nonstandard client hardcoding the POST URL while ignoring the
advertised endpoint must use HTTP replies or adopt the advertised session URL.

## Nginx trust and coordinated upgrade

The frontend Dockerfile installs frontend/config/nginx.conf into nginx:alpine.
Compose maps host.docker.internal to the host gateway; Nginx proxies API/WebSocket
requests to host.docker.internal:4110. This routing and the published ports are unchanged.
Nginx overwrites X-Real-IP and X-Forwarded-For with its observed client address.

Before upgrading a deployment that uses local sign-in behind Nginx:

1. Identify the Nginx peer address as seen by the backend. The frontend container's
   current address is a starting point, available without reading secrets:

   ```bash
   docker inspect nodeherder-frontend --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{"\n"}}{{end}}'
   ```

   The host-gateway/published-port path may translate the source address. Confirm the
   backend's immediate peer using host/container network diagnostics; a container IP
   alone does not prove what the backend sees. Trust only dedicated proxy addresses,
   never all private networks, all Docker peers or 0.0.0.0/0. Keep direct backend access
   behind the deployment's firewall/access controls.
2. `scripts/manage-stack.sh start|rebuild` does this automatically: it creates the
   Compose network, reads the network gateway (and the frontend container address when
   that service is in scope), and passes them to the backend through
   `.generated/compose.proxy-allowlist.yml`. Keep `OFFLINE_STRICT_LOCAL=true` where
   local sign-in is intended; false disables it. If you started the stack by hand, or
   want to pin peers yourself, set `TRUSTED_PROXY_CIDRS` in backend/.env instead — an
   explicit value always wins over the computed one. Do not trust all private networks,
   all Docker peers or 0.0.0.0/0.
3. Rebuild/recreate the backend and frontend together using the existing stack
   lifecycle. Confirm GET /api/auth/mode through Nginx returns local for a LAN client
   and google for an external client. Confirm local sign-in still creates a usable
   session. Forged IP headers on a direct untrusted request must not create a session.
   These are deployment checks, not checks run against household services here.
4. [Docker may assign new addresses when containers are recreated](https://docs.docker.com/compose/how-tos/networking/#updating-containers-on-the-network). Recheck the immediate
   peer and update/recreate the backend with the allowlist when it changes. A narrow
   dedicated proxy range can be used where the operator controls that network. The
   stack manager recomputes the peers on every `start`/`rebuild`, so a manual value in
   backend/.env is the only thing that can go stale.

The allowlist is empty by default. Unlisted peers with forwarding metadata, invalid
allowlist entries, malformed client IPs and listed proxies lacking a client IP cannot
select local sign-in: mode is google and offline/start returns 403 without a session.
Direct local requests from peers outside the proxy allowlist, without forwarding metadata,
retain local sign-in. A listed address must always supply a client address, including
when that address is loopback. This fail-closed
default requires configuring the proxy before upgrading to preserve proxied local login.
One X-Real-IP is authoritative; when absent, X-Forwarded-For is read right to left through
listed proxies, with at most 32 hops/4096 bytes. Unsupported Forwarded alone cannot
authorize local sign-in. No forwarded scheme header can enable Secure cookies; use the
explicit HTTPS setting described above.

Rollback restores backend/frontend images and their matching configuration together.
The prior backend trusts forged client-IP headers, so keep its port restricted while
rolling back. No deployment settings were guessed or applied here.

## Deferred authentication migration

The user deferred MCP/context authentication migration; those routes retain their
existing public access for a later review. Proxy hardening does not protect those
public routes. The constitutions remain proposed; this document claims no ratification
or approved exception.

## Node serving dependency

The scoped serve → compression override pins compression 1.8.2 to address
[GHSA-vc2v-76pw-4v95](https://github.com/expressjs/compression/security/advisories/GHSA-vc2v-76pw-4v95).
It applies to npm start; the Docker frontend serves through Nginx. Keep the override
until serve itself requires a patched release. Other dependency advisories require
separate reachability/version-migration review; this patch is not a blanket clean audit.
