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
- [x] T008 — Snapshot banner + logs IA: Settings → Logs & diagnostics with archive list/download
  (name, size, modified), header Logs shortcut, Console removed from nav, local filters/follow/export;
  IDs: FR-04/AC-09–11; done: service tests + Go handler test; size/age come from the additive
  `/api/logfiles` route.
- [x] T009 — Settings rail summaries and section title alignment; IDs: FR-08/AC-20; done:
  `Settings.spec.ts` 22 passing; all six sections still present.
- [x] T010 — Automations export all/per row with versioned JSON; IDs: FR-08/AC-22 (export only);
  done: `automation-export` test passing; duplicate/copy-ID/run-history remain open.
- [x] T011 — Panel mode: schedule night (22:30–06:30) + persistent manual override, configurable
  four night essentials, paused/last-known controls while offline, status page retained; IDs:
  FR-07/AC-17–19; done: parser unit test + desktop/mobile panel specs including persistence.
- [x] T012 — Browser coverage for notification inbox persistence, digest dismissal/reappearance,
  panel night persistence and the Overview action; IDs: AC-05–08/17/23/24.
- [x] T021 — Overview “Review in order” now scrolls and focuses the ranked list (nested scroll
  container made the hash anchor inert); browser regression added.
- [x] T022 — Browser-only Data & cleanup section: notifications, activity, log buffer and appearance,
  with confirmations; no server deletion and no dead controls.
- [x] T023 — Overview activity is one list (removed the preview/full duplicate); health-line counts
  and “Show more” deep-link the device list with `status`/`q`; device filters are URL-addressable.
- [x] T013 — Editors/deep flows consistency pass; IDs: FR-09. Assistant rebuilt on the shell
  header/tokens, history and composer; editor/creator, provider setup and Login audited and already
  use the shared patterns (no churn).
- [x] T024 — Notification bell badge: smaller, no heavy ring, touch-safe.
- [x] T025 — Tile footer: no clock glyph, device left / age right, mock options in
  `docs/design/mocks/card-footer.html`.
- [x] T014 — Server-side cleanup: explicit-cutoff history delete (`/api/history/purge`) and one-file
  archive delete (`/api/logfile` action delete), authenticated, rate-limited, audited and confirmed;
  IDs: AC-21. History UI states scope and irreversibility; log delete names the file.

## Verification

- [x] T015 — `npm test -- --runInBand`: 168 suites / 1116 tests passing.
- [x] T016 — `npm run build`: passing (Vue type-check + Vite).
- [x] T017 — `npm run lint`: 0 errors (113 pre-existing warnings); no new warnings in changed files.
- [x] T018 — Full `--project=desktop`: 207 passed, 9 skipped.
- [x] T019 — Affected mobile suites (panel/Settings/layout): 38 passed; full matrix in CI.
- [x] T020 — Plan actuals and tasks updated; PR notes carry the remaining/deferred items.

## Discovered work

- Files list size/age: extend `/api/listlogs` only if the UX needs it; current UI avoids inventing
  data.
- Panel essentials configuration: needs a persisted per-panel preference and selection UI.
- Destructive cleanup: needs its own NH-06 spec, backend endpoints and approval.
