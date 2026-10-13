# Plan: Session reliability and local sign-in

Spec: [spec.md](spec.md) | Status: Implemented; maintainer review pending | Author: Codex | Updated: 2026-10-13
Constitutions: parent/backend/frontend 1.0.0 (proposed)

## Context and design

backend/internal/auth owns JWT expiry, cookie issuance, and mode selection. Persist
cookies with Max-Age and Expires from the issued JWT expiry, not a second hardcoded TTL.
Expose a no-store mode-only endpoint in authDelegator and enforce its predicate on
local-session start. Existing API auth middleware still validates every protected call.

frontend/services/auth.service returns null only for 401; throws on other failures,
with bounded requests. Auth Vuex records validation state and a revision to ignore
checks overtaken by explicit login/logout. Deduplicate pending checks per store state.
Only user/authenticated are persisted, preserving existing nodeherder_auth data.
useHubConnection validates before fetching hub data and retries on online/focus/visibility
with listeners owned by AppShell. An unavailable banner supplies an explicit retry.
LoginPage owns initial checking and mode presentation, including failure/retry and
Google/local copy, with cleanup for recovery listeners. No new dependencies.

Alternatives: 30-day JWT/renewal rejected for current scope; auto-LAN sign-in would
change user consent/logout; hostname guessing cannot reflect backend policy.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | AC-01..07 and tasks precede production edits | PASS | PASS (agent review; evidence below) |
| NH-02 | No device command/state/event changes | N/A | PASS (agent review; evidence below) |
| NH-03 | Mode-only metadata; local start narrowed; existing HTTP explicitly accepted | PASS | PASS (agent review; evidence below) |
| NH-04 | Existing auth/store/service boundaries and additive mode contract | PASS | PASS (agent review; evidence below) |
| NH-05 | Unit/regression/browser checks below; mock hub only | PASS (design) | PASS (agent review; evidence below) |
| NH-06 | No data migration; bounded requests; listener cleanup | PASS (design) | PASS (agent review; evidence below) |
| BE-01 | Auth transport retains existing service boundary | PASS | PASS (agent review; evidence below) |
| BE-02 | Device lifetime unaffected | N/A | PASS (agent review; evidence below) |
| BE-03 | No new backend goroutines; frontend cancellation/cleanup | PASS (design) | PASS (agent review; evidence below) |
| BE-04 | Existing JWT configuration unchanged | N/A | PASS (agent review; evidence below) |
| BE-05 | Public mode payload and local-start 403 specified | PASS | PASS (agent review; evidence below) |
| FE-01 | Shared auth remains Vuex; components own presentation | PASS | PASS (agent review; evidence below) |
| FE-02 | Runtime mode/user checks and mock contracts | PASS (design) | PASS (agent review; evidence below) |
| FE-03 | Checking/unavailable/401/retry distinguished | PASS (design) | PASS (agent review; evidence below) |
| FE-04 | Named buttons, focus and desktop/mobile browser checks | PASS (design) | PASS (agent review; evidence below) |
| FE-05 | Existing cookie helper, no browser tokens/secrets | PASS | PASS (agent review; evidence below) |

## Contracts, data, rollout

GET /api/auth/mode -> 200 application/json {mode: local|google}; Cache-Control:no-store.
Local offline/start -> 403 JSON when mode predicate rejects; success unchanged.
Auth/me -> 200 user or 401 unchanged; failures no longer mean frontend logout.
Deploy backend first then frontend; old frontend ignores additive endpoint. New frontend
against old backend shows mode lookup failure with retry. Existing session cookies become
persistent on next sign-in; no forced logout or token lifetime extension. Rollback restores
old frontend/backend; issued tokens remain compatible.

## Traceability and verification

| IDs | Paths | Checks | Actual result |
| --- | --- | --- | --- |
| FR-01/AC-01 | backend/internal/auth | Cookie/JWT duration regression | PASS: 24h/custom 2h, JWT Expires and cookie flags |
| FR-02/AC-02/03/04/NFR-01 | frontend services/auth, store/auth, composables | Jest failure/status/revision tests; browser outage/retry | PASS: network/5xx/malformed/401, dedupe, late responses, 8s timeout and restored cookie |
| FR-03/AC-05/06/07 | auth.go, LoginPage, mock routes | Mode/denial Go tests; browser local/Google/failure | PASS: local/external/disabled mode; rejected start has no cookie; local/Google/retry UI |
| NFR-02/SC-02 | frontend/e2e/auth.spec.ts | Desktop/mobile, keyboard, checking | PASS: 1280×800 / 390×844, keyboard Enter and focus, screenshot inspection |
| NH-05 | both | go test ./...; go vet ./...; go build ./...; npm test -- --runInBand; npm run lint; npm run build; affected Playwright | PASS (agent review; evidence below) |

## Debt, risks, exceptions

HTTP transport and in-memory logout blacklist remain as discussed; no longer-lived JWT
introduced. IsLocalRequest trusts proxy headers: deployment must keep backend behind a
trusted proxy; this change reuses rather than replaces it. Expiry still requires login
after 24 hours. Cached UI identity is not server authorization; protected calls remain
validated. Existing development mock token verification remains a mock only.
No invented ratification, maintainer approval, or exceptions.

## Completion

Agent acceptance and final diff/rule review completed; maintainer review remains pending.
No real household device tests or deployment were performed. Updated docs/setup.md.

Actual checks (2026-10-13):

- Backend: `go test -count=1 -timeout=5m ./...`, `go vet ./...`, `go build ./...`:
  PASS. `go test -race -count=1 ./internal/auth`: PASS.
- Frontend: `npm test -- --runInBand`: PASS, 23 suites / 152 tests.
  Full Jest reports an open-handle warning despite exit 0; targeted auth suites do not.
- `npm run lint`: PASS, 0 errors / 7 existing console/unused-variable warnings.
- `npm run build`: PASS, with component naming and large-bundle warnings.
- `npm run test:e2e -- e2e/auth.spec.ts e2e/panel.spec.ts`: PASS, 20 tests across
  desktop and mobile Chromium, mock hub only. Mobile Safari is unrun.
- `git diff --check`: PASS. Visually inspected local sign-in screenshots in
  /private/tmp/nodeherder-local-login-{desktop,mobile}.png: readable, no clipping.
- Recovery online events and explicit retry have browser evidence; foreground listener
  registration/removal verified by code inspection. No claim of real-phone reproduction.

Regression history: before cookie implementation, 24h and 2h cases failed with
MaxAge=0; mode test failed because handler was absent. Added frontend regressions
initially could not compile before mode API export existed. Final tests exercise all
specified failures. A reconnect guard initially failed TypeScript (RootState has no ws);
replaced direct state access with a typed existing-module getter. A lint attempt during
Playwright startup failed while test-results was being recreated; sequential rerun passed.

Rollout: backend first, then frontend. Sign in once after deployment to replace the old
non-persistent cookie. Daily token expiry remains deliberate; no 30-day policy or renewal.


## CI follow-up — portable browser artifacts

PR #56 CI run 38445381427 failed only the desktop/mobile local-sign-in tests because
the screenshot path /private/tmp exists on macOS but not on Ubuntu. The aggregate
Run Tests job correctly propagated the E2E failure; the DeviceCard warning was unrelated.
Use Playwright testInfo.outputPath('local-login.png') so artifacts belong to each test's
managed output directory on every runner. This is a test portability correction with
no application behavior changes; AC-05/NFR-02 remain covered by the same assertions.
Full `npm run test:e2e`: PASS, all 42 desktop/mobile tests against the mock hub.
`git diff --check`: PASS. GitHub CI will rerun after the follow-up commit.
