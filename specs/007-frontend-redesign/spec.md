# Spec: Frontend redesign and customisable design system

ID: 007-frontend-redesign | Status: Draft for review | Author: Claude Code | Updated: 2026-10-09
Constitutions: parent/frontend 1.0.0, proposed, not ratified. Components: frontend
(backend contracts referenced read-only; no backend change in scope).

## Problem, scope, exclusions

The household owner reports that the whole UI needs a redesign to be "smart,
production-ready, retro and customisable", similar in spirit (not a copy) to the
Home Assistant Lovelace dashboard and the Zigbee2MQTT frontend. The automation
editor is singled out as complicated. The request authorises **documentation,
planning and reviewable design samples first**; production-code changes wait for the
owner to choose a direction (Q-01).

Evidence (see [current-state.md](current-state.md) and the five screenshots supplied
on 2026-10-09):

- Visual language is inconsistent: PrimeVue Material preset plus hard-coded colours in
  components; only 7 global CSS tokens; dark-mode toggle is the only customisation.
- Group dashboard tiles are a fixed two-column masonry ≤ 400px per group; device cards
  waste space and hide status (LQI, battery, protocol) in footers.
- No overview page comparable to Zigbee2MQTT (counts, routers/end devices, low LQI,
  recent activity); no search/filter on device list.
- Automation editor uses nested panels with no "when → if → then" overview, one action
  per trigger, unclear validation and inverted button enablement.
- Protocol support is implicit (`connection_type` `mqtt`/`http`); Wi-Fi/HTTP devices
  look identical to Zigbee devices except a logo on the About tab.

In scope: every routed screen, layout shell, dialogs, notifications, permit join,
login, theming/customisation, device protocol presentation, automation editor UX.

Exclusions: backend behaviour or wire-contract changes; new device protocols in the
backend (presentation must be ready for them, FR-07); replacing Vue/Vuex/Router/PrimeVue
(would require an ADR under FE-01); fixing the pre-existing gaps listed in
[current-state §4](current-state.md#4-transport-contracts-used-by-the-browser) except
where a redesigned screen touches them (then they become tasks, see plan).

## Scenarios and acceptance

### US-01 — Choose a design direction (P1)

Independent verification: open `samples/index.html` from a checkout without a build.

- AC-01: Given the branch is checked out, when the owner opens the sample gallery,
  then 5 distinct designs are available, each rendering every screen in FR-01 with
  representative mock data and no network access.
- AC-02: Given a sample, when the owner changes theme mode, accent, density, corner
  radius or font scale in its customiser, then the whole sample updates immediately
  and the choice survives a reload of that sample.

### US-02 — Glanceable, customisable home (P1)

Independent verification: manual procedure on the implemented dashboard with mocked WS.

- AC-03: Given dashboard groups exist, when Home opens, then each group renders as an
  "area" with tiles showing name, value with unit, icon and state, at 360px and 1440px
  widths without horizontal scrolling.
- AC-04: Given a tile for a writable binary expose, when the user activates it by click
  or keyboard, then a pending indicator shows until `deviceUpdated` confirms the new
  value; if none arrives within the timeout the tile reverts and an error is announced
  (NH-02, FE-03).
- AC-05: Given a device is offline or disabled, when its tile renders, then the state is
  conveyed by icon + text, not colour alone (FE-04).
- AC-06: Given edit mode, when the user adds/renames/reorders/removes groups or tiles, then
  changes are saved through existing `saveDashboardGroup`/`renameDashboardGroup`/
  `deleteDashboardGroup` events and destructive steps ask for confirmation (NH-06).

### US-03 — Operate devices regardless of protocol (P1)

- AC-07: Given devices with `connection_type` `mqtt` and `http`, when the device list
  or overview renders, then each shows a protocol badge (e.g. "Zigbee", "Wi-Fi/HTTP")
  and protocol-relevant diagnostics (LQI for Zigbee; last-seen/IP-style metadata for
  HTTP when present), and filtering by protocol, power source and availability works.
- AC-08: Given an unknown `connection_type` value, when it renders, then a generic badge
  with the raw value appears and no screen fails.
- AC-09: Given the overview, when it opens, then totals (devices, online/offline,
  routers or mains, battery devices, low LQI < 50, low battery < 20%) and the last N
  device updates are visible.

### US-04 — Understandable automation editor (P1)

- AC-10: Given an automation, when the editor opens, then each trigger is displayed as a
  readable sentence and a "When / If / Then" block flow on one page, without nested
  panel navigation.
- AC-11: Given an incomplete trigger (no source expose, no action target, empty value),
  when the user attempts Save, then Save is disabled and inline messages name every
  missing field; a valid automation enables Save.
- AC-11b: Given a device trigger with no conditions, when shown, then the editor warns that the backend blocks device-triggered execution without conditions (`trigger.go`: "Blocking device-triggered automation without conditions") and that it can only run manually.
- AC-12: Given unsaved edits, when the user navigates away, then they are warned and may
  discard or stay.
- AC-13: Given a trigger without expose conditions, when the user presses Run, then
  `POST /api/automation/trigger` is called once and the result is shown; repeated
  presses while pending do not send duplicates (FE-03).
- AC-14: Given an existing automation saved by the current editor, when loaded and saved
  unchanged in the new editor, then the emitted `saveAutomation` payload is deeply
  equal to the original (contract compatibility, FE-02).
- AC-15: Given schedules, when edited, then enable/disable times are shown on a 24h
  timeline with the hub's timezone label.

### US-05 — Consistent, accessible shell (P2)

- AC-16: Given any screen, when navigated by keyboard only, then all controls are
  reachable with visible focus and accessible names (FE-04).
- AC-17: Given WS state `connecting` or `disconnected`, when shown, then a persistent
  status indicator with text appears and device controls are disabled with a reason.
- AC-18: Given a customised theme, when the app reloads, then the theme is restored from
  browser storage that contains only presentation preferences (FE-05).

## Requirements and success

- FR-01: MUST redesign: shell (nav, header, status), login, overview, home/groups,
  devices dashboard, device list, device page (about/exposes/settings/metrics), automation
  list, automation editor (triggers, conditions, actions, schedules), assistant,
  console, settings, permit join, dialogs, notifications; acceptance AC-01, AC-03…AC-17.
- FR-02: MUST provide a token-based theme system (colour roles, typography, radius,
  density, elevation, motion) with light/dark modes and user-selectable accent and
  presets; AC-02, AC-18.
- FR-03: MUST keep Vuex as device truth and existing WS/HTTP contracts; AC-14.
- FR-04: MUST distinguish requested vs confirmed device state on every control; AC-04.
- FR-05: MUST provide an overview with network health and recent activity; AC-09.
- FR-06: MUST replace the automation panel stack with a single-page flow editor that
  supports all existing trigger/condition/action/schedule types; AC-10…AC-15.
- FR-07: MUST resolve protocol presentation through a registry keyed by
  `connection_type`, with a generic fallback, so future protocols (e.g. ESPHome native,
  Matter/Thread, BLE) need only a registry entry in the UI; AC-07, AC-08.
- FR-08: MUST resolve expose presentation (icon, label, formatter, control) through a
  capability registry derived from expose `type`/`access_mode`/`values`/`attributes`,
  not device model names; AC-07.
- NFR-01: Layout MUST work at 360, 768, 1280 and 1920px wide without horizontal page
  scroll; verified manually per release.
- NFR-02: Text/background colour pairs in every shipped preset MUST meet WCAG 2.1 AA
  (4.5:1 body text, 3:1 large text and UI component boundaries).
- NFR-03: Production bundle growth from the redesign MUST be ≤ 15% gzip versus the
  pre-redesign `npm run build` output, unless the plan records justification.
- NFR-04: A dashboard with 100 tiles MUST apply a 10 msg/s `deviceUpdated` stream without
  dropped updates (state reaches the store unthrottled; NH-02).
- SC-01: Owner selects one design (or a combination) recorded in Q-01.
- SC-02: All AC pass with recorded evidence in `plan.md` before merge of each phase.

## Entities and external interfaces

- Theme preferences: `{design, mode, accent, density, radius, fontScale, preset}` in
  `localStorage` key `nodeherder_theme` (presentation only, versioned; replaces
  `localStorage.theme`, which is migrated once).
- Dashboard layout preferences that must sync across browsers stay in
  `appConfig.hub.dashboardGroups` (existing contract). Any new persisted layout field
  (tile size, order, area icon) is a contract change and needs Q-04.
- No HTTP/WS/MQTT/MCP changes. Consumers of existing contracts listed in
  [current-state §4](current-state.md#4-transport-contracts-used-by-the-browser).

## Questions and assumptions

| ID | Question | Consequence | Owner | Status |
| --- | --- | --- | --- | --- |
| Q-01 | Which sample (or mix) becomes the target design? | Blocks implementation phases 2+ | Owner | Open |
| Q-02 | Are multiple actions per trigger supported by the backend engine? | Determines whether editor allows N actions | Maintainer | Code evidence: `backend/internal/automations/trigger.go` executes every action in order and joins errors; editor may allow N actions in order. Confirm intent. |
| Q-03 | Should design/theme be per-browser only, or synced via appConfig? | Contract change if synced | Owner | Assumed per-browser |
| Q-04 | May dashboard groups gain optional fields (order, icon, tile size)? | Backend contract change, separate spec | Owner/maintainer | Open; out of scope here |
| Q-05 | Pending-command timeout value | Affects AC-04 | Owner | Assumed 5 s |
| Q-06 | Is a "Lovelace-style" freely arranged card grid wanted, or areas + tiles only? | Scope of dashboard editor | Owner | Assumed areas + tiles |

Assumptions: samples are static HTML (no build) to keep review cheap; mock data mirrors
real contract shapes; "retro" means typography, palette and chrome, not reduced
usability.
