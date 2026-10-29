# Air quality v1 wire contract

Status: Proposed. Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## Representation and nullability

Resource header is the shared envelope: `contract_version` literal `air_quality/v1`, lowercase
`instance_id`/`resource_id`, integer `revision >= 1`. All operational timestamps are UTC RFC3339
ending `Z`; forecast instants are UTC, displayed in the resource timezone.

| Field | Type / meaning |
| --- | --- |
| location | Selected location label, 1–256 chars |
| latitude, longitude | Optional WGS84 origin, bounded -90..90 / -180..180; null when the provider omits it |
| timezone | Required IANA name, never `Local` |
| index_family | `european` or `us` |
| index | Optional number >= 0; the provider's overall index |
| category | Server-derived from `index_family` and `index`; `unknown` when index is null |
| dominant | Optional pollutant code with the highest contribution when the provider supplies it |
| current | Optional object of nullable pollutant values |
| hourly | Bounded array (<= 168) of timestamped pollutant values, strictly ascending |
| pollen | Six nullable species values (see below) |
| captured_at, last_checked_at, evaluated_at, freshness | As in the weather contract |
| attribution | Required provider/CAMS text, 1–512 chars |

Pollutant values are nullable numbers in µg/m³: `pm2_5`, `pm10`, `ozone`, `nitrogen_dioxide`,
`sulphur_dioxide`, `carbon_monoxide`. `uv_index` is not stored here; UV belongs to the weather
capability so the same value is not presented twice.

Pollen species are nullable numbers in grains/m³: `alder`, `birch`, `grass`, `mugwort`, `olive`,
`ragweed`. Pollen is Europe-only and seasonal. Null means not supplied or out of season; it must
never become zero. A provider that supplies a pollen array with null entries keeps them null.

## Category bands

The server evaluates the category; the browser renders the value and never recomputes thresholds.

European (EEA 2024 revision): 0–20 `good`, 20–40 `fair`, 40–60 `moderate`, 60–80 `poor`,
80–100 `very_poor`, >100 `extremely_poor`.

US: 0–50 `good`, 50–100 `moderate`, 100–150 `unhealthy_sensitive`, 150–200 `unhealthy`,
200–300 `very_unhealthy`, 300–500 `hazardous`.

A null index yields `unknown`. An index outside the family's documented range (for example >500
for US) is `unknown`, not clamped.

## Validation and bounds

- Reject negative pollutant values, non-finite numbers, invalid IANA timezones, descending hourly
  timestamps and more than 168 hourly points.
- Reject an hourly point more than 24 hours in the past relative to `captured_at` unless the
  provider explicitly supplies recent history (not used in v1).
- Require at least one current value or at least one hourly point; an entirely empty response is
  `invalid_response`, not a valid quiet state.
- `category` must match one of the declared enum values for the family.
- Missing values remain null through projection and presentation.

## Freshness and projection

`freshness` is `fresh` while `evaluated_at < captured_at + 6 hours`, otherwise `stale`. The
six-hour budget reflects the daily/bi-daily CAMS update cadence and the fact that manual refresh is
the only update path. The UI must show both the source capture time and the last check so a
provider cache check is not presented as a fresh measurement. Projection never contacts a provider.

## Identity and revision

One resource per instance with a random UUIDv4 `resource_id`, allocated on first accepted snapshot
and retained across refreshes. Each accepted refresh increments `revision`. A failed refresh
publishes nothing and leaves the last-good snapshot readable.

## Example

```json
{
  "contract_version": "air_quality/v1",
  "instance_id": "6f1c2a94-6d0f-4d64-9a2e-0b8f6a1f3a11",
  "resource_id": "d2a4b1e0-3f5a-4c1b-9e2d-7a6c5b4e3f22",
  "revision": 3,
  "location": "Nether Edge",
  "latitude": 53.3601,
  "longitude": -1.4832,
  "timezone": "Europe/London",
  "index_family": "european",
  "index": 32,
  "category": "fair",
  "dominant": "pm2_5",
  "current": {
    "pm2_5": 11, "pm10": 18, "ozone": 54, "nitrogen_dioxide": 22,
    "sulphur_dioxide": 3, "carbon_monoxide": 180
  },
  "hourly": [
    { "time": "2026-10-29T15:00:00Z", "pm2_5": 11, "pm10": 18, "ozone": 54, "nitrogen_dioxide": 22, "sulphur_dioxide": 3, "carbon_monoxide": 180 }
  ],
  "pollen": { "alder": null, "birch": null, "grass": 2, "mugwort": null, "olive": null, "ragweed": null },
  "captured_at": "2026-10-29T15:40:00Z",
  "last_checked_at": "2026-10-29T15:40:00Z",
  "evaluated_at": "2026-10-29T15:42:00Z",
  "freshness": "fresh",
  "attribution": "Air quality: CAMS ENSEMBLE data providers via Open-Meteo. Non-commercial use."
}
```
