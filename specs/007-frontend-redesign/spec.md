# Spec: Frontend redesign and customisable design system

## Mobile entity repair — 2026-10-18

Reproduction: with delayed history, read-only dialogs open as sheets then expand;
empty presence/illuminance stay sheets. Slider native-thumb travel differs from the
painted handle; pending state dims it. Energy has no requested meter visual.

- AC-51: Every entity opens full-screen on mobile from the first rendered frame,
  independent of history/loading/error/offline state. Content scrolls inside a stable
  viewport; desktop sizing depends on entity type, never an asynchronous response.
- AC-52: Brightness/temperature dragging follows the finger continuously, with a
  stable handle/fill and no layout jump on pending/confirmation. Native steps and
  keyboard accessibility remain; release sends once and cancellation sends nothing.
- AC-53: Energy shows a compact semicircular meter reading visual alongside separate
  interval consumption history. No invented maximum, percentage, or safety bands:
  absent real metadata bounds, the arc is decorative and the exact reading is labelled.
- AC-54: Environmental, state/event, alert, diagnostic and writable families share
  the same stable mobile shell, concise recent history and capability-driven controls.
  Missing history and zero consumption remain distinct. No device commands in audits.

### Visual completeness follow-up — 2026-10-18

The owner requests the originally agreed multicolour meter and any remaining
omissions. AC-55: meter arc has a green/yellow/coral decorative palette (no threshold
scale); known numeric charts show a final-sample marker and fine dashed guides;
air quality is teal and action events purple. Contact open/closed and alert/dry
colours must match their labels, with unknown visually distinct. Battery diagnostics
show the known percentage bar and badge; missing readings never become zero.
Entity choices with at most five metadata options use labelled direct buttons;
longer lists retain the shared select. Requested settings remain distinct from
confirmed state. Verify both themes and desktop/mobile against isolated mocks.

### Drag stability repair — 2026-10-19

Owner reproduction: brightness and colour-temperature handles jump backward/forward
while dragging. The old BaseSlider used native range/v-model and emitted only change;
the replacement adds a competing pointer-value writer. Unconfirmed live reports can
also clear a released draft while its command is still pending.

AC-56: one native drag owner, local preview only while dragging, and one native-unit
command on release. Unconfirmed reports must not reset an active/pending draft;
confirmation, refusal, disconnect and timeout keep their established semantics.
Brightness and mired/Kelvin temperature need mouse/touch/keyboard/cancel coverage,
including live reports during drag and while awaiting confirmation. No arbitrary
animation delay, continuous MQTT writes, or device-state optimism.

## Implementation amendment — 2026-10-17

Owner instruction: implement the accepted entity/chart designs, clean touched CRAP
scores, test, review and commit. This supersedes the investigation-only status below.
Scope is entity controls, compact/full history and necessary producer fixes; unrelated
old follow-up tasks remain deferred.

- FR-20 / AC-44: Mobile entity details fill the viewport. Writable exposes use large
  labelled controls; brightness/colour temperature match v3, read-only charts match
  environmental v5, other families match activity/diagnostic v4. Desktop adapts the
  same content. Verify at 320/390/768/1280/1920 px and short landscape.
- FR-21 / AC-45: Native bounds/steps and units determine commands. A completed gesture
  sends once; cancellation sends nothing. Drafts never replace confirmed state and
  revert on refusal/timeout/disconnect. Closing does not retry/cancel sent commands.
- FR-22 / AC-46: Full and compact history share validated typed data, calculations,
  theme and query ownership. Empty, error, partial, unknown and measured zero differ.
  Queries use explicit UTC instants, displayed local timezone, inclusive endpoints,
  10s timeout, bounded queue/cache, and late-response rejection.
- FR-23 / AC-47: Add optional HTTP `history: true` for bounded raw numeric/binary/enum
  event history. Existing absent/false requests retain aggregation/interval responses.
  Raw enums retain equal repeated reports; state transitions count only actual changes.
  Public route access is unchanged. New option permits one device and up to 32 exposes,
  at most 31 days and 10,001 returned samples per series; scan budget is 250,000 records
  per request. Budget exhaustion is an explicit error, never a successful empty result.
- FR-24 / AC-48: Energy consumption bars use nonnegative cumulative counter deltas
  only between consecutive valid samples inside one displayed bucket. Missing boundary
  baseline, reset or explicit gap makes that bucket partial/unknown. Do not allocate
  deltas across buckets or integrate power. Label partial consumption and retain meter
  total separately; no full-day total claimed without complete interval coverage.
- NFR-09 / AC-49: At most one active client history request, 32 pending distinct queries,
  10 cached entries and 10,000 retained samples per series; render at most 1,000 numeric
  points preserving extrema/gaps. No hidden-tab polling. At 50 open/close cycles there
  are no retained view-owned chart instances, handlers or requests after settlement.
- AC-50: Touched functions meet the documented complexity/CRAP gate, with measured
  coverage where supported and conservative zero-coverage upper bounds otherwise.
  No exclusions of Vue script functions. Every new/changed function has C<=10 and
  CRAP<=30; touched decision logic gets >=80% branch evidence or is refactored until
  conservative evidence plus behavioral integration tests establishes the gate.

No device ingestion/automation timing, storage schema or retention changes. No
automatic retries for commands or history. Read failures return useful errors;
existing query consumers keep compatible success shapes. Tests use mock hub/temp DB.

## Accepted entity/chart direction — 2026-10-17

The accepted direction covers brightness/colour temperature controls, environmental
history, activity/electricity, and diagnostics/settings. Initial concept images
were removed after implementation at the owner’s request; git history retains them.
Mobile entity views fill the screen; adjustable entities prioritise large controls
and minimal recent history. Read-only entities can show larger charts and condensed
useful summaries. All charts, including full device history, share truthful semantics
and the accepted visual language. Desktop adapts the same behavior to available space.

Implementation was subsequently authorised and completed. The
[implementation evidence](implementation-evidence-2026-10-18.md) records final
ownership, contract decisions, and verification. Existing acceptance IDs and
completed history remain.

### Corrective pass — 2026-10-18

Owner review of the completed views found implementation defects, not new
requirements. Reproduction against the isolated mock: an offline/missing adjustable
reading shows a large `—` and a minimum-position rail that reads as a handle above
the slider; the grip is offset from the fill boundary; the slider receives a focus
outline on open; power/alarm controls lack visible state or labels; colour temperature
lacks its warm/cool information; desktop keeps the compact chart; counter history
with no complete interval draws empty axes without explanation; raw boolean history
renders as `Unknown` beside a confirmed `ON` state; and read-only views with no samples
leave an oversized blank column. Expected behavior is unchanged: FR-20/21/22/24 require
explicit unknown states, truthful history/energy coverage, labelled accessible controls
and the accepted responsive composition, including consistent binary state words across
live and recorded values and compact device-row sliders with the same direction as the
full entity control. The
[plan](plan.md#entity-view-corrective-pass--2026-10-18) and [tasks](tasks.md#entity-view-corrective-pass-2026-10-18)
record the corrections and verification; no contract or device-semantics change.

ID: 007-frontend-redesign | Status: Main redesign merged; follow-up design proposed | Author: Claude Code / Codex | Updated: 2026-10-11
Constitutions: parent/frontend/backend 1.0.0, proposed, not ratified. Components: frontend;
original additive backend field (`DashboardGroup.order`, FR-09), with scoped backend query/
device-pipeline investigation and any demonstrated fixes in the 2026-10-11 amendment.

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

Historical acceptance: the owner reviewed the prototype gallery before implementation.
The prototypes were removed on 2026-10-12 after the main redesign was completed;
AC-01/02 below preserve the original design-selection requirements.

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
  (tile size, area icon) is a contract change and needs a new spec; `order` is specified in FR-09.
- No HTTP/WS/MQTT/MCP changes. Consumers of existing contracts listed in
  [current-state §4](current-state.md#4-transport-contracts-used-by-the-browser).

## Questions and assumptions

| ID | Question | Consequence | Owner | Status |
| --- | --- | --- | --- | --- |
| Q-01 | Which sample (or mix) becomes the target design? | Blocks implementation phases 2+ | Owner | **Resolved 2026-10-10:** owner chose the Hearth + Panel prototype and authorised implementation |
| Q-02 | Are multiple actions per trigger supported by the backend engine? | Determines whether editor allows N actions | Maintainer | Code evidence: `backend/internal/automations/trigger.go` executes every action in order and joins errors; editor may allow N actions in order. Confirm intent. |
| Q-03 | Should design/theme be per-browser only, or synced via appConfig? | Contract change if synced | Owner | Assumed per-browser |
| Q-04 | May dashboard groups gain optional fields (order, icon, tile size)? | Backend contract change | Owner/maintainer | **Resolved 2026-10-10 for `order` only:** owner asked for drag-to-reorder areas; an optional integer `order` is added (FR-09). Icon/tile size remain out of scope. |
| Q-05 | Pending-command timeout value | Affects AC-04 | Owner | Assumed 5 s |
| Q-06 | Is a "Lovelace-style" freely arranged card grid wanted, or areas + tiles only? | Scope of dashboard editor | Owner | Assumed areas + tiles |

Assumptions: samples are static HTML (no build) to keep review cheap; mock data mirrors
real contract shapes; "retro" means typography, palette and chrome, not reduced
usability.

## Amendment 2026-10-10 — implementation of Hearth + Panel

Owner decision (Q-01): implement **Hearth + Panel** in `frontend/`, responsive on desktop and
mobile; automations, panels (panel mode and dashboards) and the sign-in page must keep working;
add drag-to-reorder for Home areas; build tests first. IDs above remain valid; new IDs below.

### US-06 — Arrange Home areas (P1)

Independent verification: Playwright against the mock hub; Go unit test for the field round-trip.

- AC-19: Given Home in edit mode, when the user drags an area onto another position (mouse or
  touch), then the areas render in the new order and `saveDashboardGroup` is sent for every
  group whose `order` changed, with the new `order` values.
- AC-20: Given the same, when the user uses the area's "Move earlier/later" buttons with the
  keyboard, then the result is identical to dragging (FE-04 keyboard alternative).
- AC-21: Given groups without `order` (existing data), when Home renders, then groups appear
  in name order and the first reorder assigns `order` 0…n-1; given an older backend that drops
  the field, reorder still works for the session and no error is shown.
- AC-22: Given a reload after reordering, when Home renders, then the saved order is used.

### US-07 — Panel mode (P1)

- AC-23: Given any page, when the user opens Panel mode, then a full-screen room view shows a
  clock, one room per page (pager, arrow keys), a climate dial when the room has a temperature,
  and large buttons that toggle with the same pending/confirmed behaviour as AC-04.
- AC-24: Given panel mode, when the user selects a read-only sensor, then the normal device page
  opens; "Exit panel" returns to Home. Panel mode never contains an automation editor.

### US-08 — Sign in (P1)

- AC-25: Given an unauthenticated visitor, when any protected route is opened, then the sign-in
  page shows the logo and a "Sign in with Google" button; pressing it calls `POST /api/auth/login`
  and redirects to the returned URL (unchanged contract); while pending the button is disabled
  with a "Signing in…" label; a failure shows a plain-language error with a retry.
- AC-26: Given a return to `/?auth=success`, when the session is valid, then the user lands on
  Home (or the `redirect` query target) and the URL parameter is removed.

### US-09 — Recent activity switch (P2)

- AC-27: Given Overview, when the user turns Recent activity off, then the feed is hidden, its
  buffer is cleared and no new entries are collected; the preference is stored per browser and
  restored on reload.

### Requirements

- FR-09: MUST support reordering dashboard groups by drag and drop and by keyboard; MUST
  persist order in an optional integer `order` field on `DashboardGroup` (absent = unordered,
  sorted by name after ordered groups); AC-19…AC-22.
- FR-10: MUST provide panel mode at `/panel` sharing the device-command path of the main UI;
  AC-23, AC-24.
- FR-11: MUST keep the OAuth redirect sign-in contract (`/api/auth/login`, `/api/auth/me`,
  `?auth=success`); AC-25, AC-26.
- FR-12: MUST provide the Recent activity switch; AC-27.
- FR-13: MUST show the NodeHerder logo in the sidebar, the mobile header, the sign-in page and
  panel mode.
- NFR-06: Critical flows (sign-in, Home toggle, area reorder, automation open/edit/save, panel
  toggle, device page) MUST have automated browser tests at 1280×800 and 390×844 that run in CI
  against the mock hub, never real devices (NH-05).

## Amendment 2026-10-11 — entities, complete history redesign and code quality

The main redesign merged in #54 (`fcdebca2`). The owner now requests entity rework,
a complete replacement of history-chart UX, and CRAP-score remediation, refactoring,
appropriate patterns, SOLID responsibilities and a single source of truth for everything
touched. This turn authorises updated planning artifacts; it does not start implementation.
Earlier decisions and completed evidence remain history. This amendment supersedes the
original exclusion of metrics work for the paths needed by these outcomes.

Scope includes Home entities, device cards/expose rows/controls, panel entities, the entity
detail/history dialog, full device history and mini charts; shared formatting, capabilities,
query state, chart transformations and renderer configuration; the review's reliability,
contrast, editor and remaining-screen gaps. Backend query/error fixes are included only
where required to make history correct. No storage migration, retention change, device
protocol change, new charting/state framework or unrelated repository-wide cleanup.

### US-10 — Consistent entities (P1)

Independent verification: fixtures for every expose capability and mocked browser flows.

- AC-28: Given the same entity on Home, Devices, the device page, Panel and its detail
  dialog, each uses the same identity, label, metadata unit, binary meaning, availability,
  disabled state and confirmed value. Unknown/removed exposes show an explicit fallback.
- AC-29: Given a writable entity, all entry points use the same command lifecycle:
  disabled with a reason when unavailable/disconnected, one outstanding command per expose,
  pending distinct from confirmed, confirmation or 5-second timeout, rollback and an
  announced error. Numeric zero remains controllable; missing data never becomes zero.
- AC-30: Given entity details at 360/768/1280/1920px, current reading, available controls,
  history, status and device navigation form one coherent layout with labelled keyboard
  controls, focus containment/return and no horizontal page scrolling. History remains
  available for chartable entities even when a brightness control also exists.

### US-11 — Explore trustworthy history (P1)

Independent verification: pure history tests, backend contract fixtures and browser tests.

- AC-31: Given numeric history, users can select an entity, preset/custom range and an
  explicitly labelled display timezone (browser zone or UTC), inspect exact timestamp/value/
  unit, zoom and reset. The history layout and interactions are rebuilt, not merely recoloured.
  Mini charts are compact views of the same history model, with an entry to full history.
- AC-32: Given binary/state history, a proportional timeline shows actual state labels
  and unknown periods; repeated stored events remain visible/countable. Event counts and
  state-transition counts are labelled separately. Enum/event entities use an event list
  when supported; unsupported series explain the limitation instead of disappearing.
- AC-33: Given empty, missing, zero, malformed, sparse, duplicate-timestamp, out-of-order
  or truncated samples, charts distinguish these cases. They never invent a leading binary
  state, join across explicit missing samples or claim completeness for limited results.
  Stats derive from validated source samples, not display decimation; averages state whether
  sample-based, and unavailable stats/deltas are shown as unavailable.
- AC-34: Given rapid range/entity changes or multiple history views, only the matching
  request updates that query's state. Loading lasts until response/error/timeout; stale
  responses, cancellation and closing a view cannot overwrite a newer selection.
- AC-35: Given an offline hub, rate limit or failed query, history shows useful feedback
  and explicit retry without confusing failure with no measurements. Optional auto-refresh
  is read-only, pauses while hidden/offline, has bounded resources and cleans up on close.
  Live device state and stored history remain distinguishable.
- AC-36: Given light/dark themes, dense data and touch/keyboard use, full/mini charts have
  readable axes/units, accessible names, visible focus, text summaries and a keyboard-usable
  sample/event table. Tooltip content is text or deterministically escaped untrusted input.

### US-12 — Reliable controls and maintainable implementation (P1)

- AC-37: Given disconnected or failed automation/layout persistence, edits remain recoverable
  and are not presented as saved; manual Run has a shared in-flight lock with no retries.
  Invalid drafts cannot save; daily schedules validate times and label the actual scheduler
  zone (currently Europe/London), with the specified timeline.
- AC-38: Given each changed/new production function, before/after complexity, coverage and
  CRAP evidence is recorded. New/changed functions meet NFR-07; removed functions identify
  their replacements. Equivalent behaviour and meaningful failures remain tested.
- AC-39: Given the finished follow-up, the Assistant and remaining dialogs match the shared
  UI; required viewport, keyboard, screen-reader, reconnect, contrast and load evidence is
  recorded accurately, distinguishing passed, failed and unrun checks.

### US-13 — Timely illuminance readings (P1 investigation)

Reported device: **Living room presence sensor**, expose **illuminance**. After changing
light brightness the UI eventually changes, but appears delayed; elapsed time and the
light's identity are not yet measured. This is a reported latency symptom, not a confirmed
root cause or permission to command household devices during routine tests.

- AC-40: Given a controlled illuminance-update replay, record timestamps at raw ingress,
  accepted backend state, broadcast, browser receipt, store mutation and rendered value.
  Separate sensor reporting cadence from backend sampling and UI latency. Replays cover
  `illuminance` and `illuminance_lux`, rapid changes, a final change followed by silence,
  zero, unchanged reports, offline/reconnect and effective per-device debounce settings.
- AC-41: Given a diagnosed defect, add a failing regression before fixing it and record
  the timing contract. Proposed browser-receipt-to-render target: ≤ 250ms at p95 over 20
  controlled updates on a recorded browser. If a backend sampling change is needed, specify
  a bounded latest-value delivery deadline and automation/metrics implications before
  implementation. Do not manufacture illuminance from a brightness command or assume
  illuminance equals lux without metadata evidence. If delay is solely sensor reporting,
  record that result and provide truthful last-reported/stale feedback instead of a fake fix.

### US-14 — Reliable brightness slider and explainable automation activity (P1)

The owner also reports trouble changing brightness through the UI slider while the dial
automation works. The affected light, gesture and observed result are not yet identified.
Logs supplied for 2026-10-11 19:33:19…19:33:36 show repeated `triggering automation for
changed exposes` on device `0xa4c13894070052fc`. Separate Attic air sensor metrics logs
are not evidence that this device is the faulty light or that actions ran repeatedly.

- AC-42: Given brightness slider operations by mouse, touch and keyboard, compare the UI
  `deviceSetValue` payload, resolved target/expose, numeric type/range and MQTT command with
  a working dial-automation fixture. Verify raw brightness vs displayed percentage, min/max/
  step, release/commit semantics, exactly one command per committed gesture, pending/refused
  commands, echoed/different/no confirmation, final-value rollback and ability to set zero.
  Add a failing regression for any demonstrated defect before fixing it.
- AC-43: Given the reported device's synthetic update stream, correlate accepted reports,
  changed expose names, automation dispatches, conditions matched, actions actually executed
  and outgoing/echoed commands. Determine whether repetition is normal sensor cadence,
  duplicate delivery, unrelated-expose dispatch or a feedback loop. Debug dispatch logs must
  not be presented as proof of action execution. Preserve repeated physical events and
  ordered state; any loop/timing/logging change needs its own verified acceptance evidence
  and must avoid persisting household values or adding noisy per-report logging.

### Requirements and measurable success

- FR-14: MUST provide one shared entity presentation/capability model and one device-command
  path across all touched entity surfaces; AC-28…30.
- FR-15: MUST rebuild full and mini history experiences with a shared typed history model,
  correct semantics and accessible inspection; AC-31…36.
- FR-16: MUST isolate history query identity/status and validate external samples before
  state mutation; preserve existing consumers and document actual backend/mock contracts;
  AC-33…35.
- FR-17: MUST fix the review's persistence, manual-run, numeric rollback, schedule and
  contrast gaps and finish remaining UI verification/layout work; AC-29, AC-36…39.
- FR-18: MUST investigate delayed Living room presence sensor illuminance and remediate
  any demonstrated application defect under an explicit timing contract; AC-40, AC-41.
- FR-19: MUST investigate UI brightness-slider behaviour and reported repeated automation
  dispatches, fixing demonstrated application defects without breaking the working dial
  flow or suppressing legitimate repeated events; AC-42, AC-43.
- NFR-07: Proposed gate: each new/changed production function has CRAP ≤ 30 and cyclomatic
  complexity ≤ 10, using pinned tooling and a documented per-function coverage mapping.
  High-score touched functions are remediated before merge; function scores cannot be hidden
  by file averages, renaming, extraction into unmeasured helpers or coverage exclusions.
  Touched decision logic additionally has ≥ 80% branch coverage and meaningful failure tests.
  If a function cannot meet the gate, record a blocking scoped debt item; do not self-approve
  an exception. Thresholds are planning assumptions, not existing repository policy.
- NFR-08: Proposed history budgets: one concurrent network query per client; user-selection
  debounce 300ms; timeout 10s; optional refresh no faster than every 10s; ≤ 10 cached query
  entries, evicted on logout; ≤ 10,000 validated samples per series, with a 10,001st sample
  detecting truncation; ≤ 1,000 rendered numeric points per series preserving extrema/gaps.
  For a 10,000-point fixture, query normalization + stats + decimation takes ≤ 100ms at p95
  over 20 warmed runs on a recorded runner; report machine/browser and raw timings.
  These are client budgets, not claims of bounded backend storage scans.
- SC-03: AC-28…43 and NFR-07…08 have recorded evidence, alongside remaining original gates,
  before this follow-up is called complete.

### Interfaces and assumptions for this amendment

- Entity identity remains `(deviceId, exposeName)`; Vuex remains confirmed device truth.
  No entity/chart household data is persisted in browser storage or logs.
- Preferred history transport reuses existing `POST /api/metrics/query` through the API
  helper, keeping legacy `loadMetrics` consumers compatible. See
  [ADR-004](adr-004-history-and-quality.md) and [design](entities-and-history.md).
- Proposed defaults: retain Hearth + Panel and ApexCharts; history presets 1h/6h/24h/7d/30d,
  Today/Yesterday plus custom ranges; browser timezone default, UTC option; sample mean,
  no implicit interpolation across gaps. Entities means all surfaces listed above.
- No new backend aggregation, pagination, persisted layout schema, timezone configuration
  or export/comparison features are assumed. Any need discovered for these updates the
  contract/spec before dependent code. Query failure-vs-empty and range-boundary fixtures
  are required preparation gates; their implementation evidence is pending.


### Device-update repair contract (2026-10-12)

The owner has requested implementation of the investigated issues. For AC-41/42,
accepted changed exposes and repeated events reach the existing live-update callback
before the ordered Update call returns, independent of metrics sampling. A final report
followed by silence therefore does not need a later report or trailing timer. Writable
same-value device reports are valid live confirmation evidence; they do not retrigger
state automations or create duplicate sampled metrics. Read-only unchanged state remains
deduplicated. Outgoing commands never count as confirmation. Existing measurement
sampling intervals/overrides, command targets/payloads, seed/availability behavior and
fresh ordered automation inputs remain compatible. Network/render deadlines are separate
from this same-call processing contract. See [ADR-005](adr-005-live-device-updates.md).
