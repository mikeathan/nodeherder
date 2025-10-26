# OAuth/OIDC and Security Hardening TODOs (Prioritized)

Last updated: 2025-10-26
Branch: feature/add_authentication

Scope and context
- Goal: Securely integrate OAuth/OIDC and harden transport, sessions, WebSockets, and configuration.
- Frontend exposure: Will be served publicly via Cloudflare Tunnel. Treat Cloudflare as an additional perimeter, not a replacement for in-app authentication/authorization.

Severity key
- MUST = critical before public exposure
- SHOULD = high priority after MUST items
- NICE-TO-HAVE = optional; consider for defense-in-depth or scale

## Critical (MUST before public exposure)
1) Protect all APIs and WebSockets with authn/z
- Enforce authentication on every REST endpoint and at WS handshake (no tokens in query strings).
- Authorize per action with least-privilege scopes/roles.

2) Correct OIDC flow and token validation
- Use Authorization Code + PKCE.
- Validate JWKS signature, alg, kid; enforce iss, aud (and azp if applicable).
- Check exp/nbf/iat with small clock skew; reject "none"/weak algs.
- Use access tokens for API authorization (not ID tokens).

3) Remove hardcoded secrets and rotate
- Move MQTT and any OAuth client secrets to env vars/secret manager.
- Immediately rotate currently committed credentials.

4) Enforce HTTPS/WSS end-to-end
- Only serve over HTTPS/WSS. With Cloudflare Tunnel, browser→CF is HTTPS and tunnel is mTLS; still mark cookies Secure.
- Enable HSTS (can be set at Cloudflare edge).

5) CSRF protection for cookie-based sessions (BFF)
- If using HttpOnly session cookies, require CSRF token for state-changing requests (double-submit or equivalent) and SameSite=Lax/Strict where possible.

6) Prevent token exposure in the browser
- Prefer BFF session cookies (HttpOnly, Secure). If SPA must hold tokens, store in memory only and avoid tokens in URLs.

## High (SHOULD)
7) WebSocket hardening
- Authenticate handshake; verify Origin equals your CF domain; avoid token in query strings; close on session expiry.

8) Strict CORS and origin controls
- Allow only your Cloudflare hostname(s); restrict methods/headers; disable credentials unless required.

9) Proxy awareness behind Cloudflare
- Respect X-Forwarded-Proto/Host (or CF-Connecting-* headers) for absolute URLs, cookie flags, and OIDC redirects.
- Configure trusted proxy settings in the backend server.

10) Authorization model
- Define roles/scopes and enforce per endpoint and WS operation (least privilege, deny-by-default).

11) Don’t persist secrets in BoltDB
- Keep client secrets/refresh tokens out of settings.db. If unavoidable, encrypt at rest and restrict file perms.

12) Rate limiting and abuse controls
- Apply rate limits to auth endpoints and sensitive operations; set request size/timeouts; add basic DoS guards.

## Medium (Recommended)
13) Security headers
- CSP (prefer nonces), X-Frame-Options: DENY, X-Content-Type-Options: nosniff, Referrer-Policy: strict-origin-when-cross-origin, Permissions-Policy (least privilege), HSTS.

14) Frontend route guards and protected data loading
- Guard protected routes; avoid preloading sensitive data before auth; graceful 401/403 handling.

15) Input validation and schemas
- Validate body/params/query for all endpoints; enforce strict schemas especially for device control paths.

16) Logging and telemetry hygiene
- No tokens/secrets/PII in logs; add audit logs for auth events; centralize logs; monitor anomalies.

17) Dependency and build hygiene
- Pin versions, audit dependencies, remove unused libs; enable image/package scanning where applicable.

## Low (Nice-to-have / optional)
18) Session fixation and logout hardening
- Rotate session IDs on login/privilege change; clear cookies securely; validate OIDC state and nonce thoroughly.

19) Sender-constrained tokens (DPoP/MTLS)
- Consider if SPA must hold tokens and IdP supports DPoP; less relevant if using BFF.

20) Cloudflare Access as an outer gate
- Optionally protect routes with CF Access (SSO/MFA) and validate CF-Access-Jwt-Assertion for admin areas.

21) Secret management upgrade
- Migrate from env vars to a secret manager with rotation policies as the project grows.

## Acceptance checks (post-implementation)
- Unauth calls to any API/WS return 401/403; WS rejects unauthenticated handshake.
- Token validation rejects wrong iss/aud/azp/alg and expired tokens; nonce/state verified.
- CSRF attempts fail for cookie-authenticated POST/PUT/PATCH/DELETE.
- No tokens in storage or URLs; Secure/HttpOnly/SameSite cookies verified.
- HTTPS/WSS enforced; HSTS active; security headers present and effective.

## Cloudflare Tunnel deployment notes
- Set external base URL and OIDC redirect/logout URLs to the CF hostname.
- Enforce Secure cookies and proper SameSite; verify origin on WS; lock CORS to CF domain(s).
- Consider enabling HSTS and additional security headers at CF edge; complement with app-level headers.
