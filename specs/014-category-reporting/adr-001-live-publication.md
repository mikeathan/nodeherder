# ADR 001: Share policy resolution; isolate live publication state

Status: Proposed implementation decision | 2026-10-16

The history sampler keeps its first eligible reading and has no trailing flush.
Live diagnostics need a latest-value flush so silence cannot leave stale UI values.
A single generic debounce used by both consumers would silently change history
semantics or lose the final live reading. A whole-device throttle would delay presence
and command evidence alongside diagnostics.

Use shared, ordered category/field policy resolution, retaining existing history
configuration fields. Compose the existing numeric delta filter into a separate
live ObservationPublisher. This service owns bounded pending output and one timer
per device; raw state and automation remain on existing device lanes. The output
mutex serializes immediate and timer publication, cancellation and baseline changes;
timer callbacks never write raw state. Existing output callbacks are synchronous
and do not re-enter the publisher. A generation token rejects already-started stale
timer callbacks after rescheduling or cancellation. Hub context cancellation closes
publication. Device disable/offline drops pending values; config edits re-evaluate them using
current rules, retaining the previous publication time as the interval anchor.

No new dependencies or protocol changes. Costs: one retained numeric baseline and
last-publication time per known field, at most one pending value per field, one timer
per device with pending work. Bursts overwrite those values rather than creating
per-reading timers. Existing availability worker lifecycle is not redesigned here.
