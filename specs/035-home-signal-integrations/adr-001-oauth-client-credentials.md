# ADR-001: OAuth 2.0 client credentials for provider authentication

Status: Proposed | Date: 2026-10-29 | Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)
Rules: NH-03/04/06, BE-01/03/05, FE-02/05

## Context

Every installed provider authenticates with a static secret: `none`, `bearer`,
`optional_bearer`, `apikey_header` or `apikey_query`. The UK Fuel Finder API requires an OAuth 2.0
access token obtained with client credentials, and future providers (calendar, health) are likely
to need the same flow. Holding a manually copied bearer token is not viable: it expires, and the
user cannot refresh it from the app.

Two origins are involved: the token endpoint issues a short-lived token, and the API origin accepts
it. The current `AuthConfig` has a single `secret_slot` and the request builder binds credentials
to the declared base origin.

## Options

1. Manual long-lived bearer token pasted into the credential slot. Rejected: expires, poor
   recovery, and silently becomes a stale integration.
2. Proxy the provider through an external service. Rejected: new deployment dependency and a
   second trust boundary.
3. CSV export only. Rejected: twice-daily data, not a live integration, and a file-import subsystem
   would be needed.
4. Add a bounded client-credentials auth mode with a declared token endpoint, private client
   id/secret slots, in-memory token cache and single retry.
5. Full OAuth authorization-code flow with user consent. Rejected for this phase: browser redirects,
   callback routes and refresh tokens expand scope well beyond the provider.

## Decision

Adopt option 4.

- `AuthConfig` gains `mode: oauth2_client_credentials`, `token_url`, optional `scope` and two
  private slots (`client_id`, `client_secret`). No secret ever appears in YAML.
- The token request is `POST` form-encoded (`grant_type=client_credentials`, optional `scope`) with
  HTTP Basic credentials. `token_url` must be public HTTPS on 443 and is an independently declared,
  reviewed destination, so it gets the same destination-review treatment as a provider origin.
- The access token is cached in process memory keyed by definition revision plus token URL, with
  expiry minus a safety margin. It is never persisted, logged, returned to the browser or written
  to an error.
- The token is sent only to the provider base origin. A 401/403 triggers at most one token refresh
  and retry; a second failure produces a `reconnect` failure and preserves last-good data.
- `RequiresCredential` returns true, so setup requires both slots before search/refresh. Missing or
  rejected credentials surface as an actionable reconnect state.
- No refresh token is requested or stored. A backend restart simply fetches a new token.
- Token requests count against the operation deadline and body/time limits like any provider call.

## Consequences

- One new auth mode must be implemented consistently across compiler, request builder, credential
  lifecycle, definition editor, frontend contracts and tests.
- The token endpoint is a second egress destination; it must be disclosed during setup and
  reviewed when a provider definition changes.
- In-memory caching means a token is not shared across processes; the runtime is single-process by
  design, so this matches existing constraints.
- A provider that needs per-user authorization (not client credentials) still requires a separate
  design; this ADR does not deliver calendar/health OAuth.

## Validation and reconsideration

Tests: token request shape, scope, expiry cache, single retry, no retry loop, origin binding,
redaction in errors/logs, missing/expired/invalid credentials, and existing static auth
regressions. Reconsider if providers require authorization-code consent or refresh tokens, or if
token endpoints need mutual TLS.
