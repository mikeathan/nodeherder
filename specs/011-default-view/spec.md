# Spec: Default view on this browser

ID: 011 | Status: Complete | Owner: Codex | Updated: 2026-10-16
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

Panel has a bookmarkable URL, but visiting `/` always opens Home. Users want an
explicit browser-local default without ordinary navigation overwriting it.
Scope: Panel and main navigation pages. Excludes device details, editors/creation,
account sync, device commands, and remembering arbitrary URLs or household data.

## Scenarios and acceptance

### US-01 — Choose a startup view (P1)

Independent verification: isolated browser with mock hub.

- AC-01: Pin Panel; revisit `/` or sign in again; Panel opens. Refresh keeps it.
- AC-02: Visit another page; the pin stays. Explicit URLs and login redirect targets win.
- AC-03: Pin another main page; it replaces Panel. Unpin; `/` opens Home.
- AC-04: Missing/invalid/blocked storage falls back to Home without breaking navigation.
- AC-05: Pin controls fit 390px mobile and 1280px desktop, support keyboard activation,
  visible focus and accessible pressed state, and explain browser-local scope.
- AC-06: Detail/editor/create routes have no pin control.

## Requirements and success

- FR-01: Persist only an allowlisted main route name in browser local storage (AC-01–04).
- FR-02: Explicit pin/unpin control in the app title area and Panel dock (AC-03/05/06).
- NFR-01: No horizontal overflow at specified widths; no new dependencies or backend calls.
- SC-01: All acceptance flows pass with a mock hub; frontend tests, lint and build pass.

## Entities and external interfaces

One optional presentation preference per browser/origin. No auth bypass, no HTTP/WS
contract change, no user/device identifiers. Removing the key recovers Home.

## Questions and assumptions

Resolved: use a separate control, browser storage, one default, explicit links win.
Assumption: all primary sidebar destinations are eligible; temporary subpages are excluded.
Blocked storage preserves the choice in memory for this tab only.
