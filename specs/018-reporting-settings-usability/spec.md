# Spec: Clear reporting settings and device category overrides

ID: 018 | Status: Complete | Updated: 2026-10-17
Constitutions: parent/backend/frontend v1.0.0 proposed.

Users cannot distinguish live filtering, inherited category defaults and history.
Adding one sensor exposes too many unrelated controls. A device cannot override its
category interval without configuring every sensor individually.

- AC-01: Device categories offer shared default, every change and custom seconds;
  saved custom values affect only that device. Explicit zero overrides shared delay;
  reset removes the entry and resumes inheritance, including later shared edits.
- AC-02: Resolution matches frontend/backend: device field > shared field > device
  category > shared category > immediate. Existing specific rules remain authoritative.
  Measurements stay immediate/no minimum change unless explicitly configured.
- AC-03: Main sensor adjustment exposes a native-unit minimum-change input, a concise
  explanation of zero and last-displayed comparison, explicit reset and optional
  collapsed timing. History controls are separate and collapsed by default.
- AC-04: History category overrides follow the same hierarchy, without changing live
  settings. Existing field history rules remain accessible and preserved on live reset.
- AC-05: Controls show scope/inheritance, keyboard names/focus, validation and offline
  feedback. Desktop 1440px and mobile 390px layouts have no horizontal overflow.
- AC-06: Raw state/automation/event/confirmation behavior remains independent of live
  intervals and thresholds. No new per-report allocations/workers/timers are introduced
  beyond bounded policy lookups; existing publisher owns scheduling and shutdown.

No commands/deployment/commit. Existing settings wire maps are reused. Reset means
remove the selected override, not delete shared defaults or unrelated history settings.
