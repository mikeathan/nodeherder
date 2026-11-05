# Spec: Multi-sensor presence fusion (presentation)

ID: 047 | Status: Implemented (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

A room or home with several spatial sensors has no single answer to "is anything present, and
do the sensors agree?". Each presence tab only describes its own sensor.

In scope: a read-only Fusion view that automatically includes every spatial-capable device with
a hub-derived presence device, and combines them: occupancy as OR (any known occupied), reported
target counts as a labelled sum (not people), per-sensor zone state, online/offline status and
explicit disagreement/unknown notes.

Exclusions, stated in the view: no fused automation entity (automations keep using each
sensor's derived entities), no cross-sensor identity or same-person deduplication, no
persistence, no backend or route changes, and no inference from a sensor being silent.

## Scenarios and acceptance

### US-01 — Combined presence across sensors (P1)

- AC-01: The view discovers spatial-capable devices that have a derived presence device; no
  manual selection is required and no non-spatial device appears.
- AC-02: Combined occupancy is occupied when any sensor with a known value reports occupied,
  clear only when every known sensor reports clear, and unknown when no sensor has a known value.
- AC-03: Combined target count is the sum of known per-sensor `target_count` values, labelled as
  reported targets across sensors, not people.
- AC-04: Each sensor row shows its name, derived online/offline state, occupancy, known target
  count and the names of zones it currently reports occupied.
- AC-05: Disagreements are explicit: one sensor occupied while another is clear, count spread
  greater than one, and offline sensors are listed without being treated as clear.
- AC-06: With no derived sensor presence the view explains why and lists nothing invented.
- AC-07: Rows link to the source device's page; the view performs no writes and feeds no
  automation.
- AC-08: Keyboard operable and fits 375/768/1440 px.

## Requirements and success

- FR-01: Fusion maths MUST be pure and tested in `domain/spatial/fusion.ts`.
- FR-02: Inputs MUST come only from existing derived presence entities in the store.
- FR-03: Unknown and offline MUST never be converted to clear.
- NFR-01: No new route, request, timer or persistence.
- SC-01: Domain tests cover OR semantics, unknown, sums, disagreements and offline.
- SC-02: Full frontend checks plus browser specs pass; results recorded.

## Questions and assumptions

- Q-01: Summed counts are presented as reported targets, because sensors cannot prove they track
  different people and overlapping coverage can double-count.
- Q-02: Automations deliberately remain per sensor until a fused entity has defined room/sensor
  membership and ownership. Recorded as a known gap, not implemented here.
