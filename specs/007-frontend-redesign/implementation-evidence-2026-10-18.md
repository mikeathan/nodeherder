# Entity/history implementation evidence

Scope: FR-20–24 / AC-44–50 on `feat/entity-history-redesign`, baseline
`771d2a9b2f64529d5a927b3d7e1d2ff2f03f5f07`. The owner's implementation/review/commit
instruction supersedes the earlier investigation-only checkpoint. Constitutions
remain proposed; no ratification, maintainer review or exception is invented.

## Final ownership and review

- `domain/controls.ts` owns native numeric bounds, steps and display conversion;
  `domain/exposes.ts` owns capability/labels/value formatting. `useExposeEdit` owns
  local drafts and delegates to `useDeviceCommand`; hub Vuex remains confirmed truth.
  Both device rows and full-screen controls use `UiRange`, including cancellation.
- `useEntityDialog` selects only relevant supported controls. Mobile entity dialogs
  fill the viewport; desktop adjustable views put history beside the control.
  Read-only views include compact useful summaries, real units and sample inspection.
- `domain/history/*` owns range, numeric reduction, state duration, event count and
  counter consumption rules. Shared `HistoryView`, `NumericHistory`, `StateHistory`
  and one options adapter replace the separate legacy mini/full chart implementations.
- `contracts/history.ts` validates HTTP ingress. `history.service.ts` owns a bounded
  deduplicated request queue, timeout and reference-counted cancellation. Vuex owns
  bounded results; `useHistoryQuery` rejects stale view/session completions. Logout
  clears histories and command timers. No per-tile polling or history persistence.
- Go handlers validate/adapt, the query service resolves devices and owns scan/time
  budgets, domain collectors preserve raw reports, and the repository owns storage.
  [Wire contract](history-contract.md) is additive; legacy HTTP/WS remain compatible.
  Cache snapshots and an actual cache boundary prevent duplicate/lost split samples;
  an empty cache after reopening uses persisted recent data.

The separate cleanup pass removed unused chart/slider implementations, factories,
helpers and their orphaned tests after import scans. Comments explain ownership,
draft/confirmed separation, unit inversion, gaps, counters and lifecycle guards.
Heavy chart code now loads asynchronously instead of being globally registered.
Chart.js 4.5.1 replaces ApexCharts/vue3-apexcharts; its small colour utility is the
only transitive addition. There is no storage schema/retention change. The library
change and measured reason are recorded in ADR-004.

## Acceptance evidence

| Acceptance | Evidence |
| --- | --- |
| AC-44 layout/design | EntityViewDialog/DeviceMetrics browser specs; full-screen 390px, no overflow at 320/360/768/1280/1920 and short 500px landscape; mock [screenshots](../../docs/design-previews/README.md) visually inspected |
| AC-45 controls | Native 184/254 brightness displays 72%; Kelvin display/mired command; mouse/touch one-send gesture and cancellation; keyboard, confirmation, timeout/refusal/disconnect, missing-reading/zero tests |
| AC-46 shared history | Shared renderer/domain/service tests; full-history numeric/contact/repeated event switching, custom invalid/valid dates, explicit error/refresh, generation/session invalidation |
| AC-47 producer | API raw repeated enum/truncation/400/500 tests, domain decode tests, service empty/failure tests, repository range/reopen/scan/cancellation regressions; existing API/WS suites remain passing |
| AC-48 energy | Source-owned energy tests: endpoints, reset, explicit gap, no cross-bucket allocation, missing baseline and measured zero; meter total separate from labelled observed consumption |
| AC-49 resources | Scheduler tests: one active, bounded queue/cache, shared cancellation/deadline/reset; 50 completed chart open/close cycles on desktop/mobile; heap/DOM evidence below |
| AC-50 quality | Pinned per-function TS/Vue/mock and Go checks below; no changed function exceeds C=10 or CRAP=30; no coverage exclusion added |

## Checks and runner

Darwin arm64; Node 24.21.0, Go 1.26.1; Playwright 1.56.1 / bundled Chromium
141.0.7390.37. Browser services are isolated mock hub 4110 and frontend 4100.
No household device was commanded by these tests.

- Frontend `npm test -- --runInBand --coverage`: 42 suites / 273 tests pass.
- Frontend `npm run lint`: pass, zero errors; seven existing warnings in unrelated
  `scripts/get-version.cjs` and `scripts/start.mjs`.
- Frontend `npm run build`: type check and production build pass. Main bundle is
  about 488kB / 144kB gzip after lazy chart loading; Chart.js is a separate ~208kB /
  71kB gzip chunk, versus the earlier ~578kB / 157kB Apex chunk. The build no longer
  reports an oversized chunk.
  Existing duplicate `DeviceCard` auto-import warning remains outside this scope.
- Frontend `npm run test:e2e`: all 104 desktop/mobile browser tests pass.
- `node --test tools/server/history.test.mjs`: mock contract test passes.
- Backend `go test ./...`, `go vet ./...`, `go build ./...`: pass.
- Backend `go test -race -coverprofile=/private/tmp/nodeherder-history-go.cover
  ./internal/metrics/... ./internal/api`: pass; temporary stores/mocks only.

Earlier iteration failures were corrected: explicit select names in history, timeout
rollback expectations, and a range-test fixture using an unsupported Playwright fill
operation. A prior auth handoff failure did not recur in the completed full suite.
Local-port sandbox restrictions required the tests to run with local binding enabled.

## Quality measurement

Frontend command, from repo root after fresh coverage:

```sh
node frontend/tools/quality/history-quality.cjs frontend/coverage/coverage-final.json
```

TypeScript 5.9.2 AST, source-line-preserving Vue script extraction, separate nested
functions/callbacks; 243 changed/new functions. Maximum C=9, maximum CRAP=30,
zero violations. 134 functions have Istanbul statement evidence; 109 use explicitly
labelled zero-coverage upper bounds. Removed predecessors remain in the baseline;
baseline upper bounds are not claimed as historical measured scores. See the
[runner documentation](../../frontend/tools/quality/README.md) for exact conventions.

Touched branch coverage includes history service 100%, control conversions 97.61%,
events 100%, ranges 100%, state 95.23%, numeric 94.44%, contract 89.18%, energy
89.47%, useExposeEdit 86.36%, useHistoryQuery 88.88%. Vue rendering is verified in
Playwright; generated templates are not included in the script complexity report.

Go scorer `github.com/hbaldwin98/crap/cmd/crap@8eea294df1123778a77ecb826f0b16d95b65768d`
(reported version `v0.0.0-20260908235400-8eea294df112`), from `backend/`:

```sh
crap -format json -coverage /private/tmp/nodeherder-history-go.cover \
  -diff-base 771d2a9b2f64529d5a927b3d7e1d2ff2f03f5f07 internal/metrics internal/api
```

22 changed functions; maximum C=9 / CRAP=9, zero violations. Coverage is mapped
from the affected race suite; whole-package percentages are not substitutes for
function scores. Reports/coverage are reproducible generated artifacts kept out of git.

## Performance and lifecycle

The 10,000-point normalization/statistics/extrema-reduction benchmark includes
runtime decoding, sorting and range filtering, three warmups and 20 measured runs.
Payload: 465,281 UTF-8 bytes. p95=8.017ms against the 100ms budget. The 20 measured
runs in milliseconds (sorted, recorded to three decimals):
`7.675, 7.711, 7.735, 7.747, 7.761, 7.780, 7.803, 7.826, 7.854, 7.875,
7.878, 7.897, 7.906, 7.906, 7.908, 7.938, 7.955, 7.986, 8.017, 11.522`.
Each chart renders at most 1,000 numeric points; dense state timelines use a labelled
64-bucket duration view and recorded samples paginate at 100 rows.

Lifecycle browser tests count app-owned modal key handlers, pending history reads,
body scroll restoration and retained canvas DOM after 50 completed charts.
An initial Apex registry-only check missed real mobile retention. Forced-GC DOM
and listener counts caught it in both dev and compiled production. A chart-free
control stayed stable; changing resize/animation teardown and avoiding the Vue
wrapper did not fix it. Those attempted workarounds are not retained. The final
Chart.js renderer passes the stronger regression:

| Project | Cycle 25 heap / nodes / listeners | Cycle 50 heap / nodes / listeners |
| --- | --- | --- |
| Desktop | 15,112,152 bytes / 1,748 / 98 | 15,188,040 bytes / 1,748 / 98 |
| Mobile | 15,193,116 bytes / 1,748 / 96 | 15,271,844 bytes / 1,748 / 96 |

Owned modal handlers, pending history reads and retained canvas elements are zero;
body scroll is restored. The regression requires no more than 100 extra nodes,
five listeners or 2MiB heap growth between cycles 25 and 50. No such native-resource
growth occurred. This is evidence for the tested flow, not universal leak freedom.

## Rule review and limits

NH-01 PASS: amended spec/ADR/plan/tasks before production edits. NH-02 PASS: command
preview/confirmation stays separate, no ingestion/automation timing changes. NH-03
PASS scoped review: validated history ingress, same route access, no new secrets or
history persistence. NH-04 PASS: established service/repository/Vuex boundaries and
documented additive contract. NH-05 PASS: source-owned mock/temp-store regressions
and required checks. NH-06 PASS scoped review: units, bounds, partial/unknown data,
resource caps and code-only recovery; lifecycle evidence accompanies this review.

FE-01/02/03/04/05 PASS scoped review: shared ownership, matching runtime contract,
explicit network/command outcomes, keyboard/touch/focus/viewports/theme tokens,
existing API helper and no new sensitive persistence. BE-01/03/04/05 PASS scoped
review: layer boundaries, synchronous request-owned budgets, race tests, inclusive
ordering/empty/error contract and unchanged access. BE-02 N/A: lifetime ingestion,
automation events and sampling were not changed; full suites verify compatibility.

This completes the scoped entity/history amendment, not unrelated older 007 tasks.
Tests cover Chromium desktop/mobile emulation; physical iOS/Android, assistive-tech
sessions and real-household checks were not run. History is a bounded snapshot with
explicit refresh. Chart rendering remains a lazy optional download. Deploy backend
and frontend together; rolling back code needs no data migration.
