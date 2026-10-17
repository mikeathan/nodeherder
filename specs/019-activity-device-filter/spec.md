# Spec: Recent activity device search

ID: 019 | Status: Complete | Updated: 2026-10-17
Constitutions: parent/frontend v1.0.0 proposed.

- AC-01: Typing a partial device name immediately limits visible activity to matching
  devices, ignoring case and surrounding whitespace. Clearing restores the feed.
- AC-02: Search the entire bounded retained feed before the existing 60-row display
  limit. Order, recording, clear-all and device links remain unchanged; new matching
  events appear while filtered. Filter is local view state, never persisted/sent.
- AC-03: Accessible search/clear-filter controls, meaningful no-match versus waiting/off
  states and result count. Fits 390px mobile and 1280px desktop with theme tokens.

No backend, automation, storage, dependencies, timers or recording policy changes.
