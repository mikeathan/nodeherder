# Spec: Home signal integrations and scalable integration UI

ID: 035 | Status: Draft | Owner: NodeHerder contributors | Updated: 2026-10-29
Constitution: proposed 1.0.0 | Components: backend and frontend

## Problem, scope, exclusions

Households need the external signals that drive everyday decisions, not only bin days and a
forecast. Connections currently ships `waste_schedule/v1` and `weather_forecast/v1`; weather
already carries most day-to-day quantities but omits sunrise/sunset and UV, and there is no air
quality, pollen, official weather-warning or fuel-price capability. At the same time the
Integrations catalog already needs a weather-specific branch in
`frontend/src/components/integrations/ConnectionCatalogDialog.vue` to hide four providers behind
one type. Adding three capabilities and roughly ten providers would make the flat catalog and
the current Home tiles unusable.

Observed current behavior:

- `backend/internal/integrations/capabilities/registry.go` installs exactly two handlers;
  `frontend/src/integrations/registry.ts` installs the matching two presenters.
- `frontend/src/contracts/integrations.ts` `parseResource` falls back to the waste parser for any
  unknown contract version, so a new capability resource is rejected as malformed waste.
- `parseProviderDefinition` hard-codes `waste_schedule/v1|weather_forecast/v1` and
  `waste_collection|weather`, and branches on capability for the mapping shape.
- The catalog response already carries `types` (`ConnectionCatalogType`) but
  `ConnectionCatalogDialog.vue` renders types and providers ad hoc, with a `weather` special case.
- Home has two visually different card implementations: `EntityTile.vue` (devices) and
  `IntegrationTile.vue` (integrations), dispatched from `HomeCardItem.vue`.
- `GroupDashboard.vue` lays areas out on a CSS grid with fixed `grid-column: span 2` for areas and
  integration blocks; there is no per-card width or height model.
- A recorded Home screen shows named area sections ("Living room", "Attic room", "Garden",
  "Kitchen") interrupted by bare "Bin collections" and "Weather" cards at the same top-level
  position, with no section header. They read as ungrouped orphans and their wider card also
  changes the row rhythm.
- There is no map or coordinate presentation anywhere in `frontend/src`.

In scope:

1. Add sunrise, sunset and UV to the weather resource and its shipped providers.
2. Add an `air_quality/v1` capability with pollutants, AQI and pollen.
3. Add a `weather_alerts/v1` capability with official warnings.
4. Add a `fuel_prices/v1` capability for nearby forecourts.
5. Redesign the Integrations page and catalog around integration types with providers nested
   inside them, and give configured entries a grouped, searchable, status-led list.
6. Redesign Home as Home Assistant-style sections with one shared card anatomy and a searchable
   add-card sheet. Every card belongs to a section, including integrations: new integration cards
   land in a default, renameable "Services" section, and a card may later be moved into any area
   section. A section may hold one card or many and switch between card and list views.
7. Add an optional, opt-in map on integration detail pages.

Out of scope: automatic or scheduled provider refresh (the deferred coordinator from
`specs/034-weather-integration/plan.md`), integration values as automation triggers or conditions,
device control, radar/satellite imagery, civil emergency alerts (there is no public API; the
Emergency Alerts service is broadcast-only), historical charts of integration data, multi-household
tenancy, and non-JSON provider formats beyond the single bounded Atom link discovery in
[ADR-002](adr-002-alert-feed-discovery.md).

Assumptions to validate:

- A1. The household keeps the existing public-HTTPS-egress constraint; no private or metadata
  destinations are introduced.
- A2. A UK Fuel Finder client credential can be obtained (GOV.UK One Login plus client
  credentials). Unverified; see Q2.
- A3. Met Office warnings require feed discovery because the GeoJSON snapshot URL embeds a
  per-update UUID. See [provider research](provider-research.md) and Q1.
- A4. Manual refresh remains the update model for this phase, so alert and air-quality data are
  only as current as the last explicit refresh. See Q5.
- A5. Existing homes accept a Home Assistant-style section layout; no user has asked to keep the
  current area/standalone-card composition unchanged.

## Scenarios and acceptance

### US-01 — Weather gains daily planning signals (priority P1)

Independent verification: synthetic provider fixtures, model validation tests, presenter tests.

- AC-01: Given a refreshed weather resource whose provider supplies sunrise, sunset and UV, when
  the connection results render, then each daily row shows local sunrise and sunset times and the
  daily maximum UV, and hourly detail shows the UV value, all labelled with their meaning.
- AC-02: Given a provider that does not supply those quantities, when the resource is normalized,
  then the fields are null, the resource is accepted, and the UI omits them rather than showing
  zero; a UV value outside 0–20, an invalid local time or a negative daily maximum is rejected as
  `invalid_response`.
- AC-03: Given a weather snapshot saved before this change, when it is projected, then it remains
  readable with no migration and renders without the new fields.

### US-02 — Air quality and pollen (priority P1)

Independent verification: fake-provider setup and refresh tests, contract tests, viewport checks.

- AC-04: Given the air-quality catalog, a user selects a provider, searches a town or UK postcode,
  explicitly confirms a location and saves; the wizard discloses every search recipient.
- AC-05: Given a valid response, results show the index category, the dominant pollutant when the
  provider supplies it, PM2.5, PM10, ozone, NO2, SO2 and CO with units, and every pollen species
  the provider returned, in the resource timezone.
- AC-06: The index band (good/fair/moderate/poor/very poor/extremely poor for the European index,
  good/moderate/unhealthy-sensitive/unhealthy/very unhealthy/hazardous for the US index) is
  derived on the server from the declared index family; a missing index yields `unknown` and the
  UI says so, and the browser does not invent band thresholds.
- AC-07: Given an out-of-season pollen value, a null pollutant or a provider failure, then missing
  values are omitted and a failed refresh preserves and labels last-good data; out-of-season pollen
  is never shown as zero.

### US-03 — Official severe weather warnings (priority P1)

Independent verification: recorded GeoJSON fixtures, feed-discovery rejection tests, projection tests.

- AC-08: Given a refresh with no active warnings, the resource reports an explicit `none` state,
  distinct from an unavailable provider or a failed refresh.
- AC-09: Each active warning shows the normalized severity (`minor`, `moderate`, `severe`,
  `extreme`, `unknown`), the provider level (`yellow`, `amber`, `red`) when supplied, warning type,
  headline, valid-from/valid-to in local time, affected areas and instruction text as plain text.
- AC-10: A warning whose `valid_to` is at or before the evaluation time is never projected as
  active, even if the provider snapshot still contains it; read-time projection does not contact
  the provider.
- AC-11: For the Met Office provider, discovery follows at most one `rel="related"` link from the
  declared feed and refuses a different origin, a missing link, an oversized body, a redirect or a
  parse error, leaving the last-good snapshot published.
- AC-12: A missing or rejected credential produces an actionable reconnect state; no API key,
  warning body or provider HTML is logged or rendered as markup.

### US-04 — Nearby fuel prices (priority P2)

Independent verification: OAuth stub tests, response fixtures, sorting and formatting tests.

- AC-13: Setup selects a location by town or UK postcode and queries stations within a declared
  radius of that origin; the radius and origin are visible before saving.
- AC-14: Results list stations sorted by distance ascending, then by the cheapest displayed grade,
  showing brand, distance in metres/kilometres, and price per litre for every grade the provider
  supplied; unsupported grades are omitted.
- AC-15: Prices are integer pence per litre; a negative, fractional or missing price rejects that
  station's price rather than silently rounding, and a station with no valid price is shown without
  prices or omitted, never as zero.
- AC-16: The client obtains a token from the declared token endpoint, uses it only for the declared
  provider origin, caches it in process memory, and retries once after a 401; client secrets and
  tokens never enter the browser payload, YAML, logs or errors.
- AC-17: Missing, expired or rejected client credentials produce a clear reconnect state and
  preserve last-good data.

### US-05 — Integrations scales by type, not provider (priority P1)

Independent verification: component tests and recorded desktop/mobile browser checks.

- AC-18: The catalog groups providers under integration types; a type with one available provider
  shows one setup action, a type with several shows an explicit provider choice, and unavailable
  providers or disk errors stay visible without blocking other types.
- AC-19: Configured integrations are grouped under their type with a non-colour-only status
  (ok/stale/error/disabled), name, provider and location subtitle, and last refreshed time.
- AC-20: A search field filters entries by name, type, provider and location; an empty result
  explains what is configured instead of showing a blank panel.
- AC-21: The detail page uses tabs at 768px and above and sequential labelled sections below;
  Refresh and Configure are keyboard reachable and the active tab is announced.
- AC-22: Type and provider icons resolve through one registry with a neutral fallback; generic
  catalog and entry components contain no provider-specific branches.
- AC-23: Adding a second provider of a type already present does not create a second top-level
  section or duplicate the type heading.

### US-06 — Home follows the Home Assistant section model (priority P1)

Independent verification: layout schema tests, component tests, 320/375/768/1024/1440 checks.

- AC-24: Home renders named sections with a header (name, card count, collapse, view switch, add
  card, overflow menu) and 1/2/3/4 columns below 480, 480–1023, 1024–1439 and at 1440+ CSS pixels.
- AC-25: Device and integration cards share one frame and anatomy (badge, name, state, status);
  existing saved layouts load with default card geometry and no loss of placements or order.
- AC-26: The add-card sheet is one searchable list grouped by integration type and device, marks
  sources already placed in that section, keeps single selection for integration results and
  multi-selection for device values, and is fully operable by keyboard.
- AC-27: Card width and height can be changed without dragging, and dragging keeps the full card
  preview and persisted order on desktop and touch.
- AC-28: Status and errors use icon plus text, focus rings are visible, touch targets are at least
  44px, and no card relies on hover.
- AC-33: Every card renders inside a section; no card occupies the top-level section grid on its
  own, and existing saved standalone (`kind: card`) blocks render inside the implicit Services
  section without being rewritten just by loading Home.
- AC-34: New integration cards are placed in a single default "Services" section, which is
  renameable, reorderable and removable like an area; the app never creates one section per
  provider or per capability.
- AC-35: A section containing one card keeps its header, count, add, collapse, view and overflow
  affordances; an empty section shows an add-card empty state and is never deleted automatically
  because it is empty. Auto-hiding a header is not applied.
- AC-36: Sections share one layout grammar. Areas keep the two-column inner grid with device
  cards; Services uses the same grid and changes only card geometry and view. No section or
  capability gets a bespoke layout, and an integration card moved into an area behaves identically.
- AC-37: At section widths of two or more card columns an integration card spans the full section;
  compact integrations use one column. Rows wrap, never scroll horizontally, and card content
  ellipsizes or wraps using shared rules at 320px.
- AC-38: Sections reflow 1/2/3/4 across the page while each section keeps its internal card
  columns, so card position is preserved when the viewport changes. Dense placement is optional
  and off by default.
- AC-39: Card width (one or two columns) and height (two to five units) are persisted per
  placement. Changing either is possible without pointer dragging, and existing cards without
  stored values render with their source default.
- AC-40: The Services section defaults to two page columns wide on desktop and lays its cards out
  in three to four inner columns. It must not render one card per row while horizontal space is
  available, and six compact service cards must occupy at most two inner rows at 1024px and wider.
- AC-41: Integration cards use compact bodies on Home: at most two detail rows plus a “+N more”
  indicator. Full lists remain on the detail page, so a single card cannot grow without bound.
- AC-42: A section does not gain a third grouping level. Type headings never appear inside a
  section; instead an explicit “Split into sections by type” action creates ordinary, independent
  sections once, which can then be renamed, reordered, collapsed and removed.
- AC-43: Any section can be collapsed. A collapsed section keeps its header with name, card count
  and worst status, hides its cards, and persists the collapsed state; expanding restores the
  previous order and sizes.
- AC-44: A tall section never overlaps, covers or reflows another section's cards. It may only
  increase the row height and leave whitespace beside shorter neighbours. Dense gap-filling
  placement remains an optional setting and is off by default.
- AC-45: Home cards use an explicit width of one or two card columns and a bounded height of two
  to five units of a shared base unit. A card never grows beyond its span; overflow becomes a
  “+N more” hint, and no card scrolls horizontally.
- AC-46: When a card is placed, the smallest span that fits its content is chosen (device tiles are
  three units, or four when the label or device name needs the extra line). The user can raise the
  height in edit mode without dragging, and the choice persists.
- AC-47: Sections have independent, content-driven heights and are top-aligned. Two sections in
  the same row may end at different heights; the shorter one keeps its cards in place and shows
  whitespace below, which is accepted rather than stretched or hidden.
- AC-48: A section may contain one card or many; a one-card section keeps its named header, count,
  add, collapse and view controls, and the card does not stretch to fill the section. There is no
  separate “single integration card” type.
- AC-49: Every section has a view of `cards` or `list`, switchable from its header without entering
  edit mode. The choice is presentation state, persisted with the layout, and applies to that
  section only.
- AC-50: In list view each card renders as one compact row with icon, name, a single primary value
  and status; rows are full width, ignore stored card width and height, remain keyboard operable,
  and open the same detail destination as the card. A wide section may lay its rows out in two
  columns.
- AC-51: Switching view preserves card order, identity, stored width and height. Returning to cards
  restores the previous geometry exactly; no placement or source data changes.
- AC-52: A section can hold device cards, integration cards, or both; removing a non-empty section
  requires confirmation and offers moving its cards to another section (Services by default) or
  deleting them with the section.

### US-07 — Coordinates are visible on demand (priority P3)

Independent verification: presenter tests, disclosure and network-observation checks.

- AC-29: Weather, alert and fuel detail pages offer an explicit "Show map" action; no map, tile or
  coordinate request is made before that action.
- AC-30: The map shows the configured origin and the available station markers or warning
  polygons, and selecting a marker exposes the same values as the corresponding list row.
- AC-31: The disclosure names the tile provider before loading; coordinates are not written to
  browser storage, logs or URLs, and closing the map leaves the list fully usable.
- AC-32: A warning with missing or over-limit geometry is still listed and readable; the map shows
  the remaining warnings rather than failing.

## Requirements and success

- FR-01: MUST extend `weather_forecast/v1` additively with optional sunrise, sunset, daily UV max
  and hourly UV (AC-01–03).
- FR-02: MUST add an `air_quality/v1` capability and at least one shipped provider (AC-04–07).
- FR-03: MUST add a `weather_alerts/v1` capability and ship the Met Office and NWS providers or
  record why a provider is unavailable (AC-08–12).
- FR-04: MUST add a `fuel_prices/v1` capability and the UK Fuel Finder provider with OAuth 2.0
  client credentials (AC-13–17).
- FR-05: MUST group the catalog and configured entries by integration type with search (AC-18–23).
- FR-06: MUST render Home as sections with one shared card frame and a searchable add-card sheet
  (AC-24–28).
- FR-07: MUST offer an opt-in map on detail pages without changing list-only behaviour (AC-29–32).
- FR-08: MUST keep every new provider host-editable YAML and reusable without application code
  changes when it fits an installed capability.
- FR-09: MUST place every Home placement inside a section, render legacy standalone blocks in an
  implicit Services section, and materialise that section only on an accepted layout save
  (AC-33–36).
- FR-10: MUST persist per-card width and height additively and reflow sections without changing
  their internal card columns (AC-37–39).
- FR-11: MUST render Services as a wide, multi-column section with compact card bodies so growth
  adds rows slowly rather than one card per row (AC-40–41).
- FR-12: MUST support collapsing sections and an explicit split-by-type action, with no
  third grouping level and no automatic reordering (AC-42–44).
- FR-13: MUST size Home cards by an explicit column span and a bounded unit height, choosing the
  smallest fitting span and truncating overflow rather than growing without limit (AC-45–47).
- FR-14: MUST use one section primitive for one or many cards, with a per-section `cards` or `list`
  view that preserves geometry across switches (AC-48–52).
- NFR-01: External payloads are bounded (body size, list lengths, polygon vertices, forecast
  horizon) and validated before publication; missing data never becomes zero (AC-02, 05–07,
  10, 15).
- NFR-02: No new public route, MQTT topic, MCP tool or unauthenticated entry point is added; new
  secrets use the existing encrypted credential slots (AC-11–12, 16–17).
- NFR-03: Provider attribution, licence and non-commercial limits are shown in setup and results
  (AC-04, 07, 14).
- NFR-04: UI changes remain keyboard operable, announce status, and fit 320/375/768/1024/1440
  without clipping (AC-21, 26–28).
- NFR-05: Refresh failures are recoverable and never overwrite a last-good snapshot (AC-07,
  11–12, 17).
- SC-01: All acceptance scenarios pass with synthetic providers, temporary encrypted stores and
  mock browser hubs; no household device or real provider account is used.
- SC-02: The catalog remains usable with four integration types and at least ten providers; adding
  a provider requires no edit to a generic component.
- SC-03: Existing waste and weather instances, results and Home placements load unchanged after
  the redesign.

## Entities and external interfaces

- Weather resource: additive optional fields only; existing consumers ignore unknown JSON fields.
- Air quality resource: one per instance, carrying the selected origin label, index family,
  current values, a bounded hourly series and pollen. Provider: Open-Meteo Air Quality, a separate
  origin from the forecast host, free for non-commercial use with CAMS attribution.
- Alert resource: one per instance, carrying active warnings with normalized severity, provider
  level, type, headline, description, instruction, affected areas, validity window and optional
  bounded polygon. Providers: Met Office NSWWS (GeoJSON after feed discovery) and NWS (GeoJSON).
- Fuel resource: one per instance, carrying the search origin coordinates, radius, currency and a
  bounded station list with distance and per-grade prices. Provider: UK Fuel Finder (OAuth 2.0).
- Home layout: adds optional section width, view and collapsed state, and optional per-card width
  and height. Existing cards default without rewriting saved data.
- Authenticated integrations API remains the only transport; see
  [contracts](contracts/) for resource shapes.

## Questions and assumptions

- Q1. Met Office NSWWS has no stable snapshot path; approve the bounded feed discovery in
  [ADR-002](adr-002-alert-feed-discovery.md), or ship UK alerts later. Consequence: without it the
  UK has no official warning provider. Owner: maintainer. Blocks US-03 for Met Office.
- Q2. Can the household obtain UK Fuel Finder credentials, and what are the rate limits and
  redistribution terms? Consequence: without credentials US-04 cannot be verified against a live
  provider; fixtures only. Owner: household maintainer. Blocks live verification, not design.
- Q3. Approve the Leaflet + OpenStreetMap dependency and the tile-provider disclosure in
  [ADR-003](adr-003-map-and-coordinate-presentation.md), or keep coordinate lists only.
  Consequence: no map without approval. Owner: maintainer. Blocks US-07.
- Q4. Confirm "civil alerts" means official weather warnings only. Assumption: yes, there is no
  public civil-alert API. Owner: requester. Non-blocking for US-03.
- Q5. Confirm manual refresh remains the model, given that warnings and prices are time-sensitive.
  Consequence: alerts are only as current as the last refresh; scheduled refresh stays deferred.
  Owner: maintainer. Non-blocking, but it limits AC-08–10 usefulness.

## Compatibility and rollback

New contract versions do not affect existing waste or weather resources. The Home layout gains
optional `width`, `view`, `collapsed`, card `width` and card `height` fields. Binaries predating
this change decode JSON into the older struct and silently drop those fields, and compute a
different Home ETag; a rollback therefore requires exporting/restoring a pre-change Home layout to
avoid losing the new presentation state. New capabilities can be removed without touching devices.
