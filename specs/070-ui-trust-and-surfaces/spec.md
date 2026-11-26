# Spec: UI trust and daily surfaces

ID: 070 | Status: Draft | Owner: engineering | Updated: 2026-11-25
Constitution: 1.0.0 (proposed) | Components: both

## Problem, scope, exclusions

The UI cannot distinguish “available” from “still reporting”, loses every transient problem after a
three-second toast, hides logs and maintenance behind weak information architecture, and gives the
network Overview no hierarchy. Panel mode (wall tablet) has no night or offline treatment. These are
one problem: the interface does not tell the truth quickly enough for someone standing in a room.

Evidence: `EntityTile.vue` derives status from availability only; `domain/devices/index.ts:66`;
`hub-state` has no snapshot timestamp; `store/modules/alerts` uses a fixed 3 s timeout with no
inbox; `ConsoleViewer.vue` has no search/export and duplicates Settings → Logger; `StorageSettings.vue`
has only Refresh; `OverviewPage.vue` renders nine competing blocks; `PanelMode.vue` has no offline
origin, paused controls, or night mode.

**In scope (all designed surfaces):**
Home cards and status strip; Overview; Devices and device detail; Integrations; notifications and
new-session digest; logs and log archive download; Settings index and available cleanup/export;
automation row actions; Panel mode; shared Ring 0 primitives; mobile layouts for all of the above.

**Out of this change:** automation run-history deletion (no data source/schema yet). Automation
editor/creator, Assistant, Login and integration setup are now aligned or audited; typed contracts
still apply if touched.

## Scenarios and acceptance

### US-01 — Know when data is old (P1)

Independent verification: domain and store tests; Home/Overview browser specs.

- AC-01: Given a device with `last_seen` older than its policy, when its tile renders, then it shows
  a quiet/stale age label and does not claim to be live.
- AC-02: Given a stale device whose availability is still `online`, when automations/commands are
  evaluated, then `deviceStatus` remains `online` and control enablement is unchanged (silence is not
  availability).
- AC-03: Given the socket is disconnected or the snapshot is older than the grace period, when any
  shell page renders, then one banner states the snapshot time and offers Retry, and tiles are marked
  as not live.
- AC-04: Given a fresh device, when it renders, then no warning styling is applied.

### US-02 — Nothing actionable disappears (P1)

Independent verification: store unit tests; notification browser spec.

- AC-05: Given an operation failure, when the toast expires, then the message remains in the inbox
  with severity, time, read state and a link/action.
- AC-06: Given the connection is lost and restored, when each transition occurs, then exactly one
  inbox item is produced (deduplicated) and the unread count updates.
- AC-07: Given a new session after items arrived, when Overview first renders, then a digest shows at
  most three items with Review and Dismiss, and no toast is replayed.
- AC-08: Given the user marks all read, when they reload, then read state persists for that browser.

### US-03 — Logs are findable and keepable (P1)

Independent verification: service tests; browser spec for download and filters.

- AC-09: Given archived files exist, when Settings → Logs opens the Files tab, then file name, size,
  age and a download action are shown using the existing protected endpoints.
- AC-10: Given the live buffer is filtered by level or text, when it updates, then only matching rows
  render and the follow state does not steal scroll while the user reads.
- AC-11: Given the user downloads the current buffer or an archive, then a file is produced and the
  object URL is released.

### US-04 — Overview answers “what first?” (P1)

Independent verification: page browser spec at desktop and mobile.

- AC-12: Given five issues, when Overview renders, then one verdict, at most three ranked actions and
  one links-only health line are visible before scrolling; the network table and four stat cards are
  absent.
- AC-13: Given no issues, then a single all-clear sentence and timestamp render.
- AC-14: Given mobile, then the same verdict and three actions render without horizontal scroll.

### US-05 — Devices and device detail are honest (P1)

- AC-15: Given the device list, then a freshness column/filter distinguishes fresh, quiet and stale
  without changing availability.
- AC-16: Given a quiet device page, then a trust header states the age and controls remain enabled;
  given offline, values are labelled last-known and controls are disabled.

### US-06 — Panel mode is a wall device (P1)

- AC-17: Given the panel at night, then a large clock and four essentials render with warm, low-glare
  styling; a manual night toggle and schedule window are available.
- AC-18: Given the hub is offline, then one banner states the snapshot age, controls show “Paused”
  and are disabled, and Status/Exit remain usable.
- AC-19: Given the panel status page, then four readings and a short attention list render at 2 m
  scale with no tables or scrolling for the common case.

### US-07 — Settings and automations do not dead-end (P2)

- AC-20: Given Settings opens, then an index lists every registered section with a live summary and
  no section is removed; detail pages keep existing controls.
- AC-21: Given the user exports a home layout or automations, then a versioned JSON file downloads;
  given cleanup actions that exist, scope and confirmation are explicit.
- AC-22: Given an automation row, then overflow exposes Run/Edit/Duplicate/Export/Copy ID/Delete with
  accessible labels.

### US-08 — Shared interaction contract (P1)

- AC-23: Every changed interactive control has an accessible name, keyboard operation and visible
  focus; status never relies on colour alone.
- AC-24: Mobile changed pages have no horizontal page scroll, 44 px primary targets and sheets for
  row actions; text scaling to 200 % does not clip actions.
- AC-25: No new timer, listener or subscription leaks after unmount; one shared clock/interval per
  concern.

## Requirements and success

- FR-01: Display freshness separately from availability; never feed silence into `deviceStatus`,
  `canControl` or automation routing. Acceptance: AC-01, AC-02.
- FR-02: Track snapshot age in hub state; mark stale shells. Acceptance: AC-03.
- FR-03: Persisted notification inbox with dedupe, unread count, digest. Acceptance: AC-05–08.
- FR-04: Logs page in Settings with archive download and a global quick-access entry. Acceptance:
  AC-09–11.
- FR-05: Calm Overview (A). Acceptance: AC-12–14.
- FR-06: Freshness in Devices and device detail. Acceptance: AC-15–16.
- FR-07: Panel day/night/offline/status. Acceptance: AC-17–19.
- FR-08: Additive Settings index; available cleanup/export; automation menu. Acceptance: AC-20–22.
- FR-09: Shared primitives and mobile rules. Acceptance: AC-23–25.
- NFR-01: One 30 s shared clock; no per-card timers; log buffer bounded; listeners disposed.
- NFR-02: No new network calls on render paths; archive fetch only on user action.
- NFR-03: No secrets/household data in localStorage beyond notification text, read state and last
  visit timestamp; logs never persisted without an explicit user download.
- SC-01: A disconnected shell shows a snapshot age within one render; a failed operation is
  retrievable after reload; Overview first viewport contains ≤ 4 blocks.

## Entities and external interfaces

- Device freshness is derived client-side from `Device.last_seen`, power source and `now`; it is not
  a wire contract.
- Hub state gains `lastSyncedAt` (client-only). No backend schema change.
- Notification persistence: browser `localStorage`, scoped by hub and user; bounded count and age.
- Logs: existing `GET /api/listlogs` and `POST /api/logfile` (protected, JSON `{file, action:"load"}`)
  consumed by a new frontend service; downloads use a released object URL.
- Automation export uses a client-side versioned envelope; import validation is not added in this
  branch.

## Questions and assumptions

- Q1: Freshness thresholds. Assumption: mains 2 h quiet / 6 h stale; battery 24 h / 48 h. Resolved
  for this implementation; a per-device expected interval is future work.
- Q2: Panel night trigger. Decision: schedule baseline 22:30–06:30, optional ambient-light source +
  threshold, and a persistent manual choice for the current night cycle. The four essentials are
  selectable from panel device controls and stored per browser.
- Q3: Offline commands. Decision: disabled, never queued (stale intent is unsafe).
- Q4: Server-side purge. Resolved with an explicit RFC3339 cutoff, authenticated/rate-limited
  endpoint, audit log and informed confirmation; log deletion is one named file at a time.
