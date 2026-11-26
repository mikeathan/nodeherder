# Tasks: UI trust and daily surfaces

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: in progress

Each task maps to requirement/acceptance IDs or a constitutional gate. Evidence is recorded only
after the check ran.

## Foundation

- [x] T001 — Freshness domain in `frontend/src/domain/devices/freshness.ts`; IDs: FR-01/AC-01/02/04;
  done: `freshness.test.ts` 5 passing, mains/battery/unknown/future cases.
- [x] T002 — Snapshot age in `hub-state` + throttled WS touch; IDs: FR-02/AC-03; done:
  `hub-state/index.test.ts` 2 passing; getter/clear covered. Banner UI in T008.
- [x] T003 — Durable notification store + persisted state + logout clear; IDs: FR-03/AC-05–08;
  done: `notifications/index.test.ts` 4 passing (dedupe, read, cap, clear).
- [x] T004 — Connection and operation producers; IDs: AC-05/06; done: code path in
  `useNotificationProducers.ts` and `ws/index.ts`; **no automated test yet** (see T012).

## Surfaces

- [x] T005 — Shared `FreshnessLabel` and tile/list/device-page integration; IDs: FR-01/FR-06/
  AC-01/15/16; done: type-check/build + desktop `devices`, `home`, `EntityTile` specs pass.
- [x] T006 — Home status strip and room/device health; IDs: FR-06/R14; done: Home desktop/mobile
  geometry specs pass.
- [x] T007 — Calm Overview (A): verdict, three actions, health line, digest, collapsed full feed;
  IDs: FR-05/AC-12–14; done: `overview` + `ActivityFeed` specs pass (details opener updated).
- [x] T008 — Snapshot banner + logs IA: Settings → Logs & diagnostics with archive list/download,
  header Logs shortcut, Console removed from nav, local filters/follow/export; IDs: FR-04/AC-09–11;
  done: `logs` service tests 3 passing; storage/settings/overview specs pass. Deviation: the
  existing `GET /api/listlogs` returns names only, so size/age are not shown.
- [x] T009 — Settings rail summaries and section title alignment; IDs: FR-08/AC-20; done:
  `Settings.spec.ts` 22 passing; all six sections still present.
- [x] T010 — Automations export all/per row with versioned JSON; IDs: FR-08/AC-22 (export only);
  done: `automation-export` test passing; duplicate/copy-ID/run-history remain open.
- [x] T011 — Panel mode: schedule night (22:30–06:30) + manual override, four controls in night,
  paused/last-known controls while offline, status page retained; IDs: FR-07/AC-17–19; done:
  desktop/mobile panel specs pass. Deviation: the four essentials are the current room's first
  four tiles, not yet configurable.
- [x] T012 — Browser coverage for notification inbox persistence (`NotificationBell.spec.ts`) plus
  affected layout/panel/home/device suites; IDs: AC-05–08/23/24. Digest dismissal and panel night
  toggle still lack dedicated specs.
- [ ] T013 — Editors/deep flows: automation editor/creator, integration setup/provider, Assistant,
  Login; IDs: FR-09 (scope deferred in spec §Problem). Not started.
- [ ] T014 — Server-side cleanup/purge UI; IDs: AC-21 (excluded by NH-06). Blocked: no endpoints
  and no named approval for destructive history/log deletion.

## Verification

- [x] T015 — `npm test -- --runInBand`: 168 suites / 1116 tests passing.
- [x] T016 — `npm run build`: passing (Vue type-check + Vite).
- [x] T017 — `npm run lint`: 0 errors (113 pre-existing warnings); no new warnings in changed files.
- [x] T018 — Full `--project=desktop`: 207 passed, 9 skipped.
- [x] T019 — Full `--project=mobile`: 216 passed.
- [x] T020 — Plan actuals and tasks updated; PR notes carry the remaining/deferred items.

## Discovered work

- Files list size/age: extend `/api/listlogs` only if the UX needs it; current UI avoids inventing
  data.
- Panel essentials configuration: needs a persisted per-panel preference and selection UI.
- Destructive cleanup: needs its own NH-06 spec, backend endpoints and approval.
