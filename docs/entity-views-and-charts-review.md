# Entity views and charts: design investigation

Date: 2026-10-17. Historical investigation; subsequent implementation and results:
[2026-10-18 evidence](../specs/007-frontend-redesign/implementation-evidence-2026-10-18.md).

Original scope: Scope: all entity detail views, their entry tiles and controls,
all five mini-chart variants, full numeric and binary history, period selection,
statistics, tooltips and loading states. Research and source review only; this is
not an implementation plan or a claim of browser/accessibility conformance.

## Revised direction after preview feedback

Latest exploration: [preview index](design-previews/README.md). Brightness v3 is
the preferred reference; temperature v3 was rejected. V4 explores alternatives
across environmental, activity/electricity and diagnostic/config/event families.
All charts should receive a modern visual redesign; a library change remains an
option to evaluate against interaction/accessibility/performance needs, not a
prerequisite decided by these images.
Environmental v4 was rejected for its chart appearance. The newer
[environmental v5 preview](design-previews/entity-views-environmental-concept-v5.png)
explores Home Assistant-inspired clean strokes, restrained area fills and simpler
axes. The owner accepted v5; sample curves are not instructions to hide real data.

The user's feedback supersedes the denser preview boards and the mobile layout
recommendations below. Mobile entity views are focused control surfaces, inspired
by Home Assistant's more-info interaction. An adjustable entity opens to its main
adjustment immediately; the device page remains the home for device information,
diagnostics and advanced settings.

- Mobile uses the full screen, with a compact title and close button, a prominent
  current value, and a large primary control occupying most of the usable space.
  Brightness should use a substantial vertical filled slider with an obvious handle;
  colour temperature gets an equally usable warm-to-cool adjustment. Control size
  adapts to short screens without pushing essential actions off screen.
- Keep only a few directly relevant controls, such as power and switching between
  brightness and colour temperature. Clearly label the selected adjustment. Changes
  show pending/failed feedback without pretending the requested value is confirmed.
- A small, secondary history chart covers the last few hours. Omit statistic cards,
  event lists, prominent freshness metadata, period toolbars and advanced settings
  from the default mobile control view. Show freshness/offline warnings when they
  materially affect whether a reading or adjustment can be trusted.
- Read-only entities can use the available space for useful, compact information:
  a larger recent-history chart and a small entity-specific summary. Temperature
  can show min/average/max for the displayed period; presence/contact can show
  known active/open duration and genuine transitions. Keep the presentation clean
  and condensed, with explicit period/coverage semantics. Do not add artificial
  controls or unrelated device diagnostics.
- Desktop can place controls beside a larger history chart and add useful summaries
  where space allows, while keeping the primary adjustment visually dominant.

The generated desktop and mobile boards in `docs/design-previews/` are earlier
explorations. In particular, the mobile board is too dense and its sliders too small;
it does not represent the accepted direction above. The revised preview is
[Mobile concept v3](design-previews/entity-views-mobile-concept-v3.png), showing
large adjustment controls with inset grips, minimal recent history for adjustable
entities, and a compact period summary for read-only temperature. Its data and chart geometry
are illustrative. No production implementation has been created yet.

## Recommendation

Use one consistent entity detail layout with content chosen by capability and
measurement semantics. Keep Hearth and Panel styling through existing theme tokens.
Use a compact desktop dialog and a phone bottom sheet that can expand for longer
content. Put the entity state first, its main control second, and history after that.
Keep advanced settings behind a labelled disclosure and provide an explicit device link.

Review all charts together. Mini and full charts should share units, state labels,
time boundaries, missing-data rules, colours and calculations. Size should change
the amount of detail, not the meaning of the data.

## Evidence and limits

Reviewed the five supplied screenshots and frontend production sources, including
`EntityViewDialog`, `useEntityDialog`, `useDialogUI`, `EntityTile`, `ExposeControl`,
`domain/exposes`, `useDeviceCommand`, every component under `components/chart`,
`DeviceMetrics`, chart utilities/contracts and the metrics Vuex module. Reviewed
existing formatter tests and the ExposeControl browser regression source.

No application was launched, no household devices were contacted, and no tests were
run. Findings below are screenshot observations and source-confirmed behavior;
actual clipping, contrast ratios, touch interaction and focus behavior still need
mock-backed browser verification. Device-specific capabilities are not an inventory
of hardware installed in this household.

## Existing problems

| Finding | Evidence and implication |
| --- | --- |
| Tile and dialog disagree | The brightness tile uses `formatExposeValue`, which normalises against the expose maximum. The dialog uses the older formatter, which appends `%` to raw brightness. This explains the screenshot's 72% versus 184%. |
| Dialog controls bypass newer command UX | `useEntityDialog` dispatches directly rather than using `useDeviceCommand`. Its offline check does not express all the shared control restrictions, pending confirmation or timeout feedback. |
| Selected entity does not determine the primary control | The dialog selects a device numeric control by iteration order. It also hides the selected entity's chart whenever the device has a brightness control. A device's unrelated capabilities should not determine the selected reading's presentation. |
| Generic enablement is unreliable | `isEnabled` disables numeric zero and requires an active binary control when controls exist. Zero is a valid value; writable numeric-only devices need their own enablement rules. |
| Incomplete control coverage | The old entity component resolver handles numeric inputs only; config menus depend on preset values. The newer `ExposeControl` supports switches, choices, sliders and numbers. |
| Timestamp wording lacks scope | The dialog displays `device.last_seen`; it is device activity, not proof that this specific reading changed then. Label it accordingly unless a per-reading timestamp exists. |
| Layouts do not share the newer dialog pattern | `useDialogUI` uses a full-screen phone presentation; `UiDialog` already has a token-based bottom sheet. Screenshots show oversized vertical content and substantial unused space. |
| Accessibility needs correction | Old control icons are clickable `div`s; the range has no explicit accessible name and removes its outline. Binary chart cells are also click-only `div`s. Keyboard and screen-reader access need explicit design and verification. |

## Design references

These are patterns to adapt, not a recommendation to replace the UI framework.

- [Home Assistant tile cards](https://www.home-assistant.io/dashboards/tile/): separate quick state/control from deeper detail. This fits the existing tile interaction model.
- [Home Assistant history graphs](https://www.home-assistant.io/dashboards/history-graph/): distinguish numeric trends from state histories and make the displayed period explicit.
- [Home Assistant energy cards](https://www.home-assistant.io/dashboards/energy/): consumption over a selected period is more useful than an unexplained cumulative-energy gauge.
- [Apple sheets](https://developer.apple.com/design/human-interface-guidelines/sheets): a sheet suits a focused task within the current context; longer workflows need more room. Avoid stacked modal flows.
- [Material bottom sheets](https://m1.material.io/components/bottom-sheets.html): useful precedent for a phone detail surface and adapting its presentation on larger screens. This reference is the older Material guidance.
- [WAI dialog pattern](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/) and [slider pattern](https://www.w3.org/WAI/ARIA/apg/patterns/slider/): define focus, keyboard and accessible-value behavior; test touch assistive technology as well.
- [WCAG 2.2](https://www.w3.org/TR/WCAG22/): accessibility baseline. Recommend 44–48 CSS px touch controls as a product target; WCAG AA's minimum target criterion is 24 CSS px with exceptions.

## Entity families to cover

| Family | Recommended view |
| --- | --- |
| Temperature, humidity, pressure, illuminance, air quality, distance | Current reading with unit; clear freshness; trend; compact period statistics. Air-quality limits only when meaningful thresholds are defined. |
| Presence and occupancy | Detected/Clear; duration in current state only if known; thin state timeline; detected duration and genuine transitions; optional activity density. |
| Contact | Open/Closed; labelled legend; time open and opening count. Preserve the expose's on/off mapping rather than treating boolean truth as universally active. |
| Lights | Labelled power control; correctly normalised brightness; colour temperature and presets when supported; colour control only where the contract supports it. Prefer a compact horizontal slider with readable output; do not make a tall slider the only way to set a value. |
| Switches and sockets | Clear On/Off control with pending feedback. Group related power/energy readings and child lock as secondary content. |
| Energy and power | Separate instantaneous W/kW, cumulative meter kWh, and consumption over a period. Show which quantity each value represents. |
| Battery and signal diagnostics | Battery percentage or low-battery state; signal in its actual unit; optional history. No invented good/bad thresholds. |
| Smoke, leak, tamper, alarm and faults | Explicit alert/clear words and icon; priority styling; history where available. Put any supported silence/test action in a clearly labelled control. |
| Enum/mode/preset exposes | Labelled select or a short segmented control; current selection; confirmation state. Category history or an event list only if recorded. |
| Numeric configuration | Label, description, unit, bounds and supported precision; number input or slider as appropriate. Configuration need not have a chart. |
| Actions/events | Timestamped event list; repeated equal events remain visible. Distinguish event counts from state transitions. |
| Unknown/new exposes | Safe fallback from metadata: human-readable label/value, supported control and clear history availability. Unsupported shapes must not crash the view. |
| Future thermostat, cover or fan capability groups | Specialist grouped controls only when actual expose metadata supports them. Do not infer a complete device interface from one property name. |

## Shared desktop and phone layout

Desktop: approximately 480–640 px wide for ordinary detail, content-driven height
with a viewport limit. A wider history view can use more space. Header contains the
entity label, device/area context and labelled close action. State and its unit form
one readable block. Controls have visible labels; history has its own heading,
period selector and explicit refresh/live status. The device link and advanced
settings follow the main task.

Phone: one column, bottom sheet with a visible close control and safe-area spacing.
Allow expansion for history or complex settings; keep header visible and body
scrollable. Avoid nested scrolling and horizontal page overflow. Short landscape
viewports must keep close and primary controls reachable. Mobile browser Back should
dismiss the view predictably; this needs deliberate integration and testing.

Preserve readable numbers and wrap long device names. Do not squeeze five statistic
columns into a phone width. Use two-column statistics or a short summary with more
detail available. Keep normal text comfortably readable, visible focus and themes
consistent, and honour reduced motion. Settings should open within the existing
surface where practical rather than create a stack of dialogs.

## Full chart audit

| Surface | Current source finding | Recommended change |
| --- | --- | --- |
| Numeric mini/full history | Smooth curves are used broadly. Numeric normalisation drops non-finite samples; there is no explicit gap representation in the reviewed numeric point type. | Straight lines for measurements; steps for discrete settings. Preserve meaningful gaps and show insufficient coverage. Do not imply measurements between outages. |
| Numeric downsampling | Utility keeps every nth point and the last point. Short spikes can disappear. Full charts sample to 350 points; mini variants to 50/60. | Preserve extrema and important transitions. Validate with synthetic spikes and gaps. Statistics should describe the intended source data, not an arbitrary display sample. |
| Numeric statistics | Average is arithmetic over samples, not duration-weighted. Full-chart statistics disappear beyond 2,000 normalised points. | Label sample averages or define time-weighted averages where appropriate. Keep useful summaries at large ranges using suitable aggregation. |
| Full numeric formatting | `NumericChart` and `AreaChart` derive units from the old name-based formatter; reviewed metric types carry no explicit unit field. | Resolve metadata consistently with the selected expose. Verify producer contracts before deciding whether frontend lookup suffices or API metadata must change. |
| Percentage mini | Displayed latest value is clamped to 0–100, but historical series uses raw points. | Use one normalisation for label, series, axis and tooltip. Bound genuine percentages; flag invalid values rather than silently hiding them. |
| Realtime mini | History's last sample is labelled “Now”; empty stats fall back to zero. | Say “Latest recorded” with time unless freshness is established. Distinguish zero, missing and stale. |
| Energy gauge | Four equal coloured bands, a default maximum and last historical sample create an unexplained gauge. It can disagree with the current state, as in the screenshot. | Replace cumulative-energy gauge with meter total plus interval consumption where available. Gauges are optional only with a meaningful labelled scale and thresholds. |
| Binary mini/full timelines | Minimum segment widths deliberately enlarge short intervals and shrink larger ones. | Keep the time axis proportional. Make narrow events selectable through a separate hit area or event list; never alter duration geometry to enlarge targets. |
| Binary statistics | “Changes”/“Triggers” sums sample counts. Repeated same-state reports can count as changes. Percentage excludes unknown duration without showing coverage. | Compute transitions separately from reports/events. Show known coverage and define percentage denominator. Use entity-specific wording such as “Time open”. |
| Binary range consistency | Mini derives its range from first data point to now; full chart uses requested bounds. | Apply the requested window consistently. A missing initial state is unknown unless a valid pre-window state can establish it. |
| Density view | Mini uses five-minute buckets; full uses fifteen minutes. An event-count threshold automatically changes representation. | Offer labelled Timeline/Activity modes. Choose buckets for range and width; state whether colour means event count or active duration, with a legend. |
| Tooltips and inspection | Binary tooltips use click-only cells and HTML strings; numeric variants use hover-oriented library tooltips. | Tap and keyboard inspection; bounded tooltip placement; exact timestamp/value/duration; a textual or tabular alternative. Review escaping of externally derived tooltip text during implementation. |
| Period controls | Full history offers many overlapping period names; mini is fixed to four hours. | Shared clear presets, e.g. 1h, 6h, 24h, 7d, plus Today/custom where supported. Distinguish rolling 24h from calendar Today and show timezone/date range. |
| Loading and refresh | `DeviceMetrics` awaits an action that sends a WS request, not its eventual response. Mini has no explicit error state; “live” subscribes to store changes without establishing recurring fetches. | Loading must track response completion, timeout or error. Separate one-shot Refresh from genuine live updates; stop live work when hidden. Never advertise refresh behavior that the data path does not provide. |
| Request ownership | Metrics results are keyed by device only and new queries delete that device's previous result. Different expose/range consumers can interfere. | Investigate request identity and stale-response handling before promising reliable simultaneous views. This is a contract/state concern as well as presentation. |
| Themes and mobile | Many chart colours, labels and tooltip themes are hard-coded for dark presentation. Full stats specify five mobile columns; date range forces no wrapping. | Use shared theme tokens, readable contrast, fewer mobile ticks, wrapping dates and responsive statistic rows. |

## Chart selection rules

Measurements use line charts; state and setting histories use steps or labelled
timelines; cumulative energy can use a labelled cumulative line; consumption uses
interval bars. Deriving consumption requires correct counter-reset handling, units,
time boundaries, timezone and missing-data policy. Do not assume all those data are
already available. Events use lists or clearly labelled count buckets. Generic enum
history is currently skipped except binary-chartable `state`; adding category
history depends on the recorded data contract.

Use neutral up/down change styling unless higher/lower has a defined good/bad
meaning. Never invent warning bands for a measurement merely to decorate a chart.
Show source coverage when it materially limits a statistic. Treat one point as a
point, an empty period as no data, and an outage as unknown.

## Suggested priorities and evidence needed next

First address trust: shared value formatting, command confirmation, units, true time
proportions, transition counts, missing data and chart request lifecycle. Then unify
the entity surface and chart layouts, followed by optional deeper analytics. These
are priorities for a later SDD spec/plan/tasks, not implementation instructions.

Before implementation, use mock-backed desktop and phone designs for a light,
numeric sensor, contact/presence sensor, socket/energy view, config expose and alert.
Exercise 320/360/390 px phones, short landscape, 768 px tablet and 1280/1920 px
desktop, plus zoom and long names. Cover keyboard focus/return, touch inspection,
light/dark themes, reduced motion, pending/timeout, offline/reconnect, deleted
entities and missing readings.

Chart fixtures should include no samples, one sample, measured zero, gaps, irregular
intervals, repeated reports, brief spikes, unknown binary state, dense activity,
large ranges, counter reset, overlapping requests, stale responses, midnight and DST
boundaries. Read-only history failures must not disable unrelated device controls.
No routine verification should command household devices.

Implementation should build on `domain/exposes`, `ExposeControl`, `useDeviceCommand`
and shared dialog/theme infrastructure, while fixing their relevant limitations.
The existing ExposeControl browser regression source explicitly records a draft
value remaining displayed after timeout; reuse alone does not resolve that issue.
Retain the current chart library unless measured constraints justify replacement.
