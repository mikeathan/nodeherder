# Spec: Session reliability and local sign-in

ID: 008 | Status: Implemented | Owner: Codex | Updated: 2026-10-13
Constitution: 1.0.0 (proposed) | Components: both

## Problem, scope, exclusions

Session cookies lack persistence, and frontend validation treats network/server failures
as logout. The local sign-in screen incorrectly promises Google authentication.
Fix browser-session persistence, failure recovery, and request-specific local presentation.
HTTPS, longer token lifetimes, token renewal, server session storage, and device behavior
are excluded. The existing HTTP/LAN access model is accepted by the user for this work.

## Scenarios and acceptance

### US-01 — Keep valid sign-in across browser restarts and outages (P1)

- AC-01: A new session cookie persists until its JWT expires (currently 24 hours),
  retaining HttpOnly, path /, and SameSite=Lax. Custom JWT durations remain aligned.
- AC-02: Network failure, timeout, 5xx, or malformed session data preserves previously
  authenticated state and shows an unavailable status with retry; 401 clears auth.
- AC-03: Retry or foreground/network recovery validates again without device commands;
  concurrent checks and responses after sign-out cannot restore stale authentication.
- AC-04: A browser without cached auth sees a checking state before the sign-in options;
  valid cookies resume the session, while failure offers retry without claiming identity.

### US-02 — Describe the actual sign-in path (P1)

- AC-05: Local requests with local mode enabled see Continue on local network and
  explicit no-Google wording. Other requests see Sign in with Google.
- AC-06: Mode detection failure shows an actionable error rather than guessed Google
  wording. Mode is checked by the backend using the same predicate as login.
- AC-07: Direct local-session start rejects requests outside that predicate with 403;
  Google and service-token lifetimes remain unchanged.

## Requirements and success

- FR-01: Cookie persistence matches server expiry; acceptance AC-01.
- FR-02: Session validation distinguishes rejected authentication from unavailable
  service and protects against stale responses; acceptance AC-02/03/04.
- FR-03: Request-specific sign-in presentation and local start use backend policy;
  acceptance AC-05/06/07.
- NFR-01: Session/mode checks terminate within 8 seconds and listeners clean up on unmount.
- NFR-02: Keyboard and mobile/desktop interaction remain accessible (390/1280 px).
- SC-01: Regression tests prove transient failure does not clear auth and 401 does.
- SC-02: Browser tests demonstrate local/Google flows and outage recovery on both viewports.

## Entities and external interfaces

Session identity remains the existing authenticated user. GET /api/auth/mode is public,
returns only {mode: local|google}, has no cookie/identity side effects and is not cached.
GET /api/auth/me keeps its existing payload and protected route. Login/logout contracts
remain compatible. Local start gains a 403 denial when local access is not offered.

## Questions and assumptions

Resolved from conversation: exclude HTTPS; accept current local mode; fix reliability
and misleading Google text. Use existing 24-hour lifetime instead of the previously
suggested 30 days. Do not automatically sign in on LAN or silently renew expired tokens.
Existing proxy-header trust and in-memory revocation are debt, not expanded here.
