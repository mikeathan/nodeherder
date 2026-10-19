# ADR-004: Shared history queries, renderer boundary and measured quality

## Implementation refinement — 2026-10-17

Owner authorised implementation. Add opt-in `history: true` to existing metrics HTTP
request. Each returned device value is `{name,type,from,to,data,partial}`. All samples use
`{timestamp,value}` with epoch milliseconds and typed numeric/boolean/string values. Legacy requests retain
their existing shapes. History limits and scan/time/device bounds are defined in
FR-23. Context and collector budget checks stop work at scan callbacks. Storage
bytes/keys and legacy interfaces remain unchanged. Enum history uses a raw collector
only for this option. Missing bucket is empty; genuine errors propagate.

Energy bars explicitly show observed within-bucket deltas and completeness; unknown
boundaries/reset/gap are not measured zero. They never silently assert full-period
consumption from incomplete data. This replaces the earlier exclusion on inferred
counter consumption for the scoped entity/history redesign.

Keep source-owned Jest/Playwright checks. Conservative zero-coverage
CRAP upper bounds cover Vue script/mock code when Istanbul source coverage is absent;
do not call upper bounds measured coverage. Keep low complexity and document before/
after functions, including removed predecessors. Final renderer selection follows
the lifecycle amendment below.

Lifecycle amendment — 2026-10-18: the isolated mobile repeated-close test found
retained DOM/listener growth with ApexCharts 4.7.0 after destruction, also with a
raw unproxied instance in the compiled production build. Dialogs without charts
stayed stable. Replace only the renderer with pinned Chart.js 4.5.1, whose canvas
instance is owned outside Vue deep reactivity and explicitly destroyed on unmount.
Retain shared transforms, units, statistics, controls, tokens and visual direction.
Chart.js supplies monotone interpolation, steps, bars, gradients and native responsive
canvas rendering. This replaces Apex/Vue-Apex runtime dependencies; resource and
bundle comparisons must be re-run. No date adapter is needed for epoch-linear axes.
Verify retained DOM/listener/heap trends, not just an empty library registry.
The adapter follows the documented [Chart.js destruction contract](https://www.chartjs.org/docs/latest/developers/api.html#destroy).

Status: Proposed | Date: 2026-10-11 | Spec: [spec.md](spec.md), FR-14…19

Implementation and verification: [2026-10-18 evidence](implementation-evidence-2026-10-18.md).
The original proposal below remains historical where superseded by the refinement.

## Context

Legacy full and mini charts use `loadMetrics` and a Vuex result map keyed only by device.
`DeviceMetrics` starts/ends use milliseconds in Go but the frontend declares optional
strings; Go binary samples contain booleans but frontend samples declare strings.
`useMiniChartData` watches another view's device result and has no bounded error completion.
`DeviceMetrics.vue` stops loading after sending, before receiving. The backend WS handler
currently ignores the requested expose; the mock honours it. These are actual mismatches.

The existing HTTP query endpoint can request multiple exposes and supplies query bounds,
units and raw numeric/binary data under `values[].value` when `aggregation` is empty.
It supports request cancellation in the browser and avoids adding WS correlation fields.
However, `QueryService.Query` currently continues on query/collector errors, which can
make failures look like empty data. There is no repository CRAP gate found in the inspected
frontend scripts or CI; Jest currently transforms TypeScript, not Vue mounting tests.

## Proposed decision

1. Reuse `POST /api/metrics/query` for the new history service via `contracts/api.ts`.
   Validate its actual raw payloads at ingress; keep old WS events compatible while any
   consumer remains. Correct touched failure handling so absence is distinct from genuine
   query failure. Keep existing HTTP error envelope/status conventions. No new public route.
2. Keep query entries in Vuex under a canonical device/exposes/range/aggregation key, with
   request generation, status, bounds and truncation. The service owns cancellation/timeout;
   composables own view subscriptions. Confirmed devices remain in the hub module.
3. Retain ApexCharts initially behind one typed renderer adapter. Pure domain functions own
   sample normalization, stats, timelines and decimation; full/mini views reuse them.
   Binary/event views can use accessible HTML/SVG and tables without another chart library.
4. Add reproducible per-function CRAP reporting for TS and Vue script blocks, plus Go for
   any touched backend functions. Pin metric conventions, versions and coverage mappings
   before measuring the baseline. Extend Jest with a compatible Vue transformer/mounting
   harness only if needed for Vue coverage; verify its Node 24/ESM compatibility first.
   Do not migrate testing frameworks merely to calculate a score.

Thresholds and budgets are proposed in NFR-07/08. Tool selection/coverage proof is a
preparation gate, not a claim that the gate already exists or scores have been measured.
Any new development dependency records purpose, compatibility, maintenance cost and lockfile
impact here during implementation. Renderer/package replacement requires an ADR amendment.

## Alternatives

| Alternative | Tradeoff |
| --- | --- |
| Extend WS with request IDs/errors | Supports current transport but requires coordinated producers, routing and mocks; avoid while existing HTTP contract meets the need |
| Serialize legacy WS queries per device | Cannot robustly distinguish delayed responses for repeated identical ranges; leaves current failure semantics ambiguous |
| Separate caches/transforms per chart | Repeats state ownership and permits full/mini disagreement |
| Replace ApexCharts immediately | Dependency and bundle cost before a demonstrated need; complete UX replacement does not require immediate library replacement |
| Gate file-average CRAP or tests alone | Hides complex untested functions and rewards coverage without reducing responsibilities |

## Consequences and recovery

The mock needs the actual HTTP query contract and empty/failure/429 cases. Backend changes
are limited to required query correctness, and receive Go regression/race checks appropriate
to their scope. The endpoint is currently registered publicly in `internal/hub.go`; this
decision does not claim authentication is present or change access. Browser calls use the
established cookie helper; no new exposure or sensitive cache persistence is introduced.

Deploy any required compatible backend fixes before the new frontend. If producer/schema
changes become necessary, revise the spec/ADR first. Roll back the frontend to #54 without
storage migration; retain legacy metrics contracts until import/consumer scans prove removal
safe. A failed backend preparation gate blocks history integration, not independent entity,
quality-tooling or mock design work. Storage scans are not bounded by the current response
limit; backend resource limits remain a measured design risk, not a solved claim.
