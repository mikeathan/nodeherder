# Plan: Remote logger hook shutdown race

Spec: [spec.md](spec.md) | Status: Fixed (unreviewed) | Updated: 2026-11-05

## Change

`backend/utils/logger.go`: add an RWMutex to `RemoteHook`, guard `emitter` in `Configure`,
snapshot it under RLock in `Fire`, return nil when the snapshot is nil, and release the lock
before `Broadcast` so emitter callbacks can call back into the hook. `RemoveRemoteLoggerHook`
disables the hook, detaches the emitter, then replaces the logrus hooks.

`backend/utils/logger_test.go`: concurrent `Fire` against repeated `Configure(nil)`/re-configure,
asserting no panic and no error.

## Verification

| IDs | Procedure | Result |
| --- | --- | --- |
| AC-01–AC-04 | new `TestRemoteHookFireDuringUnconfigureDoesNotPanic`; existing toggle/emitter-callback tests | pass |
| AC-02 | `go test ./utils/ -run RemoteHook -count=5 -race` | pass |
| AC-05 | `go test ./internal/services/ -count=15`, `-count=5 -race`, full backend suite | pass |

## Completion

Fixed and verified locally. The failing CI job was a teardown race; this removes the nil-emitter
window without changing delivery semantics.
