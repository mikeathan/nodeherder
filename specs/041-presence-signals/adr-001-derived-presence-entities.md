# ADR-001: Derived presence entities as a separate device

Status: proposed | Date: 2026-11-03
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: unreviewed draft | Rules: NH-02, NH-04, BE-01, BE-04

## Context and options

Zone occupancy, counts, dwell and "room empty for N" must be usable by automations with no
browser open. Today they exist only in the browser: filtering, track association and zone
hysteresis run in `frontend/src/domain/spatial/`. Automations read confirmed device state
from the backend, so a browser-side value cannot be a trigger.

The firmware already reports its own zone occupancy, but only for the four rectangle/polygon
zones it knows about, with its own delayed-off, and only when the user configured firmware
zones. NodeHerder's drawn zones may be overlays only, may exceed or differ from firmware
zones, and carry names the user chose.

Options considered:

1. **New automation trigger type (`spatialZone`).** Explicit, but adds a second trigger
   pipeline, a new editor, new validation, and duplicates history/metrics/MCP plumbing that
   already exists for device entities.
2. **Virtual entities appended to the source device's expose map.** No new device, but the
   source inventory belongs to its provider (ESPHome re-discovery replaces it), which would
   destroy derived entities on reconnect and mix firmware truth with derived state.
3. **A separate derived device per mapped source** with read-only exposes, registered through
   the existing provider seam. Reuses triggers, history, metrics and MCP unchanged; the source
   inventory stays provider-owned; lifecycle is explicit.
4. **Compute only while the presence tab is open.** Cannot drive automations headless.

## Decision and consequences

Take option 3. For every source device with a saved mapping and at least one zone, the hub
registers one derived device, provider `presence`, connection type `derived`, with stable
read-only entities:

| Entity | Type | Meaning |
| --- | --- | --- |
| `occupied` | binary | any track inside any detection zone |
| `target_count` | numeric | tracks inside detection zones |
| `zone_<hash>_occupied` | binary | zone occupancy with hysteresis (one per zone) |
| `zone_<hash>_count` | numeric | tracks inside that zone while it is occupied (same debounce) |
| `zone_<hash>_dwell_seconds` | numeric | seconds that zone has been continuously occupied |
| `last_zone` | enum/text | most recent occupied zone name, or empty |
| `empty_seconds` | numeric | seconds since any tracked target was seen |

`<hash>` is the first 12 hex characters of the zone UUID; the zone name travels in the entity
label and a `zoneId` attribute, so renames never change entity names. Occupancy and counts
publish on change; dwell and empty time publish every 10 s and on transitions, so history does
not grow at frame rate. Occupancy and counts are measurement category; dwell, last zone and
empty time are diagnostic, so they are not recorded as history unless a user opts in.

The **hub** is the authority for these values. The presence tab keeps client smoothing for the
dot, but zone occupancy/count shown to the user come from the derived device when it exists, so
the screen and the automation cannot disagree.

Costs and limits: a second device per mapped sensor in the device list; an engine that must
mirror the filtering rules closely enough for stable occupancy (documented and tested with the
same synthetic sequences as the browser, not shared code); derived devices must never be
commandable, never be treated as providers to enroll, and must disappear with their mapping.

Reconsider if derived entities are ever wanted for non-spatial devices or if the Home block
needs a different shape.

## Supersession

Not superseded.
