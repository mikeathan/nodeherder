# Plan: History, charts and mobile shell

Spec: [spec.md](spec.md) | ADR: [adr-006](adr-006-chart-interactions.md) | Status: Implemented; manual verification pending
Author: NodeHerder contributors | Updated: 2026-10-23
Constitutions: parent proposed 1.0.0; frontend 1.0.0; backend 1.0.0 (no backend change in scope)

## Context and design

Review correction (2026-10-23): omit an empty midnight-end bucket but retain actual inclusive-end
samples (including null reports). Daily statistics use a floating min/max bar and average/last
markers with an HTML legend and named tooltips; energy/event bars keep their baseline. Jest global
setup pins Europe/London before workers start so DST assertions are unconditional, including when
launched from UTC. Native-unit history export gets visible help; statistics export keeps displayed
units. Geometry moves to scoped CSS and an inherited height variable. The shell loader becomes a
viewport-sized themed skeleton, theme-colour metadata follows `--nh-bg`, and mouse leave/reset clear
the crosshair without clearing the touch readout on finger release. No dependency, transport,
backend, retention or device-command change. Reopen T011/T014–T017/T019/T022/T030/T031/T033; verify
midnight endpoints, UTC-launched DST tests, range/marker encoding, units/help, scoped geometry,
delayed loading and theme updates in existing companion Jest/browser files. Full frontend checks
and isolated browser regression remain gates.
The midnight browser regression also verifies refresh keeps Statistics selected while a request is
loading; only an actually decoded incompatible series switches the mode back to History.

Implementation refinement (2026-10-22): the signed-in `AppShell` will be loaded asynchronously
from `App.vue`, with a named loading status, so the new mobile navigation does not increase the
initial entry payload. Chart.js and interaction code remain deferred. The five-minute zoom minimum
is bounded by the fetched duration for a valid shorter Custom/early-Today range; zoom never widens
the query. Statistics convert all raw samples to display units before aggregation (especially
mired → kelvin); statistics CSV uses those same values/units, while raw history CSV keeps native units.

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
**intersecting local calendar day**, enumerated rather than capped. The view must not hard-code a
bucket count: iterate local midnights from the start day while the bucket start is before the range
end, also retaining a bucket starting at the end when it owns a real inclusive-end report. Step with
`setDate(+1)` then `setHours(0,0,0,0)` so each step is DST-correct.

The count is larger than "31 days ⇒ 31 or 32 dates" suggests. Both endpoints can be mid-day (+2) and a
DST transition moves a local midnight, so a 31-day (744 h) range can intersect **33** local dates —
verified in `Europe/London`: 1 Mar 2026 23:30 GMT → 2 Apr 2026 00:30 BST is exactly 744 h and touches
1 Mar…2 Apr. The widest possible count follows from midnights being at least 23 h apart:
`floor(744 / 23) = 32` midnights plus the start date, so ≤ 33. Treat that as a derived bound for
sizing/tests, never as a `MAX_BUCKETS` constant.

Bucket ownership is `[local midnight, next local midnight)`: a sample at exactly midnight belongs to
the day starting there, and the first and last buckets are marked partial when the range does not
start or end at midnight, so samples are neither dropped nor counted twice. The chart keeps `decimate`; a
`partial` series must show which window it covers, keep uncovered buckets empty, refuse export, and
offer a narrower range instead of silently changing the user's selection.

Design — one shared history stack, three presentations:

```
domain/history/          pure: range presets, decimation, stats, bucketing, CSV rows
services + store         unchanged transport and query identity
components/chart/history/
  HistoryChart.vue       owns view range, gestures, buttons and a11y; HistoryCanvas owns Chart.js
  HistoryToolbar.vue     presets, expose chips, mode switch, actions
  HistoryStats.vue       sample summary, daily range/markers and accessible statistics table
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
exposes `aria-expanded` for the drawer. `NAV_ITEMS` gains a `primary` marker; CSS filters primary links from the phone drawer, so there is still one navigation source.

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

Before implementation, PASS cites the supported design. The verification record below names actual
execution results; pending manual checks are not implied by an automated pass.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec records reproduction, ACs and exclusions; ADR-006 records the decisions | PASS | Code/automated evidence below |
| NH-02 | Read-only view of stored samples; no commands, no optimistic state; gaps/unknown preserved (AC-13) | PASS | Code/automated evidence below |
| NH-03 | No new entry point, auth change or household data; only range/expose persisted locally (documented) | PASS | Code/automated evidence below |
| NH-04 | Reuses domain/transport/store/chart boundaries; pure range/statistics/export/preferences modules + presentation components; ADR for the dependency decision | PASS | Code/automated evidence below |
| NH-05 | Companion Jest for domain/interaction helpers, real-input Playwright for Vue wiring; physical manual pass pending | PASS | Automated evidence below; physical manual pass pending |
| NH-06 | No metrics/storage/retention change; bounded work per pointer move; manifest needs no background service | PASS | Code/automated evidence below |
| BE-01 | No backend change; statistics computed client-side from the existing decoded series | N/A | N/A |
| BE-02/03 | No ingestion, lifetime, concurrency or goroutine change | N/A | N/A |
| BE-04 | No query meaning, aggregation or persistence change | N/A | N/A |
| BE-05 | No route, auth or payload change | N/A | N/A |
| FE-01 | History state stays in the existing Vuex module + composables; components remain presentational | PASS | Code/automated evidence below |
| FE-02 | Reuses `contracts/history.ts` decoding; no new payload; preferences validated on read | PASS | Code/automated evidence below |
| FE-03 | Keeps loading/error/empty/partial/offline states; no retries added; no optimistic readings | PASS | Code/automated evidence below |
| FE-04 | Named controls, keyboard operation, focus ring, 44 px targets, mobile/desktop verification | PASS | Automated evidence below; physical manual pass pending |
| FE-05 | No secrets; no new authenticated call; only non-sensitive UI preferences stored | PASS | Code/automated evidence below |

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
| FR-01, AC-01/02 | `domain/history/{range,preferences}.ts`, `HistoryToolbar.vue`, `DeviceMetrics.vue` | Companion Jest; DeviceMetrics browser reload/custom/Yesterday flows | Passed (London TZ) |
| FR-02, AC-03/10 | `DeviceMetrics.vue`, `HistoryChart.vue`, `AppShell.vue`, `BottomNav.vue` | Both-theme viewport, chart-height and overflow assertions at 1280/768/390 widths; mobile bar targets and safe padding | Passed in Chromium (both themes) |
| FR-03, AC-04/05/06 | `components/chart/history/interactions.ts`, `HistoryChart.vue` | Real mouse/wheel/keyboard, CDP horizontal/vertical pinch and one-finger readout; no extra query or command | Passed in Chromium; Safari manual pending |
| FR-04, AC-07/08/09 | `domain/history/{stats,export}.ts`, `HistoryStats.vue`, `HistoryView.vue` | London 33-day case, missing/zero/partial/decimation/unit conversion; both downloaded CSV schemas | Passed |
| FR-05, AC-11/12 | `EntityViewDialog.vue`, `UiDialog.vue`, `AppShell.vue`, manifest | Sheet drag and existing Escape/close tests, expose-vs-preference deep link, safe padding and manifest-404 boot | Passed in Chromium |
| FR-06, AC-13/14/15 | all touched files | Existing suites plus 10 000-sample truncation, ≤1 000 rendered points, sample paging, no gesture requests/commands | Passed; one RAF and binary-search readout bound pointer work |
| NFR-01/02/03 | bundle output, chart component | Build delta; keyboard, names, focus, targets and both-theme screenshots | Bundle/automated a11y passed; real-phone frame timing and physical Safari/safe-area checks unrun |

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
  budget is ≤ 4 kB gzip for the lazy adapter/interaction chunk, with no initial-entry growth.
  Build output and the unchanged dependency lists/lockfile verify the dependency and loading
  boundaries; package.json only adds a Jest timezone setup hook.
- **Existing debt untouched**: cross-device comparison (contract limits one device per series) and
  automation markers (no run history stored) remain out of scope.
- Deferred ideas from the retired chart backlog: alternate chart types, JSON/image export,
  comparison across periods, moving averages/trends/anomalies, threshold overlays, live streaming,
  cursor pagination, cache/retention resampling, and worker-based progressive rendering. These are
  future work, not new performance promises or acceptance requirements for 029.
- No exceptions requested; no approval implied.

## Execution and acceptance record — 2026-10-22

Implementation branch: `spec/029-review-fixes`; no backend, API request shape, dependency or lockfile
change. Spec/plan/task reviews preceded production edits; no maintainer approval or ratification is
claimed. Baseline was inspected at `f9cc6e62`: 235/105px chart, fixed ticks, no interaction plugin,
no Yesterday, and no full-history link in the sheet. Baseline browser inventory: 188 tests.
Baseline build passed. No baseline screenshots were captured before edits; new browser artifacts
show the resulting behaviour rather than purporting to demonstrate the old UI.

| Check | Actual outcome |
| --- | --- |
| Frontend `TZ=Europe/London npm test -- --runInBand` | 53 suites, 338 tests passed; includes the 33-date DST case, native/display conversion, malformed date bounds, null/zero/partial, decimation divergence and CSV guards |
| Frontend `npm run lint` | Passed |
| Frontend `npm run build` | Passed with Node 24; existing DeviceCard auto-import naming warning remains |
| Frontend `npm run type-check:e2e` | Passed |
| Frontend `CI=true npm run test:e2e` | 193 passed, 9 pre-existing desktop skips for phone-only cases; no failures (202 cases) |
| Final focused `CI=true npm run test:e2e -- DeviceMetrics.spec.ts layout.spec.ts` | 18 passed, no skips, after the resize fix and final preference validation; includes both themes and 1280/768/390 widths, horizontal/vertical pinch, scrubbing, keyboard/focus, paging and ≤1,000 rendered points |
| Statistics visual/export follow-up (`DeviceMetrics.spec.ts --grep 'daily statistics'`) | 2 passed, desktop/mobile; daily bars use calendar-date labels and both CSV downloads retain the specified schemas |
| Backend `go test ./...` | Passed with isolated test listeners/cache access; no backend diff |
| `git diff --check` and transport/dependency diff inspection | Passed; history request contract and package/lockfile unchanged |

The initial sandbox-only browser/backend attempts failed because local test listeners and the Go
cache were restricted. Authorized reruns passed. The new resize assertion first exposed a transient
canvas overflow; the canvas now has `max-width:100%`, and geometry assertions wait for Chart.js's
ResizeObserver. The focused rerun passed after that correction. The local interaction proof uses
the real isolated fixture chart instead of a disposable prototype, so the proof remains executable.

Bundle evidence (gzip): initial entry **141.46 → 115.96 kB**; Chart.js chunk **71.35 → 71.35 kB**;
HistoryCanvas adapter/interaction chunk **0.93 → 2.19 kB** (+1.26 kB, within +4 kB). The signed-in
shell and mobile bar are deferred. No chart dependency becomes eager and no package is added.

Visual/input evidence: real browser input checks and generated screenshots under
`frontend/test-results/`; desktop light, phone dark and desktop/mobile daily-statistics screenshots were inspected. Keyboard
focus, accessible control names, 44px targets, overflow and a synthetic 20px safe inset are asserted.
The samples table and daily table supply textual alternatives to the canvases. Unit tests remain
companion files; Vue wiring uses the existing browser runner rather than adding mounting tooling.

**Still unrun:** physical iOS Safari pinch/cancellation, real cutout safe areas, measured frame timing
on a mid-range phone, and a human desktop/mobile acceptance pass. Chromium/CDP emulation and visual
inspection do not establish those results or full WCAG conformance. T031/T033 remain open for that
manual evidence; the implementation is not represented as fully verified for merge.

## CI follow-up — deferred-shell sign-out readiness

The mobile shard 4/4 run on `af764ca4` timed out in the existing browser-default-view sign-out flow.
The shared `e2e/support/ui.ts` helper checks menu visibility once: while the deferred AppShell is
loading it can see no menu, skip opening the phone drawer, then wait forever for its hidden Sign out
button. Reproduce with an explicitly gated AppShell module in the existing `panel.spec.ts` companion
flow; make the shared helper await the attached navigation, open it only if closed, and await its
visibility. Preserve desktop sidebar behavior. This is test readiness, with no production or protocol
change. Reopen T031/T033 for the regression, both-project checks, the reported mobile shard and CI.

Evidence: the gated mobile regression failed with the exact hidden-Sign-out locator timeout before
the helper fix. Afterward, the three browser-default-view flows passed on both projects (6/6), and
the exact reported command `npm run test:e2e -- --project=mobile --shard=4/4 --reporter=list,html,json`
passed all 25 tests with zero unexpected/skipped results. `npm run lint`, `npm run type-check:e2e`
and `git diff --check` passed. No timeout increase or retry was added; the 10-second override was
used only to capture the pre-fix failure quickly. Other application suites were not rerun for this
test-only correction. GitHub CI remains a separate check after push.

## Review corrections — 2026-10-23

All seven numbered review findings are addressed:

- Midnight-end buckets omit the phantom day but preserve actual inclusive-end reports, including
  zero/null reports. The new companion regression failed before the fix (two buckets instead of
  one), then passed; the browser case also checks sample totals and refresh preserving Statistics.
- Daily statistics use floating minimum–maximum bars with average circles and last diamonds,
  distinct kind colours, a text legend/table and named tooltips. The axis fits measured values;
  energy/event baselines remain unchanged. The resulting desktop screenshot was inspected.
- Jest global setup selects London before workers; spring (23 hours), autumn (25 hours) and the
  33-date assertions are unconditional. A full test run launched with `TZ=UTC` passed all 53 suites
  and 341 tests, proving CI does not silently skip that coverage.
- History CSV retains its specified native values/units with visible help; Statistics help names
  displayed units. Both schemas/downloads remain covered by the existing browser flow.
- Canvas rules are scoped; the interactive wrapper owns a responsive inherited height variable,
  and other canvases honour their height prop. Both-theme viewport/overflow assertions pass.
- Deferred shell loading uses a themed viewport-sized skeleton and accessible loading status;
  the gated shell regression checks its geometry on desktop/mobile.
- Browser theme colour follows the active background and survives light/dark toggles and reload;
  the manifest colour remains a launch fallback.

Mouse leave/reset also clear stale crosshairs/readouts while touch release retains the reading.
Queued wheel deltas now have an intent comment. Optional config-expose ordering, the minor bucket
count allocation, and the unreachable empty samples-page label remain outside this correction.

Checks: lint passed (seven existing JS warnings), production build passed, e2e TypeScript passed,
and the focused desktop/mobile run passed 12/12. Corrected deferred-shell/default-view checks
passed 6/6. The final full browser rerun passed **201 tests, nine existing desktop phone-only skips,
zero failures (210 cases)** in 4.8 minutes. The initial build caught a Chart.js scale-options typing error; it was corrected before
these passing checks. No backend, request-contract, dependency-list or lockfile diff; backend tests
were not rerun for this frontend correction (the prior isolated backend run passed).

The first full browser run passed 199 with nine existing skips and two deferred-shell test failures:
Vite's `AppShell.vue?t=…` module URL bypassed the exact-string gate. Match its pathname (including
query variants) and assert the gated request arrived before checking the skeleton/sign-out flow.
This retains controlled loading rather than relying on an arbitrary delay. Rerun both projects and
the full browser suite for that shared regression.

Updated bundle (gzip): initial entry 116.20 kB (baseline 141.46); Chart.js unchanged at 71.35 kB;
lazy canvas/interaction adapter 2.40 kB (baseline 0.93, +1.47 kB within the +4 kB budget).
NH-01…NH-06 and FE-01…FE-05 were checked against the final changes: domain ownership, missing/zero
semantics, inclusive endpoints, unit disclosure, accessible legend/status, scoped styling, cleanup,
and unchanged transport/device-command boundaries are preserved. Physical Safari, cutout safe areas,
phone frame timing and human acceptance remain unrun; T031/T033 stay open.

Final verification: `git diff --check` passed. All reopened implementation/component-check tasks
are complete; T031/T033 remain open only for the physical/manual acceptance evidence above.

CI follow-up on `59376841`: frontend unit/lint/build, backend ordinary/race and the originally
reported browser shard 4 passed. Both shard-3 jobs exposed a readiness error in the new theme test:
after reload the HTML/meta were ready before the deferred shell's toggle mounted, so `count()`
skipped the dark-mode click. Wait for the toggle on each iteration and deliberately gate shell
loading on reload to keep the slow-loading case covered. This is a browser-test correction only;
verify both complete shard-3 project runs, e2e types and lint before pushing the follow-up.

Follow-up validation: exact desktop shard 3/4 passed 10/10; exact mobile shard 3/4 passed 10/10,
with controlled loading on both theme reloads and no retries. Lint (seven existing warnings), e2e
type checks and `git diff --check` passed. Production code is unchanged from the 201-pass full run.
