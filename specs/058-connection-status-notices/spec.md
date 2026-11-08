# Spec: Compact connection status instead of stacked notices

ID: 058 | Status: Implemented (unreviewed) | Owner: unassigned | Updated: 2026-11-06
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

`StatusBanners.vue` renders connection and session notices as a fixed panel of full-width
rows at the bottom of the screen. The panel:

- covers content and grows with each notice;
- mostly offers Dismiss, which is the only action for several states;
- resets its `dismissed` flag whenever the notice list changes, so the same notice returns
  on every state change and the user repeatedly dismisses it;
- duplicates the header connection pill, which already encodes the transport state.

Outcome: one compact header status pill is the single surface for connection health. It
shows a warning badge when there is an actionable notice, and opens a small details dialog
with the notice text and its retry action. No persistent connection banners remain.

In scope:

- header pill as a button with a warning badge and an accessible label;
- a Connection dialog with the current status, every notice text and Retry/Reconnect;
- removing the fixed connection/session/snapshot notices;
- keeping the content-level permit-join banner unchanged;
- tests for snapshot failure, transport loss/recovery and content stability.

Exclusions: permit-join banner redesign, MCP status details beyond the existing text,
transient toasts, automatic opening of the dialog, backend/API changes.

## Scenarios and acceptance

- AC-01: Given a session, snapshot or transport failure, then no persistent
  connection banner is rendered; the header pill keeps its state colour and shows a warning
  badge with an accessible label naming the issue count.
- AC-02: Given the pill, when activated, then a “Connection” dialog lists the current state
  and every notice with its retry action; Close or Escape dismisses it.
- AC-03: Given a dismissed dialog, then the state remains visible in the pill and the same
  notice is not shown again until the user reopens the dialog.
- AC-04: Given transport loss then recovery, then the pill changes to reconnecting and back
  to connected without moving page content.
- AC-05: Given a phone viewport, then the pill remains a 44 px target and the dialog uses
  the existing bottom-sheet layout.

## Requirements and success

- FR-01: Connection notices MUST render only in the header status dialog, never as a
  persistent page panel; acceptance: AC-01, AC-03.
- FR-02: The pill MUST expose the status and issue count to assistive tech and open a
  labelled dialog; acceptance: AC-01, AC-02.
- FR-03: Retry/Reconnect MUST remain available for each notice kind (session, snapshot,
  transport); acceptance: AC-02.
- FR-04: The permit-join banner MUST stay unchanged; acceptance: AC-04 (regression).
- NFR-01: The change MUST NOT move page content on connection state changes.
- NFR-02: New interactive targets MUST be 44 px on phones; the dialog is keyboard and
  Escape operable.
- SC-01: `frontend/e2e/StatusBanners.spec.ts` covers startup, snapshot retry, transport
  loss/recovery and the permit banner.
- SC-02: Full unit/build/lint/browser checks pass with recorded results.

## Questions and assumptions

- Q-01: No toast is added. The header badge/dot is the non-intrusive signal; a toast would
  reintroduce the “appears every time” complaint. Resolved.
- Q-02: The dialog is user-initiated; it never opens automatically. Resolved.
- Q-03: The permit-join banner stays because it is a time-limited action, not a connection
  notice. Resolved.
