# Plan: Home signal integrations and scalable integration UI

Spec: [spec.md](spec.md) | Status: Draft for review | Author: NodeHerder contributors | Updated: 2026-10-29
Constitutions: parent/backend/frontend 1.0.0 (proposed)
Research: [provider research](provider-research.md) · Prototypes:
[home layout options](prototypes/home-layout-options.html) ·
[home layout with real card sizes and views](prototypes/home-layout-realistic.html) ·
[integrations redesign](prototypes/integrations-redesign.html)
ADRs: [001 OAuth](adr-001-oauth-client-credentials.md) ·
[002 alert feed discovery](adr-002-alert-feed-discovery.md) ·
[003 map dependency](adr-003-map-and-coordinate-presentation.md)

## Context and observed behavior

The integrations subsystem is a declarative provider engine with two installed capabilities.
`capabilities.Registry` is the extension point; concrete types are `waste_schedule/v1` and
`weather_forecast/v1`. Provider YAML is host-editable and shipped seeds are embedded from
`backend/configs/integrations`. The runtime has no scheduler: provider I/O happens only on an
admitted setup step, refresh or status check, and page reads never call a provider.

Facts that constrain this design:

- `frontend/src/contracts/integrations.ts` `parseResource` dispatches weather then falls back to
  `parseWasteSchedule`; any new contract version would be rejected as malformed waste.
- `parseProviderDefinition` hard-codes the two capabilities/types and branches mapping validation
  on capability. The provider editor, catalog and setup wizard are otherwise capability-neutral.
- The catalog DTO already exposes `types[]` and `providers[]`, but `ConnectionCatalogDialog.vue`
  renders them ad hoc with an `type.id === 'weather'` branch and a provider `<select>`.
- Home persists `HomeBlock{Kind: group|card}`. A `card` block is a standalone placement with an
  empty name and exactly one card. `GroupDashboard.vue` renders those blocks at the same level as
  areas, which is the ungrouped result visible in the recorded screen.
- `HomeCard` has no presentation properties; card width is decided by CSS
  (`HomeCardItem.vue` forces `grid-column: 1 / -1` for integrations).
- The app already has a density system (`data-density` comfortable/compact/spacious) and theme
  tokens; no new design system is required.
- No map library exists in `frontend/package.json`.

## Design

### Home layout grammar (grouping and space)

The options prototype compares four grouping approaches; the realistic prototype renders the
chosen model at real card heights with the view switch. The plan selects **one default "Services"
section, rendered as a wide multi-column section with compact cards and a per-section view**.

The earlier draft treated Services as an ordinary-width section. That is wrong: six integrations
in one narrow column is roughly six stacked cards (~700px), which sets a tall row and leaves a gap
beside every shorter area. Section width and card density, not extra headings, control the space.

Three independent axes:

1. **Section width** — a section occupies one or two page columns. Services defaults to two on
   desktop; areas default to one. Users can change it.
2. **Card geometry** — `Width` is one or two inner columns; `Height` is two to five units of the
   shared base unit (`--nh-card-unit`, nominally 30px plus gap). Defaults come from the source
   kind and are overridable per placement. There is no `small/medium/large` enum: the two numbers
   are explicit and persist independently.
3. **Section view** — `cards` or `list`, chosen per section from its header and persisted. List is
   the density escape hatch for many integrations; cards is the spatial view for a few.

There is one section primitive. A section may hold one card or many (device, integration or
mixed), and a one-card section is normal: it keeps its name, count, add, collapse and view
controls, so it can never become the unlabelled orphan from the recorded screen. There is no
separate “single integration card” type and no “Services type”. “Services” is only the default
section name used when an integration is first added.

Why numeric spans rather than natural heights: real cards are not uniform (a long device name
wraps, bins has three rows, weather has four lines), but letting every card grow freely makes
reordering unpredictable and leaves ragged holes. Spans keep rows aligned and drag predictable
while still allowing the variation the recorded screen shows.

Layout rules:

- Page grid reflows 1/2/3/4 section columns. A two-column Services section spans two of them and
  uses `repeat(auto-fit, minmax(170px, 1fr))` internally, giving three to four card columns at
  1024px and wider. Six compact cards occupy two rows, close to the height of one area of device
  tiles.
- Inside an ordinary one-column section: two card columns at ≥520px, one below. Inside a
  two-column section: three to four columns, collapsing to two and then one as the viewport
  narrows. Sections reflow; internal columns change only by breakpoint, preserving position.
- **Bounded card bodies.** Height is a span, not a suggestion. Content that does not fit is
  truncated: at most two detail rows plus "+N more"; weather shows current plus high/low; bins the
  next two collections; fuel the cheapest two; alerts the headline and expiry. Full lists stay on
  the detail page. A card therefore cannot dominate a section, and the user raises its height in
  edit mode if they want more.
- **Smallest fitting span on placement.** A device tile is three units, or four when the label and
  device name need the extra line. Integration cards default to four units, or five when they carry
  rows. The choice is computed once from the known source presentation, then stored; it is not
  re-measured on every render, so the layout does not jump.
- **Sections have independent heights.** They are top-aligned `align-self: start` items. A row may
  contain a 300px area and a 200px section; the shorter one leaves whitespace below. That is
  accepted, not stretched. Dense placement can fill gaps, off by default.
- **Section view controls density, not content.** In list view each card becomes one ~42px row:
  icon, name, a single primary value, status, and the same link. Card `Width`/`Height` are stored
  but ignored; switching back to cards restores the exact geometry. A wide section may arrange
  rows in two columns. List view does not change order, identity or source data, and the view
  control is available outside edit mode because it is presentation, not layout.
- **Section deletion is explicit.** Removing a non-empty section asks whether to move its cards to
  Services or delete them with it. The default is move, so a delete cannot silently discard
  placements.
- **No third grouping level.** Type headings never appear inside a section. Type is carried by the
  card's icon, title and colour. If a user wants real grouping, the section overflow offers
  "Split into sections by type", which creates ordinary Weather/Collections/Travel/Safety sections
  once. After that they are normal sections: rename, reorder, collapse, drag between them. The app
  never re-splits or reorders on its own.
- **Collapse per section.** The header keeps name, card count and worst status ("1 alert · 2
  stale"); collapsing hides the cards and leaves one row. State persists in the layout. This is the
  safety valve for a large Services section and applies equally to areas.
- Default card order in Services follows severity then usefulness (active alerts, weather, air
  quality, collections, fuel, transport, energy), but only when a card is first added. A manual
  drag always wins; severity never silently reorders cards, so spatial memory holds.
- Single-item section: keep the header, count, add and overflow; the card does not stretch. Empty
  section: header plus an add-card state, never auto-deleted. Auto-hiding headers is rejected.
- Cards may move between sections. New integration cards default to Services; a genuinely
  room-scoped integration (garden sensor) can be dragged into an area and behaves identically.
- Dense placement (filling gaps) is optional and off by default, mirroring the Home Assistant
  sections view. A tall section cannot overlap or cover another section because sections are
  top-aligned grid items; it can only increase row height and leave honest whitespace.

Schema additions, all additive and optional:

- `HomeBlock.Width` (`1|2`, empty means 1), `HomeBlock.View` (`cards|list`, empty means cards) and
  `HomeBlock.Collapsed` (bool).
- `HomeCard.Width` (`1|2`) and `HomeCard.Height` (`2..5`), empty means the source default.

Existing `kind: card` blocks render inside an implicit Services section without rewriting saved
data. The section becomes a real `group` block on the next accepted save, matching the read-only
legacy migration pattern in `home_cache.go`.

Affected files: `frontend/src/components/dashboards/GroupDashboard.vue` (render sections),
new `HomeSection.vue`, `HomeCardFrame.vue`, `HomeCardRow.vue`, `CardSizeControl.vue`,
`SectionWidthControl.vue`;
`HomeCardItem.vue` becomes a dispatcher; `IntegrationTile.vue` becomes a compact integration body
without card chrome; `AreaCard.vue` / `HomeStandaloneCard.vue` collapse into `HomeSection.vue`;
`HomeBlockHeader.vue` gains count, collapse, view switch, width, add and overflow; `home-picker.ts`
and `AddTilesDialog.vue` become grouped/searchable. `HomeBlock.Width`, `HomeBlock.View`,
`HomeBlock.Collapsed`, `HomeCard.Width` and `HomeCard.Height` are added in
`models/settings/home.go` and `types/home.type.ts`.

### Integrations page

- Catalog is type-first. Type cards show icon, name, purpose, configured count and worst status.
  Choosing a type reveals only its providers; provider rows show credential type, cost/limit and
  availability before the wizard starts. Remove the weather special case.
- Configured entries are grouped by type with a status dot **and** word (OK/Stale/Error/Disabled),
  name, provider · location subtitle, last checked time, refresh and overflow actions.
- Search matches name, type, provider and location. Type chips filter. Empty search explains.
- Detail keeps Results/Settings/Configuration/Help but presents them as tabs at ≥768px and as
  labelled stacked sections below. Refresh and Configure stay in the header.
- Icons resolve through one registry keyed by integration type and provider id, with a neutral
  fallback. Generic components contain no provider names.

Affected files: `ConnectionCatalogDialog.vue`, `ConnectionsPage.vue`, `ConnectionBoard.vue`,
`ConnectionWidget.vue` → entry row, `ConnectionDetail.vue`, `IntegrationSections.vue`,
`IntegrationPageHeader.vue`, new `IntegrationCatalog.vue`, `IntegrationTypeCard.vue`,
`IntegrationEntry.vue`, `integration-icons.ts`, plus `types/integrations.type.ts` additions for
type metadata (`icon`, `provider_count`, `configured_count`, `status`).

### Weather extension (additive, no version bump)

- `models/integrations/weather.go`: `WeatherDay.Sunrise`, `.Sunset` (`*string`, local `HH:MM`),
  `.UVMax` (`*float64`); `WeatherPoint.UV` (`*float64`). Nullable, never zero-filled.
- Validation: UV 0–20, clock strings `HH:MM`, sunset after sunrise when both exist, daily UV max
  ≥ 0. Invalid values reject the response as `invalid_response`.
- `capabilities/weather/mapping.go` accepts `uv`, `uv_max`, `sunrise`, `sunset`; daily clock
  selectors declare a local-time format. `normalize.go` reduces an ISO local timestamp to `HH:MM`
  in the resource timezone. Projection unchanged.
- Shipped YAML: Open-Meteo daily `sunrise,sunset,uv_index_max`, hourly `uv_index`. Met Office and
  WeatherAPI add only the fields their fixtures confirm; absent fields stay null.
- Frontend: `types/weather.type.ts`, `contracts/weather.ts`, `WeatherResults.vue`,
  `integrations/weather/presentation.ts` (compact value uses UV/sunset when useful).

### Air quality capability (`air_quality/v1`, type `air_quality`)

- New `models/integrations/airquality.go`: header, location, optional coordinates, timezone,
  `index_family` (`european|us`), `index`, server-derived `category`, `dominant`, current
  pollutant values, bounded hourly series, pollen per species, attribution, freshness.
- Category bands are evaluated on the server from the declared family; the resource carries the
  category so the browser never invents thresholds. Missing index → `unknown`; out-of-season pollen
  → null.
- New `internal/integrations/capabilities/airquality/` (`adapter.go`, `mapping.go`,
  `normalize.go`, `sources.go`) registered in `service.go`.
- Provider `backend/configs/integrations/air_quality/open_meteo.yaml`, origin
  `https://air-quality-api.open-meteo.com` (separate from the forecast host). Setup uses the
  existing `location_select` flow with Open-Meteo geocoding and Postcodes.io, disclosed as
  recipients. No key. CAMS + Open-Meteo attribution required.
- Frontend presenter `frontend/src/integrations/air-quality/` with a compact widget (category +
  dominant + one pollen row) and a results component (current pollutants, hourly, pollen).
- Because the origin differs from `api.open-meteo.com`, air quality is its own instance and the
  user selects a location once for it. Sharing one location across instances is out of scope.

### Alerts capability (`weather_alerts/v1`, type `alerts`)

- New `models/integrations/alerts.go`: status `active|none`, warnings[], attribution, freshness.
  Warning: id, normalized severity (`minor|moderate|severe|extreme|unknown`), provider level
  (`yellow|amber|red|unknown`), type, headline, description, instruction, areas, valid window,
  issued/modified, optional bounded polygon.
- Projection drops warnings whose `valid_to <= evaluated_at`, orders by severity then valid-from,
  and never contacts the provider. An empty set is `none`, not unavailable.
- New `internal/integrations/capabilities/alerts/` adapter/mapping/normalize/sources.
- Providers:
  - `met_office`: `https://data.hub.api.metoffice.gov.uk/nswws/v1.1`, `apikey_header` (the header
    name is `apikey`), GeoJSON snapshot reached through one bounded Atom link discovery
    ([ADR-002](adr-002-alert-feed-discovery.md)).
  - `nws`: `https://api.weather.gov`, no key, stable `/alerts/active` JSON for US installations.
- Frontend presenter with severity-led card body, expiry display and plain-text instruction.

### Fuel prices capability (`fuel_prices/v1`, type `fuel_prices`)

- New `models/integrations/fuel.go`: origin label/coordinates, radius, currency, stations[] with
  id, name, brand, address, postcode, optional coordinates, distance in metres, prices in integer
  pence per litre per grade, price update time; attribution and freshness.
- New `internal/integrations/capabilities/fuel/` adapter/mapping/normalize/sources. Distance is
  computed in normalization with a haversine helper from a new optional `capabilities.Context`
  origin, then sorted by distance then cheapest displayed grade.
- Price validation: integer pence, 0–1000; invalid station prices are dropped, never rounded or
  zero-filled; a station with no valid price is shown without prices or omitted.
- Auth: new `oauth2_client_credentials` mode ([ADR-001](adr-001-oauth-client-credentials.md)):
  `token_url`, optional `scope`, private `client_id`/`client_secret` slots, in-memory token cache,
  one 401 retry, origin-bound forwarding, redacted errors. `RequiresCredential` returns true.
- Provider `backend/configs/integrations/fuel_prices/fuel_finder.yaml`. The Fuel Finder query
  shape, rate limits and terms are unverified until credentials exist (Q2); fixtures and the YAML
  are completed against a captured response.

### Cross-cutting backend changes

- `capabilities.Context` gains an optional origin (lat/lon) sourced from the setup checkpoint; no
  credential is added to the context.
- `Definition.AuthConfig` gains `token_url` and `scope`; compiler, request builder, credential
  lifecycle, definition editor and frontend contracts update together.
- `Definition.Request` gains an optional `discover` block (Atom link discovery) used by the Met
  Office provider; the compiler bounds it to one hop, same origin, public HTTPS.
- `Definition.Request` gains an optional bounded `headers` map for fixed, non-secret request
  headers. It is needed by NWS, which requires an identifying `User-Agent`. Only public constants
  are allowed: no credential material, no newlines, bounded count and length, and values are
  visible in the reviewed YAML.
- Registry registration in `service.go`; presenter registration in `integrations/registry.ts`.
- `contracts/integrations.ts`: extend `parseResource` dispatch, capability/type enums,
  `parseProviderDefinition` mapping branches, and add per-resource parsers.
- `backend/configs/integrations/embed.go` adds the new directories.

### Map (optional, phase 8)

Coordinate-bearing resources (weather location, fuel origin/stations, alert polygons) can be shown
on demand. Leaflet is loaded lazily on detail pages only, behind an explicit "Show map" action and
a disclosure naming the tile provider. No tiles, no map on Home, no coordinates in browser
storage, logs or URLs. See [ADR-003](adr-003-map-and-coordinate-presentation.md).
Warnings with missing or over-limit geometry are still listed; the map renders the rest.

## Alternatives considered

| Decision | Chosen | Rejected and why |
| --- | --- | --- |
| Home grouping | One section primitive; “Services” is only the default section | Per-type sections multiply headers and push areas down; bare blocks (current) have no context; folding all integrations into rooms scatters house-level data |
| Single-item section | Keep header | Auto-hide makes the page jump and removes add/overflow |
| Section layout | One grammar, geometry and view | Per-capability layouts duplicate CSS and break drag consistency |
| Catalog | Type → provider two-step | Flat provider list with per-type special cases (current) does not scale |
| Alerts UK source | Feed discovery, one hop | Hard-coding the snapshot UUID is stale by design; full XML engine is unnecessary |
| Fuel auth | OAuth client credentials | Manual bearer expires and is unusable; CSV export is twice daily and not an integration |
| Map | Opt-in Leaflet on detail | Static map URLs leak location too and add a vendor; a coordinate list alone answers the "where" question but not the spatial one |

## Constitution Check

Design review only; no execution evidence is claimed. Planned checks are listed under
Traceability.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 035, ADRs, provider research, prototypes; Q1–Q5 unresolved where dependent | PASS (open questions recorded) | Pending |
| NH-02 | Information-only; no device pipeline, command or automation input changes | N/A | Pending |
| NH-03 | Authenticated API only; bounded egress; secrets in existing slots; token never leaves server; coordinates treated as household data; opt-in tiles | PASS | Pending |
| NH-04 | Capability/presenter contracts; additive resources; ADRs for auth, discovery and dependency; API unchanged | PASS | Pending |
| NH-05 | Acceptance IDs mapped to tests/procedures; synthetic providers; no device commands | PASS (design) | Pending |
| NH-06 | Atomic last-good publication; bounded lists/polygons; freshness and expiry; retention/attribution; rollback notes | PASS | Pending |
| BE-01 | Adapters perform no I/O/storage; services and repositories unchanged | PASS | Pending |
| BE-02 | `DeviceLifetimeService` untouched | N/A | Pending |
| BE-03 | Existing worker/context ownership; OAuth token cache bounded; discovery one hop; race tests planned | PASS | Pending |
| BE-04 | Additive fields and new contract versions; no query/retention change; fixture compatibility | PASS | Pending |
| BE-05 | Existing authenticated routes; provider errors sanitized; no MCP/assistant change | PASS | Pending |
| FE-01 | Existing Vuex/services; presenters own capability mapping; new components are presentational | PASS | Pending |
| FE-02 | New resources validated at ingress; contracts owners updated; mocks updated | PASS | Pending |
| FE-03 | Loading/empty/error/stale states specified; no optimistic provider state | PASS | Pending |
| FE-04 | Tabs, search, drag alternatives, focus rings, 44px targets, 320–1440 checks | PASS (design) | Pending |
| FE-05 | No secrets in bundles; coordinates not persisted in browser; map opt-in and disclosed | PASS | Pending |

No exception is requested. Pending maintainer approval in Q1–Q3 is BLOCKED for the dependent
phases, not a claimed approval.

## Contracts, data, rollout

| Contract | Version | Change | Consumers |
| --- | --- | --- | --- |
| Weather resource | `weather_forecast/v1` | Additive optional `sunrise`, `sunset`, `uv`, `uv_max` | Existing weather results, Home card |
| Air quality resource | `air_quality/v1` | New ([contract](contracts/air_quality.v1.md)) | New presenter, Home card |
| Alerts resource | `weather_alerts/v1` | New ([contract](contracts/alerts.v1.md)) | New presenter, Home card |
| Fuel resource | `fuel_prices/v1` | New ([contract](contracts/fuel_prices.v1.md)) | New presenter, Home card |
| Provider definition | `nodeherder.integration/v1` | Add `oauth2_client_credentials`, `token_url`, `scope`, `discover`, `headers` | Compiler, editor, setup, runtime |
| Home layout | version 1 | Additive optional section `width`/`view`/`collapsed` and card `width`/`height`; no block shape change | Home, Panel, settings export |
| HTTP API | unchanged | No new routes; existing authenticated integrations endpoints only | Web client, tests |

- Persistence: no new database or bucket. New resources use the existing encrypted instance
  snapshots. `HomeBlock.Width`/`View`/`Collapsed` and `HomeCard.Width`/`Height` are optional;
  absent means the default.
- Migration: loading Home renders legacy `kind: card` blocks in the implicit Services section
  without writing. The first accepted save materialises the group. This mirrors the existing
  read-only legacy behavior in `home_cache.go`.
- Rollback: pre-redesign binaries decode the saved JSON into the older struct, silently drop the
  new `width`, `view`, `collapsed` and card `height` fields, and compute a different Home ETag. A
  rollback can therefore cause one revision conflict and would erase the new presentation state on
  the next save; restore an exported pre-change layout, as already documented for weather reference
  kinds. New capability resources are ignored by older binaries because their definitions are not
  installed; remove instances before rolling back if a definition remains.
- Deployment order: backend first (capabilities + definitions), then frontend. Existing waste and
  weather instances need no action. Provider YAML updates are seeded only when the host file is
  absent; existing host files require the reviewed-update flow.

## Traceability and verification

| Requirement/acceptance | Paths | Test/procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01–03 | `models/integrations/weather.go`, `capabilities/weather/*`, `configs/integrations/weather/*`, `types/weather.type.ts`, `contracts/weather.ts`, `integrations/weather/*` | Weather model validation table, fixture normalization, presenter render, legacy snapshot projection | Pending |
| FR-02, AC-04–07 | `models/integrations/airquality.go`, `capabilities/airquality/*`, `configs/integrations/air_quality/*`, `integrations/air-quality/*` | AQI band table per family, null/unknown/out-of-season fixtures, fake-provider setup/refresh, contract ingress | Pending |
| FR-03, AC-08–12 | `models/integrations/alerts.go`, `capabilities/alerts/*`, `configs/integrations/alerts/*`, `operations` discovery | Expiry projection, empty vs unavailable, discovery negatives (origin/size/parse/redirect), credential errors | Pending |
| FR-04, AC-13–17 | `models/integrations/fuel.go`, `capabilities/fuel/*`, `operations` OAuth, `configs/integrations/fuel_prices/*` | OAuth stub (token, 401 retry, origin binding, redaction), distance sort, price validation, error recovery | Pending |
| FR-05, AC-18–23 | `components/integrations/*`, `integrations/registry.ts`, `integration-icons.ts`, `types/integrations.type.ts` | Component tests, search/filter, status semantics, tabs vs stacked, no provider branch | Pending |
| FR-06, FR-09, FR-10, FR-11, FR-12, FR-13, FR-14, AC-24–28, 33–52 | `GroupDashboard.vue`, `HomeSection.vue`, `HomeCardFrame.vue`, `HomeCardRow.vue`, `HomeCardItem.vue`, `home-picker.ts`, `AddTilesDialog.vue`, `models/settings/home.go` | Layout schema tests (width/height/view/collapsed/legacy standalone), wide-section height check at 1024/1440, cards-vs-list geometry preservation, compact body bounds, collapse persistence, split action, picker tests, drag/reorder, viewport checks | Pending |
| FR-07, AC-29–32 | `components/integrations/IntegrationMap.vue`, `integrations/map/*`, ADR-003 | No-request-before-opt-in, disclosure, marker/list parity, unusable geometry fallback | Pending |
| NFR-01–05 | Capability validators, contracts, transport limits | Bounds tables, malformed fixtures, secret redaction, 320–1440 and keyboard checks | Pending |

Proposed component checks, run from the component directories:

- Backend: `go test ./...`, `go build ./...`, `go vet ./...`, race on
  `./internal/integrations/...` after shared changes.
- Frontend: `npm test -- --runInBand`, `npm run lint`, `npm run build`; browser checks for Home and
  Integrations at 320/375/768/1024/1440.
- Fixtures are synthetic; no household device, real provider account or live API key is used.
  Fuel Finder and Met Office live verification remain manual, credential-gated checks.

## Debt, risks, exceptions

- **Manual refresh (Q5).** Alerts, air quality and fuel prices are time-sensitive; without the
  deferred coordinator they are only as current as the last refresh. The UI must show the checked
  time prominently. Pulling scheduling forward is a scope decision, not silently included.
- **Met Office discovery.** ADR-002 adds one narrow transport capability. Risk: a discovered URL
  could target another origin. Mitigation: one hop, same declared origin, public HTTPS, bounded
  body, no redirects, no recursion, failure preserves last-good.
- **Fuel Finder.** Credentials, rate limits and redistribution terms are unverified (Q2). Risk:
  the provider YAML may need adjustment after a fixture capture. Mitigation: complete fixtures
  before claiming acceptance; keep the capability independent of the provider.
- **Map dependency.** Leaflet + OSM tiles add a runtime dependency and disclose approximate
  location to the tile provider. Mitigation: lazy load, explicit opt-in, no Home map, no browser
  persistence; ADR-003 records the cost.
- **Open-Meteo licensing.** Free hosted use is non-commercial; attribution is mandatory for both
  forecast and air quality. The UI must show it.
- **WeatherAPI retention.** The unresolved 24-hour retention limitation from spec 034 remains; this
  plan must not add new providers that depend on that data.
- **Home layout change.** Risk of losing placements or breaking drag on touch. Mitigation:
  additive size field, render-only implicit section, regression tests over saved fixtures, and
  rollback guidance.
- **Payload bounds.** Alert polygons and station lists need explicit limits (for example 2000
  vertices, 100 stations, 7-day air-quality horizon) to keep snapshots and the browser bounded.
- **No automation integration.** None of these signals can trigger an automation. This is a known
  product limitation, not delivered here.

## Completion

Acceptance evidence, actual check results, changed documentation and prototype links are added
here when implementation runs. Unresolved: Q1–Q5, live provider verification, and the deferred
scheduling/data-policy work inherited from spec 034.
