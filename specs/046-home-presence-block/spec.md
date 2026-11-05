# Spec: Home presence block

ID: 046 | Status: Implemented (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: backend settings/API + frontend Home

## Problem, scope, exclusions

Presence information lives on each device's Presence tab. A Home dashboard can only place
single expose tiles, so answering "is anyone home / which zone is occupied" needs several tiles
and still lacks zone context.

In scope: one Home presence card per spatial-capable device, rendered from the hub-derived
presence entities, plus Home layout persistence/identity/admission for the new card kind.

Exclusions: a new route, live canvas or editing on Home, Panel-mode rendering (Home only),
presence aggregation across devices (fusion is a later increment), and values computed in the
browser.

## Scenarios and acceptance

### US-01 — Place and read a presence card (P1)

- AC-01: The tile picker offers exactly one "Presence" choice for a device that can carry a
  spatial mapping; devices without that capability offer none.
- AC-02: The card shows the device name and availability, hub-derived occupied/clear/unknown,
  target count when known, last zone when present, and empty seconds when clear.
- AC-03: Zone chips come from derived zone entities and their `zoneName`/`zoneId` attributes;
  unknown occupancy is shown as unknown, never as clear.
- AC-04: Without a derived device or before presence exists, the card explains that presence
  has not been derived instead of inventing values.
- AC-05: Removing, moving and layout editing use the existing Home mechanics; a missing device
  remains an explicit removable placeholder.
- AC-06: Presence cards are excluded from Panel mode, because Panel projects device controls only.
- AC-07: The backend accepts a presence card that references an existing device with no expose,
  rejects mixed device/expose or integration references, and keeps source identity device-wide.
- AC-08: The card fits 375/768/1440 px and remains keyboard accessible.

## Requirements and success

- FR-01: The card MUST read only derived presence entities through the existing store.
- FR-02: Layout persistence MUST validate the new kind additively; existing layouts stay valid.
- FR-03: Picker admission MUST be limited to `supportsSpatialMapping` devices.
- NFR-01: No new route, polling loop or data exposure.
- SC-01: Backend validation tests and frontend identity/picker/panel/title tests.
- SC-02: Full checks plus Home and devices browser specs pass; results recorded.

## Questions and assumptions

- Q-01: One presence card per source device is enough; per-zone cards can be added later from
  the derived device's entities.
- Q-02: Panel mode intentionally omits presence until a wall-display design exists.
