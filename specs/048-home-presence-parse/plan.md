# Plan: Decode Home presence cards

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Updated: 2026-11-05

## Change

`frontend/src/contracts/home/index.ts`, `readHomeCard`: add a `presence` case that rejects
integration fields and expose, requires a bounded `device_id`, and returns
`{ id, panel, kind: 'presence', device_id }`. No other decoder branch changes.

`frontend/src/__tests__/contracts/home/index.test.ts`: parse a server-shaped presence card,
assert the decoded shape, and assert rejection when expose or integration fields are present.

## Verification

| IDs | Paths | Procedure | Result |
| --- | --- | --- | --- |
| AC-01, AC-03–AC-05 | `contracts/home/index.ts`, contract tests | accepted presence decode + device/integration branches unchanged | pass |
| AC-02 | contract tests | presence with expose/instance/resource/service throws | pass |
| SC-01 | all | `npm test` (749 tests), build, lint (0 errors), `type-check:e2e`, `home.spec.ts` (20 passed) | pass |

## Completion

Fixed and verified. A stored presence card now decodes, so affected Home dashboards load again
without data migration; the card can be kept, moved or removed. Server validation and admission
are unchanged.
