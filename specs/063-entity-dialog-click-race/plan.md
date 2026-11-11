# Plan: Entity dialog opens during a hub-state refresh

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft | Updated: 2026-11-11
Constitutions: parent 1.0.0 (proposed); frontend 1.0.0 (proposed)

## Context and design

Observed facts from the failed CI run (43046100850, desktop shard 2/4) and local repeats:

- The failure screenshot showed the Home page with **no dialog**.
- The trace contained no `/metrics/query` request and ended on the dialog/canvas wait.
- Temporary instrumentation logged `NH tile open false false 0x00158d0005a23c38 brightness`:
  `props.editing` false and `device.value` **undefined** at click time.
- `EntityTile.openHistory` returned early on `!device.value`, so no dialog event was
  emitted. The tile remains rendered while the hub store clears for a refresh/reconnect.

Design:

1. **Open by identity.** `EntityTile.openHistory` only checks edit mode. It still emits the
   existing dialog event with the device ID and expose name; the title fallback is the
   device ID. `EntityViewDialog` already resolves the device/expose and shows its
   unavailable state if it is still missing.
2. **Readiness, not retries.** The leak test waits for the dialog and the history canvas
   separately, each with a 15 s timeout. The history service itself allows a 10 s request,
   so the test no longer gives up while the app is still legitimately loading. No click
   retry or test-wide retry masks a dropped open.
3. Keep history fetch, dialog close, and leak assertions unchanged.

Files:

| Path | Change |
| --- | --- |
| `frontend/src/components/entity/EntityTile.vue` | open by identity without the transient device guard |
| `frontend/e2e/EntityViewDialog.spec.ts` | separate dialog/canvas waits above the 10 s history budget |

Alternatives rejected:

- **Retry the click until a dialog appears:** masks a real dropped click instead of fixing
  the open path.
- **Dispatch a synthetic click in the test:** bypasses pointer behavior and would not catch
  the regression.
- **Only increase the test timeout:** the app emitted no dialog event at all, so a longer
  wait could never pass.
- **Retain stale devices during a refresh:** changes store lifecycle and could show removed
  devices as current; the dialog's existing unavailable state is the correct fallback.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 063 records reproduction, root cause and acceptance | PASS | PASS |
| NH-02 | No device command, state, automation or history semantics change | PASS | PASS |
| NH-03 | No new entry point, data exposure, secret or fetch | PASS | PASS |
| NH-04 | Existing dialog event contract; no new dependency or abstraction | PASS | PASS |
| NH-05 | Repeated leak test, exact failed shard, mobile dialog suite, unit/build checks | PASS | PASS |
| NH-06 | No persistence, timer or subscription change | PASS | PASS |
| FE-01 | Tile keeps presentation ownership; dialog keeps device resolution | PASS | PASS |
| FE-02 | No API/WS payload or type change | PASS | PASS |
| FE-03 | History loading/error/empty states are unchanged | PASS | PASS |
| FE-04 | Keyboard/pointer opening remains the existing button interaction | PASS | PASS |
| FE-05 | No secrets or assistant content involved | PASS | PASS |

## Contracts, data, rollout

No wire, storage or configuration change. The only behavior change is that a tile click is
no longer dropped during a transient empty device lookup; the dialog either resolves the
device or shows its existing unavailable state.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, AC-02 | `EntityTile.vue` | repeated leak test; dialog spec | pass |
| FR-02, AC-05 | `EntityTile.vue`, `EntityViewDialog.vue` | dialog spec desktop/mobile; leak bounds unchanged | pass |
| AC-03, NFR-02 | `EntityViewDialog.spec.ts` | separate 15 s dialog/canvas waits | pass |
| AC-04 | leak test | 5/5 repeats; desktop shard 2/4; mobile dialog suite | pass |
| NFR-01 | all | diff review: no dependency/network/store change | pass |

## Verification (recorded)

- Before the fix, `npx playwright test e2e/EntityViewDialog.spec.ts --project=desktop -g
  "closing 50 completed views" --repeat-each=5` failed 2/5; with instrumentation the
  handler ran with `device=false`.
- After the fix, the same 5-repeat command passed **5/5**.
- `npm run test:e2e -- --project=desktop --shard=2/4` (the exact failed shard) →
  **59 passed, 9 skipped**.
- `npx playwright test e2e/EntityViewDialog.spec.ts --project=mobile` → **39 passed**.
- `npm test -- --runInBand` → **146 suites / 821 tests passed**.
- `npm run lint` → **0 errors, 23 pre-existing warnings**.
- `npm run build` → pass; `npm run type-check:e2e` → pass.

## Debt, risks, exceptions

- The store can still be briefly empty during a refresh; the dialog now handles that
  instead of the tile silently dropping the click.
- The dialog title prop remains unused by the dialog heading; the fallback keeps the event
  contract safe without changing presentation.
- No ADR: small, reversible interaction fix with an existing event contract.

## Completion

Implemented on `fix/e2e-entity-dialog-flake`. Acceptance evidence is recorded above; no
exceptions are requested.
