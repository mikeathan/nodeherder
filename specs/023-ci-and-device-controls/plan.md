# Plan: Faster CI and cleaner device controls

Spec: [spec.md](spec.md) | Status: Implemented locally; CI rollout pending | Author: Codex | Updated: 2026-10-20
Constitutions: parent/frontend 1.0.0 (proposed; no ratification or approval claimed)

## Context and design

CI: `.github/workflows/ci.yml` partitions frontend-e2e into desktop/mobile × shards
1/2, with fail-fast disabled. Keep one worker and fullyParallel false; isolated
runners own separate mutable hubs. Use list/HTML/JSON reporters and unique timing/
failure artifact names. Other jobs, permissions, triggers, deadlines, cancellation,
aggregate and tagging remain intact. `docs/ci.md` records costs/reproduction/limits.
Increasing workers against one hub or reducing assertions would weaken evidence.

UI: DeviceDashboard uses CSS columns (19rem preferred width), break avoidance and
shared spacing. DeviceCard retains its original compact flex footer; offline panels
size to content. ExposeRow/ExposeControl provide thin horizontal rails, round handles,
44px inputs and responsive value/preset alignment. UiRange's large vertical entity
design stays unchanged. Card headers expose clear native links; EntityViewDialog's
friendly-name RouterLink closes only for ordinary activation and keeps modal cleanup.

Remove all visible command progress from ExposeControl, EntityAdjustment,
EntityPowerButton, EntityTile, PanelButton and UiToggle. Retain aria-busy, normal
opacity while awaiting confirmation, disabled guards and confirmed toggle text.
Shared edit/command composables still own drafts, confirmation, duplicate prevention
and timeout errors. Genuine offline/unavailable styling and data/save loading remain.
ESLint ignores generated Playwright reports/results as it already ignores dist.

Planning: `docs/sdd/README.md#branch-scope` owns the current-branch reuse rule;
AGENTS.md and the SDD skill route to it. Consolidate only this work's six unmerged
drafts into this directory, preserving final acceptance and evidence. Future branch
iterations update these files; completed earlier work remains history.

## Constitution Check

Original design/spec/tasks preceded implementation; this consolidation records final
scope at the user's request. Local checks passed; hosted rollout remains pending.

| Rule | Evidence | Design | Final local review |
| --- | --- | --- | --- |
| NH-01 | Final scenarios, implementation/tasks and consolidated evidence | PASS | PASS; no new application behavior in consolidation |
| NH-02 | Shared command/edit logic and native payloads unchanged | PASS | PASS: gestures, confirmation, timeout, duplicate and offline cases |
| NH-03 | Existing auth/routes, isolated mocks, read-only CI jobs | PASS | PASS: no household device commands or exposure changes |
| NH-04 | Existing components/router/reporters; aggregate contract retained | PASS | PASS: no new dependencies or ownership changes |
| NH-05 | Companion regressions, partition execution, frontend/workflow checks | PASS | PASS locally; hosted target pending |
| NH-06 | Existing deadlines/cleanup; no new timers or persistence | PASS | PASS: dialog cleanup, bounded runners, reversible presentation |
| FE-01 | Presentation in components; shared composables remain state owners | PASS | PASS: shared sources unchanged |
| FE-02 | Existing wire payloads/routes | PASS | PASS: no schemas or ingress handlers changed |
| FE-03 | aria-busy, preview/confirmed separation and errors retained | PASS | PASS: held-confirmation and timeout/recovery tests |
| FE-04 | Native links, keyboard/focus, 44px targets, widths/modes | PASS | PASS: geometry, screenshots, modifier-click and Back checks |
| FE-05 | No secrets/raw HTML/new auth or persistence | N/A | N/A |

Backend rules N/A: no Go/backend interface/storage changes; all backend CI checks stay.

## Contracts, rollout, risks

Required status remains `Run Tests`, verified with read-only branch-protection/ruleset
API calls; settings unchanged. Existing PR/main/merge-group events and main tag gate
stay. Only E2E job/artifact display names change. Roll back workflow/presentation
independently; no deployment sequencing, migration or data recovery needed.

More runners add installation cost and can queue. Large files remain indivisible;
use retained timings before adding shards. CSS columns read down columns and can
rebalance on content updates; source/keyboard order stays aligned. No exceptions.

## Traceability and verification

| IDs | Source / evidence | Actual result |
| --- | --- | --- |
| FR-01, AC-01 | Workflow, existing Playwright config; baseline plus four fresh-server partition runs | Initial 164-case inventory/executed union exact; 155 passes / nine existing skips |
| FR-01, AC-02/03 | actionlint, aggregate Bash cases, workflow comparison, docs/ci.md | PASS locally; hosted execution/uploads/timing pending |
| FR-02, AC-04/05 | DeviceDashboard.spec.ts, ExposeControl.spec.ts; light/dark screenshots and geometry | PASS: natural heights/compact footers, rails/targets/values, specified widths |
| FR-02, AC-06 | DeviceDashboard.spec.ts, EntityViewDialog.spec.ts | PASS: online/offline/Enter/Back/new-tab/missing; no navigation commands |
| FR-03, AC-07/08 | Card/row/entity companions plus EntityTile.spec.ts, PanelButton.spec.ts | PASS: no progress decoration/fades; held-confirmation, native commands and timeout recovery |
| FR-04, AC-09 | Workflow/AGENTS/skill and consolidated docs | PASS: one directory, repaired links, skill validator, whitespace; earlier completed specs untouched |
| SC-01 | Frontend Jest/lint/build and full isolated mock Chromium suite | 282 Jest tests; 173 browser passes / nine existing skips; lint/build pass |
| NFR-01 | Revised GitHub workflow after push | Unrun locally; compare actual hosted partition and total times |

## Recorded execution evidence

CI baseline (2026-10-19): 155 browser passes / nine existing skips, 3.7m. Each
fresh-server partition used `CI=true PLAYWRIGHT_JSON_OUTPUT_FILE=test-results/timings.json
npm run test:e2e -- --project=<project> --shard=<shard>/2 --reporter=list,html,json`.

| Partition | Local elapsed including startup/reporting | Passed | Skipped |
| --- | --- | --- | --- |
| desktop 1/2 | 54.4s | 43 | 9 |
| desktop 2/2 | 51.4s | 30 | 0 |
| mobile 1/2 | 66.7s | 52 | 0 |
| mobile 2/2 | 54.2s | 30 | 0 |

All initial partitions exited zero; executed identities equaled full discovery exactly
once. JSON included outcomes/config/durations. This validates isolation, not concurrent
Ubuntu scheduling. `actionlint@v1.7.12 -shellcheck=` passed; ShellCheck unavailable
locally, retained in CI. Extracted aggregate Bash passed all-success and rejected
failure/cancel/skip/empty. Non-E2E workflow compared identical to HEAD.

Final UI (2026-10-20): `npm test -- --runInBand` passed 42 suites / 282 tests;
`npm run lint` passed with seven existing warnings; `npm run build` passed TypeScript/
Vite with the existing duplicate DeviceCard auto-import warning. Final isolated run:
`CI=true APP_VERSION=slider-qa E2E_EXTERNAL_SERVERS=0 E2E_BASE_URL=http://localhost:4200
E2E_HUB_API=http://localhost:4210/api PLAYWRIGHT_JSON_OUTPUT_FILE=test-results/no-progress-timings.json
npm run test:e2e -- --reporter=list,html,json`: **173 passed / nine existing desktop
skips**, 182 cases, zero unexpected/flaky results (4.0m).

The QA copy at `/private/tmp/nodeherder-device-sliders-qa-qqndb4xu` matches production
sources byte-for-byte; test copies only substitute mock ports. Config uses separate
ports/cache/font allowlist. User's live servers on 4100/4110 were never reused/stopped.
UiRange, useExposeEdit and useDeviceCommand remain unchanged from HEAD.

Regression evidence: earlier footer height 67.375px failed compactness; device-row
outputs were misaligned by 86px; pending text moved downstream cards by 17.39px.
Final cases verify accepted behavior without those defects. The delayed indicator was
then removed entirely after user feedback. Two new panel cases initially selected
the wrong default room; explicit room selection fixed test setup, and both pass.
Initial isolated QA font access and test focus/new-tab assumptions were corrected
without weakening command or layout assertions. Screenshots reviewed in light/dark.

## Completion

Implementation is locally verified; hosted CI execution/artifact uploads and the
180s target require the PR workflow. Backend/container checks and non-Chromium
engines were not rerun locally for frontend/CI changes; their CI gates remain enabled.
Consolidation changes documentation only; recheck links, skill validity, whitespace
and final test inventory, without rerunning unchanged application suites.

Consolidation checks: final discovery contains 182 cases exactly once across four
partitions (59/32 per viewport). Relative documentation links resolve; stale draft
references are absent; `git diff --check` passes. The skill's bundled validator passes
using `/usr/local/bin/python3`; the default/Homebrew and system Python lack its PyYAML
dependency. No dependencies were installed. Earlier completed spec files are untouched.
