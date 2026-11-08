# Tasks: Consolidated, consistent settings

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: complete (unreviewed)

## Preparation

- [x] T001 — Resolve grouping, History/System merge and rebuild depth with the owner;
  recorded in `spec.md` (US-01…US-04) and `plan.md` alternatives. IDs: NH-01.
- [x] T002 — Read parent/frontend constitutions and `.agents/rules/vue-engineer.md`;
  classified the change as frontend-only. IDs: NH-01, NH-04.
- [x] T003 — Fix the browser-test fixture (`docs/hub_state.json`: `sleepTimeout`,
  `logger.level`) and the mock's wrong `saveHistoryConfig` target. IDs: AC-06, AC-11;
  done: `npm test` 768 passed, `e2e/StorageSettings.spec.ts` reload case passes.

## Coverage and implementation

- [x] T004 — `bytes` quantity + `formatBytes` in `src/utils/units.ts`, cases in
  `src/__tests__/utils/units.test.ts`. IDs: NFR-04, AC-09.
- [x] T005 — `src/domain/settings/policy.ts` (labels, bounds, payload builders that never
  mutate input) + `policy.test.ts`. IDs: FR-04, AC-07, AC-08, AC-11, AC-14.
- [x] T006 — `src/components/hub/settings/sections.ts` + `sections.test.ts` covering ids,
  grouping, aliases and defaults. IDs: FR-01, AC-02…AC-04.
- [x] T007 — Rebuild `Settings.vue` on the registry; delete `src/mixins/useTabComponents.ts`.
  IDs: FR-01, AC-01…AC-05, AC-17.
- [x] T008 — Add `StorageSettings.vue`, delete `HistorySettings.vue` and
  `SystemSettings.vue`. IDs: FR-02, AC-06…AC-10.
- [x] T009 — Rebuild `LoggerSettings.vue` and `AssistantSettings.vue` on shared controls
  with policy builders. IDs: FR-03, FR-04, AC-11, AC-14.
- [x] T010 — `useMcpControls.ts` + `useMcpControls.test.ts`; `MCPSettings.vue` with
  `UiChip` status and `.nh-kv` rows. IDs: FR-05, AC-12, AC-13.
- [x] T011 — Section headings for Device Defaults and the device settings tab; removed the
  nested card in `DeviceConfigurationSettings.vue` and `DevicePage.vue`. IDs: FR-03, AC-16.
- [x] T012 — Tokenised borders/heights in the reporting components (`--nh-touch-h`,
  `--nh-border-w`, `--nh-border-strong`), preserving the spec 018 selectors. IDs: AC-16,
  AC-18.

## Handoff

- [x] T013 — New browser specs: `e2e/Settings.spec.ts`, `e2e/StorageSettings.spec.ts`,
  `e2e/LoggerSettings.spec.ts`, `e2e/MCPSettings.spec.ts`,
  `e2e/AssistantSettings.spec.ts`; `e2e/layout.spec.ts` now sweeps all six sections.
  IDs: AC-01…AC-18, SC-01.
- [x] T014 — Recorded in `plan.md`: `npm test` 768 passed; lint 0 errors / 26 pre-existing
  warnings; `npm run build` pass; `npm run type-check:e2e` pass; 50 Playwright tests
  passed across desktop and mobile. Backend `go test ./...` **unrun** (no backend change).
  Gate: NH-05.
- [x] T015 — Self-review: removed unused `assistantHost`, `SETTINGS_KINDS` and the unused
  section `description`; fixed the stacked-mode flex-basis and the duplicated assistant
  help line; regenerated `components.d.ts`. Gates: NH-04, NH-05.

## Follow-up review corrections

- [x] T016 — Replace the Sensor adjustments `fieldset`/`legend` (and the nested history
  group) with in-card headings, keeping `role="group"` and the accessible names the
  spec 018 browser suite depends on. IDs: AC-16a; done: heading renders inside the card,
  group names and pickers verified, `legend` count 0.
- [x] T017 — Give settings sections and cards their own icon meanings distinct from the
  integration and navigation glyphs; add `settings*` entries to the icon registry and
  use them in the rail and card headers. IDs: AC-16b; done: `sections.test.ts` enforces
  unique, non-reused icons; visual check of all five hub sections.
- [x] T018 — Make the mock/browser-test ports overridable so runs do not collide with a
  local dev stack. IDs: NH-05; done: visual and contract checks ran on spare ports
  against a running dev hub without touching it.
- [x] T019 — Re-run checks: `npm test` **769 passed / 140 suites**; `npm run lint`
  0 errors (26 pre-existing warnings); `npm run build` pass.

## Review pass

- [x] T020 — Fix retention data loss: `canPersistRetention` refuses a policy built from
  the pre-snapshot placeholder, the card stays locked until the snapshot loads and is
  blank rather than showing `0`. IDs: AC-07, NH-02, NH-06; done: unit regression test +
  browser check that the saved frame keeps `sleepTimeout: {12, hours}`.
- [x] T021 — Clear a stale MCP timeout error when a later status frame arrives. IDs:
  AC-13; done: regression test fails against the old watcher and passes now.
- [x] T022 — Delete the components orphaned by the rewrite (`input/ActionButton.vue`,
  `input/InputBox.vue`, `controls/TimeInterval.vue`), the unused `settingsPolicy` icon and
  its import, and the unnecessary `displayRetentionInterval` export. IDs: NH-04.
- [x] T023 — Move `sections.test.ts` under `__tests__/components/hub/settings/` to mirror
  its source path. IDs: NH-05.
- [x] T024 — Abort a superseded/unmounted storage-usage read so a late reply cannot
  overwrite newer data or write after unmount. IDs: NH-06.
- [x] T025 — Final checks: `npm test` **771 passed / 140 suites**; `npm run lint`
  0 errors / 23 warnings; `npm run build` pass; `npm run type-check:e2e` pass.
