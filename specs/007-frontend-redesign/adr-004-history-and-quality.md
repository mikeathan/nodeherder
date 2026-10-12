# ADR-004: Shared history queries, renderer boundary and measured quality

Status: Proposed | Date: 2026-10-11 | Spec: [spec.md](spec.md), FR-14…19

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
