# Provider research

Checked 2026-10-29 against provider documentation and a validated community client. These are
proposed extensions, not capabilities the current compiler already supports. Numeric limits and
terms remain provider-controlled and must be rechecked before claiming compliance.

## Summary

| Capability | Provider | Origin | Auth | Format | Free access | Verified |
| --- | --- | --- | --- | --- | --- | --- |
| Weather (extend) | Open-Meteo | `api.open-meteo.com` | none | JSON | Non-commercial, attribution | Yes: documented variables |
| Air quality | Open-Meteo Air Quality | `air-quality-api.open-meteo.com` | none | JSON | Non-commercial, CAMS attribution | Yes: documented variables and bands |
| Alerts (UK) | Met Office NSWWS v1.1 | `data.hub.api.metoffice.gov.uk` | `apikey` header | Atom feed → GeoJSON | DataHub subscription | Yes: feed behaviour confirmed by a live-validated client |
| Alerts (US) | NWS | `api.weather.gov` | none, identifying `User-Agent` | GeoJSON | Public | Yes: documented |
| Fuel prices (UK) | Fuel Finder | `developer.fuel-finder.service.gov.uk` | OAuth 2.0 client credentials | JSON REST | Government service, registration required | Partial: auth documented, query shape not |

## Open-Meteo forecast (weather extension)

- Existing provider already requests `current`, `hourly` and `daily` from `/v1/forecast` with
  `timeformat=unixtime` and `timezone=auto`.
- Additional documented variables:
  - hourly: `uv_index`, `uv_index_clear_sky`.
  - daily: `sunrise`, `sunset`, `uv_index_max`, `daylight_duration`, `sunshine_duration`.
- Sunrise/sunset are returned as local ISO 8601 timestamps (for example
  `2026-10-29T06:58`), aligned with the requested timezone. Normalize to a local `HH:MM` clock and
  keep the resource timezone; do not convert to UTC.
- Free hosted access is for non-commercial use and requires attribution. The existing provider
  already states this in setup copy.
- Decision: add only `sunrise`, `sunset`, `uv_index_max` and hourly `uv_index`. Duration fields are
  not needed for the acceptance scenarios.

## Open-Meteo Air Quality

- Endpoint: `https://air-quality-api.open-meteo.com/v1/air-quality`.
- Separate origin from the forecast host. The current definition model has one `base_url` per
  provider; a bounded companion request must be same-origin, so air quality cannot be folded into
  the weather definition. It becomes its own capability instance.
- Parameters: `latitude`, `longitude`, `hourly`, `current`, `timezone=auto` or an IANA name,
  `timeformat=unixtime|iso8601`, `forecast_days` (0–7, default 5), `domains=auto|cams_europe|cams_global`,
  `cell_selection=land|sea|nearest`.
- Variables:
  - pollutants: `pm10`, `pm2_5`, `carbon_monoxide`, `nitrogen_dioxide`, `sulphur_dioxide`, `ozone`.
  - indices: `european_aqi`, `us_aqi` plus per-pollutant variants.
  - UV: `uv_index`, `uv_index_clear_sky` (also available here; we keep UV in weather to avoid
    duplicating it).
  - pollen: `alder_pollen`, `birch_pollen`, `grass_pollen`, `mugwort_pollen`, `olive_pollen`,
    `ragweed_pollen`, unit grains/m³, **Europe only, during pollen season, 4-day forecast**.
- European AQI bands from the EEA 2024 revision: 0–20 good, 20–40 fair, 40–60 moderate, 60–80 poor,
  80–100 very poor, >100 extremely poor. The overall index is the maximum of the five pollutant
  indices.
- US AQI bands: 0–50 good, 50–100 moderate, 100–150 unhealthy for sensitive groups, 150–200
  unhealthy, 200–300 very unhealthy, 300–500 hazardous.
- Attribution: CAMS ENSEMBLE data providers plus Open-Meteo. Required.
- Decision: evaluate the category server-side from the declared family; store the category in the
  resource so the browser never recomputes thresholds.

## Met Office NSWWS (severe weather warnings, UK)

Documented behaviour, corroborated by a live-validated community integration (August 2026) because
the public docs are partly stale:

- Base URL: `https://data.hub.api.metoffice.gov.uk/nswws/v1.1`. Fixed public host, not a personal
  URL. Requires a Weather DataHub subscription to the Warnings product.
- Auth header is **`apikey`** (the older docs say `x-api-key`). This matches the existing
  `apikey_header` mode, whose header name is already `apikey`.
- **There is no stable "issued" path.** `/objects/feed` returns an Atom feed; its `rel="related"`
  link points at the current GeoJSON snapshot, whose URL contains the feed's own UUID and changes
  whenever the feed updates. The snapshot must therefore be discovered on every refresh.
- Snapshot is GeoJSON `FeatureCollection`. Warning `properties` include: `issuedDate`,
  `modifiedDate`, `weatherType[]`, `warningLevel` (`YELLOW|AMBER|RED`), `warningImpact`,
  `warningLikelihood`, `warningStatus`, `warningHeadline`, `whatToExpect[]`, `whatShouldIDo`,
  `warningFurtherDetails`, `warningId`, `warningVersion` (string), `validFromDate`, `validToDate`,
  `affectedAreas[]{regionName, regionCode, subRegions[]}`. Geometry is always `MultiPolygon`.
- Snapshot excludes EXPIRED and CANCELLED warnings and is ordered by risk then `validFromDate`.
  Empty is `{"type":"FeatureCollection","features":[]}`.
- Errors: 403 bad key, 404 stale/unknown snapshot URL, 429 rate limited.
- Engine impact: the JSON+JSON-Pointer model cannot parse the Atom feed. Options are (a) bounded
  one-hop link discovery (selected, [ADR-002](adr-002-alert-feed-discovery.md)), (b) no UK
  provider, or (c) a general XML engine (rejected as unnecessary). The snapshot URL changes, so a
  captured `state.*` URL would go stale and cannot be used across refreshes.
- Region filtering is left to the presentation layer in the community client; our normalization
  should keep affected regions and let the user choose a region or polygon filter later. The first
  release keeps all active warnings and shows the affected areas.

## NWS (US alerts)

- `https://api.weather.gov/alerts/active` (with `area=`/`point=` filters), GeoJSON, no API key.
- Requires a descriptive `User-Agent`. **Our request model has no arbitrary header field**, and
  `apikey_header` only sends `apikey`. A bounded, non-secret `headers` map (or a dedicated
  `user_agent` field) is required before this provider can ship; add it to the same compiler
  extension as discovery, restricted to fixed public values with no credential material.
- US only. Useful only for installations in NWS coverage; the UK provider is the primary one.

## UK Fuel Finder

- Government service: “Fuel Finder”, published under The Motor Fuel Price (Open Data)
  Regulations 2025. Data updates within 30 minutes of a price change.
- Access options: CSV download (twice daily), email subscription, or REST API.
- API authentication: **OAuth 2.0 client credentials**. Requires a GOV.UK One Login and a
  registered client. The API returns JSON.
- Data includes current retail prices by fuel type, forecourt details (address, operator, brand),
  amenities, opening hours and update timestamps.
- Not verified: exact query parameters (area, coordinates, radius or pagination), response field
  names, rate limits, and redistribution/attribution terms. These require a captured response and
  must be resolved in T002 before acceptance.
- Engine impact: new `oauth2_client_credentials` mode
  ([ADR-001](adr-001-oauth-client-credentials.md)). Token endpoint and API origin are
  different hosts, so the token call needs its own declared, allowed destination; the access token
  is then sent only to the API origin.
- Fallback if credentials are unavailable: keep the capability and provider YAML unshipped, or
  ship the provider disabled with a clear "credentials required" state. Do not scrape a
  third-party site.

## Cross-provider notes

- All new providers are HTTPS on port 443 and fit the existing destination rules.
- New origins must be disclosed during setup: Open-Meteo geocoding, Postcodes.io, the air-quality
  host, the Met Office feed/snapshot host, the NWS host and the Fuel Finder token/API hosts.
- No provider requires an inbound webhook, so the pull-only model still holds.
- Manual refresh remains the update model; none of these providers is polled on a timer.
