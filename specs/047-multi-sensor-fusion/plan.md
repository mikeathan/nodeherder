# Plan: Multi-sensor presence fusion (presentation)

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Context and design

Every source device already has a derived `presence-<id>` device in the store. A fusion view can
read those entities client-side without new transport, timers or persistence.

- `domain/spatial/fusion.ts`: `FusedSensor` input and `fuseSensors` returning combined occupancy
  (OR with unknown), summed known targets, online/total, per-sensor zone occupancy and
  disagreement/offline notes. Pure and tested.
- `SpatialFusionPanel.vue`: builds inputs from `useHub().devices` + `derivedPresenceDeviceId` +
  `derivedPresenceValues`, renders the summary, sensor rows, zone chips and links; explicit
  boundaries (no fused automation entity, targets not people).
- New Fusion view in the presence tab; no backend, store or layout change.

## Traceability

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| FR-01–FR-03, AC-02–AC-05 | `fusion.ts` | `fusion.test.ts`: OR, all-clear, unknown, sums, disagreement notes, offline not clear | pass |
| AC-01, AC-06–AC-08 | `SpatialFusionPanel.vue`, `view.ts` | type-check/build; discovery and rendering reviewed; browser not driven yet | pass (type) |
| SC-02 | all | `npm test` (747 tests), build, lint (0 errors), `type-check:e2e`, `devices.spec.ts` (18 passed) | pass |

## Completion

Implemented and checked for presentation-only fusion: automatic discovery of spatial sensors
with derived presence, OR occupancy with unknown preserved, summed reported targets, occupied
zones, disagreement/offline notes and device links. A fused automation entity and room/sensor
membership remain explicitly out of scope (Q-02), and no real multi-sensor data has been
inspected; the view is type/unit-tested, not browser-driven.
