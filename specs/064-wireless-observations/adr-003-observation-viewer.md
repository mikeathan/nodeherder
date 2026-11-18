# ADR-003: Wireless viewer geometry

Status: proposed (no review recorded) | Date: 2026-11-11
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: none recorded |
Rules: FE-01, FE-03, FE-04

## Context and options

The viewer must render observed identities in a way that reads as motion. Two facts rule
out a straightforward reuse of the existing spatial renderer.

**Physics.** A single ESPHome proxy reports RSSI only. RSSI maps to a distance via
log-distance path loss (`d = 10^((A - rssi) / (10n))`), which yields a **circle, not a
point**. BLE 5.1 direction finding (AoA/AoD) requires antenna arrays and CTE-capable
transmitters; the ESP32 does not provide it. With three or more proxies, multilateration
produces a real position. Below that, placement must be explicitly bearing-less.

**Accuracy.** RSSI→distance error is 1–3 m indoors, and 10 dB ≈ 3× distance, so ordinary
±6 dB packet jitter moves a blip twofold. Rendering raw packets as moving dots produces a
snowstorm of noise, not motion.

Options:

1. **Reuse `SpatialCanvas.vue`.** It is millimetre-Cartesian with zones, workspace
   geometry, sensor cones, heat cells, and editing handles. The wireless view has none of
   those. Extending it would add a bearing-less mode to a node designed around bearings,
   and would couple two unrelated features.
2. **Shared primitive module.** Extract a generic polar/cartesian drawing kernel used by
   both. Attractive in principle, but the spatial canvas is built around workspace
   projection and editing gestures, so the extracted surface would be large and shallow.
3. **Separate component sharing only primitives (chosen).** New polar renderer for RSSI
   rings; share only `radarViewBox`/`rotatePoint` (`domain/spatial/radar.ts`), the view-id
   pattern (`domain/spatial/view.ts`), the composable shape
   (`composables/spatial/useSpatialView.ts`), and `utils/units.ts` formatting.

## Decision and consequences

Choose **option 3**. The viewer is a polar range display with logarithmic rings, plus a
packet table. Aggregate state renders on the canvas; raw packets render in the table.

Consequences:

- Some drawing code is not shared. Accepted: the two views answer different questions
  with different data shapes.
- Placement below three sources is explicitly labelled range-only and bearing-less, so
  the UI never implies a position it cannot compute.
- The canvas consumes filtered, aggregated snapshots; the table exposes per-packet
  detail. This split is the mitigation for RSSI jitter.
- A future trilateration path upgrades placement without changing the component
  contract: the snapshot gains a position and the canvas prefers it when present.
- No mm-space concepts enter the wireless view, and no bearing concepts enter the spatial
  view.

Verification: recorded keyboard operation, accessible names, and 390px/1280px checks;
unit tests for projection math and for the "no bearing" labelling.

## Supersession

None.
