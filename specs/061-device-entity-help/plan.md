# Plan: Entity help for device configuration

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft | Updated: 2026-11-10
Constitutions: parent 1.0.0 (proposed); backend 1.0.0 (proposed); frontend 1.0.0 (proposed)

## Context and design

Observed facts:

- ESPHome discovery has no description field. `Describe` already normalizes untrusted
  metadata and `deviceEntity` writes bounded attributes; `object_id` is available on every
  discovery message but is currently discarded.
- The shared `Entity.Description` carries the provider description for Zigbee2MQTT and the
  discovered name for ESPHome. Reusing it for ESPHome help would break
  `exposeDisplayLabel`'s fallback for opaque provider IDs, so help must be a separate
  concept.
- `ExposeRow` is the single row renderer for provider entities; `ProviderEntitySection`
  supplies search and pagination; `DevicePage` selects the configuration section.
- `.nh-card` sets `overflow: hidden`, so an absolutely positioned tooltip would be clipped
  and a hover-only tooltip is unavailable on touch. An inline disclosure is the reliable
  interaction.
- The ESPHome discovery API has **no default value**. `EncodeCommand` already validates
  every number/select/switch/light/text command against the descriptor, and
  `useDeviceCommand.send` already owns pending state, the confirmation timeout, offline
  refusal and the no-retry rule. Reset must reuse that path.
- Factory defaults are only in the firmware YAML (`initial_value`, `initial_option`,
  `restore_mode`) and the manufacturer docs, so they require curated data just like help.

Design:

1. **Backend carries the fact.** Add `ObjectID` to `Descriptor`, populate it through the
   existing `GetObjectId()` capability check, validate it with the same bounded text rule
   as names and device classes, and write it as `attributes.object_id` in `deviceEntity`.
   No text and no device-model logic enter Go.

2. **Frontend owns presentation resolution.** New cohesive module
   `src/domain/entityHelp/`:
   - `types.ts`: `EntityHelp`, `EntityHelpKind`, `EntityHelpSource`, `EntityHelpResolver`.
   - `catalog.ts`: pure data. `ENTITY_HELP_BY_OBJECT_ID` holds vendor/component knowledge
     sourced from official documentation; `ENTITY_HELP_BY_DEVICE_CLASS` holds generic
     standards-based explanations. Every entry is `{ text, source? }`.
   - `resolve.ts`: `providerEntityHelp` (reuses `showsExposeDescription`),
     `catalogEntityHelp`, and `resolveEntityHelp` over an ordered `EntityHelpResolver[]`;
     uses the shared identity normalizer from `src/domain/entityKey.ts`.
   - `index.ts`: the public surface.
   The resolver is the single resolution path (SSOT); sources conform to one function
   type, so adding a source or entry does not modify consumers (OCP/DIP).

   Precedence: provider description → `object_id` → normalized label → `device_class`.
   Provider help wins because it is authoritative for that bridge; the label fallback
   covers older ESPHome records that predate the attribute.

3. **UI reuses the resolved value.** `ExposeRow` resolves once and:
   - renders provider help inline through the existing `<small>` path (AC-05, AC-06);
   - renders catalog help through a native button and an inline panel with a unique id,
     `aria-expanded` and `aria-controls` (AC-01, AC-02, AC-07, AC-08).
   A new opt-in `show-help` prop flows `DevicePage` → `ProviderEntitySection` → `ExposeRow`
   and is enabled only for the ESPHome configuration section. Unknown entities therefore
   stay unchanged (AC-04).

4. **Defaults are separate data, not help text.** New `src/domain/entityDefaults/`:
   - `types.ts`: `EntityDefaultValue`, `EntityDefault`, `EntityDefaultTarget`.
   - `catalog.ts`: `ENTITY_DEFAULTS_BY_OBJECT_ID` sourced from the firmware YAML
     (`initial_value`, `initial_option`, `restore_mode`) and manufacturer docs.
   - `resolve.ts`: `resolveEntityDefault` validates the entry against writability,
     control type, range/step and select options; `entityDefaultTargets` filters a list;
     `resetDefaultsMessage` builds the confirmation summary.
   `normalizeEntityKey` moves to `src/domain/entityKey.ts` so help and defaults share one
   identity rule. Only documented defaults are eligible; buttons, unsupported entities and
   the disruptive `mmWave Bluetooth` switch are excluded (AC-14, AC-16, AC-17).

5. **UI sends resets through the existing command path.** A new
   `src/composables/device/useEntityReset.ts` resolves targets from the whole card (not
   just the current page), disables the action when the device is offline or any target is
   pending, confirms with a summary naming the affected settings, then calls
   `useDeviceCommand.send` per target. `ProviderEntitySection` gains an opt-in `show-reset`
   and renders the button from that composable. The success notice says "requested", never
   "confirmed" (AC-12, AC-13, AC-15, AC-18).

6. **Tests.** Go table cases for capture, validation and the written attribute. Jest cases
   for help normalization/precedence/catalog integrity and for default applicability,
   filtering and confirmation copy. The device page browser test covers both the
   disclosure and a reset request.

Files:

| Path | Change |
| --- | --- |
| `backend/internal/deviceproviders/esphome/constants.go` | `attributeObjectID` |
| `backend/internal/deviceproviders/esphome/entity.go` | `Descriptor.ObjectID`; bounded capture |
| `backend/internal/deviceproviders/esphome/device_sink.go` | write `object_id` attribute |
| `backend/internal/deviceproviders/esphome/entity_test.go` | capture + invalid input cases |
| `backend/internal/deviceproviders/esphome/device_sink_test.go` | attribute assertion |
| `frontend/src/types/device.type.ts` | `ExposeAttributeKeys.ObjectID` |
| `frontend/src/domain/entityKey.ts` | shared `normalizeEntityKey` |
| `frontend/src/domain/entityHelp/{types,catalog,resolve,index}.ts` | help resolver and data |
| `frontend/src/domain/entityDefaults/{types,catalog,resolve,index}.ts` | defaults data, validation and summary |
| `frontend/src/components/entity/ExposeRow.vue` | resolve help, disclosure, styles |
| `frontend/src/composables/device/useEntityReset.ts` | reset targets, availability and confirmed bulk action |
| `frontend/src/components/device/provider/ProviderEntitySection.vue` | `show-help` passthrough and reset button |
| `frontend/src/components/device/DevicePage.vue` | enable help and reset on configuration |
| `frontend/src/__tests__/domain/entityHelp/{resolve,catalog}.test.ts` | help unit coverage |
| `frontend/src/__tests__/domain/entityDefaults/{resolve,catalog}.test.ts` | defaults unit coverage |
| `frontend/e2e/devices.spec.ts` | disclosure and reset browser cases |

Alternatives rejected:

- **Backend-hosted descriptions:** duplicates browser copy, needs a cross-language source
  of truth and links, and gives Go a presentation responsibility (BE-01/FE-01).
- **Runtime scraping of vendor docs:** adds a network dependency, latency, failure modes
  and an external trust boundary for text that changes rarely (NH-03/NFR-01).
- **Synthesizing from min/max/options:** truthful about shape but not purpose; it would put
  boilerplate on every row and still not explain "Assume Present Timeout" (AC-04).
- **Device-model or project-name checks in components:** violates NFR-02 and makes every
  new device a code change.
- **Reusing `Entity.Description` for ESPHome help:** it already holds the discovered name
  and is the fallback label for opaque provider IDs; overwriting it would break labels.
- **Hover-only tooltip:** clipped by `.nh-card`, unavailable on touch, weaker for long text
  and source links (AC-09).
- **Reset to min/zero/first option:** none of those is a factory default; it would silently
  change detection behaviour and still be wrong. Only documented defaults are eligible.
- **Infer defaults from the first observed state:** that is the last user value, not a
  default, and would make reset meaningless after the first session.
- **Fetch the firmware YAML at runtime:** adds network, latency, parsing and trust
  boundaries for data that changes rarely (NH-03/NFR-04).
- **Reset every writable entity:** buttons are stateless, unsupported entities have no
  value, and `mmWave Bluetooth` reboots the sensor. Eliminating those is a data decision,
  not a component special case (AC-16).
- **Backend-hosted defaults:** would duplicate the data and add a protocol change for a
  browser-only affordance; the existing command path already validates values (NFR-04).

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 061 with AC-01…AC-18 and resolved Q1–Q7 | PASS | PASS |
| NH-02 | Reset reuses `useDeviceCommand.send`: pending state, confirmation timeout, offline refusal and no retry; help changes no command path | PASS | PASS |
| NH-03 | Static text only; no new entry point, secret, household data or runtime fetch | PASS | PASS |
| NH-04 | Additive attribute; resolvers depend on function types, not concrete sources; catalogs are data | PASS | PASS |
| NH-05 | Go cases + help/defaults Jest cases + desktop/mobile device-page browser checks; results below | PASS (design) | PASS |
| NH-06 | No persistence, timer, subscription or retention change; bulk reset requires confirmation and never claims confirmed state | PASS | PASS |
| BE-01 | Adapter change stays inside the ESPHome provider; no transport or service change | PASS | PASS |
| BE-02 | No `DeviceLifetimeService` or update-semantics change | N/A | N/A |
| BE-03 | No goroutine, timer or shared mutable state added | PASS | PASS |
| BE-04 | Additive attribute only; no query/storage schema or defaults change | PASS | PASS |
| BE-05 | No route, payload, authentication or MCP change | N/A | N/A |
| FE-01 | Pure domain resolver; row remains presentational; no shared state change | PASS | PASS |
| FE-02 | Attribute is optional; existing records and mocks remain valid | PASS | PASS |
| FE-03 | Reset requests stay distinct from confirmed state; rows keep their existing pending/confirmation/timeout behaviour | PASS | PASS |
| FE-04 | Native help button, accessible reset button, explicit reset confirmation, focus-visible and mobile checks | PASS (design) | PASS |
| FE-05 | Plain text, no HTML, static `https` source links with `noopener noreferrer` | PASS | PASS |

## Contracts, data, rollout

- `Expose.attributes.object_id`: optional string, additive JSON. Producer: ESPHome adapter.
  Consumer: frontend resolver. Older persisted devices lack it and fall back to the
  normalized label, so no migration or compatibility window is required.
- No new HTTP/WS/MQTT/MCP field or persistence change. Reset sends existing JSON scalar
  values through the existing `deviceSetValue` WS event; `EncodeCommand` on the backend
  remains the validation authority. Rollback is reverting the branch; no stored data refers
  to the feature.
- Both catalogs ship with the bundle. Source links are static official documentation and
  are rendered as plain anchors.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-03, AC-10 | `entity.go`, `device_sink.go`, `constants.go` | Go: `TestDescribeDiscoveredEntities`, `TestDeviceSinkPublishesOnlyBoundedPublicMetadata` | pass |
| FR-02, FR-03, AC-01, AC-04, AC-05, AC-11 | `domain/entityHelp/*` | Jest: `resolve.test.ts` | pass |
| FR-05, AC-02, AC-10 | `catalog.ts` | Jest: `catalog.test.ts` | pass |
| FR-04, AC-07, AC-08 | `ExposeRow.vue`, `DevicePage.vue` | browser: info button, Enter toggles `aria-expanded`, help text visible | pass |
| AC-06 | `ExposeRow.vue`, `ProviderEntitySection.vue` | diff review: provider description remains inline, not duplicated | pass |
| AC-09 | `ExposeRow.vue` | browser: 1280 px and 375 px, `scrollWidth <= clientWidth` | pass |
| FR-06, FR-09, AC-14, AC-16, AC-17 | `domain/entityDefaults/*` | Jest: `resolve.test.ts`, `catalog.test.ts` | pass |
| FR-07, FR-08, AC-12, AC-13, AC-15 | `ProviderEntitySection.vue` | browser: confirm, send two defaults, skip unknown/button | pass |
| AC-18 | `ProviderEntitySection.vue` | browser: reset control at 1280 px and 375 px, no overflow | pass |
| NFR-01, NFR-02, NFR-03, NFR-04 | all | diff review: no dependency, fetch, device-model check, new command protocol or persistence change | pass |

## Verification (recorded)

- `cd backend && go test ./...` → **54 packages passed**, including
  `node-herder/internal/deviceproviders/esphome`.
- `cd backend && go build ./...` → pass (no output).
- `cd backend && go vet ./...` → pass (no output).
- `cd frontend && npm test -- --runInBand` → **144 suites / 802 tests passed**.
- `cd frontend && npm run lint` → **0 errors, 23 pre-existing warnings** (unchanged set).
- `cd frontend && npm run build` → pass (`vue-tsc` + Vite).
- `cd frontend && npm run type-check:e2e` → pass.
- `cd frontend && npx playwright test e2e/devices.spec.ts --project=desktop` →
  **11 passed**, including the configuration-help and reset cases.
- `cd frontend && npx playwright test e2e/devices.spec.ts --project=mobile` →
  **11 passed**, including the 375 px overflow assertions.

## Debt, risks, exceptions

- **Catalog drift:** vendor firmware can rename entities, change defaults or change select
  options. Mitigation: conservative wording, source links, one entry per stable `object_id`,
  and per-value validation against the current metadata. A default that no longer fits is
  skipped, not coerced. Version-aware catalogs can be added later if drift is observed.
- **Variant-specific defaults:** `update_speed` and some option sets differ by sensor
  variant. The option-presence check prevents an invalid command, but the catalog may not
  cover every variant yet; unknown/inapplicable entries simply do not reset.
- **Batch volume:** the catalogue currently covers at most a handful of settings per device;
  each reset uses one existing command and the backend queue remains bounded. No retry or
  parallel flood is introduced.
- **Disruptive entities:** `mmWave Bluetooth` is excluded because toggling it reboots the
  sensor; stateless buttons and unsupported update entities are excluded by control type.
  They remain individually controllable.
- **Reset feedback:** the success notice says "requested", not "confirmed"; per-row pending
  state and the existing confirmation timeout remain the source of truth (NH-02/FE-03).
- **Renamed entities:** `object_id` is derived from the ESPHome name, so a user rename
  loses the match. The label fallback covers missing attributes, not renames. Recorded in
  the spec.
- **Global key assumption:** two components could reuse an `object_id` with different
  meaning. The current catalog uses distinctive names; if a collision appears, scope the
  key by project or device class rather than adding component checks.
- **Uncatalogued entities show nothing:** deliberate (AC-04). The alternative is invented
  or boilerplate text, which is worse than silence.
- **English-only:** no i18n system exists; adding one is a separate change.
- **No ADR:** the wire change is additive and reversible, and the boundary decision
  (facts in Go, presentation in Vue) follows existing ownership. Revisit with an ADR only
  if a second consumer needs the text server-side.

## Completion

Implemented on `feat/device-entity-help`. Acceptance evidence is recorded above; no
exceptions are requested. The help catalog covers documented settings only and resolves to
no help for unknown entities (AC-04). The defaults catalog covers documented defaults only;
unknown or now-inapplicable values do not reset, and stateless/disruptive entities are
excluded (AC-14, AC-16).
