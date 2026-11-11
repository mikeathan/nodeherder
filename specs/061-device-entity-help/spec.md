# Spec: Device configuration help and defaults

ID: 061 | Status: Draft | Owner: agent draft | Updated: 2026-11-10
Constitution: 1.0.0 (proposed) | Components: backend, frontend

## Problem, scope, exclusions

ESPHome configuration entities appear with only a discovered name. Names such as
"Assume Present Timeout" or "Exit Threshold Pct" do not explain what the setting changes,
and a wrong value can alter presence behaviour for the whole room.

Facts observed in the current tree:

- `ExposeRow.vue` already renders `expose.description` as a one-line subtitle, and
  `ProviderEntitySection.vue` gates it with `showsExposeDescription`.
- The ESPHome adapter sets `e.Description = d.Name` (`device_sink.go:161`). Because the
  description equals the display label, `showsExposeDescription` suppresses it, so ESPHome
  rows never show help.
- The ESPHome native API (`third_party/esphome_api/proto/api.proto`) has **no description
  field**. It exposes `object_id`, `name`, `icon`, `device_class`, `entity_category`,
  `unit_of_measurement`, `min/max/step`, `options`, `assumed_state`, `disabled_by_default`.
- `Describe` (`entity.go`) keeps `device_class` and category but drops `object_id` and
  `icon`; `deviceEntity` writes only label, entity kind, device class and control limits.
- Zigbee2MQTT genuinely supplies a description (`BridgeExpose.Description`, copied in
  `CreateEntityFromExpose`). It is already shown inline and must keep working.
- Manufacturer documentation is the authoritative semantic source for vendor entities,
  for example "Assume Present Timeout: How long, in seconds, to hold occupancy after a
  target is lost without a valid exit being detected. Default is 30 minutes."
- The same API exposes **no default value**: numbers carry only min/max/step/unit, selects
  carry options, switches/lights carry `assumed_state`, and buttons carry nothing. The
  current report is the last state, not a factory default.
- Defaults exist in the firmware YAML (`initial_value`, `initial_option`,
  `restore_mode: RESTORE_DEFAULT_OFF/ALWAYS_OFF`) and in the manufacturer documentation.
  NodeHerder never receives that configuration.
- Configuration rows have no reset affordance. A user who tunes a timeout or threshold
  must remember the original value or reflash/factory-reset the device.

Scope: a provider-agnostic help resolver in the frontend, backed by a data catalog keyed
by stable entity identity, plus an accessible disclosure for configuration entities; and a
card-level "Reset to defaults" that sends catalogued defaults through the existing device
command path. The ESPHome adapter additionally carries the canonical `object_id` so
matching does not depend on the display label.

Exclusions: no runtime documentation or firmware-config scraping, no backend-hosted copy
of the text or defaults, no generated descriptions/defaults from ranges/options, no
localization, no new command protocol, no reset of stateless buttons or unsupported update
entities, no help for Z2M entities that already have a provider description, and no
firmware flashing or factory-reset workflow.

## Scenarios and acceptance

### US-01 — Understand a configuration setting (priority P1)

Independent verification: `src/__tests__/domain/entityHelp/resolve.test.ts`,
`src/__tests__/domain/entityHelp/catalog.test.ts`, and the device page in `e2e/devices.spec.ts`.

- AC-01: Given an ESPHome configuration entity whose `object_id` or normalized label has
  a catalog entry, when the user activates the info control, then the row reveals the
  help text.
- AC-02: Given a catalog entry with a source, then the disclosure shows a link that opens
  the source in a new tab with `rel="noopener noreferrer"`.
- AC-03: Given an ESPHome entity, matching uses its `object_id`; when the attribute is
  absent (older record), matching uses the normalized label.
- AC-04: Given no provider description and no catalog match, then no info control is shown
  and no text is invented.
- AC-05: Given a Zigbee2MQTT entity with a provider description, then it remains inline
  and is not duplicated in a disclosure.
- AC-06: Given an ESPHome entity whose description equals its label, then no duplicate
  inline description is rendered.

### US-02 — Accessible interaction (priority P1)

Independent verification: manual keyboard pass and the device page browser test.

- AC-07: The info control is a native button with an accessible name that includes the
  entity label, `aria-expanded`, and `aria-controls` referencing the revealed text.
- AC-08: The control is keyboard operable and shows a visible focus indicator that does
  not rely on colour alone.
- AC-09: At the existing mobile breakpoint the disclosure expands inside the card without
  horizontal scrolling.

### US-03 — Generic extension (priority P2)

Independent verification: resolver unit cases.

- AC-10: Adding a description for another entity requires only a catalog entry; the
  resolver and row component do not change.
- AC-11: Device-class entries apply to any provider's expose that reports that class.

### US-04 — Reset documented settings (priority P1)

Independent verification: `src/__tests__/domain/entityDefaults/` and the device page
browser test.

- AC-12: Given the ESPHome configuration card and at least one catalogued default, when
  the user activates "Reset to defaults", then a confirmation names the affected settings
  and states that values update only after device confirmation.
- AC-13: On confirm, each applicable setting is sent through the existing device command
  path with its documented default; no submitted value is presented as confirmed state.
- AC-14: An entity with no catalogued default, or whose type, range or options make the
  default inapplicable, is skipped.
- AC-15: When the device is offline/disabled or any target command is pending, the reset
  control is unavailable; individual refusals are reported without claiming success.
- AC-16: Stateless buttons, unsupported update entities and settings whose change reboots
  the sensor are never included.
- AC-17: Adding a default for another entity requires only a catalog entry; component
  logic does not change.
- AC-18: At the existing mobile breakpoint the control and confirmation remain usable
  without horizontal scrolling.

## Requirements and success

- FR-01: ESPHome discovery MUST retain a bounded `object_id` on the entity contract;
  acceptance: AC-03, AC-10.
- FR-02: The resolver MUST prefer a provider description, then `object_id`, then the
  normalized label, then `device_class`; acceptance: AC-01, AC-03, AC-05, AC-11.
- FR-03: An unmatched entity MUST resolve to no help; acceptance: AC-04.
- FR-04: The UI MUST render non-provider help as an accessible disclosure and MUST keep
  provider descriptions inline; acceptance: AC-01, AC-02, AC-05, AC-07, AC-08.
- FR-05: Catalog text MUST be plain text with an optional source link; no raw HTML;
  acceptance: AC-02.
- FR-06: Defaults MUST resolve from a data catalog keyed by `object_id` (label fallback)
  and MUST be validated against the expose's writability, control type, range and options;
  acceptance: AC-13, AC-14, AC-17.
- FR-07: Reset MUST send values through `useDeviceCommand`, preserving pending state,
  confirmation timeout, offline refusal and no-retry rules; acceptance: AC-13, AC-15.
- FR-08: Reset MUST require an explicit confirmation that names the affected settings;
  acceptance: AC-12.
- FR-09: Reset MUST exclude stateless actions, unsupported entities and settings with
  disruptive side effects; acceptance: AC-16.
- NFR-01: Help resolution MUST be synchronous and in-memory, with no network call or new
  dependency.
- NFR-02: Both catalogs MUST be data keyed by normalized entity identity; components MUST
  NOT check device models or project names.
- NFR-03: Help MUST NOT alter device command, confirmation, history or persistence
  semantics; reset MUST use the existing command semantics without adding retries or
  optimistic confirmation.
- NFR-04: Reset MUST NOT add a backend command, protocol or storage change; the existing
  `deviceSetValue` path and backend validation remain authoritative.
- NFR-05: Help and reset controls MUST remain keyboard operable and mobile-safe.

## Entities and external interfaces

- `Expose.attributes` gains optional `object_id` (string, valid UTF-8, bounded by the
  existing entity-name limit). Additive and backward compatible; records without it use
  the label fallback.
- `EntityHelp`: `{ text, kind: 'provider' | 'catalog', source?: { label, url } }`.
- `EntityHelpResolver`: `(expose: Expose) => EntityHelp | null`.
- Help catalog: `Record<normalizedKey, { text, source? }>` for `object_id` and `device_class`.
- `EntityDefault`: `{ value: number | string | boolean, source? }`.
- `EntityDefaultTarget`: `{ expose, value, source? }` for one applicable reset command.
- Defaults catalog: `Record<normalizedKey, EntityDefault>` for `object_id`; resolvers are
  pure and validate the value against the current expose metadata.
- Reset reuses the existing browser `deviceSetValue` WS event with the existing JSON scalar
  types. No new HTTP/WS/MQTT/MCP payload, field or persistence change.

## Questions and assumptions

- Q1: Where does the authoritative description come from? The ESPHome API has none;
  vendor documentation and firmware configuration define vendor entities, and standard
  device classes define generic ones. Resolution: curated catalog data plus the existing
  provider description; no generated semantics.
- Q2: Why not synthesize text from min/max/options? It describes the control shape, not
  its purpose, and would add boilerplate to every row. Resolution: no synthesized text.
- Q3: Why not store the copy in the backend? It is presentation copy with source links and
  would duplicate browser content; the backend should carry facts. Resolution: frontend
  resolver.
- Q4: Where do reset defaults come from? The API exposes none. Resolution: curated data
  from the firmware YAML (`initial_value`, `initial_option`, `restore_mode`) and the
  manufacturer docs, keyed by `object_id`, with source links.
- Q5: Why not reset to the minimum or zero? Neither is a default; it would silently change
  detection behaviour. Resolution: only catalogued defaults are eligible.
- Q6: How is firmware/variant drift handled? Every default is validated against the current
  `min/max/step/options`; an inapplicable value is skipped rather than coerced. The catalog
  holds documented defaults only. A future version-scoped catalog can refine entries.
- Q7: Why exclude `mmWave Bluetooth`? Its switch action reboots the mmWave sensor; a bulk
  reset must not trigger an unexpected reboot. It remains individually controllable.
- Assumption A1: The UI is English-only, matching the rest of the application.
- Assumption A2: `object_id` is stable for a given firmware configuration; renaming an
  ESPHome entity changes it. The label fallback covers records without the attribute but
  not renames. Recorded as a limitation.

Success criteria:

- SC-01: The configuration entities in the reported EPL device resolve help, while an
  unknown configuration entity resolves to none.
- SC-02: No new network request, dependency or persisted field.
- SC-03: Backend and frontend suites pass with new resolver, catalog and `object_id`
  coverage.
- SC-04: Reset sends only applicable catalogued defaults; unknown settings and values
  rejected by current metadata remain unchanged.
- SC-05: Reset adds no backend command and preserves offline, pending, confirmation-timeout
  and no-retry behaviour.
