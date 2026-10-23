# Plan: Responsive binary history chart

Spec: [spec.md](spec.md) | Status: Draft | Author: NodeHerder contributors | Updated: 2026-10-23
Constitutions: parent/frontend proposed 1.0.0

## Context and design

`HistoryView.vue` sends binary full-history series to `StateHistory.vue`. Today the component turns `stateSegments()` into a 2rem strip when there are up to 96 segments and a 64-bucket density strip above that threshold. It has three endpoint/midpoint labels, a legend and exact summary, but brief state changes are only narrow color slivers. `state.ts` already owns unknown intervals, expose-value semantics, exact stats and density. The history contract bounds decoded input at 10,000 samples; `RENDER_LIMIT` is 1,000 for numeric chart rendering. Existing tests include `state.test.ts`, expose palette assertions in `exposes.test.ts`, and browser history scenarios in `e2e/DeviceMetrics.spec.ts`.

The reviewed throwaway prototype demonstrates Option A and its mobile/dense variants (`../../binary-chart-review.html`). It is evidence for geometry, not production code or a passed acceptance test.

### Proposed implementation

1. Keep `stateSegments()` and `stateStats()` authoritative. Add a pure width-binning helper in `domain/history/state.ts` that consumes sorted segments and returns per-column active duration, known duration, transition count, and active percentage (`null` when no known time). Sweep segments through bins in O(N + W); do not rescan all samples for every bin. Keep input segments immutable. Unit tests prove bin-duration conservation, unknown-only bins, event counts and empty/partial cases.
2. Update only the **non-compact** branch of `StateHistory.vue` to render Option A as inline SVG. The upper lane is the metadata-defined active value; the lower lane is inactive. Use expose labels from `formatExposeValue()`, colors from `exposePalette()`, and a distinct hatch for unknown. Preserve existing summary, partial/unknown semantics and the `compact` strip branch.
3. Measure the chart host with `ResizeObserver`; disconnect on unmount. Use responsive lane proportions: active lane at 30% (wide) / 34% (narrow), inactive lane at 70% / 72%, measured from plot top. Derive the left label gutter from the active/inactive text, adapt tick spacing to plot width, and clamp endpoint labels/tooltips inside the host. Re-render on width changes without a history request.
4. In sparse mode, draw time-proportional horizontal state runs, vertical transition edges and restrained transition markers. Choose the mode from segment count versus measured plot width (initial target: aggregate when `segments > max(40, plotWidth / 4)`); keep the threshold centralized/testable. In dense mode, draw one duration-share column per CSS pixel, capped at `RENDER_LIMIT`; active percentage uses known duration only. Keep unknown visually distinct, and include its coverage in the bin tooltip so a gap cannot read as an inactive measurement.
5. Support pointer hover and mobile tap readout, plus a focusable chart interaction surface with left/right keyboard navigation and visible focus. Sparse lookups use binary search over segments; dense lookups map pointer x directly to a bin. Readouts identify the active/inactive label, time bounds, duration, or (for a bin) active share, known coverage and transition count. Avoid an always-announcing live region during pointer movement.
6. Keep exact summary values sourced from `stateStats()`, not bins. Preserve recent reports/sample paging, empty/loading/error/partial UI, full-history routes, and compact dialog display. No new package or external interface.

### Alternatives

- **Keep the existing strip and only raise its height:** rejected; long-range brief events remain sub-pixel and state labels still require legend lookup.
- **Draw every interval as SVG regardless of density:** rejected; up to 10,000 samples could produce a large DOM and a visually solid comb.
- **Extend Chart.js for categorical state lanes:** possible, but custom state labels, unknown hatching, responsive density aggregation and interaction would require chart-specific callbacks/overlays. Native SVG is a better fit for two categorical lanes, already demonstrated without a dependency; keep the transformation in tested domain code.

## Constitution Check

The constitutions are proposed, not ratified. “Before implementation” records design evidence only; all execution and merge checks remain pending.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | User-selected Option A and acceptance IDs in `spec.md`; scope excludes unrelated chart/data changes. | PASS — specified before production edits. | Pending diff/acceptance review. |
| NH-02 | Read-only visual change; state event semantics and device commands are unchanged; assert no extra commands. | PASS — no device-control path changed. | Pending source/e2e evidence. |
| NH-03 | Uses already-loaded series only; no new endpoint, persistence, logging or household-data surface. | PASS — no trust-boundary change. | Pending diff review. |
| NH-04 | Vue owns presentation; pure transform stays under `domain/history`; existing props/contracts and dependencies remain. | PASS — no API, store or chart-library replacement. | Pending diff and build review. |
| NH-05 | ACs map to existing domain and Playwright companions; deterministic mocked fixture covers 10,000 inputs and responsive modes. | PASS — verification is defined. | Pending commands/results. |
| NH-06 | Unknown/partial time remains distinct; summaries use exact segments; input cap and sample access are unchanged. | PASS — data integrity is specified. | Pending regression evidence. |
| FE-01 | `StateHistory.vue` renders; state-bin transformation remains a pure tested domain helper. | PASS — ownership preserved. | Pending review. |
| FE-02 | No request/response, WS or runtime-ingress change. | N/A — no browser contract change. | N/A unless scope changes. |
| FE-03 | Preserve empty/error/partial notices, gap handling and sample access; no additional requests. | PASS — covered by AC-08/09. | Pending browser checks. |
| FE-04 | Dynamic text labels, hatch, keyboard/focus, pointer/tap, light/dark, 320–1280px checks. | PASS — design specified. | Pending browser/manual evidence. |
| FE-05 | No secret, assistant content, storage or authenticated-call changes. | N/A — not applicable to this display-only change. | N/A unless scope changes. |

No backend constitution items apply; no backend files or device behavior are in scope.

## Contracts, data, rollout

No HTTP/WS/MQTT/MCP/configuration or persistence change. The 10,000-sample decode limit, time bounds, expose labels, state conversion, partial flag, API request queue and route selection remain unchanged. Bins are render-only and are never used for exports, sample tables or exact duration summaries. A frontend rollback restores the current strip; no data migration is needed. No new runtime dependency or bundle chunk is introduced.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02 | `StateHistory.vue` | Extend `DeviceMetrics.spec.ts`; assert both labels, lane positions and transitions using mocked contact history. | Pending — planning only. |
| AC-03, NFR-02 | `StateHistory.vue`, `DeviceMetrics.spec.ts` | Desktop/mobile Playwright viewports 1280×800, 768×1024, 390×844 and 320×844; light/dark; assert no app/page overflow and inspect screenshots. | Pending — planning only. |
| FR-02, AC-04 | `StateHistory.vue`, `exposes.test.ts`, `DeviceMetrics.spec.ts` | Keep domain palette assertions; browser-check computed token colors and labels for the contact fixture; retain existing expose-kind mappings. | Pending — planning only. |
| FR-03, AC-05/07, NFR-01 | `domain/history/state.ts`, `src/__tests__/domain/history/state.test.ts`, `DeviceMetrics.spec.ts` | Unit-test duration/coverage conservation; 10,000-point mocked alternating contact series; assert aggregated mode, at most `min(plotWidth, RENDER_LIMIT)` bins, exact totals, and no loss of unknown coverage. | Pending — planning only. |
| FR-04, AC-06 | `StateHistory.vue`, `DeviceMetrics.spec.ts` | Verify pointer hover, mobile tap, arrow-key readout, visible focus, unclipped tooltip and accessible chart name. | Pending — planning only. |
| FR-05, AC-08/09 | `StateHistory.vue`, `HistoryView.vue`, `DeviceMetrics.spec.ts`, `EntityViewDialog.spec.ts` | Re-run current empty/partial/contact/sample and compact-dialog checks; assert history inspection adds no query or device command. | Pending — planning only. |
| NH-05 | `frontend/` | `npm test -- --runInBand`; `npm run lint`; `npm run build`; `npm run type-check:e2e`; isolated Playwright projects for `DeviceMetrics.spec.ts` and `EntityViewDialog.spec.ts`. | Not run — planning only. |

Run frontend commands from `frontend/`. Browser tests must use the configured isolated/mock hubs; never run routine checks against household devices. Inspect both-theme screenshots and record actual viewport/browser versions, failures, and any unrun manual Safari check in this plan. Backend tests are not proposed because no backend behavior or files change.

## Debt, risks, exceptions

- SVG axes, event markers and tooltip content are hand-rendered; add source-owned tests for coordinate edges and prevent regressions on short and very narrow ranges.
- Some state names can be longer than Open/Closed; derive the label gutter from actual localized labels and test `Smoke detected` / `Detected` rather than assuming fixed widths.
- Dense bins intentionally summarize within-bin transitions. Exact raw sample paging remains the route to individual reports; tooltip must make aggregation explicit and include coverage.
- Avoid mounting an interactive SVG overlay with an inaccessible `role="img"` descendant. Verify the chosen group/interaction semantics with Playwright role/name/focus assertions and manual keyboard review.
- Existing palette choices are semantic and expose-name-based; do not reinterpret alert polarity or use colors as the only state signal.
- Exceptions: none requested or approved.

## Completion

Planning only. No application tests or production changes are claimed. Before implementation, confirm the reviewed full-history/compact scope and then follow the tasks in `tasks.md`. At handoff, update this table with actual test/manual results; do not mark planning evidence as execution evidence.
