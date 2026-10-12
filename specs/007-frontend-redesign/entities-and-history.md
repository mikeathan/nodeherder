# Entity and history follow-up design

Status: Proposed | Updated: 2026-10-11 | [Spec](spec.md) | [Plan](plan.md)

This is the complete entity/history rework requested after #54, not a colour-only chart
pass. Proposed product details and thresholds are explicit assumptions to validate in
mocked prototypes. No household device commands are needed to design or test it.

## Current paths and gaps

| Area | Existing paths | Observed gap |
| --- | --- | --- |
| Entity surfaces | `frontend/src/components/entity/{EntityTile,DeviceCard,ExposeRow,ExposeControl}.vue`, `components/panel/{PanelButton,PanelMode}.vue` | Shared domain exists but legacy detail controls bypass `useDeviceCommand`; metadata and binary semantics diverge |
| Entity details | `components/dialogs/EntityViewDialog.vue`, `composables/useEntityDialog.ts`, `mixins/useEntityComponents.ts` | Hard-coded colours, separate formatter/control lookup; chart hidden when brightness exists; zero can disable numeric controls |
| Device history | `components/device/DeviceMetrics.vue`, `components/chart/` | Loading follows dispatch rather than response; full/mini logic and palette duplicated |
| Query state | `store/modules/metrics/`, `composables/useMiniChartData.ts` | Device-only keys, competing query results, unbounded pending state; live switch watches results rather than providing defined refresh |
| Chart logic | `utils/chart.utils.ts`, `contracts/{chart,metrics}.ts`, `types/{metrics,chart}.type.ts` | Mixed transforms/options/HTML; wire types differ from Go; decimation can miss spikes; binary sample/transition counts differ |
| Backend | `internal/api/routes.go`, `internal/metrics/{services,query,domain,storage}/`, `store/store.go` | Reuse HTTP raw query; verify errors, bounds, units, limits against code and fixtures |

Paths above are frontend-relative unless marked backend. Legacy code can be removed only
after full/mini/device/assistant consumers have migrated and import scans pass.

## Product experience

- Home entities stay glanceable: label, device, value/unit, freshness, meaningful status,
  one explicit primary action and entry to details. Read-only entities never acquire an
  unrelated device toggle. Binary controls identify the exact expose being commanded.
- Device rows/cards and Panel reuse the same capability decisions. Details show the current
  confirmed value first, explicit pending commands, shared controls, and history below.
  On phones, details is a sheet; desktop uses a roomy dialog with focus containment/return.
- Full history gets a range toolbar, entity selector, timezone label, numeric summary or
  state/event summary, chart, accessible sample table and explicit loading/error/empty/
  partial/stale states. Presets: 1h/6h/24h/7d/30d, Today/Yesterday and a custom range.
  Default display zone is the browser's named zone; UTC is selectable. Calendar ranges
  use that zone and are DST-aware; query instants are UTC.
- Numeric history uses a clear line/area, truthful gaps, exact-value inspection and zoom/
  reset with keyboard alternatives. Avoid smoothing that implies unmeasured extrema.
- Binary history uses time-proportional state segments and explicit unknown time before
  the first report. Do not extend state across explicit missing periods. Repeated stored
  reports remain in the sample/event table even if adjacent equal segments merge visually.
- Dense event data may use a labelled density view derived from the same events. Event
  count, transition count, sample mean and known-state duration are distinct measures.
  Enum/event history falls back to an event list; unknown series explains its limitation.
- Mini charts use the same range/model/theme/stats rules with fewer labels and open full
  history. Brightness controls do not hide a sensor's history. Current device readings
  are never silently appended as stored measurements.

## Ownership and patterns

| Concern | Single owner / proposed paths | Responsibility |
| --- | --- | --- |
| Confirmed device state | existing Vuex `hub` module | Device/expose/config truth; no competing entity store |
| Entity presentation | extend `src/domain/exposes.ts`; new `domain/entities.ts` | Identity, metadata-first label/unit, exact binary mapping, capability, freshness and fallback |
| Commands | `composables/useDeviceCommand.ts` | Shared pending/confirmed/refused/timeout state; all entity/dialog controls delegate here |
| History transport | new `services/history.service.ts`, `contracts/history.ts` | API helper, decoder, timeout/cancel and useful error mapping |
| Query state | existing `store/modules/metrics/`, new `composables/useHistoryQuery.ts` | Canonical query entries, generation/status, bounded cache, view subscriptions |
| Time/series logic | new `domain/history/{range,numeric,state,stats,decimation}.ts` | Pure typed transformations; metadata semantics, bounds, gaps and statistics |
| Rendering | new `components/chart/history/`, shared chart adapter/options | Typed view-model to ApexCharts or accessible timeline/table; no transport/business rules |
| Theme | existing tokens + one chart-token/options mapping | Reactive light/dark/accent; remove duplicated hex colours and formatter maps |

Use a small capability/renderer strategy registry, pure transformation pipeline and service
adapter where they remove actual duplication. Compose functions and typed discriminated
unions; components own presentation, not query or device rules. SOLID means focused
responsibilities and narrow consumer interfaces, not inheritance or a universal chart
framework. Domain modules never import Vue/store/ApexCharts. Full/mini share the same
semantic model; dimensions/options are presentation variants.

## Query and sample semantics

Preferred existing HTTP request:

```json
{"deviceIds":["fixture-sensor"],"exposes":["illuminance"],"time":{"from":"2026-10-11T10:00:00Z","to":"2026-10-11T11:00:00Z"},"aggregation":"","filters":[],"limit":10001,"sortDesc":false}
```

Response is an array of `{expose, query_start_timestamp, query_end_timestamp, values}`;
`values[]` contains `device_id`, `unit`, `value`, optional timestamp/formatted_time.
For raw queries, `value` carries samples. Capture exact numeric/binary/enum fixtures from
`query/aggregates.go` and the producer before defining the decoder. Go raw numeric points
are `{x,y}`, binary events `{timestamp,value:boolean}`, and bounds are milliseconds.
HTTP explicit request dates are RFC3339 instants; legacy WS `from/to` are seconds. Do not
reuse the WS request type or guessed mock timestamps for HTTP.

- Validate the outer response, identity, bounds and each sample before committing. Unknown
  types become explicit unsupported series; invalid samples are reported/omitted without
  turning them into zero. Represent missing values explicitly in the internal model.
- Define canonical query identity from device, sorted expose set, exact UTC range and
  aggregation/limit. Display timezone affects formatting and calendar-range construction,
  not identity for an already fixed range. Generation checks protect against stale replies.
- Data query window is inclusive `[from,to]`; verify database and tail-cache boundaries with
  producer fixtures before dependent UI code. Shared chunk boundaries must not double-count.
  Stable timestamp sorting preserves repeated stored events; identical timestamp numeric
  points remain source evidence with a deterministic display tie rule documented in tests.
- Query cap probes 10,001 points per series; display/retain at most 10,000 and mark partial
  if exceeded. Never imply full-range statistics for truncated data; ask users to narrow
  the window. Render at most 1,000 numeric points, preserving extrema, endpoints and gaps.
- Sample mean is labelled; no inferred integral/consumption from cumulative energy counters.
  Zero-baseline percent change is unavailable. Binary durations only include known time;
  empty/unknown is not 0% measured activity. No automatic gap threshold is invented without
  reporting-cadence metadata; explicit missing values break lines, sparse histories are labelled.
- Keep loading until matching completion. Timeout after 10s; cancel superseded/closed queries;
  last valid results may remain visibly stale. Distinguish 400/401/403/429/5xx, malformed
  payload, empty data and missing entity. Retry is explicit and read-only.
- One active client request, 300ms selection debounce, batched exposes, optional refresh
  every ≥ 10s while visible/online, ≤ 10 cached entries, clear on logout. No history in
  localStorage/logs. Cancel means ignore stale results; it does not claim server work stopped.
- `QueryService.Query` must distinguish missing storage from failed reads; currently it
  continues on errors. Do not bypass that defect by showing a fake successful empty chart.

## CRAP and clean-code gate for everything touched

Inventory every touched production file/function: TS, Vue script blocks, mock JS and Go
where changed. Include service/store/composable glue and extracted helpers, not just pure
chart functions. Baseline against `fcdebca2` with identical pinned tooling and fixtures.

Use `CRAP = C² × (1 − coverage)³ + C`, with coverage as a fraction. Record which per-function
executable-statement coverage is used, complexity counting conventions, generated-SFC source
maps, stable file/function identities, tool versions and excluded generated/vendor files.
Go statement coverage and frontend coverage are reported separately, not averaged together.
No scores have been calculated for this planning update.

Proposed gate: CRAP ≤ 30 and C ≤ 10 for every new/changed function; ≥ 80% branch coverage
for touched decision logic with actual boundary/failure assertions. Already-good scores
must not worsen. For removed functions record successor scores; move/rename does not reset
the baseline. Do not add superficial wrapper tests or omit Vue functions to obtain green.

Preparation must prove source coverage for `.vue` setup callbacks, methods and emitted
interactions. Existing ts-jest alone does not do that. Evaluate a compatible Jest Vue
transformer/mounting setup; only add dependencies after documenting compatibility/cost in
ADR-004. Select a pinned Go scorer for touched Go and record exact commands; tool installation
or scores from another repository are not evidence here. If a scorer cannot measure a
language, tool support is a blocking quality task for that language, not a silent exclusion.

Refactor high scores by separating validation, orchestration, transformations and rendering;
use early returns, named policies and narrow adapters. Preserve error/empty handling and
typed contracts. CI reports per-function before/after scores and fails changed-function
violations; file-average improvement alone does not pass. Unrelated cleanup stays outside
this follow-up. Pure rendering assets without executable functions use lint/visual checks.

## Delayed Living room presence sensor illuminance

Observed report: changing light brightness eventually changes `illuminance`, after a delay.
Sensor model/ID, exact elapsed time, effective config and source reporting cadence remain
unknown. Brightness is a command to a light; illuminance is a sensor report. Keep them distinct.

Code evidence: `backend/models/settings/config.go` defaults measurement sampling to 60s.
`DeviceLifetimeService.Update` writes accepted state before `DebounceExpose`, then excludes
debounced changes from metrics and WS. It has no trailing latest-value flush in that method.
The frontend `hub/updateDevice` replaces changed expose objects and the device object.
These make backend sampling a plausible source of latency; they do not prove this household
device's effective configuration or rule out sensor cadence/another UI defect.

Investigation order:

1. Reproduce with synthetic raw reports and controlled clocks. Correlate ingress, accepted
   state, filtered broadcast, browser receipt, Vuex update and rendered value; use synthetic
   identifiers, keep timings only and avoid household payload logging.
2. Inspect effective default/device/expose sampling and source expose metadata; test
   illuminance/illuminance_lux names, real zero, rapid final change then silence, unchanged
   next reports, offline/reconnect and any rounded formatter masking changes.
3. Record root cause and timing contract. If UI receives a changed value, proposed render
   latency is ≤ 250ms p95. If backend drops the final broadcast, compare a bounded latest
   snapshot broadcast independent of metrics sampling against a per-expose trailing flush.
   Preserve ordered accepted state and all automation events, existing metrics retention/
   sampling meaning, resource bounds and shutdown; document a timing ADR before changes.
4. Add failing regression and focused CRAP remediation for every changed function. Replay
   with `backend/testing/hubharness`, mocked MQTT/WS and frontend fixtures. Passive evidence
   from a real device can be considered only in a separately scoped manual procedure;
   routine tests never change the household light's brightness.

## Brightness slider and frequent automation dispatch investigation

The owner reports that the UI brightness slider is troublesome while the dial automation
works. Compare those two paths with controlled fixtures before attributing the problem
to devices. `ExposeControl.vue` currently displays brightness as a percentage but sends
the raw slider value; its range has no metadata-derived step and its local draft clears
only on a changed confirmed value. Those are investigation points, not a diagnosed cause.

Trace gesture/input/change, device/expose selection, raw bounds/step, `useDeviceCommand`
refusal/pending, `hub/setDeviceValue`, outgoing `deviceSetValue`, resolved MQTT topic/payload,
echo, store update and rendered value. Compare minimum/zero, midpoint, maximum, repeated
gesture while pending, different/no echo, keyboard arrows and touch release with working
dial fixtures. Never send to the real light just to reproduce a routine regression.

The provided 19:33:19…19:33:36 debug lines for `0xa4c13894070052fc` originate from automation
dispatch on changed exposes; `attemptToTriggerAutomation` logs before the automation engine
matches conditions or performs actions. Do not infer executions from those lines or conflate
them with Attic air sensor metrics logs. Capture controlled counts/timings of ingress,
changed expose names, dispatches, matched trigger/action executions, commands and echoes.
Use existing `backend/testing/hubharness`, controllers/services and automation scenarios.
Check unrelated-expose reports, duplicate input, genuine repeated events, command echoes and
loop prevention. A demonstrated defect gets a failing regression and scoped CRAP remediation;
normal sensor dispatch is documented without suppressing legitimate automation inputs.
