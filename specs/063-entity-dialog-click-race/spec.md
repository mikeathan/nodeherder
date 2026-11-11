# Spec: Entity dialog opens during a hub-state refresh

ID: 063 | Status: Draft | Owner: agent draft | Updated: 2026-11-11
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

After #107 merged, the desktop E2E shard 2 leak test failed waiting for the entity
dialog's history canvas. The trace shows the tile click handler ran with
`device.value === undefined`, so `EntityTile.openHistory` returned early: no dialog and no
`/metrics/query` request were emitted. The hub store clears devices briefly during a
reconnect/refresh while the tile stays rendered. Local repeats reproduced the failure 3/5
times before the fix.

Scope: open the entity dialog by stable device/expose identity even when the store lookup
is momentarily empty, and let the existing dialog handle a device that is still absent.
Give the leak test separate readiness budgets above the history service's 10 s request
timeout.

Exclusions: changing history loading, dialog lifecycle, device store refresh semantics,
test-wide retries, or leak assertions.

## Scenarios and acceptance

### US-01 — Open from a tile during a refresh (priority P1)

Independent verification: repeated leak test, the exact failed shard, and the dialog spec
on mobile.

- AC-01: Given a tile remains rendered while its device lookup is transiently empty, when
  its body is activated, then the dialog opens with the tile's device ID and expose name.
- AC-02: Given the device is still absent when the dialog resolves it, then the existing
  unavailable state is shown and no error is thrown.
- AC-03: The leak test waits for the dialog and history canvas independently, with each
  timeout above the app's 10 s history request budget.
- AC-04: Fifty open/close cycles pass repeatedly on desktop and mobile, and the exact
  previously failing shard passes.
- AC-05: Leak bounds, history behavior, and dialog close behavior remain unchanged.

## Requirements and success

- FR-01: Remove the `!device.value` early return from `EntityTile.openHistory`; use the
  device ID as the title fallback; acceptance: AC-01, AC-02.
- FR-02: Keep emitting the existing `emitOpenEntityViewDialog` event and resolving the
  device/expose inside the dialog; acceptance: AC-05.
- NFR-01: Add no dependency, network call, storage or history/dialog lifecycle change.
- NFR-02: Readiness waits MUST exceed the history service's request budget.

## Entities and external interfaces

- `EntityTile.openHistory` keeps the existing dialog event contract.
- `EntityViewDialog` keeps owning the missing-device display.
- No API/WS/store contract change.

## Questions and assumptions

- Q1: Retry the click or fix the guard? Resolution: fix the guard. A retry hides a real
  dropped click; opening by identity is correct and makes the test deterministic.
- Assumption A1: A permanently removed device showing the dialog's existing “no longer
  available” state is acceptable and preferable to silently ignoring the click.

Success criteria:

- SC-01: The leak test passes 5/5 local repeats.
- SC-02: Desktop shard 2/4 and the mobile dialog suite pass.
- SC-03: Unit tests, lint, build and e2e type-check remain green.
