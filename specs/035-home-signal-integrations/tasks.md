# Tasks: Home signal integrations and scalable integration UI

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft — revised 2026-11-01
Each task maps to requirement/acceptance IDs or a constitutional gate. Mark done only with
evidence. No task authorizes device commands, live provider calls or a committed credential.

## Phase 0 — Decisions and fixtures

- [ ] T001 — Resolve Q1: approve or reject bounded alert feed discovery
  ([ADR-002](adr-002-alert-feed-discovery.md)). IDs: AC-11, FR-03, NH-04/NH-06. Owner: maintainer.
  Done: recorded decision; if rejected, Met Office alerts move out of scope.
- [ ] T002 — Resolve Q2 and capture Fuel Finder fixtures: auth method, endpoint shape, area query,
  rate limits, terms. IDs: AC-13–17, FR-04. Owner: household maintainer. Done: redacted request
  and response fixtures plus numeric limits, no live key in the repo.
- [ ] T003 — Resolve Q3: approve or reject the map dependency and disclosure
  ([ADR-003](adr-003-map-and-coordinate-presentation.md)). IDs: AC-29–32, FR-07, NH-04/FE-05.
  Done: recorded decision; if rejected, US-07 is removed and coordinates remain list-only.
- [ ] T004 — Confirm Q4 (civil alerts = weather warnings only) and the initial cadences/quotas for
  air quality, alerts and fuel (Q5 resolved by spec 037; the scheduler is not reimplemented).
  IDs: scope, FR-03, FR-15, NFR-05. Done: recorded answers in `spec.md`.
- [ ] T005 — Record synthetic fixtures for Open-Meteo forecast (sunrise/sunset/UV), Open-Meteo air
  quality, Met Office NSWWS feed + snapshot, NWS alerts and Fuel Finder. IDs: AC-01–17. Depends:
  T002. Done: fixtures under test data with no real address or key.

## Phase 1 — Shared backend extensions

- [x] T006 — Add optional origin to `capabilities.Context` and source it from the setup checkpoint;
  keep credentials out. IDs: AC-14, BE-01/NH-03. Paths:
  `backend/internal/integrations/capabilities/registry.go`, `runtime/*`. Done: existing waste and
  weather adapters compile and pass; origin absent for non-location providers.
- [x] T007 — Add `oauth2_client_credentials` to `AuthConfig`, compiler, request builder, credential
  lifecycle, definition editor contracts and `RequiresCredential`. IDs: AC-16–17, FR-04, NH-03.
  Paths: `backend/models/integrations/definition.go`,
  `backend/internal/integrations/{definitions,operations,runtime}/`,
  `frontend/src/types/integrations.type.ts`, `contracts/integrations.ts`. Depends: T002.
  Done: token fetch, single 401 retry, origin binding, redaction and existing auth regressions pass.
- [x] T008 — Add bounded one-hop Atom link discovery and a fixed non-secret `headers` map to
  `Request` and the operation executor. IDs: AC-11, FR-03. Paths: `definition.go`,
  `definitions/compiler.go`, `operations/{request,authentication}.go`. Depends: T001. Done:
  same-origin follow accepted; foreign origin, redirect, oversize body, missing link and parse
  error rejected with tests; `headers` accepts only bounded public values and rejects secret-like
  or multiline input.
- [x] T009 — Extend frontend ingress dispatch and provider-definition validation for new
  capabilities, auth modes and request fields, and validate catalog provider `id`,
  `integration_type`, `origin` and `disk_error`; keep the required policy refresh fields.
  IDs: FR-02–05, FR-15, FE-02. Paths: `frontend/src/contracts/integrations.ts`,
  `types/integrations.type.ts`. Depends: T006–T008. Done: unknown capability still rejects; all
  fixtures parse; existing waste/weather unchanged.

## Phase 2 — Weather signals

- [x] T010 — Extend the weather model, mapping and normalization with sunrise, sunset, daily UV max,
  hourly UV and richer daily values (feels-like, gust, direction, humidity, pressure, visibility,
  night condition, snow chance), including validation and nullable semantics. IDs: AC-01–03, 54,
  FR-01, NH-06. Paths: `backend/models/integrations/weather.go`,
  `backend/internal/integrations/capabilities/weather/{mapping,normalize}.go`. Done: model and
  fixture tests pass; legacy snapshot projects unchanged.
- [x] T011 — Update shipped provider YAML where fixtures confirm the fields, keeping
  `refresh_default_seconds` and `refresh_min_seconds`. Met Office uses the confirmed
  `midday10MWindSpeed`/`midday10MWindGust`/`midday10MWindDirection`, `middayRelativeHumidity`,
  `middayMslp`, `middayVisibility`, `maxUvIndex`, feels-like, snow-probability and night-condition
  selectors; Open-Meteo and WeatherAPI add their confirmed equivalents. IDs: AC-01–02, 54, FR-01,
  FR-15. Paths: `backend/configs/integrations/weather/*.yaml`. Done: compiler tests and definition
  reload pass; absent fields stay null.
- [x] T012 — Render the compact daily card (date, condition icon, high/low, rain/wind/UV chips)
  and the selected-day summary (sunrise, sunset, maximum UV and the full daily values), omitting
  nulls. IDs: AC-01–03, NFR-04. Paths: `frontend/src/types/weather.type.ts`,
  `contracts/weather.ts`, `components/integrations/WeatherResults.vue`, `WeatherDayCard.vue`,
  `integrations/weather/{presentation,day-card,day-detail}.ts`. Depends: T010–T011. Done:
  presenter/UI tests pass at 320/1440.
- [x] T012A — Add per-day forecast detail: selecting a daily card replaces the "Next 24 hours"
  panel with that day's summary and the same hourly-card view, including every supplied daily stat,
  hourly points for the local date, null omission, the provider's hourly horizon when the day is
  outside it, `aria-pressed` selection, a "Today" control for other days and same-day toggle-off.
  IDs: AC-54–55,
  FR-16, NFR-04. Paths: `WeatherResults.vue`, `WeatherStatGrid.vue`, `domain/weather.ts`
  (`hoursForDate`, `weatherHourlyHorizonHours`), `integrations/weather/day-detail.ts`,
  `integrations/weather/stats.ts`. Depends: T012. Done: selection/restore keyboard tests, shared
  stat tiles with icons, current fallback, hourly-by-date filtering, horizon message, mobile
  pagination.
- [x] T012B — Add inline hourly detail: hourly cards are selectable, expose `aria-pressed`, and show
  that hour's supplied temperature, condition, feels-like, rain, precipitation, humidity, wind,
  gusts, direction and UV in the shared stat grid below the strip; day switch/Today clears it.
  IDs: AC-56, FR-17, NFR-04. Paths: `WeatherResults.vue`, `WeatherStatGrid.vue`,
  `integrations/weather/{day-detail,stats}.ts`. Depends: T012A. Done: hour toggle/selection tests,
  point stats, null omission, day-switch reset, keyboard checks.

## Phase 3 — Air quality and pollen

- [x] T013 — Add `air_quality/v1` model with server-derived category bands, bounded series,
  pollen and freshness. IDs: AC-05–07, FR-02, NH-06. Paths:
  `backend/models/integrations/airquality.go`. Done: band table per family, 168-hour bound,
  top-level pollen, six-hour freshness, null/unknown and out-of-season cases covered.
- [x] T014 — Implement and register the air-quality adapter, mapping and sources. IDs: AC-04–07.
  Paths: `backend/internal/integrations/capabilities/airquality/`,
  `internal/integrations/service.go`. Done: fake-provider normalization/projection/sources,
  bootstrap registry and full backend tests.
- [x] T015 — Ship `air_quality/open_meteo.yaml` with provisional
  `refresh_default_seconds: 1800`/`refresh_min_seconds: 900` and the embed entry.
  IDs: AC-04, FR-02, FR-08, FR-15. Paths: `backend/configs/integrations/air_quality/`,
  `configs/integrations/embed.go`. Done: compile, reload, location-select and catalog tests.
- [x] T016 — Add the frontend presenter, results and sources. IDs: AC-05–07, NFR-04. Paths:
  `frontend/src/integrations/air-quality/`, `integrations/registry.ts`, contracts. Depends:
  T013–T015. Done: ingress, widget, Home source and presenter tests; browser viewport checks
  remain in T040.

## Phase 4 — Severe weather alerts

- [x] T017 — Add `weather_alerts/v1` model and projection semantics (expiry, ordering, empty vs
  unavailable, bounded polygon). IDs: AC-08–10, 32, FR-03, NH-06. Paths:
  `backend/models/integrations/alerts.go`. Done: projection, ordering, duplicate-id and
  geometry-bound tests.
- [x] T018 — Implement and register the alerts adapter, mapping and sources. IDs: AC-08–12.
  Paths: `backend/internal/integrations/capabilities/alerts/`, `service.go`. Done: fixtures for
  active/none/expired/malformed geometry, sanitized text and projection.
- [x] T019 — Ship Met Office and NWS provider YAML with provisional
  `refresh_default_seconds: 900`/`refresh_min_seconds: 300`; wire discovery.
  IDs: AC-11–12, FR-03, FR-08, FR-15. Paths: `backend/configs/integrations/alerts/`.
  Done: Met Office Atom discovery + API-key seed, NWS fixed-header/free seed, compile and
  discovery tests. Live credential verification remains deployment-specific.
- [x] T020 — Add the frontend alerts presenter with severity-led body and plain-text instruction.
  IDs: AC-09–12, NFR-04. Paths: `frontend/src/integrations/alerts/`, `integrations/registry.ts`.
  Depends: T017–T019. Fixture-verified; live provider remains T019. Done: severity/level/expiry
  rendering, no markup injection and clear/stale states.

## Phase 5 — Fuel prices

- [x] T021 — Add `fuel_prices/v1` model with station, grade prices, distance and origin. IDs:
  AC-14–15, FR-04, NH-06. Paths: `backend/models/integrations/fuel.go`. Done: price range,
  integer-only, coordinate and bounded-list tests.
- [x] T022 — Implement the fuel adapter with haversine distance and deterministic sorting.
  IDs: AC-14–15. Paths: `backend/internal/integrations/capabilities/fuel/`, `service.go`.
  Depends: T006, T021. Done: haversine, rounding, absent-price and duplicate-grade fixtures.
- [x] T023 — Ship `fuel_prices/fuel_finder.yaml` using OAuth client credentials with provisional
  `refresh_default_seconds: 3600`/`refresh_min_seconds: 900`. IDs: AC-13, 16–17, FR-04, FR-08,
  FR-15. Paths: `backend/configs/integrations/fuel_prices/`. Done: JSON-body OAuth token flow,
  two credential slots, bounded pagination, station/price join and compile/normalization tests.
  Live credential verification remains deployment-specific.
- [x] T024 — Add the frontend fuel presenter (cheapest stations, distance, grades). IDs: AC-14,
  17, NFR-04. Paths: `frontend/src/integrations/fuel/`, `integrations/registry.ts`.
  Fixture-verified; live provider remains T023. Done: ingress, sorting display and error/stale
  states.

## Phase 6 — Integrations page redesign

- [x] T025 — Normalize catalog metadata (type icon, configured count, worst status, provider
  credential/limit text) and expose it additively, including provider `id`, `integration_type`,
  `origin` and `disk_error`. IDs: AC-18–23, FR-05, FR-15, BE-05. Paths:
  `backend/internal/integrations/runtime/{dto,definitions}.go`, `frontend/src/types/integrations.type.ts`.
  Done: catalog contract/domain tests; no private data in the DTO.
- [x] T026 — Replace the flat catalog with type-first discovery and provider rows. IDs:
  AC-18, 22–23. Paths: `ConnectionCatalogDialog.vue`, new `IntegrationCatalog.vue`,
  `IntegrationTypeCard.vue`, `integration-icons.ts`. Done: type-first cards, search, provider
  policy text and an explicit no-provider state; no provider branch.
- [x] T027 — Replace board widgets with grouped, searchable entry rows and a status model that
  shows schedule state when present. IDs: AC-19–21, 23, AC-53. Paths: `ConnectionsPage.vue`,
  `ConnectionBoard.vue`, `ConnectionWidget.vue`/`IntegrationEntry.vue`. Done: grouping, search,
  empty result, schedule/status semantics and arrow reorder.
- [ ] T028 — Convert detail sections to responsive tabs while keeping deep links, anchors and the
  automatic-refresh editor's dirty draft/base ETag. IDs: AC-21, AC-53, NFR-04. Paths:
  `ConnectionDetail.vue`, `IntegrationSections.vue`, `IntegrationPageHeader.vue`,
  `ConnectionSettings.vue`, `composables/integrations/useConnectionDetail.ts`. Done: tab/stacked
  behaviour at 768, schedule draft survives switches, keyboard checks.
- [ ] T029 — Update mocks, fixtures and browser cases for the new catalog/entry semantics.
  IDs: AC-18–23, FE-02. Paths: `frontend/src/__tests__/`, browser fixtures. Done: affected cases
  pass against mock servers.

## Phase 7 — Home sections and card model

- [ ] T030 — Add optional `width`/`height` to `HomeCard` and optional `width`/`view`/`collapsed` to
  `HomeBlock`, with validation and unchanged block shape. IDs: AC-39, 43, 45, 48–49, FR-10,
  FR-12–14, BE-04/NH-06. Paths: `backend/models/settings/home.go`,
  `models/settings/home_test.go`, `frontend/src/types/home.type.ts`.
  Done: old layout fixture loads, invalid values reject, ETag behaviour recorded.
- [ ] T031 — Render every block as a section, with the implicit Services section for legacy `card`
  blocks, two-column Services width, a persistent collapse header and a per-section view switch.
  IDs: AC-33–36, 40, 43–44, 48–49, FR-09, FR-12, FR-14. Paths: `GroupDashboard.vue`, new
  `HomeSection.vue`, `AreaCard.vue`/`HomeStandaloneCard.vue` removal, `HomeBlockHeader.vue`.
  Done: saved-layout fixtures render unchanged; no write on load; collapsed and view state persist.
- [ ] T032 — Introduce the shared card frame, the list-row renderer and compact body dispatch, with
  capability-owned bodies and bounded rows. IDs: AC-25, 36, 41, 50–51, FR-06, FR-11, FR-14. Paths:
  `HomeCardItem.vue`, new `HomeCardFrame.vue`, new `HomeCardRow.vue`, `IntegrationTile.vue` → body,
  `EntityTile.vue` content. Done: cards and rows share identity and destination; list ignores
  geometry; “+N more” appears past two rows; switching restores geometry. Provider devices,
  read-only entities and stateless buttons from spec 038 render through the same frame with no
  provider branch, and Zigbee tiles are unchanged.
- [ ] T033 — Implement width/height and section-width controls, the bounded span grid and
  drag/keyboard parity. IDs: AC-27, 37–40, 45–47, 50–51, NFR-04. Paths: `HomeSection.vue`,
  `CardSizeControl.vue`, `SectionWidthControl.vue`, `GroupDashboard.vue`, `HomeCardList.vue`,
  theme CSS. Done: 320/375/768/1024/1440 checks; smallest fitting span on add, including a long
  provider label and a stateless button; no card exceeds its span; touch and keyboard reorder;
  persisted order, width, height and section width.
- [ ] T034 — Replace the two-step add dialog with one searchable, grouped add-card sheet; support
  creating a named section with one card and the explicit “Split into sections by type” action.
  IDs: AC-26, 28, 42, 48. Paths: `AddTilesDialog.vue`, `domain/home-picker.ts`,
  `MoveCardDialog.vue`, `HomeSection.vue`. Done: search, grouping, already-added state, single vs
  multi selection, keyboard operation; one-card sections keep their header; split creates ordinary
  sections once and never re-splits.
- [ ] T035 — Empty and single-item section states, non-empty section deletion with move-or-delete
  choice, no auto-hide and no auto-delete. IDs: AC-35, 48, 52. Paths: `HomeSection.vue`,
  confirmation dialog. Done: component tests and visual checks; default action moves cards.

## Phase 8 — Map (only if Q3 approves)

- [ ] T036 — Add the lazy map component, disclosure and no-request-before-opt-in behaviour.
  IDs: AC-29–31, FR-07, FE-05. Paths: `components/integrations/IntegrationMap.vue`,
  `integrations/map/*`, `package.json`, `adr-003-map-and-coordinate-presentation.md`.
  Depends: T003. Done: network observation shows no tile request before opt-in; coordinates
  absent from storage and URLs; ADR-003 distinguishes the geographic map from spec 038's indoor
  spatial canvas, which is neither reused nor modified.
- [ ] T037 — Add markers/polygons and list parity for weather, fuel and alerts; geometry fallback.
  IDs: AC-30, 32. Paths: presenters and map helpers. Depends: T036, T016, T020, T024. Done:
  marker selection mirrors list rows; unusable geometry does not fail the map.

## Phase 9 — Verification and handoff

- [x] T038 — Backend checks: `go test ./...`, `go build ./...`, `go vet ./...`, race on
  `./internal/integrations/...` and `./models/settings/...`. Gate: NH-05/BE-03. Depends:
  T006–T024, T030. Done: exact commands and results recorded.
- [x] T039 — Frontend checks: `npm test -- --runInBand`, `npm run lint`, `npm run build`. Gate:
  NH-05/FE-02. Depends: T012A, T025–T037. Done: actual results, including pre-existing warnings.
- [ ] T040 — Browser checks at 320/375/768/1024/1440 for Home sections, add-card sheet,
  Integrations catalog, entries, detail tabs and weather day detail. Gate: FE-04/NFR-04. Depends:
  T039. Done: recorded desktop/mobile runs against mock servers only.
- [ ] T041 — Update documentation: correct the stale `docs/connections.md` introduction ("first
  release supports UK bin collections"; "Nothing is added to Home, Overview, Devices or Panel"),
  and document each new capability's schedule, freshness window, attribution and quota limit,
  the weather daily-detail interaction, provider research, ADR outcomes and rollback notes.
  Gate: NH-01/NH-04, FR-15. Depends:
  T038–T040. Done: docs match shipped behaviour; unresolved limits stated.
- [ ] T042 — Final diff review against NH/BE/FE rules; report passed, failed and unrun checks,
  including the schedule tab-switch draft check, a Home provider-device/button fixture and the
  weather day-detail keyboard check. Gate: all. Depends: T041. Done: review record with unresolved
  findings and no claimed approval.

## Out of scope (recorded, not tasks)

- The shared freshness/retention policy service (spec 034 debt); automatic refresh itself is
  delivered by spec 037 and reused.
- A geographic map is separate from spec 038's indoor spatial canvas; neither is changed here.
- Integration values as automation triggers or conditions.
- Civil emergency alerts, radar/satellite imagery, historical charts of integration data.
- Public API, MQTT or MCP exposure.
