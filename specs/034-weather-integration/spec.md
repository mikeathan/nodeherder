# Spec: Weather integration

ID: 034 | Status: Implemented; final verification in progress | Owner: NodeHerder contributors | Updated: 2026-10-28
Constitution: proposed 1.0.0 | Components: backend and frontend

## Problem, scope, exclusions

Households need a local weather overview and forecast in NodeHerder. Connections currently ships WhenIsBins and manually refreshes data; Home integration sources and connection results are waste-specific. Add an information-only weather integration with a choice of Met Office, Open-Meteo and WeatherAPI.com, and a Home card opening detailed results.

Out of scope: automatic scheduling, arbitrary custom providers, weather-driven automations, device control, alerts, historical metrics, air quality, and radar. The integration remains independent of Devices.

## Scenarios and acceptance

### US-01 — Connect a provider (P1)

Independent verification: provider setup tests with simulated responses.

- AC-01: Given the weather catalogue, a user selects a shipped provider, searches and explicitly confirms a location, and can save a connection after successful validation. Ambiguous/no matches are visible; no location is silently selected.
- AC-02: Open-Meteo personal-use setup has no key field; Met Office and WeatherAPI.com display a key input and a link to obtain credentials. Invalid credentials and provider errors are actionable; secrets are never displayed in results or logs.
- AC-03: The UI describes free-tier constraints and attribution, including Open-Meteo's non-commercial restriction. No provider is promised universally most accurate.

### US-02 — View weather (P1)

Independent verification: synthetic resource, UI and viewport tests.

- AC-04: A user can place a weather card on Home and open its connection results; it shows location, condition, temperature, high/low, rain information and update age when available. Missing data is labelled or omitted, never zero-filled.
- AC-05: Results show available current conditions, hourly and daily forecasts with units, local dates and time; data is identified as forecast rather than a household measurement. Availability and horizon reflect the provider response.
- AC-06: On mobile, daily results display five compact rows per page with accessible previous/next and an accurate range; hourly forecasts scroll with keyboard/touch and buttons. On desktop cards fill the row and scroll without overlapping controls. Verify at 320, 375, 768 and 1440 CSS pixels.
- AC-07: Data, provider attribution, last successful refresh and freshness are visible. Offline, expired, empty and partial data are distinguished; a failed refresh preserves and clearly labels last-good data. Deleted/unavailable sources retain an explicit Home placeholder.

### US-03 — Refresh manually (P1)

Independent verification: fake-provider integration tests.

- AC-08: Refresh performs bounded provider work on user request, reports loading/success/failure and publishes a complete validated snapshot atomically. Page reads, startup and timers do not invoke providers. Existing bins remain unchanged.

## Requirements and success

- FR-01: Ship three selectable providers using host-editable YAML definitions like the bin integration, and location setup (AC-01–03).
- FR-02: Provide Home and detailed forecast views using available fields (AC-04–07).
- FR-03: Provide manual refresh and recoverable failure behavior (AC-07–08).
- NFR-01: External payloads and saved data are validated; requests have bounded time/size, secrets and location data are protected, and no new public access is added (AC-01–03, 07–08).
- NFR-02: Keyboard, focus and responsive behavior work at specified widths; timezone and missing values are explicit (AC-04–07).
- SC-01: All eight acceptance scenarios pass with synthetic providers and recorded UI checks; no household devices are used.

## Entities and external interfaces

A connection owns provider choice, private credentials and selected coordinates; a versioned weather resource owns normalized forecast data, timestamps, units and attribution. The authenticated integrations API serves it to Connections, Home and Panel. Home layout holds a stable source reference, not a copy of forecast data. Provider contracts are external and independently versioned.

## Decisions and limits

- Met Office is initially restricted to GB locations with Europe/London calendar semantics. Open-Meteo and WeatherAPI.com support their returned location/timezone coverage.
- The shared YAML grammar supports synchronous location selection, an explicitly declared public search origin, fixed key transports and one bounded companion GET. See the plan and ADR for ownership and failure behavior.
- Capability/source and frontend presentation interfaces keep generic Home/Connections consumers independent of payload schemas.
- Public API documentation and synthetic contract fixtures support automated verification. Live provider/account calls were not performed; no key is committed or read from chat. Free quotas/terms and upstream availability remain externally controlled.
- No promise is made that every provider supplies every mockup field or the same forecast horizon. The app labels or omits unavailable quantities.

### Location and operation recovery

UK postcodes must resolve to selectable coordinates, while town results include geographic context to distinguish matching names. Checking an operation must use the current owner revision and recover from a concurrent revision change by showing the latest state without starting another lookup.

Met Office forecasts with minute-precision UTC timestamps must normalize successfully, alongside timestamps that include seconds; invalid timestamps must still be rejected.

Home integration cards must remain readable when weather and other results are placed side by side, on both desktop and narrow screens. Cards stay independently movable.
The Home picker selects one integration result at a time so each selection clearly creates an independent card; device values remain multi-select.
The Home picker offers the complete bin collection schedule, not individual bin services; existing individual service placements remain readable.

Integration cards in panel mode must remain readable at desktop and mobile widths, including weather forecasts without summary rows.

Home device tiles use a shared content-responsive layout that shows complete measurement and device names and remains readable on narrow screens.

### Home drag previews

While an area or standalone card is dragged, its content remains visible at a useful size and the potential placement is identifiable. Reordering and moving cards retain their existing save, cancellation and keyboard behavior on desktop and touch screens.

### Review corrections

Provider weather conditions and Home icons reflect actual weather; hourly forecasts show the next 24 hours; navigation has descriptive accessible names and delayed provider loading cannot overwrite a name already typed. WeatherAPI usage must comply with its retention and disclosure terms.

Integration removal uses capability-neutral wording. Once removed, its cached instance must not continue to supply Home cards; existing placements show an unavailable state and remain removable. Removing a provider definition explains its shared scope and that integrations using it must first be removed.

Integration errors use consistent, accessible status styling on Home and the Integrations board, while non-error stale or disabled statuses remain distinguishable.

### Integration code cleanup

Weather icons use the established icon registry, date and time formatting use shared utilities with explicit timezone semantics, and integration composables and backend integration packages remain cohesive and readable without behavior regressions. Review all integration production sources for complexity, duplication and demonstrably dead code, including shared network and URL logic. Preserve destination validation, credential-origin binding, DNS protections and provider compatibility.

### Browser regression follow-up

The Home layout must keep content readable without excessive empty space as device cards grow, and area dragging must preserve the full preview and persist the intended order on desktop and touch devices. Integration browser tests must select the intended provider and verify current removal and card presentation semantics when multiple providers are available.
