# Spec: Device service retirement and launcher CI coverage

ID: 016 | Status: Complete | Updated: 2026-10-16
Constitution: v1.0.0 proposed | Components: backend, CI tooling

## Problem, scope and exclusions

Successful bridge removal deletes storage but retains DeviceProcessor services,
availability workers and publisher shutdown registrations. Long-running device churn
therefore retains resources. The local launcher's five lifecycle regressions also
lack CI execution. Fix ownership and cleanup without changing device commands,
automation semantics, history, UI or transport payloads. No commit or deployment.

## Acceptance

- AC-01: stack-tools runs scripts/dev.test.mjs, required by the aggregate CI gate.
- AC-02: Successful removal runs behind in-flight device-lane work. Serialize deletion
  against publisher output so an older callback cannot reinsert the deleted device.
  Stop pending publication, unregister its shutdown callback, cancel/wait availability
  workers and evict the processor entry and queued creation events.
- AC-03: Repeated removal is safe. Old service references cannot process updates or
  be re-enabled by configuration. A subsequent accepted rediscovery uses a fresh
  service; preserve existing registration semantics without an unbounded blacklist.
- AC-04: Failed storage removal reports failure and leaves the original service,
  monitoring and pending publication operational. Unrelated devices remain independent.
- AC-05: Hub cancellation stops publishers and availability workers. Close is
  idempotent; no callback survives completed retirement. Synchronization tests use
  channels/fake clocks and run under the race detector, never household devices.

## Interfaces and quality

Internal removal callback wiring only; existing success/failure messages stay intact.
Native launcher unchanged. Resources bounded by active services/known exposes.
Removal may wait for synchronous existing I/O; lane start has a bounded deadline.
No new per-reading goroutines, timers or global locks; no new dependencies.
