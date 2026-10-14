# Spec: Stable authentication page handoff

ID: 009 | Status: Implemented | Owner: Codex | Updated: 2026-10-14
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

Follow-up to completed spec 008. App.vue switches to AppShell as soon as auth is
committed, before asynchronous router navigation finishes. The shell's RouterView
then mounts LoginPage again on the still-current login route, restarting checks
and moving the full-height login layout inside the shell. The user reports that
flicker is more obvious after the checking-state change.
Initial app mounting also precedes router readiness, so OAuth cleanup/navigation
can race initial navigation and have its URL changes overwritten.
Fix presentation ownership and initial route readiness during the handoff. Backend,
cookie policy, device semantics, theme, and session persistence are excluded.

## Scenarios and acceptance

### US-01 — Complete sign-in without remounting the login screen (P1)

- AC-01: With a valid cookie or successful OAuth return and delayed destination
  loading, the original standalone login stays mounted until navigation completes;
  no shell, nested login, or sign-in options appear during that wait.
- AC-02: Once the protected destination is ready, the shell appears and login is removed.
- AC-03: Signed-out, failed-validation/retry, local/Google sign-in, logout, retained
  sessions during outages, and protected-route access retain spec 008 behavior.

## Requirements and success

- FR-01: Shell visibility requires both authenticated state and a protected current
  route; acceptance AC-01/02.
- FR-02: Preserve existing session checks and navigation contracts; acceptance AC-03.
- NFR-01: No additional login mount or mode check during successful handoff;
  verify at desktop 1280x800 and mobile 390x844 with a delayed route module.
- SC-01: Browser regression fails before the fix and passes after it on both viewports.

## Entities and external interfaces

No changed network/persistence contracts. Existing Vuex identity and router metadata
remain authoritative. No new controls; existing status, focus, and retry remain.

## Questions and assumptions

No unresolved contract/security questions. Slow route loading is a controlled
reproduction of the observed rendering race, not an assumption about hub latency.

## Requested cleanup acceptance

- AC-04: Concurrent initial/recovery/retry checks share one in-flight operation;
  disposal removes listeners and late results do not update page-owned state.
- AC-05: Unexpected session-check rejection settles loading with an actionable retry
  error; success, failure, hidden/foreground recovery and disposal have unit coverage.
- FR-03: Separate session coordination from presentation using established composable
  ownership, readonly presentation state and named derived state; AC-04/05.
- NFR-02: New session-coordination callables meet C <= 10 and CRAP <= 30 using
  documented per-callable statement coverage mapping; baseline unknown coverage
  must be labelled rather than invented. Evidence: quality.md.
