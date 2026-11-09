# Plan: Consolidated, consistent settings

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: agent draft | Updated: 2026-11-08
Constitutions: parent 1.0.0 (proposed); frontend 1.0.0 (proposed)

## Context and design

Observed facts (see spec for evidence): the settings host wraps every section in one
`UiCard` while section components draw their own bordered panels; four sections are
legacy PrimeFlex/PrimeVue views with raw config keys as labels; History and System
duplicate retention; `MCPSettings` fakes completion with `setTimeout(1000)`; two settings
views mutate Vuex getter results in place; byte formatting lives in a component;
`.nh-kv` is unused; the browser-test fixture misspells `sleepTimeout` and omits
`logger.level`.

Ownership after the change:

| Concern | Owner |
| --- | --- |
| Section registry, grouping, URL resolution | `src/components/hub/settings/sections.ts` |
| Rail rendering + section outlet | `src/components/hub/settings/Settings.vue` |
| Retention policy rules/validation/view mapping | `src/domain/settings/policy.ts` |
| MCP pending lifecycle | `src/composables/useMcpControls.ts` |
| Byte formatting | `src/utils/units.ts` (`bytes` quantity) |
| Saved settings | Vuex `hub` module via existing actions |

Design:

1. **Sections registry** replaces `src/mixins/useTabComponents.ts` (its only consumer).
   `SETTINGS_SECTIONS` declares `{ id, title, group, load }` with literal ids:
   `appearance`, `device-defaults`, `storage`, `logger`, `mcp`, `assistant`. Groups are
   `browser` ("This browser") and `hub` ("Hub"). `resolveSettingsSection(query)` maps a
   raw `?section=` value to a section, aliasing the retired `history` and `system` ids to
   `storage`, defaulting unknown/absent values to the first section. Because ids are
   literals, retitling never breaks a link.
2. **Host** renders `<nav aria-label="Settings sections">` with one caption + `<ul>` per
   group, then the selected section component directly — no outer `UiCard`, no
   `"Saved on the hub"` note (the group caption carries that scope). Sections are still
   async components, so first paint is unchanged in kind.
3. **Section shape** (every hub section): `<section aria-labelledby>` with one `<h2>`, a
   short `.nh-help` intro, then `UiCard :heading-level="3"` blocks per topic. Cards hold
   their own content; no component nests a panel inside a card.
4. **Storage & retention** merges History + System. A `Retention` card holds two
   labelled interval controls (cleanup interval, keep history for) built from
   `domain/settings/policy.ts`; a `Usage` card renders `GET /system/storage` rows with
   `formatQuantity(bytes, 'bytes')`, a total row, and a disk line. Retry/refresh is a
   `UiButton`; loading/error/empty are explicit states.
5. **Logger** uses `UiToggle` + `UiSelect` with `UiField` labels and help. Level options
   stay the existing six values. Saving builds the complete `LoggerSettingsType` from
   current values and dispatches — no getter mutation.
6. **MCP server** shows a `UiChip` status (`Disabled` muted, `Running` ok, `Stopped`
   warn) plus `.nh-kv` identity rows (name, version, clients) when enabled.
   `useMcpControls()` dispatches start/stop/restart, marks that action pending, asks for
   `loadMCPStatus`, clears pending when `hub/mcpStatus` changes, and clears after a
   bounded `MCP_STATUS_TIMEOUT_MS` with an error notice; it removes its watcher and timer
   on scope dispose.
7. **Assistant** is a single `UiField` with label, help and inline URL validation; save
   builds a new `AssistantSettingsType`.
8. **Device defaults** keeps `DeviceConfigurationSettings` but owns its section heading
   (`Device defaults` / `Device settings`) and stops being wrapped in a card, so the
   reporting panels no longer sit inside a card. `DevicePage.vue` drops its duplicate
   `UiCard` title for the same tab. `SensorReportingSettings` keeps its DOM contract
   (headings, fieldsets, `details>summary` text, control labels) and only has its
   hardcoded borders/heights replaced with tokens.
9. **Units** gains a `bytes` quantity in `formatQuantity` (B/KB/MB/GB, two decimals above
   MB); `SystemSettings`'s private formatter is deleted with the component.
10. **Fixture**: `docs/hub_state.json` gains `sleepTimeout` (correct spelling) and
    `logger.level: "info"`. Frontend contract/fixture only; no backend change.

Files:

| Path | Change |
| --- | --- |
| `src/components/hub/settings/sections.ts` | new registry + resolution |
| `src/components/hub/settings/Settings.vue` | grouped rail, no outer card, stable ids |
| `src/components/hub/settings/StorageSettings.vue` | new (retention + usage) |
| `src/components/hub/settings/HistorySettings.vue` | delete |
| `src/components/hub/settings/SystemSettings.vue` | delete |
| `src/components/hub/settings/LoggerSettings.vue` | rebuild |
| `src/components/hub/settings/MCPSettings.vue` | rebuild |
| `src/components/hub/settings/AssistantSettings.vue` | rebuild |
| `src/components/hub/settings/DeviceDefaultSettings.vue` | delete (registry loads the component) |
| `src/domain/settings/policy.ts` | new: labels, validation, payload builders |
| `src/composables/useMcpControls.ts` | new: pending/refresh lifecycle |
| `src/utils/units.ts` | `bytes` quantity |
| `src/components/settings/DeviceConfigurationSettings.vue` | own heading, tokenised borders |
| `src/components/settings/SensorReportingSettings.vue` | tokenised borders/heights only |
| `src/components/settings/ReportingRateControl.vue`, `ReportingIntervalInput.vue`, `ReportingSensorPicker.vue` | tokenised borders/heights |
| `src/components/device/DevicePage.vue` | remove duplicate card title for the settings tab |
| `src/mixins/useTabComponents.ts` | delete |
| `docs/hub_state.json` | fixture spelling + logger level |
| `src/__tests__/domain/settings/policy.test.ts` | new |
| `src/__tests__/domain/settings/sections.test.ts` | new |
| `src/__tests__/composables/useMcpControls.test.ts` | new |
| `src/__tests__/utils/units.test.ts` | bytes cases |
| `e2e/Settings.spec.ts` | new: rail, groups, ids, aliases |
| `e2e/StorageSettings.spec.ts` | new |
| `e2e/LoggerSettings.spec.ts` | new |
| `e2e/MCPSettings.spec.ts` | new |
| `e2e/AssistantSettings.spec.ts` | new |
| `e2e/layout.spec.ts` | add `/settings?section=...` sweep |

Alternatives rejected:

- A shared `UiTabs`/`UiSection` primitive: the settings rail is the only vertical rail and
  the section heading is a plain `h2`; adding two primitives for one screen fails NFR-02.
- Sub-tabs per group: extra navigation level for six destinations.
- Keeping History and System separate: the same retention numbers would stay in two
  places with different labels.
- A new `/api/settings/storage` endpoint: the merged section can read the existing
  endpoint and the existing store getter.
- Fixing `sleepTmeout` on the backend model: the backend JSON tag is already correct; the
  defect is only in the browser-test fixture.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 059 with AC-01…AC-18 and resolved questions | PASS | PASS |
| NH-02 | No device command added; retention/logger/MCP frames unchanged; MCP keeps requested vs confirmed status separate | PASS | PASS |
| NH-03 | No new entry point, no secret or household data in the bundle; assistant URL stays non-executable text | PASS | PASS |
| NH-04 | Reuses store actions, `contracts/api`, `domain/` + `composables/` boundaries; no new dependency | PASS | PASS |
| NH-05 | Traceability table plus unit/browser checks; failures/unrun recorded | PASS | PASS |
| NH-06 | No persistence change beyond fixture; MCP timer/watch disposed on unmount; usage rounds are read-only | PASS | PASS |
| FE-01 | Sections presentational; policy/view mapping in `domain/`, lifecycle in a composable | PASS | PASS |
| FE-02 | `HistorySettingsType`, `LoggerSettingsType`, `MCPStatusType`, `AssistantSettingsType`, storage payload consumed as-is | PASS | PASS |
| FE-03 | MCP requested→confirmed pending state; loading/error/offline states per section | PASS | PASS |
| FE-04 | Labels/help on every control, `aria-current` rail, `role="alert"` errors, 44 px targets, 1280/390 checks | PASS | PASS |
| FE-05 | No `VITE_*` secret, no new browser persistence, assistant text not rendered as HTML | PASS | PASS |

## Contracts, data, rollout

- WS frames sent by the UI are unchanged: `saveHistoryConfig` (`{ sleepTimeout, expireAt }`),
  `saveLoggerConfig` (`{ enableRemoteLogger, level }`), `saveAssistantConfig` (`{ url }`),
  `startMCP`/`stopMCP`/`restartMCP`/`loadMCPStatus` (empty payloads). Covered by the
  existing `commands.test.ts` wire-contract test, which must stay green.
- `GET /system/storage` is consumed unchanged (same fields, same 15 s backend cache).
- No persistence, configuration or deployment change. The only data edit is the frontend
  test fixture described above; it is not shipped.
- Rollback: the section registry and URL aliases are additive; reverting the branch
  restores the previous host without touching hub state.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01…AC-05 | `sections.ts`, `Settings.vue` | `sections.test.ts` (5 cases); `e2e/Settings.spec.ts` (5 tests, desktop + mobile) | pass |
| FR-02, AC-06…AC-10 | `StorageSettings.vue`, `policy.ts`, `units.ts` | `policy.test.ts` (8); `units.test.ts` bytes case; `e2e/StorageSettings.spec.ts` (5 tests × 2 projects) | pass |
| FR-03, AC-11, AC-14 | `LoggerSettings.vue`, `AssistantSettings.vue`, `policy.ts` | `policy.test.ts`; `e2e/LoggerSettings.spec.ts` (2), `e2e/AssistantSettings.spec.ts` (2) × 2 projects | pass |
| FR-03, FR-05, AC-12, AC-13 | `MCPSettings.vue`, `useMcpControls.ts` | `useMcpControls.test.ts` (6); `e2e/MCPSettings.spec.ts` (2) × 2 projects | pass |
| FR-04 | `LoggerSettings.vue`, `AssistantSettings.vue`, `StorageSettings.vue`, `policy.ts` | `policy.test.ts` purity cases assert inputs are unchanged; builder returns a new object | pass |
| FR-06, AC-15, AC-18 | all sections | offline case in each section spec; 44 px targets via `--nh-touch-h` | pass |
| AC-16, AC-17 | `Settings.vue`, `DeviceConfigurationSettings.vue`, `DevicePage.vue` | nested-card assertion in `e2e/Settings.spec.ts`; `layout.spec.ts` sweep over all six sections | pass |
| NFR-01 | store/contracts untouched | `npm test`: 768 tests, 140 suites | pass |
| NFR-02, NFR-03 | no new dependency; async sections preserved | `npm run build` succeeded; dependency tree unchanged | pass |
| NFR-04 | `units.ts` | `bytes` cases in `units.test.ts` | pass |
| SC-02 | all | see recorded checks below | pass |

## Verification (recorded)

From `frontend/`:

- `npm test -- --runInBand` → **768 passed, 140 suites**.
- `npm run lint` → **0 errors**, 26 warnings (all pre-existing, none in changed files).
- `npm run build` → pass (`vue-tsc` + `vite build`); `components.d.ts` regenerated by Vite.
- `npm run type-check:e2e` → pass.
- `npx playwright test --project=desktop --project=mobile e2e/Settings.spec.ts e2e/StorageSettings.spec.ts e2e/LoggerSettings.spec.ts e2e/MCPSettings.spec.ts e2e/AssistantSettings.spec.ts e2e/SensorReportingSettings.spec.ts e2e/layout.spec.ts`
  → **50 passed** (desktop + mobile), including the unchanged spec 018 reporting suite.
- Visual check: all six sections screenshotted at 1280 px and `storage` at 390 px;
  fixed a stacked-mode `flex-basis` defect that produced a ~256 px gap, and the
  assistant field's duplicated help line.
- Follow-up (AC-16a): the Sensor adjustments `<fieldset>`/`<legend>` was replaced with a
  `<section role="group">` + `<h3>`, so the heading no longer straddles the card border.
  The same change was applied to the nested history group (`<div role="group">` +
  `<h4>`). The `group` role and its accessible name are preserved because
  `e2e/SensorReportingSettings.spec.ts` locates both blocks that way; verified with a
  temporary browser check on dedicated ports (`role="group"` names, the pickers, and
  `legend` count 0).
- Follow-up (AC-16b): settings icons no longer reuse integration/navigation glyphs.
  Added `settingsDevices`, `settingsStorage`, `settingsUsage`, `settingsLogger`,
  `settingsServer`, `settingsAssistant`, `settingsRetention` and `settingsPolicy` to the
  icon registry and used them for the section cards and rail entries; `sections.test.ts`
  asserts the icons are unique and excludes the integration/nav glyphs.

Defects found and fixed while verifying (each at its owning layer):

- `hub/init` called `clear()`, so an MCP status pushed by the socket before the HTTP
  snapshot was discarded, leaving every session without status. `init` now commits a new
  `resetSnapshot`; `hub/clear` still clears everything on logout. Regression test in
  `src/__tests__/store/modules/appconfig/index.test.ts`.
- The dev/E2E mock wrote `saveHistoryConfig` to `appConfig.history` instead of
  `appConfig.hub.history`, so saved retention was lost on reload. Fixture-only fix in
  `tools/server/ws.mjs`.
- `docs/hub_state.json` misspelled `sleepTimeout` and omitted `logger.level`.
- The browser-test ports were hardcoded, so any run failed with `EADDRINUSE` when a
  local dev stack was already using 4100/4110. `tools/server/config.mjs` now reads
  `NODEHERDER_MOCK_PORT`/`NODEHERDER_MOCK_CORS_ORIGIN` (defaults unchanged), and the
  mock OAuth redirect uses the configured origin. This let the follow-up visual and
  `role="group"` checks run on spare ports without touching a running dev hub.

Backend (`go test ./...`) was **not run**: no backend file changed. This is recorded as
unrun rather than implied.

## Review findings (fixed)

- **Data loss in Storage & retention.** `hub/history` reports the `createAppconfig()`
  placeholder (`0 days` for both intervals) until the HTTP snapshot lands. Editing one
  field then persisted the untouched field as zero; the hub reads a zero retention as
  "keep nothing" and a zero cleanup interval as "run continuously". Added
  `canPersistRetention` in `domain/settings/policy.ts`, which refuses any policy with an
  invalid interval, and gated the card on the loaded snapshot. While unloaded the inputs
  are disabled and blank rather than showing a fabricated `0`. Regression test:
  `a placeholder policy cannot be persisted…`; browser check confirmed the saved frame
  carries `sleepTimeout: {12, hours}`.
- **Stale MCP error.** A timeout message survived a later successful status frame because
  the watcher only cleared it while an action was pending. The watcher now clears both.
  Regression test: `a late status frame clears the timeout message…` (verified to fail
  against the previous logic).
- **Dead code from the rewrite.** `MCPSettings` was the only consumer of
  `input/ActionButton.vue`, which in turn orphaned `input/InputBox.vue` and
  `controls/TimeInterval.vue`; all three are deleted. Also removed the unused
  `settingsPolicy` icon entry (and its `mdiDatabaseSettings` import) and stopped exporting
  `displayRetentionInterval`, which had no external caller.
- **Test ownership.** `sections.test.ts` tested a component module from
  `__tests__/domain/settings/`; moved to `__tests__/components/hub/settings/` to mirror the
  source path, per the repository's test-file ownership rule.
- **Unmount/races in the usage read.** The storage fetch now owns an `AbortController`:
  a superseding Refresh aborts the previous request and an aborted read reports no error,
  so a slow reply cannot overwrite newer data or write after unmount.

Lint warnings fell from 26 to 23 with the dead components removed (0 errors).

Review-pass verification (latest): `npm test` **771 passed / 140 suites**; `npm run lint`
0 errors; `npm run build` and `npm run type-check:e2e` pass. Browser suites were re-run
against an isolated mock on spare ports because a local dev stack owned 4100/4110: all
settings and reporting specs pass, with `E2E_HUB_API` and the `routeWebSocket` URL matched
to the spare port. Those two URLs are hardcoded per spec file, so a mismatched port makes
the disconnect cases silently miss the socket; the failures observed during the review
were that mismatch only and passed once aligned.

CI on PR #103 (`feat/settings-consolidation`, run 42523468770) confirms that on the
canonical ports: **17 of 17 checks pass**, including `Frontend Tests, Lint and Build`
(1m12s), the full desktop and mobile E2E matrix (shards 1–4 each), `Backend Tests and
Build`, `Backend Race Detection`, both container builds and the deployment/tooling checks.
`Tag Build` is skipped because this is not a tag build.

Selected commands (from `frontend/`): `npm test -- --runInBand`; `npm run lint`;
`npm run build`; `npm run type-check:e2e`; `npx playwright test --project=desktop e2e/Settings.spec.ts e2e/StorageSettings.spec.ts e2e/LoggerSettings.spec.ts e2e/MCPSettings.spec.ts e2e/AssistantSettings.spec.ts e2e/SensorReportingSettings.spec.ts e2e/layout.spec.ts`
plus `--project=mobile` for the layout/touch/viewport behaviour. Backend is untouched, so
`go test ./...` is not required by this design; if the fixture change surprises the mock
server, the reason is recorded instead of implied.

## Debt, risks, exceptions

- `SensorReportingSettings.vue` remains structurally dense (category rows, fieldsets,
  details). This change tokenises it and removes the nested card but does not redesign
  it; a deeper information-architecture change would renegotiate the AC set in spec 018.
- The MCP section cannot enable/disable the server: `hub/mcp.enabled` is a persisted
  configuration value with no save frame in this UI. Start/Stop keep their current
  meaning; the read-only status chip is reported honestly as Disabled/Running/Stopped.
- Cleanup interval has a one-minute UI floor so a browser test cannot request a
  near-zero interval; the backend goroutine would otherwise spin. Recorded as validation,
  not as a new backend rule.
- `docs/hub_state.json` is shared by the mock server and `commands.test.ts`; the fixture
  change is verified by running the full Jest suite.
- Prior nested-card/h2 debt in `DevicePage.vue` is corrected because the same component
  is affected; no other device-page redesign is included.

## Completion

Acceptance evidence and actual command results are recorded in `tasks.md` and this plan's
traceability table as each check runs. No review approval is claimed.
