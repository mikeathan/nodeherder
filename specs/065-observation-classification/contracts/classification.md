# Contract: classification on the `observations` snapshot

Status: Draft (spec 065) | Changed: additive | Consumers: browser only

## Producer

The observation service adds classification to each `IdentityView` it already broadcasts.
Classification is computed on the aggregation worker from stored daily summaries plus the
in-memory window; it is never persisted and never computed per render.

Cadence follows the existing snapshot: at most one broadcast per 250 ms, on change only. A
class change counts as a change, so a reclassification reaches the browser without waiting
for an identity update.

## Payload

Additive fields on the existing identity object:

```json
{
  "key": "addr:aa:bb:cc:dd:ee:01",
  "kind": "adv",
  "stable": true,
  "decision": "unknown",
  "class": "personal",
  "confidence": 0.72,
  "evidence": {
    "dutyCycle": 0.34,
    "hoursObserved": 41,
    "crossings": 27,
    "nightPresence": false
  },
  "changedAt": "2026-11-10T06:12:00Z",
  "previousClass": "infrastructure"
}
```

| Field | Type | Meaning |
| --- | --- | --- |
| `class` | string | `infrastructure`, `personal`, `neighbour`, `transient`, `learning`, or a provisional value |
| `provisional` | boolean | True when the class came from the first-packet pre-filter, not the classifier |
| `confidence` | number 0–1 | Margin over the runner-up class; 0 while learning |
| `evidence.dutyCycle` | number 0–1 | Fraction of observed hours present, over the window |
| `evidence.hoursObserved` | integer | Hours of evidence behind the class |
| `evidence.crossings` | integer | Distance-band crossings, the mobility signal |
| `evidence.nightPresence` | boolean | Heard between 00:00 and 05:00 UTC |
| `changedAt` | RFC 3339 UTC, optional | When the class last changed |
| `previousClass` | string, optional | The class before that change |

### Semantics

- `learning` means insufficient evidence, not "unknown class". A consumer MUST render it as a
  learning state with `hoursObserved`, never as a class.
- **Provisional classes** (`provisional: true`) are the day-one pre-filter, available before any
  history exists. There are two: `mobile-unidentifiable` for a rotating address, and
  `likely-neighbour` for a device never heard strongly. A consumer MUST render them as
  provisional and MUST NOT present them as settled. They exist so the first day is not the full
  census, and they are superseded by a real class once evidence exists.
- `class` is **advisory**. An operator triage decision (`decision: "known" | "ignored"`)
  takes precedence for display, and the class is still carried so the UI can show both.
- `neighbour` is a heuristic, not a measurement. With one source it cannot be separated from
  "far inside the property"; consumers MUST word it as likely, never as fact.
- `confidence` below roughly 0.2 means the classes are nearly tied; the UI should present the
  class as uncertain rather than assert it.

## Browser ingress

`src/contracts/observations.ts` validates the new fields. All are optional so a payload from a
backend without this feature still parses; a missing class is treated as `learning`, never
guessed. An unrecognised class value rejects the identity rather than coercing it, because a
coerced class would be a false claim about behaviour.

## Summary history route

`GET /api/wireless/summaries?key=<identityKey>&days=<n>`

Protected. Returns daily summaries newest first, bounded to the 30-day retention window and to
a maximum of 30 rows. Used by the Timeline pane. Contains aggregates only: no payloads, no
addresses beyond the identity key the caller already supplied.

Class transitions are retained on the same 30-day boundary, so every transition the UI can
show is one whose supporting evidence still exists.

## Compatibility

Additive only. An older browser ignores the fields. A backend without summaries emits
`class: "learning"` for every identity, which the current UI renders as a learning state, so
the two versions interoperate in both directions.
