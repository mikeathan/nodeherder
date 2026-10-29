# Fuel prices v1 wire contract

Status: Proposed. Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## Representation and nullability

Shared envelope with `contract_version` literal `fuel_prices/v1`, `instance_id`, `resource_id`,
`revision`. All operational timestamps are UTC RFC3339 ending `Z`.

| Field | Type / meaning |
| --- | --- |
| origin | `{ label, latitude, longitude }`; the selected search location, required |
| radius_metres | Integer 100–50000 |
| currency | Literal `GBP` in v1 |
| stations | Array, <= 100, ordered by `distance_metres` then cheapest displayed grade |
| captured_at, last_checked_at, evaluated_at, freshness | As in the weather contract |
| attribution | Required provider text, 1–512 chars |

Station:

| Field | Type / meaning |
| --- | --- |
| station_id | Provider site id when supplied, otherwise a stable local id from the canonical tuple; 1–128 chars |
| name | Required, <= 160 chars |
| brand | Optional, <= 80 chars |
| address | Optional, <= 240 chars |
| postcode | Optional, <= 16 chars |
| latitude, longitude | Optional WGS84 |
| distance_metres | Integer >= 0, computed from the resource origin |
| prices | Array of `{ grade, label, pence_per_litre, updated_at }` |
| prices[].grade | `e10`, `e5`, `b7`, `b10`, `hvo`, `other` |
| prices[].label | Provider label for display (for example `E10`, `B7 Premium`), <= 40 chars |
| prices[].pence_per_litre | Integer 0–1000 |
| prices[].updated_at | Optional UTC instant |

## Units and validation

- Prices are integer pence per litre. A negative, fractional, zero or out-of-range value invalidates
  that price entry; the station is kept without that grade. A station with no valid price is shown
  without prices, never as `0`.
- `distance_metres` is computed by the normalizer with a haversine calculation from the resource
  origin and the station coordinates. A station with unusable coordinates is kept but sorted after
  those with a distance and shown without one.
- The origin and radius are set at setup and remain visible; a refresh does not silently change
  them.
- Duplicate `station_id` values are an ingress error, not merged.
- Currency is fixed to `GBP`; a different currency is rejected in v1 rather than converted.

## Sorting

1. Stations with a distance first, ascending.
2. Within equal distance, the cheapest valid displayed grade ascending.
3. Stations without a distance last, preserving source order.
4. Grades within a station are ordered `e10`, `e5`, `b7`, `b10`, `hvo`, `other`.

Sorting is a read-time projection over stored content; it does not change `revision` or identity.
The UI must label the grade for every price; no "from" price is invented across grades.

## Freshness

`fresh` while `evaluated_at < captured_at + 1 hour`, otherwise `stale`. Fuel Finder updates within
30 minutes of a price change, but manual refresh is the only update path, so the checked time is
always shown.

## Identity and revision

One resource per instance. `resource_id` is a random UUIDv4 allocated on the first accepted
snapshot. `station_id` uses the provider id when present; otherwise it is derived from the
canonical tuple (provider id, name, postcode, rounded coordinates) so a station keeps identity
across refreshes even if its price changes.

## Example

```json
{
  "contract_version": "fuel_prices/v1",
  "instance_id": "5a4b3c2d-1e0f-4a9b-8c7d-6e5f4a3b2c10",
  "resource_id": "1f2e3d4c-5b6a-4978-8a9b-0c1d2e3f4a5b",
  "revision": 2,
  "origin": { "label": "Nether Edge", "latitude": 53.3601, "longitude": -1.4832 },
  "radius_metres": 5000,
  "currency": "GBP",
  "stations": [
    {
      "station_id": "7f3c2a1b-9d8e-4c5f-8a6b-1c2d3e4f5a6b",
      "name": "Shell Abbeydale",
      "brand": "Shell",
      "address": "Abbeydale Road, Sheffield",
      "postcode": "S7 2QF",
      "latitude": 53.3552,
      "longitude": -1.4921,
      "distance_metres": 1240,
      "prices": [
        { "grade": "e10", "label": "E10", "pence_per_litre": 143, "updated_at": "2026-10-29T13:50:00Z" },
        { "grade": "b7", "label": "B7", "pence_per_litre": 150, "updated_at": "2026-10-29T13:50:00Z" }
      ]
    }
  ],
  "captured_at": "2026-10-29T14:02:00Z",
  "last_checked_at": "2026-10-29T14:02:00Z",
  "evaluated_at": "2026-10-29T14:05:00Z",
  "freshness": "fresh",
  "attribution": "Fuel prices: UK Fuel Finder (Department for Energy Security and Net Zero)."
}
```

Note: v1 stores integer pence. A provider value of `142.9` is rejected rather than rounded, because
rounding would invent precision the contract does not model. If tenths become necessary, a new
contract version carries them explicitly.
