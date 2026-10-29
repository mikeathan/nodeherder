# Weather alerts v1 wire contract

Status: Proposed. Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## Representation and nullability

Shared envelope with `contract_version` literal `weather_alerts/v1`, `instance_id`, `resource_id`,
`revision`. Timestamps are UTC RFC3339 ending `Z`.

| Field | Type / meaning |
| --- | --- |
| status | `active` when one or more warnings are present; `none` when the provider confirms no active warnings |
| warnings | Array, <= 50, ordered by severity then `valid_from` |
| attribution | Required provider text, 1–512 chars |
| captured_at, last_checked_at, evaluated_at, freshness | As in the weather contract |

Warning:

| Field | Type / meaning |
| --- | --- |
| warning_id | Provider warning id when supplied, otherwise a stable local id derived from the canonical tuple; 1–128 chars |
| severity | `minor`, `moderate`, `severe`, `extreme`, `unknown` |
| level | `yellow`, `amber`, `red`, `unknown` (provider scale) |
| type | Provider warning type text (for example `Wind`, `Thunderstorm`), <= 80 chars |
| headline | Required plain text, <= 300 chars |
| description | Optional plain text, <= 4000 chars |
| instruction | Optional "what to do" plain text, <= 2000 chars |
| areas | Array of affected area names, <= 50, each <= 120 chars |
| valid_from, valid_to | Required UTC instants; `valid_to` must be after `valid_from` |
| issued_at, modified_at | Optional UTC instants |
| polygon | Optional bounded GeoJSON `MultiPolygon`/`Polygon` coordinates; null when unusable |

Severity normalization is provider-owned: the Met Office `warningLevel` maps YELLOW→`minor`,
AMBER→`severe`, RED→`extreme` with the exact level retained, and NWS `severity` maps
Minor/Moderate/Severe/Extreme directly. A provider value outside the mapping is `unknown`, but the
warning is still shown.

## Status and expiry

- `status: none` requires `warnings: []` and a successful provider response. It is distinct from an
  unavailable provider or a failed refresh, which keep the last-good snapshot and its status.
- Projection drops any warning whose `valid_to <= evaluated_at`. This is read-time only: it does
  not change `revision`, `captured_at` or `last_checked_at`.
- The Met Office snapshot already excludes EXPIRED and CANCELLED warnings; local filtering is still
  required because a snapshot is only as current as its capture.
- A provider response containing only expired warnings projects to `status: none`.

## Geometry bounds

- At most 2000 vertices across all rings of one warning; at most 500 warnings' worth of total
  vertices per resource.
- Coordinates must be finite WGS84 values (-180..180 longitude, -90..90 latitude). A geometry that
  fails validation is dropped from that warning and `polygon` becomes null; the warning remains
  listed.
- Geometry is presentation data only. It never affects filtering or status in v1.

## Validation

- Reject missing `valid_from`/`valid_to`, `valid_to <= valid_from`, empty headline, over-limit
  strings/lists, duplicate `warning_id` values and non-UTF-8 text.
- Strip control characters. Provider HTML is not rendered; text is escaped by the client.
- Require `attribution` and a valid IANA-free presentation (alerts carry no timezone field; times
  are displayed in the household's configured zone).

## Freshness

`fresh` while `evaluated_at < captured_at + 1 hour`, otherwise `stale`. Warnings are time-critical;
the browser shows the last checked time and never implies live monitoring while manual refresh is
the only update path.

## Identity and revision

One resource per instance. `resource_id` is a random UUIDv4 allocated on first accepted snapshot.
`warning_id` is provider-supplied when present; otherwise it is derived from the canonical tuple
(provider id, type, headline, valid_from) so a warning keeps identity across refreshes. A warning
that disappears is simply absent; there is no tombstone in v1.

## Example

```json
{
  "contract_version": "weather_alerts/v1",
  "instance_id": "0b6c8f9e-7d51-4a2e-8b3c-1f2d3e4a5b60",
  "resource_id": "9c1d2e3f-4a5b-6c7d-8e9f-0a1b2c3d4e5f",
  "revision": 4,
  "status": "active",
  "warnings": [
    {
      "warning_id": "79bb1df2-8983-4bbd-888a-66da9c099a1c",
      "severity": "severe",
      "level": "amber",
      "type": "Wind",
      "headline": "Strong winds may cause disruption on Thursday",
      "description": "A deep area of low pressure brings gusts of 60-70 mph.",
      "instruction": "Secure loose objects and allow extra travel time.",
      "areas": ["Yorkshire & Humber", "North East England"],
      "valid_from": "2026-10-30T06:00:00Z",
      "valid_to": "2026-10-30T18:00:00Z",
      "issued_at": "2026-10-29T10:12:00Z",
      "modified_at": "2026-10-29T10:12:00Z",
      "polygon": { "type": "MultiPolygon", "coordinates": [[[[-1.2, 53.4], [-1.1, 53.5], [-1.0, 53.4], [-1.2, 53.4]]]] }
    }
  ],
  "attribution": "Warnings: Met Office National Severe Weather Warning Service.",
  "captured_at": "2026-10-29T15:41:00Z",
  "last_checked_at": "2026-10-29T15:41:00Z",
  "evaluated_at": "2026-10-29T15:45:00Z",
  "freshness": "fresh"
}
```
