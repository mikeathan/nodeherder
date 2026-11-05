# ADR-001: Spatial target frames as the live spatial contract

Status: proposed | Date: 2026-11-03
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: unreviewed draft | Rules: NH-02, NH-04, BE-01, BE-02, FE-02

## Context and options

The Spatial tab needs coordinates at a usable rate and free of axis skew. Today
`publishSpatialFreshness` forwards each mapped axis separately, at most every 250 ms per
entity, and reuses the general `deviceUpdated` event
(`backend/internal/services/device_lifetime.go:173-206`,
`backend/internal/controllers/hub-controller.go:846`). A composed point can therefore mix
axes up to 250 ms apart, which is ~250 mm of sideways error at walking speed.

The ordinary observation path is not a candidate: mapped coordinate entities are
ESPHome `diagnostic` entities, so the default live interval is 30 s
(`backend/models/settings/reporting-defaults.json`), and that interval exists to bound
UI/metrics traffic for every consumer. Automation and metrics already read accepted state
before any of this, so changing their cadence is neither needed nor wanted.

Options considered:

1. **Tell users to set a 0 s live-interval override per coordinate entity.** No code, but
   changes every consumer's cadence, still sends axes independently, needs six overrides
   per device, and re-breaks whenever entities are renamed.
2. **Per-connection subscription** (`spatialStreamStart`/`Stop`), streaming only to
   sockets that ask. Most precise; adds per-connection state, request correlation, lease
   cleanup and a second emission path, and contradicts the current "clients already
   receive this device's updates" model.
3. **Hub-level refcounted live mode.** Middle ground, but still introduces acquire/release
   state and its own failure modes for a payload measured in hundreds of bytes.
4. **Purpose-built frame event emitted whenever a device has a saved mapping**, at most
   every 100 ms, change-driven. No new connection state, no new route, complete
   synchronized payload, additive contract.
5. **Client-side pairing only.** Cheap, but cannot recover axes the server already
   separated, and leaves the 250 ms per-axis cap in place.

## Decision and consequences

Take option 4: a new `spatialTargets` event carrying complete per-target frames
([contracts/spatial-targets.md](contracts/spatial-targets.md)), emitted only for devices
with a saved mapping, change-driven, capped at 10 Hz with a trailing flush, stopping when
the device is offline or removed.

Rationale: frames remove skew at the source, keep the ordinary observation policy
untouched (NH-02), need no connection bookkeeping, and leave older browsers working
because the fallback path still exists. The cost is a small amount of always-on traffic
for mapped devices (≤10 frames/s, ≤1 KB each, nothing when nothing changes) and one more
event name to keep truthful.

Verification: contract tests with a fake clock for rate cap, trailing flush, no-change
suppression, offline stop and shutdown cancellation; browser tests for preference of the
newest source and for fallback labelling.

Reconsider if a mapped device's idle traffic becomes measurable, if a client needs a
different cadence per view, or if a second live consumer (for example a Home widget) needs
subscription-style control. Then option 2 becomes worth its bookkeeping.

## Supersession

Not superseded. Supersedes nothing; `deviceUpdated` remains the general device event.
