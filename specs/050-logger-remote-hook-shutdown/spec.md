# Spec: Remote logger hook shutdown race

ID: 050 | Status: Fixed (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: backend logging

## Problem

Backend tests intermittently panicked during teardown:

```
panic: runtime error: invalid memory address or nil pointer dereference
node-herder/utils.(*RemoteHook).Fire(...)
    backend/utils/logger.go
...DeviceLifetimeService.startAvailabilityMonitoringLocked.func1()
```

`RemoteHook.Fire` read `emitter` without synchronisation while `RemoveRemoteLoggerHook`
replaced it with nil (`Configure(newRemoteHook().emitter)`) and replaced the logrus hook set.
A background availability ticker logging during teardown could pass the enabled check, then
call `Broadcast` on a nil emitter. The race affected any package that tears down the remote
hook while device goroutines are still finishing (observed in `internal/services`).

## Scope and acceptance

- AC-01: `Fire` never calls a nil emitter; removing/unconfiguring the hook while goroutines log
  is safe and does not panic.
- AC-02: `Configure` and `Fire` may run concurrently without a data race (`-race` clean).
- AC-03: Emitter callbacks may still disable the hook without deadlocking; delivery semantics
  for an enabled hook are unchanged.
- AC-04: Removal disables delivery before detaching the emitter; late in-flight calls either
  deliver once to an already-detached emitter or return without delivery, never panic.
- AC-05: Stress tests cover concurrent Fire/Configure and repeated services runs, plus full
  backend suites and race checks.

## Requirements and success

- FR-01: Guard the emitter with a read/write lock or equivalent atomic snapshot; nil is a no-op.
- FR-02: Do not call the emitter while holding the lock (callbacks may reconfigure).
- SC-01: New concurrency test; `go test ./internal/services/ -count=15` and `-race` repeated runs.
- SC-02: Full backend build/tests and CI pass.

## Questions and assumptions

- Q-01: The panic is a shutdown race in the hook, not a device-lifetime logic error; background
  logging during teardown is legitimate, so the hook must tolerate removal.
