# Plan: History, charts and mobile shell

Spec: [spec.md](spec.md) | ADR: [adr-006](adr-006-chart-interactions.md) | Status: Draft
Author: NodeHerder contributors | Updated: 2026-10-22
Constitutions: parent proposed 1.0.0; frontend 1.0.0; backend 1.0.0 (no backend change in scope)

## Context and design

Observed facts (verified in code):

| Area | Current behaviour | Path |
| --- | --- | --- |
| History entry | Device page tab with two `<select>`s; dialog shows a fixed 4 h with no route onward | `components/device/DeviceMetrics.vue`, `components/dialogs/EntityViewDialog.vue` |
| Range presets | `1h…30d`, `today`, `custom`; **no Yesterday**; browser zone only | `domain/history/range.ts` |
| Chart | 235 px (105 compact); three x ticks forced; no zoom/pan/readout; y-axis hidden when compact | `components/chart/history/{NumericHistory,options}.ts` |
| Rendering | `decimate` keeps gaps and extrema within `RENDER_LIMIT` 1 000; decode caps at 10 001 | `domain/history/numeric.ts`, `contracts/history.ts` |
| Query state | One serialised request, 300 ms debounce, generation/epoch guards, 10 s timeout | `services/history.service.ts`, `composables/useHistoryQuery.ts` |
| Stats | `numericStats` (min/max/avg/last) exists; no bucketing, no export | `domain/history/numeric.ts` |
| Statistics API | Whole-window aggregation; no interval/bucket parameter; one device per series | `backend/internal/metrics/query/{request,aggregates}.go` |
| Shell | Sidebar ≥960 px, drawer below; `--nh-bottom-nav-h` unused | `components/shell/AppShell.vue`, `NavSidebar.vue`, `tokens.css` |
| Platform | No `viewport-fit=cover`, no manifest, safe-area insets inert | `index.html`, `public/` |
| Data volume | Already bounded four times, see below — no new dropping needed | backend + decoder + renderer |

Data volume is already handled at four layers; 029 must not add a fifth:

| Layer | Bound | Behaviour when exceeded |
| --- | --- | --- |
| Repository scan | `maxScannedRows` 250 000 rows / 8 s budget | Hard error: "history scan budget exceeded; narrow the range" |
| Query limit | `maxQueryPoints` 10001, ascending | `data[:limit]` keeps the **oldest** 10 001 samples |
| Decoder | `HISTORY_LIMIT` 10 000 | Rejects >10 001, slices to 10 000, sets `partial` |
| Renderer | `RENDER_LIMIT` 1 000 | `decimate` keeps gaps and per-bucket extrema; may return nothing when too fragmented |

Consequences for this change: statistics bucket the raw samples (0–10 000) into one bucket per
**intersecting local calendar day**. A valid range of at most 31 days can touch **32 dates** (for
example 1 Jan 12:00 → 1 Feb 12:00 covers 1 Jan…1 Feb), so the bucket count is ≤ 32, not 31. Bucket
ownership is `[local midnight, next local midnight)`: a sample at exactly midnight belongs to the day
starting there, and the first and last buckets are marked partial when the range does not start or
end at midnight, so samples are neither dropped nor counted twice. The chart keeps `decimate`; a
`partial` series must show which window it covers, keep uncovered buckets empty, refuse export, and
offer a narrower range instead of silently changing the user's selection.

Design — one shared history stack, three presentations:

```
domain/history/          pure: range presets, decimation, stats, bucketing, CSV rows
services + store         unchanged transport and query identity
components/chart/history/
  HistoryChart.vue       owns the Chart.js instance, our interaction plugin, buttons, keyboard, a11y
  HistoryToolbar.vue     presets, expose chips, mode switch, actions
  HistoryStats.vue       min/avg/max/last strip
  HistorySamples.vue     existing table, extracted
  HistoryView.vue        composes them; keeps `compact` prop for the sheet
components/device/DeviceMetrics.vue   full page: wide layout, desktop density
components/dialogs/EntityViewDialog.vue   compact chart + "View full history"
components/shell/{AppShell,BottomNav}.vue mobile bar, drawer retained
```

Ownership rules: domain modules stay framework-free and never import Vue/store/Chart.js; the chart
component owns rendering and gestures; the page owns layout and persistence; the store keeps owning
query identity. No second source of truth for navigation or history state.

Per-viewport contract (this is the "mobile focused, desktop full" requirement):

| Width | Layout | Chart | Controls | Density |
| --- | --- | --- | --- | --- |
| ≥1024 px | Full width, no 64 rem cap on the history route, samples table visible | 360–420 px, y-axis + end marker | Chips inline, keyboard shortcuts, legend for multi-expose | Detailed: stats strip, table, export |
| 720–1023 px | Single column, content padding reduced | 300 px | Chips wrap, no wheel-only affordances | Medium: stats strip, table collapsed |
| <720 px | Bottom bar, sheet flow, no sidebar | 220–260 px, compact axes | 44 px buttons, pinch, one-finger readout, horizontally scrolling chips | Focused: stats strip, table behind a disclosure |

Navigation (mobile): the bottom bar uses existing routes only — **Home** (`GroupDashboard`),
**Devices** (`Devices`), **Automations** (`Viewer`), **Assistant** (`Assistant`), and **More**, which
opens the existing drawer. History is not a destination because it is device-scoped. A tab is current
when the route matches it or its `includes` list; More is current for any route outside the four and
exposes `aria-expanded` for the drawer. `NAV_ITEMS` gains a `primary` marker plus a derived drawer
list, so there is still one navigation source.

Statistics availability: `numeric` series get the History/Statistics switch; `binary` and `enum`/
event series keep the state timeline and recent-event list without a switch; energy counters keep the
existing observed-consumption summary. `numericStats` only understands numbers, so offering it for
other types would silently produce an empty view.

Export contract: history = `timestamp,value,unit`; statistics =
`bucket_start,bucket_end,min,average,max,last,samples,unit`. Both follow the **selected** range (zoom
is a view transform), write ISO-8601 timestamps with offset, leave gap and empty-bucket cells empty
rather than `0`, and are disabled while the series is partial or empty.

Interaction (ADR-006): one local Chart.js plugin supplies the gestures, all funnelling into one
`applyRange(from, to)` that clamps to 5 min…31 days and never leaves the loaded window. Gestures are
split by pointer type so no input does two things at once:

| Input | Action |
| --- | --- |
| Mouse drag on the plot | select a range and zoom to it |
| Mouse wheel | zoom around the pointer |
| Single-finger drag (touch) | move the time/value readout only — never zoom (AC-05) |
| Two-finger pinch (touch) | zoom around the midpoint |
| Double-click / double-tap | reset to the selected preset |
| `−  +  Reset` buttons, `+ - 0`, arrows | the same range actions, keyboard and touch parity |

Touch range selection is deliberately not offered by drag: pinch and the buttons cover it, and
reserving one-finger drag for scrubbing avoids the conflict AC-05 describes. `touch-action` keeps all
gestures from scrolling the page. No gesture library is added: `chartjs-plugin-zoom` was rejected
because it pulls `hammerjs` (unmaintained since 2017). Range policy stays in
`domain/history/range.ts`.

Zooming must rebuild the chart options, not just move the axis. `components/chart/history/options.ts`
captures the range in `afterBuildTicks` (ticks at from/mid/to) and in the tick formatter
(`historyTime` chooses time vs date labels from the duration), so changing `scales.x.min/max` alone
would leave ticks outside the new window and labels formatted for the old span. The plugin therefore
emits the new range to `HistoryChart.vue`, which stores it as the view range and recomputes
`historyOptions(...)` with the new `from`/`to`; `HistoryCanvas` applies that options object as it
already does. One source of truth, and tick positions and labels always match the visible window.

Statistics (ADR-006): `domain/history/stats.ts` buckets the **raw decoded** series by local calendar
day, producing `{ start, end, min, avg, max, last, count }` with `null` for empty buckets, reusing
`numericStats` per bucket. The decimated render set is never used, because it keeps extrema but would
skew averages and counts. Buckets outside the covered sample window stay empty; a partial series names
the covered window, warns that newer samples may be missing, and disables export.

Mobile shell: `NAV_ITEMS` gains a `primary` marker; `AppShell` renders `BottomNav` (five items) below
720 px and keeps the drawer. `--nh-safe-bottom` token + `viewport-fit=cover`; the manifest lives in
`public/` and is referenced from `index.html`.

Compatibility: `HistoryView` keeps its props (`series`, `expose`, `compact`) and the sheet keeps a
compact variant; existing e2e selectors (`role="dialog"`, "Refresh history", `.nh-history-canvas`,
sample table) are preserved or mapped, and new behaviour is additive. Nothing in the backend,
transport or store contract changes.

## Constitution Check

Before implementation, PASS cites the supported design; execution evidence stays pending until the
tasks run. Final column is filled before merge.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec records reproduction, ACs and exclusions; ADR-006 records the decisions | PASS | Pending |
| NH-02 | Read-only view of stored samples; no commands, no optimistic state; gaps/unknown preserved (AC-13) | PASS | Pending |
| NH-03 | No new entry point, auth change or household data; only range/expose persisted locally (documented) | PASS | Pending |
| NH-04 | Reuses domain/transport/store/chart boundaries; one new pure module + presentational components; ADR for the dependency decision | PASS | Pending |
| NH-05 | Companion Jest tests per source file, Playwright gestures with real input, recorded manual pass | PASS | Pending |
| NH-06 | No metrics/storage/retention change; bounded work per pointer move; manifest needs no background service | PASS | Pending |
| BE-01 | No backend change; statistics computed client-side from the existing decoded series | N/A | N/A |
| BE-02/03 | No ingestion, lifetime, concurrency or goroutine change | N/A | N/A |
| BE-04 | No query meaning, aggregation or persistence change | N/A | N/A |
| BE-05 | No route, auth or payload change | N/A | N/A |
| FE-01 | History state stays in the existing Vuex module + composables; components remain presentational | PASS | Pending |
| FE-02 | Reuses `contracts/history.ts` decoding; no new payload; preferences validated on read | PASS | Pending |
| FE-03 | Keeps loading/error/empty/partial/offline states; no retries added; no optimistic readings | PASS | Pending |
| FE-04 | Named controls, keyboard operation, focus ring, 44 px targets, mobile/desktop verification | PASS | Pending |
| FE-05 | No secrets; no new authenticated call; only non-sensitive UI preferences stored | PASS | Pending |

## Contracts, data, rollout

- No HTTP/WS/MQTT/MCP/persistence change; the statistics view is derived from data already fetched.
- New browser storage: `nh-history-pref:<deviceId>` =
  `{ version: 1, preset, expose, custom?: { from, to } }`, validated on read (known preset, expose
  present on the device, custom bounds finite, ordered and ≤31 days) with a `24h` + stored-expose
  fallback; unreadable or unknown-version entries are ignored rather than thrown.
- New static files: `public/manifest.webmanifest`, an SVG icon, `theme-color` and
  `viewport-fit=cover` in `index.html`. No service worker in this change.
- Tokens: add `--nh-safe-bottom`; `--nh-bottom-nav-h` starts being used.
- Deployment: UI-only. Rollback restores the previous frontend image; stored preferences are
  additive and ignored by older code, so no migration or cleanup is required.
- Docs: retire the root `charts_todo.md` (superseded), update `docs/` where the history flow is
  described, and mark spec 007 T209–T211 superseded with a pointer to 029 while T212
  (Assistant layout, palette/contrast, legacy cleanup) and T213 stay open.

## Traceability and verification

| IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01/02 | `domain/history/range.ts`, `HistoryToolbar.vue`, `DeviceMetrics.vue` | `src/__tests__/domain/history/range.test.ts`, `src/__tests__/components/chart/history/*` | Pending |
| FR-02, AC-03/10 | `DeviceMetrics.vue`, `HistoryChart.vue`, `AppShell.vue`, `BottomNav.vue` | Playwright at 1280×800, 768×1024 and 390×844; no horizontal scroll assertion | Pending |
| FR-03, AC-04/05/06 | `components/chart/history/plugin-interactions.ts`, `HistoryChart.vue` | Playwright mouse drag, wheel, CDP two-pointer pinch, keyboard; Jest for clamp helpers | Pending |
| FR-04, AC-07/08/09 | `domain/history/stats.ts`, `HistoryStats.vue`, `HistoryView.vue` | Jest DST/empty/partial buckets; Playwright statistics + export | Pending |
| FR-05, AC-11/12 | `EntityViewDialog.vue`, `AppShell.vue`, `index.html`, `public/manifest.webmanifest` | Playwright sheet drag/Escape/safe area; manifest fetch assertion | Pending |
| FR-06, AC-13/14/15 | all touched files | Existing Jest + Playwright suites unchanged, plus budget assertions | Pending |
| NFR-01/02/03 | bundle output, chart component | `npm run build` size delta; interaction timing trace; a11y assertions (focus, names, contrast tokens) | Pending |

Planned checks (to run in the tasks, not yet executed): from `frontend/` —
`npm test -- --runInBand`, `npm run lint`, `npm run build`, `npm run type-check:e2e`,
`CI=true npx playwright test`; from `backend/` — unchanged `go test ./...` (proves no accidental
backend edit). No household device or broker is used; browser tests stay on the mock hub and the
isolated Go fixture.

## Debt, risks, exceptions

- **Gesture complexity on touch** is the main risk. Mitigation: one clamped `applyRange` path, real
  input-event tests, `touch-action` rules, and buttons as the guaranteed fallback.
- **Pinch vs browser zoom** on iOS Safari requires `touch-action: none` on the plot; iOS may still
  intercept some gestures above the page. Mitigation: buttons remain primary, and the manual pass
  records behaviour.
- **Statistics fidelity**: client-side buckets are only as complete as the fetched series. History
  mode is ascending-only and the backend keeps the **oldest** 10 001 samples, so a truncated response
  can lack recent days; the view must name the covered window, keep uncovered buckets empty and
  disable export (AC-07/08). Server-side bucketing with an interval parameter is follow-up debt; if
  long-range statistics become a primary use, that is the point to write it.
- **Layout regressions** from removing the 64 rem cap and adding a bottom bar: covered by viewport
  assertions in existing layout tests.
- **No new dependency**: the interaction plugin is our code in the lazily loaded chart chunk. The
  budget is ≤ 4 kB gzip there and an unchanged initial entry; a Jest assertion pins that the chart
  chunk imports nothing new.
- **Existing debt untouched**: cross-device comparison (contract limits one device per series) and
  automation markers (no run history stored) remain out of scope.
- No exceptions requested; no approval implied.
