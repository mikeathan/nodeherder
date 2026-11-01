# Spec: Extensible IoT control center

ID: 038 | Status: Design draft | Updated: 2026-10-30
Constitution: 1.0.0 (proposed) | Components: backend and frontend

## Problem and scope

NodeHerder manages Zigbee2MQTT devices but cannot yet onboard, monitor or operate ESPHome native API devices. Provide a protocol-independent device and entity experience, first exercised by ESPHome and the specified Everything Presence Lite, while preserving existing devices, automations and state semantics. Research and scope are recorded in [research.md](research.md).

In scope: a hardened Go-native ESPHome client based on `mycontroller-org/esphome_api@v1.4.0`; multiple manually configured devices; supported discovery; encrypted and unencrypted native connections; dynamically discovered entities; live observations; supported commands; connection health and recovery; existing automation integration; generic controls and reusable spatial display. Cameras are represented safely when unsupported; camera streaming, BLE and tracking history are excluded unless separately justified. The user selected patching the first package; that choice does not establish that its patched version is already mature or safe.

**UI scope (user direction):** preserve existing layouts, navigation, colors, typography and visual designs. Add ESPHome support to the existing Devices flow and use existing controls, themes and device settings patterns. The Devices page offers **Add ESPHome device** for IP-based onboarding; Zigbee pairing continues through its existing **Permit join** flow, not this action. New UI is limited to ESPHome onboarding, connection/settings information, discovered capabilities and spatial presentation needed for the new device type; it must fit those existing patterns. Keep the backend provider architecture extensible without presenting an unsupported generic provider picker. Do not redesign Zigbee devices or shared pages.

**Follow-up:** ESPHome device pages must keep long readings, configuration and diagnostic lists navigable with bounded, searchable, alphabetically sorted sections. Other device pages remain unchanged. The connection dialog must not claim no ESPHome devices exist while loading or after an error; if the registry contains a device without a matching saved enrollment, identify the mismatch without inventing connection settings or deleting data.

**Live-state follow-up:** Reports received while the initial hub snapshot is in flight must be reconciled with the snapshot, including reports for a device not yet present in the browser registry. Buffered data must be bounded, must not invent devices or entities, and an overflow must be visible and recoverable. A page refresh must not be required to receive later readings.

**Spatial setup follow-up:** Make the absence of a saved target mapping visible before the graph. Explain how to select the discovered X, Y and active entities and save them; do not infer or invent target bindings from device names.

Live device updates must not reset unsaved mapping selections or the current page of an ESPHome entity section. Saving a partially filled mapping, identical X/Y fields, or entities no longer available in the device inventory must show an actionable error and must not send a settings update.

When discovered labels unambiguously identify numbered targets and axes, warn immediately and reject a mapping that mixes target numbers, places a numbered target in a differently numbered row, or uses a Y-labelled coordinate as X (or vice versa). Do not guess from unrecognized labels.

For unambiguous numbered target labels, each mapping selector offers its matching target and axis. Devices without a recognizable matching label retain manually selectable unrecognized fields. Existing invalid selections remain visible for correction. Repeated mapped coordinate freshness may be broadcast at most four times per second per entity so movement is responsive while output remains bounded.

Review corrections: interruption of an enrollment edit or removal must restore the old session after it exits, unless shutdown or another operation supersedes it. An entity with an unsupported text pattern remains discoverable but cannot be written; other valid entities continue to work. Ordinary uncategorized switches and lights belong with controls; explicit device categories remain authoritative. Clearly distinguish device configuration from hub settings. Inventory-dependent sorting should not repeat for every value update. Binary active state may be change-only, so it cannot safely be expired on a timer without a reporting guarantee; spatial coordinates still expire and connection loss invalidates the view.

Spatial view must provide an expandable, keyboard-accessible help guide using the existing integration-help hierarchy, with clear diagrams and instructions for mapping, interpreting the plane, and diagnosing missing targets. The guide must be responsive and must not change device configuration.

## Acceptance scenarios

- AC-01: A signed-in user adds multiple distinct devices; valid metadata and entity discovery populate the existing device experience without duplicating identities or exposing encryption keys. Invalid or unavailable targets show actionable failures.
- AC-02: A discovered sensor report updates confirmed state and applicable automations before UI/metrics sampling. Equal repeated events remain events. Multiple browser clients do not create extra device subscriptions.
- AC-03: A user or automation commands a writable entity through the same protocol-independent routing; unsupported, invalid, unauthorized or offline commands fail visibly. A sent command is never presented as confirmed without a subsequent report, and ambiguous commands are not automatically replayed.
- AC-04: A disconnected device is visibly unavailable; reconnect uses bounded backoff, re-discovers and re-subscribes without duplicate sessions, and resumes updates. Shutdown joins owned workers without leaks or unbounded queues.
- AC-05: Existing Zigbee device identity, events, automation actions, commands, API/WS consumers and persisted data remain compatible; rollback and recovery are documented and tested.
- AC-06: The responsive, keyboard-accessible UI supports searching, device details, diagnostics and capability-driven state/control; unknown types remain safe and unavailable data is not displayed as live.
- AC-07: Spatial targets appear only when a mapping with known units, identity, freshness and disappearance semantics is supported by evidence or explicitly configured. The view supplies orientation, scale, grid and offline/no-target states; incomplete or stale coordinates are not represented as current positions.
- AC-08: Configuration and keys survive restart under a documented encrypted storage/backup and deletion policy. Secrets never appear in browser read responses, logs, metrics or unprotected public endpoints. Unsafe targets and untrusted metadata are rejected or safely handled.
- AC-09: Automated tests use fake transports, clocks and isolated storage. Physical-device discovery and commands are separately scoped manual checks, never routine tests.
- AC-10: Against the pinned patched Go client, invalid keys and failed handshakes release sockets, concurrent observations pass race detection, and oversized/malformed frames fail without panic or unbounded allocation. Context cancellation and close stop owned work within the documented deadline.
- AC-11: A provider implemented with a test double can register and deliver state and supported commands without modifying the ESPHome or Zigbee provider. Optional discovery is not required for a provider lacking it.
- AC-12: ESPHome onboarding, detail, capability controls and spatial view follow the existing Devices UI and the scope in [UI design](ui-design.md), including keyboard, focus, confirmed/pending and error states; existing pages and Zigbee device appearance remain unchanged.

## Requirements and success

- FR-01: Registration and lifecycle must support independent providers without forcing irrelevant interfaces; identity and entity metadata stay stable across reconnects (AC-01, AC-04, AC-05).
- FR-02: Discovered capabilities determine visible state and supported commands; the authoritative state remains shared with existing consumers (AC-01–AC-06).
- FR-03: Native API session management supports encryption, bounded recovery, clean shutdown and explicit diagnostics (AC-01, AC-04, AC-08).
- FR-04: Browser and automation command paths use validated, authorized routing without changing existing Zigbee semantics (AC-03, AC-05).
- FR-05: Spatial support is protocol-independent and displays only evidence-backed data (AC-07).
- FR-06: Maintain a pinned, reviewable Go patch of the selected client with upstream attribution, reproducible regression tests and an explicit update/ownership process (AC-10).
- FR-07: Keep provider lifecycle, command routing, capability rendering and spatial presentation in cohesive, independently testable modules; share existing state and controls rather than duplicate them (AC-02, AC-03, AC-06, AC-07, AC-11, AC-12).
- NFR-01: Queue sizes, retry/timeout bounds, ordering, health transitions, stale-state rules and shutdown deadlines must be specified and verified in the design before production implementation (AC-02–AC-04, AC-07, AC-09).
- NFR-02: Versioned persistence, HTTP and WS changes must specify producers, consumers, migration, rollback and recovery; access expansion requires explicit security review (AC-05, AC-08).
- SC-01: Fake-server acceptance demonstrates two concurrent devices, disconnect/reconnect, encrypted and plaintext handshakes, discovered sensors and commands, without household network access (AC-01–AC-04, AC-09).
- SC-02: Existing backend/frontend suites and affected race, security, accessibility and viewport checks pass; actual results are recorded (AC-05, AC-06, AC-08, AC-09).

## Entities and interfaces

A provider owns protocol communication; devices own stable identities and connection health; entities describe observed values and possible operations; commands are requests and reports provide confirmation. The existing application device store and event channels remain authoritative unless the eventual ADR demonstrates a safe replacement. Spatial mapping must include units and validity rather than relying on ESPHome-specific labels in shared UI code.

## Open investigation

- Q-01: **Resolved choice:** patch `mycontroller-org/esphome_api@v1.4.0` in Go. Its fitness for deployment remains unverified; pass AC-10 and the full client acceptance gate before integration is enabled.
- Q-02: What exact entities, availability signals, update cadence and coordinate semantics does the stated 1.5.0 firmware expose? Obtain version-specific source or an explicitly scoped read-only inventory; do not invent a mapping.
- Q-03: How will existing public device-context/MCP routes avoid exposing newly added private device metadata or commands? Resolve before connecting new devices to those consumers.
- Q-04: Define safe permitted network targets, DNS/rebinding behavior, encryption-key storage and deployment/recovery independently of optional external integrations. Resolve before onboarding implementation.
- Q-05: Specify stable identity, entity migration, ordering/backpressure and UI/automation command contracts against actual existing consumers before implementation.

## Review corrections — 2026-10-31

- AC-13: A connected provider remains available without changed entity reports; transport failure still marks it offline. A delayed keepalive response never blocks ordered observations.
- AC-14: Automation buttons serialize null with no value field; non-null button payloads fail validation. Stateless UI commands prevent immediate duplicate submissions and label submission without claiming execution.
- AC-15: Management retains last loaded data on polling failure, keeps mutation errors until explicitly retried, and ignores mutation completions from a closed dialog generation. Conflicting mutations are disabled.
- AC-16: Validated manufacturer/model/firmware/project metadata appears in About; discovered device classes supply generic entity presentation without changing stable IDs or Zigbee presentation.

## Spatial workspace follow-up — 2026-10-31

User direction expands the Spatial tab into a live dashboard with dedicated Zone Editor,
Room Builder, Entry/Exit & Assumed Presence, and Heatmap & Analytics tools. Each tool
owns its contextual help; first-zone guidance belongs beside the editor and empty dashboard.

- AC-17: Live dashboard shows fresh targets, recent trails, sensor placement, optional
  user-calibrated field of view, room objects, named typed zones, connection status and
  geometric occupancy previews. Missing coordinates/offline never imply a clear room.
- AC-18: Create, select, rename, move, resize and remove rectangular detection/exclusion/
  entry zones; edit polygon vertices with keyboard-accessible numeric alternatives.
  Slots are bounded to four detection, two exclusion and two entry zones. First-zone
  instructions explain sensor-relative millimetres, saving and verification.
- AC-19: Room Builder edits an outline, walls/doors/furniture and sensor placement using
  the same canvas. Layouts and zone drafts persist in existing per-device hub settings;
  saving a layout never claims to configure sensor firmware. Deletion is undoable until save.
- AC-20: Entry help accurately describes the EPL best-effort occupancy hold, not people
  counting or dedicated crossing events. Explicitly mapped discovered entities supply
  confirmed assumed presence, remaining time and existing configuration controls.
- AC-21: Heatmap and activity summaries use bounded, one-second, in-memory observations
  collected while this Spatial tab is open; ranges are last 5/15/60 minutes with actual
  collection coverage. No invented historical data, browser persistence or zero-filled
  offline periods. Clear requires confirmation; leaving/device change discards observations.
- AC-22: Each tool supplies keyboard-operable contextual help and controls at 1440/768/375px;
  existing mapping help and troubleshooting remain available. Geometry validation rejects
  nonfinite/out-of-bounds/degenerate inputs before persistence, including backend admission.

Requirements FR-05/07 extend to AC-17–22. Sensor-origin coordinates and room coordinates
are millimetres; rotation is degrees. Workspace settings are optional/versioned, additive
and remain compatible with existing spatial mappings. No new endpoints or dependencies.
Physical-zone scope uses the explicit-binding default described in AC-23 below.
Reference: Everything Smart zone creation, Live Dashboard, Zone Editor, Room Builder,
Entry/Exit & Assumed Presence, and Heatmap & Analytics guides supplied by the user.
The current entry guide explicitly says this is a heuristic occupancy hold and exposes
no entry/exit counters. Documentation has inconsistent firmware limits/defaults: use
actual discovered metadata for writable bounds and capability availability.

Physical-control assumption: use both overlays and explicit sensor bindings (the
suggested default); no inferred entity IDs or unsolicited commands. AC-23: a selected
rectangle can bind four distinct readable/writable numeric length entities. Preview
exact native-unit requested values; block unavailable/duplicate/out-of-range/step-
misaligned/unsupported/pending/offline submissions before sending any. Apply is
explicit and non-atomic; show requested/unconfirmed state, existing timeout feedback
and reported values. An optional reported occupancy entity appears separately from geometric previews. Never retry automatically or claim firmware persistence from
submission. Polygon writes require an evidenced firmware serialization contract and
are unavailable here; provide guidance to the supported manufacturer tool instead.

## Spatial review corrections — 2026-10-31

- AC-24: Spatial saves retain the current draft and Undo until the backend confirms
  persistence. Rejection, disconnect and a bounded acknowledgement timeout show a
  useful error without publishing requested settings as saved. Replies from an old
  device/view generation cannot update a new view. Edits made while saving remain
  editable and survive the acknowledgement; duplicate submissions are disabled.
- AC-25: A rectangle corner drag retains its original opposite anchor across axis
  crossings. Numeric and pointer editing share pure tested rectangle geometry.
- AC-26: Known spatial distances use the shared quantity formatter; zone colors use
  central theme tokens. Polygon intersection/area validation explains collinear
  cases, adjacent edges and tolerance in both Go and TypeScript.

The settings WebSocket request remains `saveDeviceConfigOverride` with an optional
`requestId` UUID-sized string. Correlated requests receive `deviceConfigSaveResult`
with `{requestId, deviceId, success, error?}` only after validation/persistence returns.
Legacy requests keep their existing operationSuccess/operationFailed responses.
Replies contain no room geometry, secrets or tracking positions. This is an additive
contract on the existing authenticated socket; no new public routes or device writes.

## Mapping interaction stability follow-up (2026-11-01)

- AC-22: Given saved Target 1 and Target 2 mappings and continuing coordinate reports,
  choosing Not mapped in any Target 2 selector must take effect and remain editable.
  A native dropdown's pending choice must survive reports before its change event.
- AC-23: Clearing all three fields and saving must persist an empty Target 2 row,
  preserve Target 1 and row numbering, and restore the empty row on reopening.
  Failed saves keep the editable draft; partial rows still require correction.
- Scope: browser mapping interaction and regression coverage only. No firmware commands,
  inferred bindings, new persistence contracts or changes to coordinate semantics.
