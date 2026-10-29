# Tasks: Home signal integrations and scalable integration UI

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: draft
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
- [ ] T004 — Confirm Q4 (civil alerts = weather warnings only) and Q5 (manual refresh this phase).
  IDs: scope, FR-03, NFR-05. Done: recorded answers in `spec.md`.
- [ ] T005 — Record synthetic fixtures for Open-Meteo forecast (sunrise/sunset/UV), Open-Meteo air
  quality, Met Office NSWWS feed + snapshot, NWS alerts and Fuel Finder. IDs: AC-01–17. Depends:
  T002. Done: fixtures under test data with no real address or key.

## Phase 1 — Shared backend extensions

- [ ] T006 — Add optional origin to `capabilities.Context` and source it from the setup checkpoint;
  keep credentials out. IDs: AC-14, BE-01/NH-03. Paths:
  `backend/internal/integrations/capabilities/registry.go`, `runtime/*`. Done: existing waste and
  weather adapters compile and pass; origin absent for non-location providers.
- [ ] T007 — Add `oauth2_client_credentials` to `AuthConfig`, compiler, request builder, credential
  lifecycle, definition editor contracts and `RequiresCredential`. IDs: AC-16–17, FR-04, NH-03.
  Paths: `backend/models/integrations/definition.go`,
  `backend/internal/integrations/{definitions,operations,runtime}/`,
  `frontend/src/types/integrations.type.ts`, `contracts/integrations.ts`. Depends: T002.
  Done: token fetch, single 401 retry, origin binding, redaction and existing auth regressions pass.
- [ ] T008 — Add bounded one-hop Atom link discovery and a fixed non-secret `headers` map to
  `Request` and the operation executor. IDs: AC-11, FR-03. Paths: `definition.go`,
  `definitions/compiler.go`, `operations/{request,authentication}.go`. Depends: T001. Done:
  same-origin follow accepted; foreign origin, redirect, oversize body, missing link and parse
  error rejected with tests; `headers` accepts only bounded public values and rejects secret-like
  or multiline input.
- [ ] T009 — Extend frontend ingress dispatch and provider-definition validation for new
  capabilities, auth modes and request fields. IDs: FR-02–05, FE-02. Paths:
  `frontend/src/contracts/integrations.ts`, `types/integrations.type.ts`. Depends: T006–T008.
  Done: unknown capability still rejects; all fixtures parse; existing waste/weather unchanged.

## Phase 2 — Weather signals

- [ ] T010 — Extend the weather model, mapping and normalization with sunrise, sunset, daily UV max
  and hourly UV, including validation and nullable semantics. IDs: AC-01–03, FR-01, NH-06.
  Paths: `backend/models/integrations/weather.go`,
  `backend/internal/integrations/capabilities/weather/{mapping,normalize}.go`. Done: model and
  fixture tests pass; legacy snapshot projects unchanged.
- [ ] T011 — Update shipped provider YAML where fixtures confirm the fields. IDs: AC-01–02,
  FR-01. Paths: `backend/configs/integrations/weather/*.yaml`. Done: compiler tests and definition
  reload pass; absent fields stay null.
- [ ] T012 — Render sunrise, sunset and UV in results and the compact card, omitting nulls.
  IDs: AC-01–03, NFR-04. Paths: `frontend/src/types/weather.type.ts`,
  `contracts/weather.ts`, `components/integrations/WeatherResults.vue`,
  `integrations/weather/presentation.ts`. Depends: T010–T011. Done: presenter/UI tests pass at
  320/1440.

## Phase 3 — Air quality and pollen

- [ ] T013 — Add `air_quality/v1` model with server-derived category bands, bounded series,
  pollen and freshness. IDs: AC-05–07, FR-02, NH-06. Paths:
  `backend/models/integrations/airquality.go`. Done: band table per family, null/unknown and
  out-of-season cases covered.
- [ ] T014 — Implement and register the air-quality adapter, mapping and sources. IDs: AC-04–07.
  Paths: `backend/internal/integrations/capabilities/airquality/`,
  `internal/integrations/service.go`. Done: fake-provider setup/refresh and shared-registry tests.
- [ ] T015 — Ship `air_quality/open_meteo.yaml` and the embed entry. IDs: AC-04, FR-02, FR-08.
  Paths: `backend/configs/integrations/air_quality/`, `configs/integrations/embed.go`. Done:
  compile, reload and location-select tests.
- [ ] T016 — Add the frontend presenter, results and sources. IDs: AC-05–07, NFR-04. Paths:
  `frontend/src/integrations/air-quality/`, `integrations/registry.ts`, contracts. Depends:
  T013–T015. Done: ingress, widget, Home source and viewport checks.

## Phase 4 — Severe weather alerts

- [ ] T017 — Add `weather_alerts/v1` model and projection semantics (expiry, ordering, empty vs
  unavailable, bounded polygon). IDs: AC-08–10, 32, FR-03, NH-06. Paths:
  `backend/models/integrations/alerts.go`. Done: projection and geometry-bound tests.
- [ ] T018 — Implement and register the alerts adapter, mapping and sources. IDs: AC-08–12.
  Paths: `backend/internal/integrations/capabilities/alerts/`, `service.go`. Done: fixtures for
  active/none/expired/malformed geometry.
- [ ] T019 — Ship Met Office and NWS provider YAML; wire discovery. IDs: AC-11–12, FR-03, FR-08.
  Paths: `backend/configs/integrations/alerts/`. Depends: T008, T001. Done: feed + snapshot
  fixtures, credential error and 403 handling.
- [ ] T020 — Add the frontend alerts presenter with severity-led body and plain-text instruction.
  IDs: AC-09–12, NFR-04. Paths: `frontend/src/integrations/alerts/`, `integrations/registry.ts`.
  Depends: T017–T019. Done: severity/level/expiry rendering, no markup injection.

## Phase 5 — Fuel prices

- [ ] T021 — Add `fuel_prices/v1` model with station, grade prices, distance and origin. IDs:
  AC-14–15, FR-04, NH-06. Paths: `backend/models/integrations/fuel.go`. Done: price range,
  integer-only and bounded-list tests.
- [ ] T022 — Implement the fuel adapter with haversine distance and deterministic sorting.
  IDs: AC-14–15. Paths: `backend/internal/integrations/capabilities/fuel/`, `service.go`.
  Depends: T006, T021. Done: sort/rounding/absent-price fixtures pass.
- [ ] T023 — Ship `fuel_prices/fuel_finder.yaml` using OAuth client credentials. IDs: AC-13,
  16–17, FR-04, FR-08. Paths: `backend/configs/integrations/fuel_prices/`. Depends: T002, T007.
  Done: recorded fixture-driven YAML, credential and reconnect tests.
- [ ] T024 — Add the frontend fuel presenter (cheapest stations, distance, grades). IDs: AC-14,
  17, NFR-04. Paths: `frontend/src/integrations/fuel/`, `integrations/registry.ts`. Done:
  ingress, sorting display and error/stale states.

## Phase 6 — Integrations page redesign

- [ ] T025 — Normalize catalog metadata (type icon, configured count, worst status, provider
  credential/limit text) and expose it additively. IDs: AC-18–23, FR-05, BE-05. Paths:
  `backend/internal/integrations/runtime/{dto,definitions}.go`, `frontend/src/types/integrations.type.ts`.
  Done: catalog contract tests; no private data in the DTO.
- [ ] T026 — Replace the flat catalog with type-first discovery and provider rows. IDs:
  AC-18, 22–23. Paths: `ConnectionCatalogDialog.vue`, new `IntegrationCatalog.vue`,
  `IntegrationTypeCard.vue`, `integration-icons.ts`. Done: component tests, no provider branch.
- [ ] T027 — Replace board widgets with grouped, searchable entry rows and a status model.
  IDs: AC-19–21, 23. Paths: `ConnectionsPage.vue`, `ConnectionBoard.vue`,
  `ConnectionWidget.vue`/`IntegrationEntry.vue`. Done: search, grouping, empty result and status
  semantics tests.
- [ ] T028 — Convert detail sections to responsive tabs while keeping deep links and anchors.
  IDs: AC-21, NFR-04. Paths: `ConnectionDetail.vue`, `IntegrationSections.vue`,
  `IntegrationPageHeader.vue`. Done: tab/stacked behaviour at 768 and keyboard checks.
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
  geometry; “+N more” appears past two rows; switching restores geometry.
- [ ] T033 — Implement width/height and section-width controls, the bounded span grid and
  drag/keyboard parity. IDs: AC-27, 37–40, 45–47, 50–51, NFR-04. Paths: `HomeSection.vue`,
  `CardSizeControl.vue`, `SectionWidthControl.vue`, `GroupDashboard.vue`, `HomeCardList.vue`,
  theme CSS. Done: 320/375/768/1024/1440 checks; smallest fitting span on add; no card exceeds
  its span; touch and keyboard reorder; persisted order, width, height and section width.
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
  `integrations/map/*`, `package.json`. Depends: T003. Done: network observation shows no tile
  request before opt-in; coordinates absent from storage and URLs.
- [ ] T037 — Add markers/polygons and list parity for weather, fuel and alerts; geometry fallback.
  IDs: AC-30, 32. Paths: presenters and map helpers. Depends: T036, T016, T020, T024. Done:
  marker selection mirrors list rows; unusable geometry does not fail the map.

## Phase 9 — Verification and handoff

- [ ] T038 — Backend checks: `go test ./...`, `go build ./...`, `go vet ./...`, race on
  `./internal/integrations/...` and `./models/settings/...`. Gate: NH-05/BE-03. Depends:
  T006–T024, T030. Done: exact commands and results recorded.
- [ ] T039 — Frontend checks: `npm test -- --runInBand`, `npm run lint`, `npm run build`. Gate:
  NH-05/FE-02. Depends: T025–T037. Done: actual results, including pre-existing warnings.
- [ ] T040 — Browser checks at 320/375/768/1024/1440 for Home sections, add-card sheet,
  Integrations catalog, entries and detail tabs. Gate: FE-04/NFR-04. Depends: T039. Done:
  recorded desktop/mobile runs against mock servers only.
- [ ] T041 — Update documentation: `docs/connections.md` (new capabilities, attribution, limits),
  provider research, ADR outcomes and rollback notes. Gate: NH-01/NH-04. Depends: T038–T040.
  Done: docs match shipped behaviour; unresolved limits stated.
- [ ] T042 — Final diff review against NH/BE/FE rules; report passed, failed and unrun checks.
  Gate: all. Depends: T041. Done: review record with unresolved findings and no claimed approval.

## Out of scope (recorded, not tasks)

- Scheduled provider refresh and the shared freshness/retention policy service (spec 034 debt).
- Integration values as automation triggers or conditions.
- Civil emergency alerts, radar/satellite imagery, historical charts of integration data.
- Public API, MQTT or MCP exposure.
