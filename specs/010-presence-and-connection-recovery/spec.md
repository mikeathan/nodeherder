# Spec: Presence recovery and truthful connection notices

ID: 010 | Status: Implemented | Owner: Codex | Updated: 2026-10-16
Constitution: 1.0.0 (proposed) | Components: backend, frontend

## Problem, scope, exclusions

An isolated replay with the real Living room presence recipe reproduces OFF after five
minutes followed by a missed ON when a dark illuminance report arrives after presence.
Current trigger selection ignores condition-only deltas. Current live attribution is
unconfirmed; this fix addresses the demonstrated application defect. The frontend also
renders a red connection warning during every healthy initial connection and moves page
content when it disappears. An HTTP snapshot failure incorrectly overwrites socket status.
Fix these behaviors and the directly touched lifecycle/quality gaps. No household commands,
deployment, recipe threshold/timing edits, persistence migration, new dependency or auth
policy change. Specs 005/006/008/009 remain historical compatibility requirements.

## Scenarios and acceptance

### US-01 — Enter a dark room reliably (P1)

Independent verification: fake-clock/MQTT replay through real lifetime updates and recipe.

- AC-01: Presence=true followed by illuminance<=45 publishes exactly one ON to
  Living Room Light/set, including after the configured five-minute OFF.
- AC-02: Illuminance before/with presence and threshold/zero boundaries work; while
  identical reports remain deduplicated, each changed relevant reading is evaluated
  through existing trigger policy. This preserves historical repeated ON requests while
  presence is true and readings remain dark; no new once-per-eligibility policy is added.
- AC-03: OFF delay, return cancellation, no restart on repeated false, fresh state before
  sampling, source disable and unrelated measurement isolation remain. Repeated physical
  events still execute; changing conditions must not replay cached button events.
- AC-04: Failed actions remain observable through existing errors; manual triggers remain
  isolated from device evaluation. Each selected trigger evaluates each condition at most once;
  action failures do not install a new suppression latch for later changed inputs.

### US-02 — Arrive without false outage warnings or layout movement (P1)

Independent verification: mock-hub browser tests, 1280x800 and 390x844, plus fake timers.

- AC-05: OAuth/cookie handoff, authenticated reload and delayed successful validation or
  first socket connection show no outage notice. The stable header may say Connecting.
- AC-06: Actual socket loss/failed handshake shows recovery status; a first attempt still
  pending after 10 seconds is failed and retried with bounded backoff. Recovery clears
  the warning. Repeated retry clicks never create concurrent sockets.
- AC-07: A failed initial HTTP snapshot has its own retry message, preserves open socket
  truth, and clears on successful reload without requiring a new socket.
- AC-08: Session failure, actual outage, snapshot failure and their recovery do not move
  the content. Notices are keyboard accessible, dismissible, readable without color,
  and leave navigation available at both viewports and in full-screen panel mode.
- AC-09: Logout/unmount cancels retry/connection timers and pending data loading; stale
  socket callbacks and late HTTP/session results cannot reopen or corrupt the new link.

## Requirements and success

- FR-01: Select state triggers when their primary property or a condition dependency changes;
  retain changed-delta/event semantics and existing MQTT recipe/wire formats; AC-01…04.
- FR-02: Separate transport startup/loss/retry from session validation and snapshot loading;
  use established Vuex ownership and authenticated HTTP helpers; AC-05…09.
- NFR-01: Changed production callables have complexity <=10 and CRAP <=30 (mapped coverage
  where supported, otherwise a verified conservative upper bound);
  record baseline, tools, coverage basis and limitations per callable. Split responsibilities
  when complexity is high; meaningful failure coverage is required, not score-only tests.
- SC-01: Before-fix regressions fail and after-fix acceptance and required component checks
  pass; all automated dependencies are isolated from household equipment.

## Entities and external interfaces

Presence and illuminance are independent reported state; light commands are requests,
not confirmation. No synthetic sensor values or repeated command retries are introduced.
Condition declarations are the single source of dependency truth; no eligibility cache
is added. Selection is stateless; accepted sensor state remains owned by the lifetime service.
Time conditions still require an incoming trigger/dependency change; no scheduler changes.
Browser connection states and snapshot status are internal; backend HTTP/WS/MQTT payloads,
auth cookies, routes, metrics sampling/units/retention and storage remain unchanged.

## Questions and assumptions

User authorizes implementation on a new branch. Preserve historical relevant-update behavior
with generic dependency declarations and no eligibility cache; event triggers require new events.
No claimed maintainer ratification, exception or review approval.
Current household report capture is outside automated verification and remains a limitation.
