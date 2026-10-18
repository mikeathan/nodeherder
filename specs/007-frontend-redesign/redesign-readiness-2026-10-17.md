# Entity and chart redesign: implementation readiness

Date: 2026-10-17 | Status: Investigation and proposed implementation design
Scope at investigation: accepted entity previews and all compact/full charts.
Implementation was subsequently authorised; final changes/checks are recorded in
[implementation evidence](implementation-evidence-2026-10-18.md). The source findings
and proposed gates below describe the pre-implementation checkpoint.
Authority: [spec](spec.md), [plan](plan.md), project/component constitutions and
Go/Vue/integration rules. Constitutions remain proposed, not claimed ratified.

## Agreed product direction

The owner accepted these references, with each image's sample data illustrative:

| Reference | Accepted scope |
| --- | --- |
| [Mobile v3](../../docs/design-previews/entity-views-mobile-concept-v3.png) | Brightness and colour temperature only; temperature panel rejected |
| [Environmental v5](../../docs/design-previews/entity-views-environmental-concept-v5.png) | Temperature, humidity, illuminance, air quality; clean stronger strokes, subdued fills and simple axes |
| [Activity/electricity v4](../../docs/design-previews/entity-views-activity-energy-concept-v4.png) | Presence/contact timelines, power and energy |
| [Diagnostics/settings v4](../../docs/design-previews/entity-views-diagnostics-settings-concept-v4.png) | Diagnostics, alerts, metadata-driven controls and events |

Mobile entity views fill the screen. Adjustable entities prioritise a large control
with an inset grip, confirmed value, directly relevant controls and small four-hour
history. Read-only entities can show a larger chart and compact relevant summaries.
Desktop uses the same data/control ownership with more chart space; adapt the accepted
design language rather than treating the earlier desktop image as a final layout.
All themes, density settings and font scaling must remain usable. Images do not
authorise invented capabilities, thresholds, readings or histories.

Reuse this deferred follow-up in 007; preserve the original merged redesign and
completed backend repairs. Unrelated automation, assistant and deployment work does
not become part of this redesign merely because old 007 tasks mention it. No new
task execution or implementation is authorised by this investigation.

## Source-confirmed gaps and necessary repairs

| Area | Current evidence | Necessary outcome |
| --- | --- | --- |
| Entity dialog | `EntityViewDialog.vue` uses legacy formatting, hides all charts on devices with brightness, and renders unrelated device controls | Select content by the selected expose and supported capabilities; consistent formatting; keep selected-entity history |
| Command flow | `useEntityDialog.ts` dispatches directly; its numeric-zero/binary enablement is unsuitable | Delegate through the shared command path; preserve valid zero, disabled/offline/reconnect restrictions and confirmation |
| Shared controls | `ExposeControl.vue` clears draft only when confirmed data changes, not when sending fails/times out; step is not metadata-derived | Separate editing/requested/confirmed states; reset draft after refusal/timeout; validate bounds/step; no command per pointer movement |
| Semantics | `domain/exposes.ts` already owns many labels/ranges/capabilities, but older chart/formatter paths duplicate them | Extend that owner, including bidirectional raw/display conversions; avoid a second semantic registry |
| Type safety | `types/device.d.ts` contains broad `any` and questionable enum/optional declarations; `metrics.type.ts` says string bounds/binary values whereas Go emits milliseconds/booleans | Narrow touched boundaries with discriminated history types and runtime decoding; fix relevant device types without rewriting all contracts |
| History lifecycle | `DeviceMetrics.vue` stops loading after WS dispatch; `useMiniChartData.ts` has no bounded error completion and its live switch only watches results | Loading follows matching response/error/timeout; define real refresh and cancellation behavior |
| Query isolation | `metrics` Vuex results are keyed by device and new queries delete previous results; WS controller ignores requested expose, mock honours it | Canonical query identity and stale-response protection; production-shaped mocks |
| Numeric rendering | Five mini variants and full charts duplicate transforms/options; every-nth-point downsampling can miss spikes | Shared semantic model and extrema/gap-preserving reduction; one theme/options adapter |
| State rendering | Timeline minimum widths distort durations; sample counts are called changes/triggers; unknown-time denominator not shown | Proportional intervals, distinct reports/events/transitions, visible coverage and correct binary mapping |
| Enum events | `ExposeTimeRangeMetricsResult.Collect` merges equal strings; its `Flush`/finalisation also needs boundary tests | A raw repeated-event representation for action history; category intervals are insufficient evidence for button-event counts |
| Energy | Generic gauge uses a default scale and historic latest value; no bucketed-consumption contract was found in inspected metrics code | Distinguish meter total, measured period total and derived interval use; never present counter readings as interval consumption |
| Backend failures | `QueryService.Query` continues after collector/query errors, including real read failures; existing test explicitly expects ignored errors | Differentiate legitimate no-history from failure; preserve current consumers with compatible semantics |
| Backend resources | Query context is not propagated to repository scans; limit is applied after collection | A browser abort/response cap alone does not bound server work; validate ranges and measure/limit scan work before shipping large ranges |
| Modal lifecycle | `UiDialog` has focus support but phone style is a 92dvh sheet; legacy `useDialogUI` has a separate Escape/resize path | An explicit entity full-screen variant; one close/focus/scroll lifecycle; preserve other dialog layouts |
| Retained resources | `PickSlider.vue` has a timeout without unmount cleanup; chart cells install document listeners; command service owns a permanent singleton subscription | Explicit ownership and disposal rather than assuming component removal cleans everything |

Inspected sources include backend HTTP/WS routes, query/aggregate/limit/storage code,
frontend components/composables/stores, mock server, source-owned tests, installed
chart packages and current CI. These are static findings, not executed regressions.

## Target architecture and single sources of truth

```text
Device WS updates -> existing hub Vuex -> expose/domain selectors -> entity UI
User gesture -> local draft -> useDeviceCommand -> existing deviceSetValue
Device echo -> hub update -> command settlement -> confirmed UI

History view -> useHistoryQuery -> metrics Vuex query entries
                         -> history service -> existing HTTP query
HTTP response -> runtime decoder -> pure history transforms -> typed view model
                                                  -> shared chart/timeline renderer
```

- Confirmed device truth stays in `hub`; history stays in `metrics`. History never
  overwrites current state; dragging never writes a competing device value to the store.
- Transport and handles stay in services/composables, outside serialisable Vuex data.
  Share in-flight requests by key; use subscriber ownership so closing one consumer
  does not cancel another. Bound the queue as well as active requests/cache entries.
- One metadata policy owns labels, units, binary mapping, range, precision, presets,
  permissions and supported controls. Both row and large detail controls consume it.
- One history pipeline owns ordering, window clipping, gaps, coverage, summaries,
  transitions and decimation. Renderers own only dimensions, theme and inspection.
- Use small typed pure functions and explicit variants, not inheritance or a generic
  plugin framework. New modules must remove actual duplication. No Vue/store/chart
  imports in the history calculation modules.

## Change map

Paths below are relative to `frontend/` unless prefixed `backend/`. New names are
proposals to be finalised with the implementation tasks, not files already created.

| Ownership | Existing files to change/reuse | Proposed additions and purpose |
| --- | --- | --- |
| Entity shell | `src/components/dialogs/EntityViewDialog.vue`, `DialogHost.vue`, `src/components/ui/UiDialog.vue`, `src/contracts/dialog-events.ts` | Focused entity content variants under `components/entity/`; explicit full-screen phone variant; idempotent dismissal and route-close behavior |
| Entity model | `src/domain/exposes.ts`, `src/types/device.d.ts`, `src/composables/useEntityDialog.ts` | Only add `domain/entities.ts` if selection/grouping deserves its own owner; selected expose and related controls from explicit capability rules |
| Large and row controls | `src/components/entity/ExposeControl.vue`, `src/composables/useDeviceCommand.ts`, `src/domain/commands.ts` | A reusable presentation-only range component under `components/ui/`; a numeric edit composable if row/detail share substantial draft logic |
| Formatting consumers | `EntityTile.vue`, `ExposeRow.vue`, `DeviceCard.vue`, `components/panel/PanelButton.vue`, `PanelMode.vue`, `modules/formatters/sensor-formatter.ts` | Migrate touched semantics to existing domain owner; preserve deliberate tile/panel quick actions and exact command target |
| History ingress | `src/contracts/api.ts` reuse; `src/types/metrics.type.ts`, `src/contracts/metrics.ts` | `src/contracts/history.ts` decoder/wire types; `src/services/history.service.ts` abort/timeout/error adapter |
| History state | `src/store/modules/metrics/{index,state}.ts`, `src/composables/useMiniChartData.ts` | `src/composables/useHistoryQuery.ts`; keyed statuses/generations, shared requests and bounded memory cache |
| Calculations | `src/utils/chart.utils.ts`, `src/contracts/chart.ts`, `src/utils/date.utils.ts` | `src/domain/history/{range,numeric,state,stats,decimation}.ts`; add energy/event modules only for their specific policies |
| Rendering | `src/components/chart/{BaseChart,NumericChart,NumericChartHeader,BinaryChart,DynamicBinaryChart,DateRangeDisplay}.vue`, `types/AreaChart.vue`, `mini/*.vue`, `src/mixins/useChartComponents.ts` | Shared typed history renderer/options under `components/chart/history/`; compact/detail variants; state/event inspection without HTML-string tooltip assembly |
| Full history | `src/components/device/DeviceMetrics.vue`, period types/contracts | Reuse query/model/renderer from entity histories; full-view range controls and accessible sample inspection |
| Appearance | `src/assets/styles/{tokens,themes,base}.css`, theme adapter as needed | One chart token mapping, theme-reactive options; avoid hard-coded dark colours in components |
| Backend history correctness | `backend/internal/metrics/{services/query_service.go,query/*.go,domain/device.go,storage/metrics.go}`, `backend/store/store.go`, `backend/internal/api/routes.go` | Compatible query validation/errors, event representation, bounded query execution; consumer interfaces only where needed |
| Mock parity | `tools/server/{routes,metrics-mock,state,ws}.mjs`, `e2e/support/` | Add real `POST /api/metrics/query` shape and deterministic failure/boundary fixtures; existing mock has GET/WS history instead |

Retire `useDialogUI`, `useEntityComponents`, legacy `BaseSlider`/`FillSlider`/
`PickSlider` and superseded chart code only after all actual consumers migrate and
import scans pass. Do not maintain two transports or calculation paths indefinitely.
Do not delete old WS handlers while a remaining consumer uses them.

## Controls: exact behavior to implement

Brightness displays percent but sends native bounds, e.g. 72% of a 0–254 range is
about 183; the rounded integer/raw value is the command, not `72` by assumption.
Colour-temperature metadata determines whether the native unit is mired or Kelvin:
convert only when appropriate, with inverse direction and correctly rounded steps.
The current formatter assumes mired for every `color_temp`; that needs fixing.
Validate min < max, finite values, step and presets; an invalid range must not render
a broken slider or send a command. Binary values use expose mappings, including
inverted contacts. Unknown binary values must not silently become the safe/off side.

During drag, keep a local preview. Commit once on completed gesture/keyboard change;
cancelled gestures restore the last confirmed state and send nothing. Reconcile
incoming readings without snapping the thumb mid-drag. Preserve existing refusal of
a second same-expose command while pending; no hidden command queue or retries.
Refused, failed, timed-out and disconnected edits return to confirmed state with
clear feedback. Closing a view cancels its draft, not a command already sent.
Pending command ownership must survive an ordinary view close but clear on session
reset/logout/device removal as appropriate, with bounded timers and no stale toast.

Do not rewrite MQTT/automation handling or the completed live-publication repair.
Verify target/value payloads and confirmations with mock echoes, including the dial
regression path, before claiming controls are reliable.

## History contracts, gaps and backend work

Retain the existing proposal to use `POST /api/metrics/query` through `fetchWithAuth`
for new histories. Request uses deviceIds, exposes, RFC3339 UTC bounds, aggregation,
filters, limit and sortDesc. Response is expose groups with millisecond query bounds,
device identity, unit and raw `values[].value`. Legacy WS bounds are seconds and
raw result bounds milliseconds; do not mix these shapes. Decode `unknown` at ingress.

This existing route is currently registered publicly in `backend/internal/hub.go`.
Reuse does not establish authentication or permit broader exposure; preserve current
access and avoid adding household logs/browser persistence. HTTP errors must remain
useful without surfacing internal storage details. Verify API and MCP consumers of
the query service before adjusting errors or optional fields.

Three preparations are needed before complete history integration:

1. **Query correctness and resources:** capture producer fixtures; distinguish a
   known absent bucket from real read failure; validate time ordering/range/limits;
   define context propagation and scan/work bounds with the repository boundary.
   Response truncation is applied after collection today. Check DB/tail overlap and
   inclusive `[from,to]` bounds, then use deterministic clipping/ordering. Treat
   aborted frontend requests as abandoned results until server cancellation is proven.
2. **Repeated event history:** the current enum collector is state-interval oriented
   and merges equal strings. Propose an opt-in raw-event query representation that
   retains every stored action report and timestamp, while leaving legacy interval
   consumers compatible. Document the exact additive wire shape in ADR-004 before
   implementation. Existing merged responses cannot reconstruct lost repeats.
   Regression-test collector finalisation, including constant enum values, before
   relying on category history. This does not require changing automation event input.
3. **Energy semantics:** use an existing explicitly period-scoped reading where
   available. For cumulative-only devices, define interval counter deltas separately:
   unit conversion, timezone bucket bounds, baseline at period start, resets and
   unavailable intervals. Unknown baseline/gap/reset must not become zero or negative
   use. Never infer kWh by merely relabelling a W plot. The prior plan excluded inferred
   consumption; accepted bars require an explicit scoped amendment/ADR policy before
   dependent code. Meter-total history remains useful until this contract is proven.

No storage migration is proposed. Read existing stored samples; any newly required
representation is additive. If producer investigation shows missing retained events
or baselines, show unavailable/partial history and record the limitation rather than
fabricating the preview. New retention/schema behavior needs a separate decision.

Freshness must say device seen when sourced from `device.last_seen`; a true per-expose
reading timestamp is not present in the inspected device type. A recorded sample's
timestamp is not automatically the current reading's timestamp. Do not invent staleness
thresholds without reporting metadata. Preserve measured zero and unknown gaps.

## Chart library and appearance

Installed sources: ApexCharts **4.7.0**, vue3-apexcharts **1.8.0**. Installed types
support `monotoneCubic`, straight and step curves; the wrapper calls destroy on
unmount. The library is not Chart.js. [Official stroke options](https://apexcharts.com/docs/options/stroke/)
and [lifecycle methods](https://apexcharts.com/docs/methods/) support the proposed
adapter direction, but latest online documentation is not proof for every installed
version feature; verify against installed implementation/types and the prototype.

Recommend retaining ApexCharts behind one narrow adapter for a mock-backed visual
and interaction spike. Establish the environmental v5 stroke/fill/grid/axis treatment
first, and compare the built chart with the accepted image. Use monotone interpolation
only where it preserves observed bounds and appropriate meaning; power spikes and
settings/state histories need different treatment. Pretty rendering cannot justify
deleting real variation or bridging gaps. Explicitly labelled aggregation is different
from changing raw measurements. Summaries use validated source data, not rendered points.

Use a separate proportional state-timeline renderer and event list/table where a
generic library is unsuitable. Inspection works with tap/keyboard and stays inside
the viewport; no string-interpolated household HTML tooltips. Shared compact/detail
renderers have different density, not different statistical meaning.

Replace the package only if the installed-version spike fails agreed appearance,
mobile inspection, accessible alternatives, lifecycle or measured performance.
Then record alternatives, bundle cost and migration in ADR-004; do not ship two
chart packages as a permanent solution. No package install/upgrade has been done.

## Performance and leak prevention

Carry forward 007's proposed client budgets, subject to actual benchmark evidence:
one active history request per client with a bounded/deduplicated queue; 300ms range
debounce; 10s timeout; optional visible refresh at least 10s apart; at most 10 cache
entries; probe 10,001 samples and retain at most 10,000 per series with partial status;
at most 1,000 rendered numeric points preserving extrema/endpoints/gaps. Also cap
series per query/view so entry limits do not conceal unbounded sample totals.

Pure normalisation/stats/decimation for 10,000 samples: proposed p95 <=100ms over
20 warmed runs on a recorded runner. Record payload size, render time, long tasks
and bundle delta; these budgets are not measured results. Mount heavy full-history
charts only when visible/expanded; avoid deep-watching the entire metrics map or
rebuilding chart instances on every device report. Coalesce presentation updates
without changing automation/device event streams. Do not create per-tile polling.

| Resource | Owner and required cleanup |
| --- | --- |
| HTTP controller/deadline | history service request handle; clear timeout on all terminal paths; abort when last subscriber leaves |
| Query entry/queue | metrics owner; invalidate generation on supersede/reset; evict bounded unused data and queued requests; clear on logout |
| Refresh/visibility/resize | view/query subscription; pause while hidden/offline, unsubscribe on close/route change; one shared resize observer only if needed |
| Chart instances | renderer; wrapper destroy plus verification; no manual double-destroy; discard tooltip/observer references |
| Slider gesture/marker | control; release capture/listeners and timers on cancellation/unmount; clear drafts on identity change |
| Pending command/subscription | app command owner; one subscription, explicit reset/dispose integration, bounded deadline, no forgotten state after logout |
| Dialog scroll/focus | shared modal owner; restore body scroll and opener only when still valid, close once, no duplicate Escape handlers |

Verify 50 repeated open/change/close cycles after warmup: no retained entity chart
instances, growth in owned listeners/timers, pending history handles or queued work;
record retained heap trend with controlled GC where available. Count application-owned
resources as well as browser heap. Test late responses after close and logout. No
claim of zero leaks or zero bugs is justified before those checks.

## Verification and clean-code gates

Existing Jest transforms TS, not Vue SFC mounting. Existing Playwright uses isolated
mock hub port 4110 and Vite 4100, desktop 1280x800 and phone 390x844. Extend these;
never test routine controls on household devices. Tests must use source-owned names:

- Extend `src/__tests__/domain/{exposes,commands}.test.ts`,
  `composables/useDeviceCommand.test.ts` and `utils/chart.utils.test.ts` as relevant.
  New `domain/history/numeric.ts` gets `src/__tests__/domain/history/numeric.test.ts`,
  service/composable tests follow their source paths. Do not add issue-named tests.
- Add meaningful Vue interaction coverage if required with a compatible pinned Jest
  mounting/transform setup; document dependency cost. Browser specs should belong
  to production owners, e.g. `EntityViewDialog.spec.ts`, `DeviceMetrics.spec.ts`.
- Backend regressions belong to `query_service_test.go`, `request_test.go`,
  `aggregates_test.go`, `metrics_test.go`, `routes_test.go`; collector tests belong
  to `domain/device_test.go` if a new companion is needed. Update existing fixtures
  and other consumers when additive contracts change.

Carry the earlier *proposed* NFR-07 gates: changed functions C<=10, CRAP<=30,
>=80% branch coverage for touched decision logic; establish pinned tooling and
Vue source mappings before claiming results. No CRAP runner is configured in the
inspected frontend scripts/CI. Do not reduce this to TS-only utility coverage or
inflate tests that mirror implementation. Record any inability to measure it as
unresolved work, not a passed gate. Constitution checks remain separate from scores.

Required cases: min/max/zero/missing/invalid metadata, native brightness/temperature
conversion, decimal steps, inverted/unknown binary values, pending/refused/timeout,
cancelled gestures, concurrent WS updates during drag, offline/disabled/disconnect,
removed entities, stale responses, malformed/empty/partial/error/429 histories,
repeated equal events, counter reset/missing baseline, sparse/duplicate timestamps,
spikes/gaps, DST/midnight, full/mini consistency, theme changes and route/logout cleanup.

Viewport evidence: 320/360/390 px phones, short landscape, 768 px tablet,
1280/1920 px desktop, zoom/font scaling. Verify full-screen mobile, large controls,
44px touch targets, keyboard/focus return, readable contrast in light/dark, reduced
motion, chart inspection and no modal interference with Panel room navigation.

Planned commands: frontend `npm test -- --runInBand`, `npm run lint`, `npm run build`,
`npm run test:e2e` with Node 24; scoped Go tests/race while iterating, then backend
`go test ./...`, `go vet ./...`, `go build ./...` on go.mod toolchain. Record actual
commands/results, expected mock ports and compiler/package versions. Planning-only
investigation ran none of these application checks.

## Delivery order and unresolved preparation

1. Amend 007 acceptance/ADR-004/contracts for the accepted visual references, full-screen
   mobile, scoped event/energy behavior, bounds and lifecycle; update only relevant
   T200–T214 dependencies rather than activating unrelated old tasks.
2. Capture backend-shaped fixtures, establish quality baseline and create the actual
   shared renderer/control spike with mocks. Validate desktop adaptation as part of
   this work, not another generated image assumed to be executable evidence.
3. Fix scoped query/error/event prerequisites and domain semantics; implement shared
   query lifecycle and command-draft reconciliation with regressions.
4. Build shared controls/renderers; migrate entity dialog and full device history;
   check tile/panel/device-row compatibility. Remove superseded paths after import scans.
5. Run acceptance, accessibility, lifecycle, performance and component gates; document
   results and remaining limitations. Keep wire changes additive and roll frontend
   back independently of stored data if required.

Accepted images resolve visual direction, not exact backend contracts. Remaining
preparation decisions are raw-event representation, consumption/reset/baseline policy,
server scan/cancellation bounds and proof of installed renderer quality. These can be
worked out in code/fixtures and the SDD artifacts; no further general permission is
needed to prepare them. Dependent implementation must wait for those contracts to be
specified. Current task ends at this investigation/report.

## Rule check and actual evidence

| Rule | Design disposition | Execution evidence |
| --- | --- | --- |
| NH-01 | Existing 007 reused; accepted references recorded; contract amendments identified before production edits | Investigation only |
| NH-02 | Confirmed/draft/pending distinct; exact native commands; no automation/input throttling changes | Regression execution pending |
| NH-03 | Existing ingress/access helpers; no new route/exposure or household persistence/logging | Runtime decoder/access review pending |
| NH-04 | Existing Vuex/service/domain/repository boundaries; additive event contract/ADR preparation | Contract amendment pending |
| NH-05 | Source-owned unit/browser/Go checks and acceptance cases specified | Not run |
| NH-06 | Units/time/unknown/reset policies, bounded ownership, additive storage-compatible approach | Energy/server-budget definitions pending |
| FE-01 | One hub truth, metrics query owner and presentation-only components | Implementation pending |
| FE-02 | Explicit HTTP vs WS decoder types and producer-shaped mocks | Producer fixture execution pending |
| FE-03 | Query/command loading, failure and confirmation states | Implementation pending |
| FE-04 | Shared tokens, full-screen variant, keyboard/touch/viewports/contrast checks | Browser evidence pending |
| FE-05 | Existing API helper; no untrusted HTML tooltip rendering/secrets/history persistence | Implementation review pending |
| BE-01 | Query logic stays service/domain/repository; handlers adapt/validate | Implementation pending |
| BE-02 | Preserve completed live publication and repeated automation events; raw history change scoped separately | Compatibility regressions pending |
| BE-03 | Context/scan ownership and concurrency/race checks identified | Bounds design and execution pending |
| BE-04 | Query ordering/window/limits/error/event semantics require producer fixtures; no migration assumed | Contract preparation pending |
| BE-05 | Document current public route and existing API/MCP consumers; compatible responses | Consumer validation pending |

No approval, ratification, performance result or completed implementation is claimed.
