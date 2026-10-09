# Device and protocol support (current and future)

Part of [plan 007](plan.md). Documents how devices reach the UI today and how the
redesign keeps presentation protocol-agnostic.

## How devices arrive today

| Path | `connection_type` | Code | Typical devices |
| --- | --- | --- | --- |
| Zigbee2MQTT over MQTT (`zigbee2mqtt/<friendly_name>`, `bridge/*` topics) | `mqtt` | `backend/internal/mqtt/`, `internal/controllers/handlers.go` (`deviceHandler`), bridge handlers (permit join, rename, remove, interview, logging) | Zigbee sensors, lights, plugs, dials, alarms |
| HTTP push `POST /api/collect` (JSON, `utils.ParsePayload` extracts id) | `http` | `backend/internal/api/routes.go` `DataCollectorHandler` → `HubController.Enqueue(id, payload, "http")` | Wi-Fi devices: ESP32/ESP8266 (ESPHome HTTP request, Tasmota rules, Arduino), Shelly scripts, any script |

Both paths converge on `DeviceProcessor.CreateOrUpdateDevice` → per-device lanes →
`DeviceLifetimeService` (spec 005/006). Devices with a Zigbee2MQTT bridge definition get
typed exposes from `bridge/devices`; others (incl. all HTTP devices) get exposes inferred
from payload keys (`createExpose`) with category whitelists. Bridge-only operations
(permit join, interview, remove/block, LQI) only apply to `mqtt` devices.

Browser view: `Device.connection_type` (string), `power_source` (`battery`|`mains`),
`exposes` with `type`/`access_mode`/`category`/`values`/`attributes`. Today only
`ConnectionType.vue` looks at the protocol (logo image).

## Protocol registry (FR-07)

`src/registry/protocols.ts`:

```ts
type ProtocolDescriptor = {
  id: string;                 // matches connection_type
  label: string;              // "Zigbee", "Wi-Fi · HTTP"
  icon: IconRef;              // token-coloured glyph
  diagnostics: DiagnosticKey[]; // which chips to show: 'lqi' | 'battery' | 'lastSeen' | 'rssi' | 'ip'
  supports: { permitJoin?: boolean; interview?: boolean; remove?: boolean; rename?: boolean };
  docsUrl?: string;
};
const registry: Record<string, ProtocolDescriptor> = {
  mqtt: { id: 'mqtt', label: 'Zigbee', diagnostics: ['lqi','battery','lastSeen'],
          supports: { permitJoin: true, interview: true, remove: true, rename: true } },
  http: { id: 'http', label: 'Wi-Fi · HTTP', diagnostics: ['rssi','battery','lastSeen','ip'],
          supports: {} }, // to verify: rename/interview/remove are Zigbee2MQTT bridge requests
};
export const resolveProtocol = (t: string) => registry[t] ?? genericProtocol(t); // AC-08
```

Device actions (rename/interview/remove) are shown only when `supports` allows them,
which fixes today's offering of Zigbee bridge actions on HTTP devices. Diagnostic chips
render only when the underlying expose exists (`linkquality`, `battery`, `rssi`/
`wifi_rssi`, `ip`), so HTTP devices that report RSSI get a Wi-Fi signal chip for free.

### Future protocols (no UI rewrite needed)

| Candidate | Backend work (separate spec) | UI work |
| --- | --- | --- |
| Native MQTT devices (Tasmota/ESPHome MQTT, Shelly Gen2 MQTT) | topic adapter producing `connection_type: 'wifi-mqtt'` | registry entry |
| ESPHome native API | adapter service | registry entry, `rssi` chip |
| Matter / Thread | bridge adapter (e.g. matter.js server) | registry entry; commissioning flow analogous to permit join (`supports.pair`) |
| Bluetooth LE (via ESPHome proxy / Theengs) | adapter | registry entry, `rssi` chip |
| Z-Wave (Z-Wave JS UI MQTT) | adapter | registry entry, inclusion = permit join |
| Cloud/webhook integrations | HTTP collector variants | registry entry |

The UI contract that must hold for any new protocol: unique `id`, `friendly_name`,
`availability`, `last_seen`, typed `exposes`. Anything protocol-specific belongs in
`exposes` (diagnostic category) so capability rendering keeps working.

## Capability registry (FR-08)

`src/registry/capabilities.ts` resolves an expose to presentation + control using only
generic metadata, consolidating `sensor-formatter.ts`, `device.config.ts` filters and
`DeviceExposes.vue` branching:

| Expose | Detected by | Tile/readout | Control |
| --- | --- | --- | --- |
| Switch/state | binary, writable, measurement | on/off icon glow | toggle (pending-aware) |
| Brightness/level | numeric, writable, `value_min/max` attrs | % bar | slider |
| Colour temperature | name `color_temp` + min/max | warm→cool gradient | slider |
| Preset/enum | enum or numeric with `values` | label | segmented / select |
| Read-only numeric | numeric, read | value + unit + sparkline | — |
| Binary sensor | binary, read (contact, occupancy, smoke, tamper) | state words (Open/Closed, Detected/Clear) | — |
| Action/event | name `action*` | last event + time | — (automation source) |
| Diagnostic | category diagnostic (linkquality, battery, voltage) | chips | — |
| Config | category config | Device page → Settings only | input by type |

Name-specific knowledge (icons, friendly labels, value words) lives in one keyed table
with a generic fallback (unknown expose → generic icon, raw name title-cased, raw
value + unit). This is what makes new devices "just work".
