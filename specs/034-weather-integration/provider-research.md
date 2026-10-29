# Weather provider and YAML investigation

Checked 2026-10-28 against provider documentation. These extensions are proposed, not capabilities the current compiler already supports.

| Provider | Location | Authentication | Forecast | Free access |
| --- | --- | --- | --- | --- |
| Open-Meteo | [Geocoding](https://open-meteo.com/en/docs/geocoding-api) searches names/postcodes; returns numeric ID, coordinates and timezone on a separate origin. | None for free personal use. | [Forecast](https://open-meteo.com/en/docs) returns current, hourly and daily in one request, using parallel arrays and units; up to 16 days. | [Non-commercial](https://open-meteo.com/en/pricing); attribution required. |
| Met Office Global Spot | Coordinates select nearest site; separate location search needed. [Overview](https://datahub.metoffice.gov.uk/docs/f/category/site-specific/overview). | [apikey header](https://datahub.metoffice.gov.uk/support/faqs); subscribe to product. | GeoJSON; separate hourly (48h), three-hourly (7d) and daily (7 entries including one past) calls. | [360 calls/day](https://datahub.metoffice.gov.uk/pricing/site-specific); attribution required. |
| WeatherAPI.com | [Search](https://www.weatherapi.com/docs/search-api/) returns location candidates and coordinates; [timezone appears in forecast, not search](https://www.weatherapi.com/docs/location-object/). | [key query parameter](https://www.weatherapi.com/docs/authentication/). | [Forecast](https://www.weatherapi.com/docs/forecast-api/) returns current and nested daily/hourly; [free plan](https://www.weatherapi.com/pricing) offers three days. | 100,000 calls/month; verify terms/attribution. |

## Existing YAML compatibility

The [WhenIsBins definition](../../backend/configs/integrations/waste_collection/whenisbins.yaml) declares metadata, one HTTPS base URL, auth, policy, inputs, named requests, search/select/status setup, refresh and mapping. It is compiled/validated, host-editable and reloadable, with secrets held separately. Weather must use the same lifecycle and per-provider YAML; it need not use waste-specific fields.

Required shared extensions to design and test:

1. Bounded, explicitly declared approved origins for location search and forecast; retain DNS/private-network/redirect protections and restrict credentials to the intended origin. A geocoder is a separate data recipient; disclose it. Direct coordinate entry is an alternative.
2. Header-key and query-key auth modes with secret slots; query secrets must be redacted from errors/logging and never exposed in YAML, browser DTOs or redirects.
3. Synchronous location search/select with numeric candidate IDs, bounded coordinates/timezone and disambiguated labels, selected choice retained and fenced to setup. Current flow requires string ID/label and lookup submission/status.
4. A bounded list of named refresh requests with atomic publication and origin/credential handling; Met Office hourly and daily failures preserve the last-good snapshot. Existing single-request waste remains compatible. Specify revision fencing, timeout, cancellation, limits and conditional-response semantics before coding.
5. Capability-owned declarative mappings for parallel arrays, GeoJSON periods and nested days/hours; validate lengths, timestamps, optional fields, units and condition codes. Specify bounded transformations rather than executable templates.
6. Extend provider form/YAML editor, compiler, previews and mocks consistently. Test existing WhenIsBins definitions and workflows unchanged, using source-owned fixture tests.

## Existing implementation constraints verified

`backend/internal/integrations/operations/request.go` checks the approved origin before constructing requests, binds saved credentials to that origin, rejects newline characters, and currently emits bearer credentials only. The HTTP client rejects redirects and protects against private destinations. Extend those checks for API keys; do not bypass them. `backend/internal/integrations/definitions/compiler.go` currently accepts only `none`, `bearer` and `optional_bearer`, while `frontend/src/contracts/integrations.ts` validates the same modes and `ConnectionSetup.vue` labels only bearer as required. Credential removal logic in `runtime/controls.go` also treats bearer as required. All producers, consumers and tests must change together.

The current worker in `runtime/coordinator.go` performs at most an admitted request and one renewal, then publishes inside a repository transaction after checking the operation owner. Multi-request weather refresh requires explicit changes to this state machine and its operation persistence, not parallel hidden calls in a capability. `flows/searchselect/flow.go` accepts only string candidate identifiers and labels; Open-Meteo's numeric IDs and coordinates need validated selection representation and identity fencing.

## Still to establish

The [Met Office example client](https://github.com/MetOffice/weather_datahub_utilities/blob/main/site_specific_download/ss_download.py) confirms the base `https://data.hub.api.metoffice.gov.uk/sitespecific/v0/point/`, the suffixes `hourly`, `three-hourly`, `daily`, latitude/longitude query parameters, and `apikey` header. This `/v0/` path is separate from the unrelated blended probabilistic v2 migration. The [hourly glossary](https://datahub.metoffice.gov.uk/docs/glossary?models=mo-spot-1hr&sortOrder=ALPHABETICALLY) and [daily glossary](https://datahub.metoffice.gov.uk/docs/glossary?models=mo-spot-daily&sortOrder=ALPHABETICALLY) document parameters, but exact serialized field names and interpretation still need response fixtures.

Still establish whether subscriptions cover both hourly and daily, location search source, and provider-specific terms. Complete response fixtures and YAML grammar examples before marking design gates complete. Forecasts need not include every mockup field; do not derive unsupported rain probabilities. Check Open-Meteo timestamps and daylight-saving boundaries with fixtures.
