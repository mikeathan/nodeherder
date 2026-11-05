# Contract: `spatialTargets` WebSocket event

Status: Draft (spec 040) | Changed: additive | Consumers: browser only

## Producer

`DeviceLifetimeService` emits one frame per change on the device's lane, after accepted
state and automation inputs are current and after ordinary observation output. Emission
never alters raw state, automation deltas, metrics sampling or `deviceUpdated` payloads.

Cadence rules:

- At most one frame per 100 ms per device (`SPATIAL_FRAME_INTERVAL`).
- A change suppressed by the interval schedules one trailing flush so the final value is
  not lost.
- No frame when no mapped value changed since the previous frame.
- One frame is emitted on device seed. A saved mapping change renders with the next report;
  it is deliberately not emitted from the settings callback, which is not serialized with
  device updates and could read a torn mix of axes.
- One final frame when the last active target clears: every mapped row is reported with
  `active: false` and no axes (not an empty `targets` array), so consumers can tell a cleared
  slot from a device with no mapping.
- Emission stops when the device is unavailable, removed or disabled, and all pending
  timers are cancelled on shutdown.

Only devices with a saved spatial mapping that maps at least one `x`/`y` pair emit frames.

## Payload

```json
{
  "id": "ep-lite-living-room",
  "at": "2026-11-03T10:15:22.480Z",
  "targets": [
    {
      "id": 1,
      "active": true,
      "x": { "value": 1234, "unit": "mm" },
      "y": { "value": 2870, "unit": "mm" }
    }
  ]
}
```

| Field | Type | Meaning |
| --- | --- | --- |
| `id` | string | Existing device identifier (`UpdatePackage.Id`). Not a new identity. |
| `at` | RFC3339 string | Server time when the frame was built. |
| `targets` | array | 0–3 entries, one per saved mapping row that is complete and valid. |
| `targets[].id` | integer 1–3 | Mapping row number. A slot, not a person or a track. |
| `targets[].active` | boolean | Current value of the mapped active entity. |
| `targets[].x`, `y` | object | `value` (finite number) and `unit` as reported by that entity. Present only when the axis value is finite and the unit is a known length. |

Rules:

- One target entry per mapping row whose `x`, `y` and `active` names are all present and
  distinct (X and Y differ, and Active names neither axis), and whose active entity is a
  readable binary sensor. Rows with a missing, non-binary or unreadable active entity are
  omitted entirely: a position without a trustworthy active state could be shown as live.
- `active` reflects the accepted flag. Axes are attached only while `active` is true, and
  only for axes that are numeric, finite and carry a known length unit. A cleared slot must
  not ship a stale position, and consumers must require `active` plus both axes before
  drawing a live position.
- Units are per axis; the browser converts with its existing length table rather than
  assuming X and Y agree.
- Coordinates are the accepted device state, never a prediction. The browser applies hold,
  prediction and smoothing; the server never does.
- Frame silence means "nothing changed". LD2450 firmware publishes coordinates only on
  change, so a still person can legitimately produce no frames. Consumers must not treat
  silence as absence; availability and an explicit `active: false` carry that meaning.

## Consumer

`frontend/src/store/modules/ws/index.ts` maps the event to `hub/updateSpatialTargets`.
The spatial composables prefer the newest frame by `at` and fall back to `deviceUpdated`
for mapped entities when the backend does not send frames (older server) or when frames
stop while the device stays online.

## Errors, compatibility, recovery

- No error replies: the event is unsolicited and additive. A server that does not send it
  is indistinguishable from a device without a mapping, which the fallback handles.
- No persistence, no schema or settings change, no new route, no new authentication
  surface. Frames travel on the existing authenticated socket to clients that already
  receive `deviceUpdated` for the same device.
- Recovery: after a socket reconnect the client resumes from the next frame; it does not
  replay history. A missed frame is corrected by the next one.
