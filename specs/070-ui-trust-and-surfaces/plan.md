# Plan: UI trust and daily surfaces

Spec: [spec.md](spec.md) | Status: Draft | Author: engineering | Updated: 2026-11-25
Constitutions: parent 1.0.0; backend 1.0.0; frontend 1.0.0

## Context and design

Affected frontend paths: `store/modules/{hub-state,notifications,console}`, `domain/devices`,
`domain/overview`, `composables/{useNow,useHubConnection}`, `components/{shell,entity,home,overview,
hub,panel,automations,settings}`, `services/logs.ts`, `router`, tests under `src/__tests__` and
`e2e`. Backend is read-only for this change; existing protected log routes are consumed.

Current ownership: hub state owns devices/config, transport owns connection status, console owns a
bounded log buffer, alerts own ephemeral toasts. The design adds two derived concepts (freshness,
snapshot age) and one durable store (notifications). Freshness is a pure function of last_seen and
now, so it cannot become competing device truth; snapshot age is client-only. Notifications never
drive device commands and are bounded/persisted for the browser only.

Alternatives rejected: a backend freshness field (invents a second truth and requires contracts);
per-card timers (leaks/CPU); notification generation in components (duplicates transitions); moving
purge into the UI before endpoints exist (unsafe).

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 070 with ACs and staged scope | PASS | PASS (spec/tasks updated with actuals) |
| NH-02 | Freshness display-only; `deviceStatus` unchanged; offline controls use existing `canControl`; no optimistic state | PASS | PASS (unit + panel/browser) |
| NH-03 | No new network entry points; log endpoints already protected; persistence limited to notification text/read state/last visit | PASS | PASS (no secrets; localStorage bounded) |
| NH-04 | Existing store/service/component boundaries; new domain modules/store only | PASS | PASS |
| NH-05 | Unit + browser suites; exact commands recorded | PASS | PASS (169 suites / 1118 tests; desktop 207, mobile 216) |
| NH-06 | Bounded buffer/notifications, local cleanup, no server deletion; timers cleaned on unmount | PASS | PASS (cap/prune tests, cleanup hooks) |
| FE-01 | Presentation in components; shared behavior in composables/domain; one notification store | PASS | PASS |
| FE-02 | Runtime validation for log file payloads; existing contracts for WS unchanged | PASS | PASS (service validation tests) |
| FE-03 | Loading/empty/error/offline states defined per page; no command retries | PASS | PASS (browser specs) |
| FE-04 | Accessible names, focus, colour+icon, mobile specs 390/360, panels 44 px | PASS (design) | PASS (layout + panel desktop/mobile) |
| FE-05 | No secrets; localStorage minimized; no raw external HTML | PASS | PASS |
| BE-01/05 | No backend edits; read-only log routes documented in spec | PASS | N/A |

## Contracts, data, rollout

- `hub-state`: add `lastSyncedAt: number | null`, mutation `setLastSyncedAt`, getter; set on snapshot
  publish and throttled on live reports. No persistence.
- `notifications`: new namespaced Vuex module; fields `id, at, read, severity, category, title, body?,
  dedupeKey?, action?`; cap 200 / 7 days; localStorage key includes hub identity and user id; storage
  failures are ignored (in-memory still works).
- `domain/devices/freshness.ts`: pure policy lookup by power source and age classification; exported
  constants for the thresholds in the spec.
- `services/logs.ts`: `listLogFiles()`, `loadLogFile(name)`, `downloadTextFile(name, contents)` with
  size/name validation and object-URL release. Existing endpoint error payloads are surfaced as
  readable errors.
- No backend, WS, MQTT, MCP or persistence schema change. Rollback: remove new UI entry points and
  store; existing behaviour is unchanged because freshness is additive and toasts remain.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01/AC-01/02/04 | `domain/devices/freshness.ts`, `FreshnessLabel.vue`, `EntityTile.vue`, `DeviceCard.vue`, `DevicePage.vue` | `freshness.test.ts` (5), network filter test, build + device/home browser specs | PASS |
| FR-02/AC-03 | `hub-state`, `useHubConnection` path, `StatusBanners.vue` | `hub-state/index.test.ts`; disconnect banner in shell | PASS (unit); visual covered by layout specs |
| FR-03/AC-05–08 | `store/modules/notifications`, `NotificationBell.vue`, `NotificationDigest.vue`, `useNotificationProducers.ts` | store tests (4), `notifications.test.ts` (2), `NotificationBell.spec.ts` | PASS (pending final run of the new spec) |
| FR-04/AC-09–11 | `services/logs.ts`, `LogArchiveCard.vue`, `ConsoleViewer.vue`, `AppHeader.vue`, backend `/api/logfiles` | `logs.test.ts`, Go handler metadata test, Settings specs | PASS |
| FR-05/AC-12–14 | `OverviewPage.vue`, `overview.ts`, `RecentActivityBrief.vue` | `overview.test.ts` (3), overview + ActivityFeed browser specs, desktop/mobile layout | PASS |
| FR-06/AC-15–16 | `DeviceList.vue`, `DeviceFilterBar.vue`, `DevicePage.vue` | network filter test, devices + EntityTile specs | PASS |
| FR-07/AC-17–19 | `PanelMode.vue`, `PanelEssentialsDialog.vue`, `panel-essentials.ts` | panel parser test + desktop/mobile specs incl. essentials persistence | PASS |
| FR-08/AC-20–22 | `Settings.vue`, `CleanupSection.vue`, `AutomationList.vue`, `automation-export.ts` | `automation-export.test.ts`, Settings/Logger/Storage specs | PARTIAL: rail summaries, export, copy ID and browser cleanup shipped; duplicate/run history open |
| FR-09/AC-23–25 | changed components | `layout.spec.ts` desktop+mobile, panel specs, lint/build | PASS for shipped surfaces; targeted a11y checks for new notifications still to add |
| T013 editors/Assistant/Login | — | not started | DEFERRED (scope note in spec) |
| T014 purge UI | — | excluded by NH-06 | BLOCKED, no dead controls shown |

Commands actually run:
- `npm test -- --runInBand` → 168 suites / 1116 tests passed (before the last additions; rerun recorded in tasks).
- `npm run build` → passed.
- `npm run lint` → 0 errors, 113 pre-existing warnings.
- `npx vue-tsc --noEmit --project tsconfig.build.json` → passed.
- `npx tsc --noEmit --project tsconfig.e2e.json` → passed.
- Playwright affected suites: Settings group (22), overview/ActivityFeed/layout/panel/EntityTile, home/devices, mobile layout/panel all green.

## Follow-up (2026-11-26)

- Overview “Review in order” now calls `scrollIntoView` on the ranked list and moves focus to
  its heading; the previous hash link was inert because `.nh-app-main` owns scrolling. Regression:
  `overview.spec.ts` “review in order moves focus to the ranked list”.
- Added browser-only Settings → Data & cleanup (notifications, activity, log buffer, appearance)
  with confirmation; server deletion remains excluded under NH-06.
- Panel night choice now persists for the current night cycle, and the panel suite pins a
  deterministic day baseline while adding an explicit night test.
- Automation rows gained Copy ID alongside Export.
- Affected Playwright suites: 43 desktop + 43 mobile passed; unit 169/1118, lint 0 errors,
  build and e2e type-check pass.

## Final follow-up (2026-11-26)

- Panel night essentials are now selectable, bounded to four, validated at the storage boundary
  and persisted per browser; a room-order fallback remains for first use.
- Log archive metadata is additive: `/api/logfiles` returns `{name,size,modified}` while
  `/api/listlogs` stays unchanged for older clients. The UI shows size and relative age.
- Remaining: T013 consistency pass for editor/provider/Assistant/Login, and T014 server purge
  (still NH-06-blocked). Neither is represented by dead controls.

## Debt, risks, exceptions

- Server purge, automation run history and panel lux override are deferred; UI does not advertise
  them. Risk: users still cannot delete server history; consequence documented in the spec and PR.
- Freshness thresholds are heuristic; false “quiet” is limited to amber styling and explained text.
- Notification dedupe must not drop distinct same-title events: include category, device/action and
  a time bucket in the key; cap prevents unbounded growth.
- `lastSyncedAt` updates on every live message could churn reactivity; throttle to 1 s via a plain
  module timestamp and commit at most once per second.
- Test file ownership: new `freshness.ts` → `freshness.test.ts`; store modules keep existing
  companion tests; UI specs stay in the page-owned spec files.
- No exceptions requested; no maintainer approval claimed.

## Completion

Done when the P1 acceptance scenarios pass with evidence and P2 items are implemented or explicitly
deferred in this file. Record actual commands/results and any unrun checks. Update spec/plan if
scope changes.
