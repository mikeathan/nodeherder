# ADR-005: Live device updates independent of metrics sampling

Status: Proposed | Date: 2026-10-12 | Requirements: FR-18/19, AC-40…43

## Decision

Publish every accepted changed expose and every repeated physical event through the
existing `OnDeviceUpdated`/`deviceUpdated` path during the same ordered device-lane
Update call. Sampling applies only to the metrics payload. There is no sampling wait,
trailing timer, retry or synthetic acknowledgement. The final change reaches the output
callback before Update returns, including when no later report arrives. Network/render
latency remains separately measured; this is not a wall-clock transport guarantee.

Keep unchanged state deduplicated for automation and metrics. A reported unchanged
writable expose may be forwarded to live consumers as device confirmation evidence;
read-only unchanged state remains deduplicated. Outgoing MQTT commands never confirm
state. Unknown exposes and stopped devices remain ignored. Existing availability and
last-seen semantics, seed behavior, sampling intervals/overrides, normalized measurement
values and automation-before-output ordering remain compatible. No new timer or worker.

Demote routine automation dispatch/disabled and metrics write logs to trace. Dispatch
messages say dispatched, not triggered; actual matching/action diagnostics and errors
remain intact. No household values are added to logs.

## Alternatives and recovery

Increasing command timeout leaves final readings permanently stale. Disabling metrics
sampling changes storage volume. A trailing flush introduces timer ownership and latency;
publishing live updates within the existing lane is simpler and supports confirmations.

Existing WS fields and MQTT payloads are unchanged. Older UI consumers remain compatible;
backend rollback restores old sampling behavior without storage migration. Faster output
also invokes existing device persistence more often; validate controlled throughput before
claiming broader performance acceptance. No live-device configuration changes.

## Verification

Synthetic-clock regressions cover default numeric sampling, final change then silence,
unchanged writable confirmation without automation/metrics, zero, repeated events,
disabled/unknown/availability paths and overrides. Existing dial scenarios and affected
race suites verify automation compatibility. Browser latency, frontend draft rollback
and SFC quality preparation remain separate follow-up work.
