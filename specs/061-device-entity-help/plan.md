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
   - `resolve.ts`: `normalizeEntityKey` (lowercase, non-alphanumerics to `_`, collapsed),
     `providerEntityHelp` (reuses `showsExposeDescription`), `catalogEntityHelp`, and
     `resolveEntityHelp` over an ordered `EntityHelpResolver[]`.
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

4. **Tests.** Go table cases for capture, validation and the written attribute. Jest cases
   for normalization, precedence, label fallback, device class, unknown entities and
   catalog integrity.

Files:

| Path | Change |
| --- | --- |
| `backend/internal/deviceproviders/esphome/constants.go` | `attributeObjectID` |
| `backend/internal/deviceproviders/esphome/entity.go` | `Descriptor.ObjectID`; bounded capture |
| `backend/internal/deviceproviders/esphome/device_sink.go` | write `object_id` attribute |
| `backend/internal/deviceproviders/esphome/entity_test.go` | capture + invalid input cases |
| `backend/internal/deviceproviders/esphome/device_sink_test.go` | attribute assertion |
| `frontend/src/types/device.type.ts` | `ExposeAttributeKeys.ObjectID` |
| `frontend/src/domain/entityHelp/{types,catalog,resolve,index}.ts` | resolver and data |
| `frontend/src/components/entity/ExposeRow.vue` | resolve help, disclosure, styles |
| `frontend/src/components/device/provider/ProviderEntitySection.vue` | `show-help` passthrough |
| `frontend/src/components/device/DevicePage.vue` | enable help on configuration |
| `frontend/src/__tests__/domain/entityHelp/{resolve,catalog}.test.ts` | unit coverage |

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

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 061 with AC-01…AC-11 and resolved Q1–Q3 | PASS | PASS |
| NH-02 | Diff review: no command, confirmation, event or automation path changed | PASS | PASS |
| NH-03 | Diff review: static text only; no new entry point, secret, household data or runtime fetch | PASS | PASS |
| NH-04 | Additive attribute; resolver depends on a function type, not concrete sources; catalog is data | PASS | PASS |
| NH-05 | Go cases + Jest cases + desktop/mobile device-page browser checks; results below | PASS (design) | PASS |
| NH-06 | Diff review: no persistence, timer, subscription or retention change | PASS | PASS |
| BE-01 | Adapter change stays inside the ESPHome provider; no transport or service change | PASS | PASS |
| BE-02 | No `DeviceLifetimeService` or update-semantics change | N/A | N/A |
| BE-03 | No goroutine, timer or shared mutable state added | PASS | PASS |
| BE-04 | Additive attribute only; no query/storage schema or defaults change | PASS | PASS |
| BE-05 | No route, payload, authentication or MCP change | N/A | N/A |
| FE-01 | Pure domain resolver; row remains presentational; no shared state change | PASS | PASS |
| FE-02 | Attribute is optional; existing records and mocks remain valid | PASS | PASS |
| FE-03 | No control, request or optimistic-state change | PASS | PASS |
| FE-04 | Native button, accessible name, `aria-expanded`/`aria-controls`, focus-visible, mobile check | PASS (design) | PASS |
| FE-05 | Plain text, no HTML, static `https` source links with `noopener noreferrer` | PASS | PASS |

## Contracts, data, rollout

- `Expose.attributes.object_id`: optional string, additive JSON. Producer: ESPHome adapter.
  Consumer: frontend resolver. Older persisted devices lack it and fall back to the
  normalized label, so no migration or compatibility window is required.
- No HTTP/WS/MQTT/MCP, configuration or persistence change. Rollback is reverting the
  branch; no stored data refers to the feature.
- Catalog data ships with the bundle. Source links are static official documentation and
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
| NFR-01, NFR-02, NFR-03 | all | diff review: no dependency, fetch, device-model check, command or persistence change | pass |

## Verification (recorded)

- `cd backend && go test ./...` → **54 packages passed**, including
  `node-herder/internal/deviceproviders/esphome`.
- `cd backend && go build ./...` → pass (no output).
- `cd backend && go vet ./...` → pass (no output).
- `cd frontend && npm test -- --runInBand` → **142 suites / 792 tests passed**.
- `cd frontend && npm run lint` → **0 errors, 23 pre-existing warnings** (unchanged set).
- `cd frontend && npm run build` → pass (`vue-tsc` + Vite).
- `cd frontend && npm run type-check:e2e` → pass.
- `cd frontend && npx playwright test e2e/devices.spec.ts --project=desktop` →
  **10 passed**, including the new configuration-help case.
- `cd frontend && npx playwright test e2e/devices.spec.ts --project=mobile` →
  **10 passed**, including the 375 px overflow assertion.

## Debt, risks, exceptions

- **Catalog drift:** vendor firmware can rename entities or change defaults. Mitigation:
  conservative wording, source links, and one entry per stable `object_id`; version-aware
  catalogs can be added later if drift is observed.
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
exceptions are requested. The catalog intentionally covers documented settings only and
resolves to no help for unknown entities (AC-04).
