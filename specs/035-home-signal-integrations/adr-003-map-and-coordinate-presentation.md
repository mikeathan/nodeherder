# ADR-003: Opt-in map with Leaflet and OpenStreetMap tiles

Status: Proposed; blocks map phase until approved | Date: 2026-10-29
Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)
Rules: NH-03/04/06, FE-04/05

## Context

Weather stores a selected location, fuel prices are only meaningful near an origin, and alerts
carry warning polygons. The household asked for “anything with a map that can show coordinates”,
but the immediate use is unclear. The frontend currently has no mapping dependency, and household
coordinates are sensitive: they can reveal occupancy and routines.

A map is therefore a presentation feature with a privacy cost, not a data requirement. Every
resource is usable as a list with coordinates and distances.

## Options

1. No map: show coordinates and distance in the list. Lowest cost, no dependency, no disclosure.
2. Static map image from a provider: leaks the location to the image host on every render and adds
   a vendor. Rejected.
3. Leaflet plus public OpenStreetMap tiles, loaded lazily on detail pages behind an explicit
   action. Small dependency, familiar interaction, no API key. Tile requests reveal an approximate
   viewport to the tile provider.
4. Self-hosted tiles: no third-party disclosure, but adds a service, storage and operations burden
   far beyond the feature.
5. Dependency-free SVG relative plot of origin/stations: no geography and no basemap, so it answers
   “which is nearer” but not “where”. Rejected as the primary map, though it remains a possible
   list enhancement.

## Decision

Adopt option 3, phased and opt-in:

- The map appears only on a connection detail page, inside Results, and only after the user
  activates “Show map”.
- No map on Home, in the board, or in Panel mode. No map is loaded during page reads or refreshes.
- Leaflet is dynamically imported, so it is absent from the initial bundle.
- Before loading, the UI names the tile provider and states that tiles are fetched from it. The
  disclosure is plain text, not a hidden tooltip.
- Attribution is shown on the map. Tile URL, zoom and provider are fixed in code; no arbitrary
  tile URL comes from YAML or the API.
- Coordinates and the current viewport are never written to browser storage, logs, URLs, analytics
  or assistant context. The map is destroyed when the section closes.
- Resources expose only what a map needs: an origin point, station points and warning polygons.
  Polygons are bounded (for example 2000 vertices) and simplified or omitted beyond the limit;
  a warning with unusable geometry remains listed.
- Default remains list-only. A user who never opts in makes no tile request.

## Consequences

- New runtime dependency (`leaflet`) and a third-party request path from the browser.
- Household coordinates become visible to the tile provider for the time the map is open. The
  disclosure and opt-in are the mitigation; there is no way to remove this while using public tiles.
- The app must work fully offline and without the map; tests must prove no request before opt-in.
- If the household later wants an always-on map, self-hosted tiles or a local basemap require a new
  decision.

## Validation and reconsideration

Tests: no tile request before activation; disclosure present and accessible; map renders for
weather, fuel and alerts; marker selection matches the list row; unusable/oversized geometry does
not break the page; coordinates absent from storage, URLs and logs; keyboard close and focus
return; reduced-motion and 320px layout. Reconsider if the dependency size, tile terms or privacy
requirements change, or if self-hosted tiles become available.
