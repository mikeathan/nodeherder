# Opt-in entity history contract

Producer: `backend/internal/api/history.go` → metrics services/repository. Consumer:
`frontend/src/contracts/history.ts` → shared history service/Vuex/renderers. Mock:
`frontend/tools/server/history.mjs`. ADR: [004](adr-004-history-and-quality.md).

`POST /api/metrics/query`, JSON, using the existing API helper and route access.
The existing route is public; this change does not claim new authentication or
broaden it. Absent/false `history` retains legacy HTTP behavior and WS is unchanged.

```json
{
  "history": true,
  "deviceIds": ["fixture-device"],
  "exposes": ["temperature"],
  "time": { "from": "2026-10-17T10:00:00Z", "to": "2026-10-17T14:00:00Z" },
  "aggregation": "",
  "filters": [],
  "limit": 10001,
  "sortDesc": false
}
```

One device, 1–32 exposes, positive explicit range ≤31 days, ascending raw samples,
no filters/lookback/aggregation, limit 1–10,001 per series. Bounds are inclusive.
The response is an array of expose groups; each has at most one device value:

```json
[
  {
    "expose": "temperature",
    "query_start_timestamp": 1792231200000,
    "query_end_timestamp": 1792245600000,
    "values": [{
      "device_id": "fixture-device",
      "unit": "°C",
      "timestamp": 0,
      "value": {
        "name": "temperature", "type": "numeric",
        "from": 1792231200000, "to": 1792245600000,
        "data": [{ "timestamp": 1792231200000, "value": 11.2 }],
        "partial": false
      }
    }]
  }
]
```

All timestamps are epoch milliseconds. Numeric values are numbers, binary values
booleans and enum values strings. Repeated equal reports are retained. Empty data
is an empty array, missing exposes are omitted, and empty units may be omitted.
The client validates identity/type/window/scalars, retains ≤10,000 samples and marks
truncation partial. Equal timestamps retain their original order. No household
history is persisted to browser storage.

Invalid input returns 400; existing rate limiting returns 429; read/cancellation/
scan-budget failures return 500 with a generic actionable JSON `error`. Missing
storage buckets are legitimate empty history. One device scan shares an 8-second
context deadline and 250,000-record callback budget. Browser requests time out at
10 seconds and cancel when the last subscriber leaves. No automatic retry.

Energy `energy` in Wh/kWh is a cumulative meter. Consumption bars contain only
nonnegative deltas between adjacent valid samples inside each bucket. Missing
endpoints, resets and explicit gaps make a bucket partial; unobserved consumption
is null. Neither counter deltas spanning buckets nor power integration are invented.
Events count raw reports; a truncated unscanned tail stays unknown. Binary geometry
may merge equal adjacent states without discarding raw reports or inflating durations.

No persistence schema/key/retention migration. Deploy the compatible backend with
the frontend; rollback is code-only. Existing stored data remains readable.
