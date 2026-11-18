# Contract: `observations` WebSocket event

Status: Draft (spec 064) | Changed: additive | Consumers: browser only

## Producer

The observation service (`internal/observations`) emits one snapshot after aggregating a
bounded batch of normalised observations. It is composed at startup beside the spatial
recorder (`backend/internal/hub.go:137-143`) and broadcasts through the existing
authenticated WebSocket hub.

Cadence rules:

- A snapshot is emitted only when aggregated state changed since the previous snapshot.
- At most one snapshot per 250 ms globally, regardless of inbound observation rate
  (NFR-02). A change suppressed by the interval schedules one trailing flush so the final
  value is not lost.
- Identities whose operator decision is `ignored` are excluded before broadcast.
- Identities with `stable: false` are included but marked, and never counted in the
  arrivals field.
- Emission stops on context cancellation; the worker is joined during shutdown.

Observations never pass through `HubController.Ingest` and never alter device state,
automation inputs, metrics sampling, or `deviceUpdated` payloads.

## Payload

```json
{
  "at": "2026-11-11T10:15:22.480Z",
  "truncated": false,
  "sources": [
    {
      "id": "ep-lite-living-room",
      "name": "Living room proxy",
      "media": ["ble.adv"],
      "facets": ["kind", "vendor", "addressType"]
    }
  ],
  "identities": [
    {
      "key": "ibeacon:2f234454-cf6d-4a0f-adf2-f4911ba9ffa6:1:2",
      "kind": "ibeacon",
      "vendor": 76,
      "stable": true,
      "label": "Front door tag",
      "decision": "known",
      "lastSeen": "2026-11-11T10:15:22.000Z",
      "packets": 1841,
      "rssi": -62,
      "sources": [{ "id": "ep-lite-living-room", "rssi": -62 }]
    },
    {
      "key": "fp:9c02ab41",
      "kind": "adv",
      "stable": false,
      "decision": "unknown",
      "lastSeen": "2026-11-11T10:15:20.000Z",
      "packets": 12,
      "rssi": -81,
      "sources": [{ "id": "ep-lite-living-room", "rssi": -81 }]
    }
  ],
  "arrivals": 0
}
```

| Field | Type | Meaning |
| --- | --- | --- |
| `at` | RFC 3339 UTC | Snapshot production time |
| `truncated` | bool | True when the snapshot hit its cardinality cap |
| `sources` | array | Observing devices that contributed at least one identity |
| `sources[].media` | string array | Media declared by the source, e.g. `ble.adv` |
| `sources[].facets` | string array | Filter facets this source declares |
| `identities` | array | Aggregated identities, most recently seen first |
| `identities[].key` | string | Stable grouping key; `fp:` prefix marks an unstable fingerprint |
| `identities[].stable` | bool | False when derived from a rotating address |
| `identities[].decision` | `unknown` \| `known` \| `ignored` | Operator triage state |
| `identities[].packets` | integer | Count since the current process started |
| `identities[].rssi` | integer or absent | Best (highest) RSSI observed in the window; absent when unknown |
| `identities[].sources[].rssi` | integer or absent | Per-source RSSI; absent when unknown |
| `arrivals` | integer | Count of `stable` identities first seen within the current session |

RSSI is dBm. An absent value means "not observed", which is distinct from a measured
zero (NH-02/FE-03). No distance or position is carried: a single source cannot compute
one, and the contract must not imply otherwise.

## Identity keys

| Kind | Key shape | Stable |
| --- | --- | --- |
| iBeacon | `ibeacon:<uuid>:<major>:<minor>` | yes |
| Eddystone | `eddystone:<namespace>:<instance>` | yes |
| BTHome | `bthome:<device-id>` | yes |
| Public address | `addr:<AA:BB:CC:DD:EE:FF>` | yes |
| Rotating address | `fp:<hash>` | no |

Consumers MUST treat the key as opaque. Changing key derivation for an existing kind is
a breaking change: it would orphan persisted decisions.

## Browser ingress

`frontend/src/contracts/observations.ts` validates the payload before it reaches state.
Unknown `decision` values, a non-array `identities`, a missing `at`, or a non-boolean
`stable` reject the whole snapshot; application state is unchanged and the rejection is
logged without household data. Unknown facet names are ignored rather than rejected, so a
newer backend can add facets without breaking an older browser.

## Actions (additive, browser to hub)

| Action | Payload | Effect |
| --- | --- | --- |
| `saveObservationIdentity` | `{ requestId?, key, decision, label? }` | Upsert one triage decision |
| `deleteObservationIdentity` | `{ requestId?, key }` | Remove the decision, reverting to `unknown` |

Results reuse the existing correlated `deviceConfigSaveResult` pattern rather than
inventing a second acknowledgement channel.

## Compatibility

Additive only. A hub with no observing source broadcasts nothing and shows no
navigation item, so an un-upgraded browser is unaffected. A store written before this
change reads back with no decisions, which presents every identity as `unknown` — the
correct default.
