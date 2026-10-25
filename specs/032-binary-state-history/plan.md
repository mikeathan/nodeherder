# Plan: Responsive binary history chart

Spec: [spec.md](spec.md) | Status: Implemented; manual Safari review pending | Author: NodeHerder contributors | Updated: 2026-10-23
Constitutions: parent/frontend proposed 1.0.0

## Context and design

`HistoryView.vue` sends binary full-history series to `StateHistory.vue`. Today the component turns `stateSegments()` into a 2rem strip when there are up to 96 segments and a 64-bucket density strip above that threshold. It has three endpoint/midpoint labels, a legend and exact summary, but brief state changes are only narrow color slivers. `state.ts` already owns unknown intervals, expose-value semantics, exact stats and density. The history contract bounds decoded input at 10,000 samples; `RENDER_LIMIT` is 1,000 for numeric chart rendering. Existing tests include `state.test.ts`, expose palette assertions in `exposes.test.ts`, and browser history scenarios in `e2e/DeviceMetrics.spec.ts`.

The reviewed throwaway prototype demonstrates Option A and its mobile/dense variants (`../../binary-chart-review.html`). It is evidence for geometry, not production code or a passed acceptance test.

### Proposed implementation

1. Keep `stateSegments()` and `stateStats()` authoritative. Add a pure width-binning helper in `domain/history/state.ts` that consumes sorted segments and returns per-column active duration, known duration, transition count, and active percentage (`null` when no known time). Sweep segments through bins in O(N + W); do not rescan all samples for every bin. Keep input segments immutable. Unit tests prove bin-duration conservation, unknown-only bins, event counts and empty/partial cases.
2. Update only the **non-compact** branch of `StateHistory.vue` to render Option A as inline SVG. The upper lane is the metadata-defined active value; the lower lane is inactive. Use expose labels from `formatExposeValue()`, colors from `exposePalette()`, and a distinct hatch for unknown. Preserve existing summary, partial/unknown semantics and the `compact` strip branch.
3. Measure the chart host with `ResizeObserver`; disconnect on unmount. Use responsive lane proportions: active lane at 30% (wide) / 34% (narrow), inactive lane at 70% / 72%, measured from plot top. Derive the left label gutter from the active/inactive text, adapt tick spacing to plot width, and clamp endpoint labels/tooltips inside the host. Re-render on width changes without a history request.
4. In sparse mode, draw time-proportional horizontal state runs, vertical transition edges and restrained transition markers. Choose the mode from segment count versus measured plot width (initial target: aggregate when `segments > max(40, plotWidth / 4)`); keep the threshold centralized/testable. In dense mode, draw one duration-share column per CSS pixel, capped at `RENDER_LIMIT`; active percentage uses known duration only. Keep unknown visually distinct, and include its coverage in the bin tooltip so a gap cannot read as an inactive measurement.
5. Support pointer hover and mobile tap readout, plus a focusable chart interaction surface with left/right keyboard navigation and visible focus. Sparse lookups use binary search over the **visible** segments; dense lookups map pointer x directly to a bin. Readouts identify the active/inactive label, time bounds, duration, or (for a bin) active share, known coverage and transition count. Avoid an always-announcing live region during pointer movement.
6. Add one **same-chart activity focus** mode. A pure `recentActivityWindow()` helper selects the latest ten state transitions, pads them by one minute on each side, enforces an 8-minute minimum where the loaded range allows, and clamps to `series.from/to`. Label it `Zoom to activity`, give it a 44px minimum target, and show it only when that cluster would occupy less than 50 CSS pixels on the overview. Toggling it changes local display bounds only: not the selected period, series, full-range stats, samples, exports, route or request count. The 44px-minimum control becomes `Back to full range`; the focused bounds are shown while active.
7. When `stateStats().duration / selected duration < 50%`, place a concise known/no-data duration note beside the plot. This supplements—not replaces—the hatched unknown timeline, and continues to describe the full selected period while focused.
8. Keep exact summary values sourced from full-series `stateStats()`, never from the bins or focus viewport. Preserve recent reports/sample paging, empty/loading/error/partial UI, full-history routes, and compact dialog display. No new package or external interface.

### Alternatives

- **Keep the existing strip and only raise its height:** rejected; long-range brief events remain sub-pixel and state labels still require legend lookup.
- **Draw every interval as SVG regardless of density:** rejected; up to 10,000 samples could produce a large DOM and a visually solid comb.
- **Extend Chart.js for categorical state lanes:** possible, but custom state labels, unknown hatching, responsive density aggregation and interaction would require chart-specific callbacks/overlays. Native SVG is a better fit for two categorical lanes, already demonstrated without a dependency; keep the transformation in tested domain code.
- **Render a permanent second detail chart:** rejected after review; it duplicates the same history and reads as two competing charts. Keep one plot and let the user temporarily focus its x-range.
- **Widen short events on the overview:** rejected; it falsifies durations and can make a rare sensor hit look longer than it was. Use the local focus view instead.

## Constitution Check

The constitutions are proposed, not ratified. “Before implementation” records design evidence; “Before merge” records the final local code review and verification, not external maintainer approval.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | User-selected Option A plus the approved one-chart activity-focus follow-up in `spec.md`; no unrelated history behavior. | PASS — specified before production edits. | PASS — AC-10…12 and final scope reviewed. |
| NH-02 | Read-only visual change; no state/event or device-command logic changed; browser test observes no command. | PASS — no device-control path changed. | PASS — focus/reset sent no device command. |
| NH-03 | Uses already-loaded history only; no new endpoint, persistence, logging or household-data surface. | PASS — no trust-boundary change. | PASS — focus is local-only and emitted no additional history request. |
| NH-04 | Vue owns SVG presentation; pure binning stays in `domain/history`; existing props/contracts/dependencies remain. | PASS — no API, store or chart-library replacement. | PASS — build and final diff show no contract/dependency changes. |
| NH-05 | ACs map to source-owned Jest and Playwright companions; fixtures cover 10,000 points and rare-event focus. | PASS — verification was defined. | PASS — final check results are recorded below; Safari remains unrun. |
| NH-06 | Unknown/partial time remains distinct; summaries use exact segments; local focus never changes the loaded period. | PASS — integrity requirements specified. | PASS — focus clamps to loaded bounds; full-period stats/export bounds remain unchanged. |
| FE-01 | Component owns presentation; pure transform remains independently tested in the existing state companion. | PASS — ownership preserved. | PASS — diff keeps state logic pure and chart interactions in Vue. |
| FE-02 | No request/response, WS or runtime-ingress change. | N/A — no browser contract change. | N/A — contract paths unchanged. |
| FE-03 | Existing empty/error/partial, gaps, sample access and request queue preserved. | PASS — covered by AC-08/09. | PASS — DeviceMetrics and EntityViewDialog browser suites pass. |
| FE-04 | Dynamic labels, hatching, keyboard/focus, pointer/tap, the new focus/reset control, both themes and required widths. | PASS — design specified. | PASS — accessible zoom/reset and coverage note verified in both browser projects; Safari remains unrun. |
| FE-05 | No secret, assistant content, storage or authenticated-call changes. | N/A — not applicable to this display-only change. | N/A — no sensitive-operation path changed. |

No backend constitution items apply; no backend files or device behavior are in scope.

## Contracts, data, rollout

No HTTP/WS/MQTT/MCP/configuration or persistence change. The 10,000-sample decode limit, time bounds, expose labels, state conversion, partial flag, API request queue and route selection remain unchanged. Bins and activity focus are render-only; focus never changes the selected series/range used by summaries, sample tables or exports. A frontend rollback restores the current strip; no data migration is needed. No new runtime dependency or bundle chunk is introduced.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02 | `StateHistory.vue`, `DeviceMetrics.spec.ts` | Assert expose labels, state transitions and inactive-lane proportions from mocked contact history. | PASS — desktop/mobile projects; 320/390/768/1280px, both themes. |
| AC-03, NFR-02 | `StateHistory.vue`, `DeviceMetrics.spec.ts` | Responsive axes/no-collision/page-overflow assertions and screenshot inspection. | PASS — both projects across all four widths/themes; phone and wide screenshots inspected. |
| FR-02, AC-04 | `StateHistory.vue`, `exposes.test.ts`, `DeviceMetrics.spec.ts` | Preserve semantic token colors and contact labels; run existing kind-palette unit coverage. | PASS — expose unit suite and browser computed-color assertions. |
| FR-03, AC-05/07, NFR-01 | `domain/history/state.ts`, `src/__tests__/domain/history/state.test.ts`, `DeviceMetrics.spec.ts` | Verify exact duration/coverage conservation and a 10,000-point fixture bounded by `min(plotWidth, RENDER_LIMIT)`. | PASS — 8 state tests; exact 10,000-sample active/known/transition totals and bins passed in both browser projects. |
| FR-04, AC-06 | `StateHistory.vue`, `DeviceMetrics.spec.ts` | Verify pointer hover, mobile tap, arrows, visible focus, tooltip and accessible names. | PASS — sparse and aggregated browser inspection passed. |
| FR-05, AC-08/09 | `StateHistory.vue`, `HistoryView.vue`, `DeviceMetrics.spec.ts`, `EntityViewDialog.spec.ts` | Preserve empty/partial/contact/sample and compact dialog behavior; no query/command from inspection. | PASS — full targeted DeviceMetrics + EntityViewDialog suites passed in both projects. |
| FR-06, AC-10/11/12, NFR-03 | `state.ts`, `StateHistory.vue`, `state.test.ts`, `DeviceMetrics.spec.ts` | Sparse burst in 24h/6h; focus/reset one chart; verify 1h desktop context, phone responsiveness, period/full stats/coverage/request count. | PASS — 8 transitions and 4 min detected focus in 6h/24h; 1h desktop remains overview, phone adapts; `<1%` coverage displays honestly; selected range/stats preserved and no extra request/command. |
| NH-05 | `frontend/` | `npm test -- --runInBand`; `npm run lint`; `npm run build`; `npm run type-check:e2e`; `CI=true npm run test:e2e -- --project=desktop --project=mobile e2e/DeviceMetrics.spec.ts e2e/EntityViewDialog.spec.ts`. | PASS — 53 suites/352 tests; lint 0 errors (7 pre-existing script warnings); build and e2e type-check passed; 95 browser tests passed, 9 existing desktop-only skips, 0 failures. |

Commands ran from `frontend/`. Browser tests used the configured isolated/mock hubs; no household hub/device was used. Desktop/mobile Playwright ran Chrome at the configured 1280×800 and Pixel 7 390×844 baselines, with explicit responsive assertions at 320×844, 390×844, 768×1024 and 1280×800. Screenshots were inspected in both themes. Physical Safari/iOS and a real-device pass remain unrun. Backend tests were not run because no backend behavior/files changed.

## Debt, risks, exceptions

- Activity focus must not silently become another range preset. Keep the selected range and its full-range summary visible, and clearly label the local focus bounds plus the reset action.
- Sparse mode must filter to visible segments before drawing; aggregation remains O(N + W), and unknown stays patterned in overview and focus.
- SVG axes, event markers and tooltip content are hand-rendered; add source-owned tests for coordinate edges and prevent regressions on short and very narrow ranges.
- Some state names can be longer than Open/Closed; derive the label gutter from actual localized labels and test `Smoke detected` / `Detected` rather than assuming fixed widths.
- Dense bins intentionally summarize within-bin transitions. Exact raw sample paging remains the route to individual reports; tooltip must make aggregation explicit and include coverage.
- Avoid mounting an interactive SVG overlay with an inaccessible `role="img"` descendant. Verify the chosen group/interaction semantics with Playwright role/name/focus assertions and manual keyboard review.
- Existing palette choices are semantic and expose-name-based; do not reinterpret alert polarity or use colors as the only state signal.
- Exceptions: none requested or approved.

## Review follow-up (2026-10-23)

Self-review of the first implementation found the component had grown to ~840 lines mixing
layout, axis, interaction and two render modes, plus duplicated duration/percentage
formatting. The follow-up restructures it without changing acceptance:

- `domain/history/format.ts` (new, tested) owns every user-facing duration, moment and
  coverage string; `StateHistory`, `StateHistoryChart`, `StateHistoryCompact`,
  `HistoryView` and the numeric `options.ts` axis now reuse it instead of local copies.
- `stateLayout.ts` (new, tested) owns lane geometry, gutters, dense/focus thresholds and
  axis ticks; `stateAxisLabel` was folded into the shared formatter.
- `StateHistory.vue` is now an orchestrator (~220 lines) rendering `StateHistorySvg.vue`
  (presentational), `StateHistoryChart.vue` (interaction/resize) and
  `StateHistoryCompact.vue` (dialog strip).
- `stateTransitions()` in `domain/history/state.ts` replaces three copies of transition
  scanning (stats, activity window, SVG markers).
- Pointer inspection is coalesced per animation frame; zoom/reset announce the new range
  through a polite status region; test-only duplicate `data-*` metrics were removed in
  favour of asserting rendered summary text.
- Browser fixtures are now registered per test with `route.fallback()` instead of shared
  module-level flags, so sharding can be enabled safely.
- `.vue` linting was previously inert (`files: ['*.vue']` matched only the root, and a
  removed `vue/script-setup-uses-vars` rule aborted root-level runs). The ESLint config
  now parses `<script setup lang="ts">` and lints every SFC with essential correctness
  rules. This surfaced and fixed five pre-existing errors (`Range.vue` prop shadowing,
  `Selection.vue`/`MenuDropdown.vue` array/object defaults, `EntityViewDialog.vue`
  duplicate `title`, `ExposeSelectionDialog.vue` missing attribute whitespace).

Verification for this follow-up: 55 Jest suites / 369 tests, lint 0 errors (28 warnings,
all pre-existing rules on newly linted SFCs), build and e2e type-check pass, and the
desktop/mobile `DeviceMetrics` + `EntityViewDialog` suites pass 95 tests with 9 existing
desktop-only skips. Physical Safari/iOS remains unrun.

## Execution notes

- The pre-implementation baseline browser pass was 2/2 projects with no overflow; its screenshots show the original thin strip.
- The first domain regression run failed as expected because `stateHistoryBins` did not yet exist. The first integration run then exposed the old e2e role locator and a focus-ring assertion that relied on `:focus-visible` after programmatic focus. The test now targets the new named inspector group and focus gets a visible ring whenever focused; rerun passed.
- The first e2e type-check found that this tsconfig does not include iterable DOM collections. The test uses `Array.from`; the final type-check passed.
- First iteration: 53 suites / 348 tests; lint 0 errors with 7 pre-existing script warnings; build and e2e type-check passed; 93 desktop/mobile browser tests passed with 9 existing desktop-only skips.
- Activity-focus follow-up: 53 suites / 352 tests; lint 0 errors with 7 pre-existing script warnings; build and e2e type-check passed; 95 desktop/mobile browser tests passed with 9 existing desktop-only skips and 0 failures. This includes the eight-transition/four-minute fixture at 6h/24h, the 1h desktop overview, 44px action/reset targets, `<1%` coverage label, full-range focus/reset checks, and the 10,000-sample regression.
- Final focused regression rerun after the reset target-size assertion passed 2/2 desktop/mobile projects.
- Final focused e2e rerun after asserting the reset target size passed 2/2 desktop/mobile projects.

## Completion

Option A now includes a one-chart recent-activity focus. The overview remains time-proportional; the accessible action zooms the same plot to a recent transition cluster and offers a direct reset. Full-range statistics, coverage, selected period, sample access and export bounds are preserved. Automated acceptance and local code/rule review are complete. Physical Safari/iOS remains unrun. No backend/API/WS/persistence/dependency changes or approved exceptions.
