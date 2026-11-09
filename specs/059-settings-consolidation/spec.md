# Spec: Consolidated, consistent settings

ID: 059 | Status: Draft | Owner: agent draft | Updated: 2026-11-08
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

The settings page grew in two eras. `Settings.vue` and the Device Defaults section were
rewritten for spec 007; History, Logger, MCP Server and Assistant are untouched legacy
views that still use PrimeFlex row grids, PrimeVue `Tag`/`Select`, and raw configuration
keys as user-visible labels (`sleepTimeout`, `enableRemoteLogger`, `url`). The host wraps
every section body in one `UiCard`, so Device Defaults renders a card inside a card and
two `h2`s with different meanings. Section ids are derived from display titles, so a
rename silently breaks `?section=` links.

Facts observed in the current tree:

- `Settings.vue` builds `SECTIONS` from `settingsTabComponents` (title slug ids) and
  renders `<UiCard :title="current.title">` around each section. `DeviceConfigurationSettings`
  and `SensorReportingSettings` declare their own bordered panels, so borders nest.
- History and System show overlapping retention data: History edits
  `hub.history.{sleepTimeout,expireAt}`; System reads `GET /system/storage`, whose
  `metricsRetentionDays`/`positionsRetentionDays` duplicate the same policy read-only.
- `HistorySettings.vue` and `AssistantSettings.vue` mutate the object returned by a Vuex
  getter before dispatching (`historySettings.value[key].value = value`,
  `assistantSettings.value[key] = value`), which writes store state outside a mutation.
- `MCPSettings.vue` awaits `setTimeout(1000)` after start/stop/restart, has no pending
  state, and passes an action that returns `undefined` to `ActionButton`, so
  `handleResponse` always shows success.
- `SystemSettings.vue` implements its own B/KB/MB/GB formatter, duplicating a
  responsibility that `.agents/rules/vue-engineer.md` assigns to `src/utils/units.ts`;
  `units.ts` has no byte quantity, which is why the rule was not followed.
- `base.css` defines `.nh-kv`, a labelled key/value grid, used by zero settings screens.
- `docs/hub_state.json` (the browser-test backend fixture) stores the history cleanup
  interval under the typo `sleepTmeout`, and its `logger` object has no `level`, so the
  dev/E2E fixture cannot exercise those fields.
- No Jest test covers these settings components; `SensorReportingSettings.spec.ts` is the
  only functional settings browser suite and covers `?section=device-defaults` only.

Scope: the settings host, its section navigation, and the five hub sections
(Device defaults, Storage & retention, Logger, MCP server, Assistant), plus the Appearance
section's container. This is presentation plus the state/feedback corrections listed
below; no new backend endpoint, message type, or configuration field.

Exclusions: device-page settings tab behavior beyond reusing its component
(`DevicePage.vue` keeps its own tab bar), assistant chat behavior, log storage, and the
meaning of any persisted configuration value.

Assumptions: the hub remains the source of truth for all hub settings; the browser keeps
owning Appearance. No real household device is contacted by any check here.

## Scenarios and acceptance

### US-01 — Find a setting quickly (priority P1)

Independent verification: `e2e/Settings.spec.ts` and the domain unit test.

- AC-01: Given `/settings`, the rail shows two labelled groups, "This browser" and "Hub",
  and the Hub group lists Device defaults, Storage & retention, Logger, MCP server and
  Assistant in that order.
- AC-02: Given any section selected, the browser URL contains that section's stable id
  (`appearance`, `device-defaults`, `storage`, `logger`, `mcp`, `assistant`), and loading
  that URL directly selects the section and marks its rail link `aria-current="page"`.
- AC-03: Given the previous ids `history` or `system` in the URL, the Storage & retention
  section opens instead of falling back to Appearance.
- AC-04: Given an unknown section id, the first section opens and the page still renders
  without error.
- AC-05: Every rail entry is a link reachable with Tab, shows visible focus, and each
  group caption is associated with its link list for assistive technology.

### US-02 — Understand and change what the hub keeps (priority P1)

Independent verification: `e2e/StorageSettings.spec.ts`.

- AC-06: Given the Storage & retention section, a "Retention" card shows the cleanup
  interval and history retention as human labels with help text, and the current values
  from `hub/history` (falling back to the documented defaults when the hub omits a field).
- AC-07: Given a changed cleanup interval or retention value, the section dispatches
  `hub/saveHistorySettings` with a `HistoryConfig` whose other field is unchanged; the
  store getter object is not mutated in place.
- AC-08: Given a non-positive or non-numeric value, the field reports an inline error via
  `role="alert"` and no frame is sent.
- AC-09: Given the usage card, sizes come from `GET /system/storage` and are formatted by
  `units.ts`; loading, failure and empty-entry states are distinguishable, and Refresh
  re-reads without a page reload.
- AC-10: Given the hub socket is not connected, retention controls are disabled and an
  explanatory status message is shown; usage stays readable only if it already loaded.

### US-03 — Control the hub services without guessing (priority P2)

Independent verification: `e2e/LoggerSettings.spec.ts`, `e2e/MCPSettings.spec.ts`,
`e2e/AssistantSettings.spec.ts`.

- AC-11: Given the Logger section, "Send hub logs to this browser" and "Minimum level"
  use shared controls with help text, and changing either dispatches
  `hub/saveLoggerSettings` with the full `LoggerSettingsType`.
- AC-12: Given the MCP server section, a status chip distinguishes Disabled / Running /
  Stopped using token tones and text (never colour alone), and the identity rows show
  name, version and connected clients only when the server is enabled.
- AC-13: Given Start, Stop or Restart, the pressed control shows a pending state, the
  section requests `hub/loadMCPStatus` and clears pending when `hub/mcpStatus` changes or
  after a bounded wait; no fixed sleep is used to fake completion.
- AC-14: Given the Assistant section, the URL field has a label and help text, an invalid
  URL shows an inline error and sends nothing, and a valid change dispatches
  `hub/saveAssistantSettings` without mutating the getter object.
- AC-15: Given a section with inputs and the socket disconnected, hub controls are
  disabled and one status message explains why; Appearance stays editable because it is
  stored in this browser.

### US-04 — A readable page at any width (priority P2)

Independent verification: `e2e/layout.spec.ts` and per-section specs.

- AC-16: Given 1280 px or 390 px width, no settings section scrolls horizontally, cards
  do not nest inside other cards, and each section has exactly one `h2` for its own
  content.
- AC-16a: Given the Sensor adjustments and History sensor adjustments blocks, their
  headings sit inside the card like every other card heading — no `<legend>` straddling
  a card border — while each block keeps `role="group"` with its accessible name.
- AC-16b: Given any settings card or rail entry, its icon is chosen for that setting and
  does not reuse a glyph that already means an integration, a navigation destination or
  a device action elsewhere in the shell.
- AC-17: Given the rail at ≤860 px, it becomes a horizontally scrollable row, remains
  operable by keyboard, and the selected entry stays visible.
- AC-18: Given a section, status and error text is accompanied by text or an icon, and
  every control keeps a 44 px touch target.

## Requirements and success

- FR-01: MUST group the rail into "This browser" and "Hub" with stable, non-derived
  section ids and alias the retired `history`/`system` ids; acceptance: AC-01…AC-04.
- FR-02: MUST merge History and System into one Storage & retention section with an
  editable retention card above a read-only usage card; acceptance: AC-06…AC-10.
- FR-03: MUST rebuild Logger, MCP server, Assistant and the Device defaults section on
  shared primitives (`UiCard`, `UiField`, `UiToggle`, `UiSelect`, `UiChip`, `.nh-kv`,
  `units.ts`) with human labels and help text; acceptance: AC-11…AC-14, AC-16.
- FR-04: MUST save hub settings by dispatching complete settings objects built from the
  current values, never by mutating a getter result; acceptance: AC-07, AC-11, AC-14.
- FR-05: MUST replace the MCP fixed sleep with a status refresh, bounded pending state
  and cleanup on unmount; acceptance: AC-13.
- FR-06: MUST distinguish Missing/Invalid/Unavailable in text and expose disabled reasons;
  acceptance: AC-08, AC-10, AC-15, AC-18.
- NFR-01: MUST keep every existing WebSocket frame type and payload shape unchanged
  (`saveHistoryConfig`, `saveLoggerConfig`, `saveAssistantConfig`, `startMCP`, `stopMCP`,
  `restartMCP`, `loadMCPStatus`); acceptance: existing `commands.test.ts` stays green.
- NFR-02: MUST add no dependency and no new global CSS primitive; a new shared need is
  added once in `base.css` or a `src/components/ui` component.
- NFR-03: MUST keep the page's first paint cost unchanged in kind: sections stay lazily
  loaded async components, and no new polling or interval is introduced on mount.
- NFR-04: MUST format byte sizes through `src/utils/units.ts` so no section re-implements
  a unit table; acceptance: unit test for the new quantity.
- SC-01: Settings sections render one card per topic with a single `h2`, verified by the
  browser suites at 1280 px and 390 px.
- SC-02: Unit, lint, build and the affected browser suites pass; any unrun check is
  recorded in `plan.md` as unrun rather than implied.

## Entities and external interfaces

- `HistorySettingsType` (`sleepTimeout`, `expireAt`: `TimeInterval`) — unchanged shape and
  wire frame (`saveHistoryConfig`). Presentation adds labels and a unit choice; the value
  is an integer duration with a unit from `TimeUnits`.
- `LoggerSettingsType` (`enableRemoteLogger`, `level`) — unchanged. Level values stay the
  existing set (`trace`, `debug`, `info`, `warn`, `error`, `fatal`) that the backend
  `SetLogLevel` path already receives.
- `MCPStatusType` (`running`, `enabled`, `name`, `version`, `connectedClients`) and the
  `startMCP`/`stopMCP`/`restartMCP`/`loadMCPStatus` frames — unchanged. `enabled` is read
  from `hub/mcpConfig` as today; the mock's `mcpStatus` payload still omits it, so the
  section must not depend on it.
- `AssistantSettingsType` (`url`) — unchanged shape; `isValidHttpUrl` remains the only
  validation rule.
- `GET /system/storage` (`entries[].label/bytes/retentionDays`, `dataDirBytes`,
  `diskFreeBytes`, `diskTotalBytes`, `metricsRetentionDays`, `positionsRetentionDays`) —
  unchanged; only the browser formatting moves to `units.ts`.

## Questions and assumptions

- Q1: Should the backend expose a single "storage policy" endpoint that combines
  retention and usage? Consequence: removes the last duplicate read. Resolution: out of
  scope; the read-only usage card keeps reading `/system/storage` unchanged.
- Q2: Should cleanup interval accept seconds for testing? Consequence: a near-zero
  interval would busy-loop the cleanup goroutine. Resolution: the UI floor is one minute
  and the spec records it as validation, not a backend change.
- Assumption A1: `?section=device-defaults` must keep working because
  `SensorReportingSettings.spec.ts` links to it; the stable id is preserved verbatim.
- Assumption A2: fixing the `sleepTmeout` typo and adding `logger.level` to
  `docs/hub_state.json` changes only the development/browser-test fixture, not the
  backend contract. This is required for the storage and logger browser checks.
