# Entity design previews

Generated with the built-in image generation tool for visual discussion. Sample
values, geometry, timestamps and statistics are illustrative, not validated data.
Capability examples do not establish support on any installed device.

Current boards:

- [Brightness and colour temperature, v3](entity-views-mobile-concept-v3.png):
  brightness is the preferred reference; the temperature panel was rejected.
- [Environmental refined history, v5](entity-views-environmental-concept-v5.png):
  temperature, humidity, illuminance and air quality with stronger clean strokes,
  restrained area fills, fewer time labels and horizontal guides. V4 was rejected
  for its jagged traces and chart appearance. The owner accepted v5.
- [Activity and electricity, v4](entity-views-activity-energy-concept-v4.png):
  presence, contact, instantaneous power and interval energy consumption.
- [Diagnostics, alerts and settings, v4](entity-views-diagnostics-settings-concept-v4.png):
  battery/signal, leak (also a pattern for smoke/tamper/faults), enum/numeric
  controls and repeated button events. Settings are capability illustrations;
  actual names, values, ranges and writable properties must come from metadata.

Earlier desktop and mobile files remain as explorations. Mobile adjustable views
are full-screen and control-first, with large controls and small recent history.
Read-only views can include compact useful summaries alongside history.

Charts are being redesigned for a modern, polished appearance, readable typography,
restrained fills, fine grid lines, clear units and accessible inspection. The implementation uses Chart.js behind a shared typed renderer after mobile
retention checks identified an ApexCharts teardown problem.
Touch interaction, accessibility, accuracy and resource ownership are verified by
source-owned tests and the isolated browser suite. Generated curves/timelines are not an accuracy specification.
At investigation time the frontend used ApexCharts/vue3-apexcharts. The completed
implementation replaces them with Chart.js 4.5.1; see the verification record. The v5
preview illustrates calm sample histories; production must preserve actual spikes,
gaps and samples. Use explicit labelled aggregation when reducing detail, and do
not discard meaningful variation merely to imitate the preview's contours.

V5 edit prompt brief: keep the four environmental layouts; replace scratchy traces
and rectangular grids with confident rounded strokes, understated fading area
fills, three horizontal guides and three time labels, inspired by Home Assistant
sensor history. Use simple illustrative histories and retain compact summaries.

Prompt briefs for the v4 boards:

- Environmental: four full-screen dark mobile views; alternative left-aligned
  temperature reading; recent line chart; condensed range/average rows; consistent
  layouts for humidity, illuminance and CO₂; no gauges or artificial controls.
- Activity/electricity: four dark mobile views; thin proportional presence/contact
  timelines and useful summaries; power line chart and socket toggle; interval
  energy bars with cumulative meter total clearly separate; modern restrained styling.
- Diagnostics/settings: battery step history and signal in its actual unit; leak
  state/history; metadata-driven mode/number controls; button event bars/list with
  repeated events retained; comfortable targets and compact meaningful content.

See the [investigation](../entity-views-and-charts-review.md) for source findings and
data/interaction requirements and the [implementation evidence](../../specs/007-frontend-redesign/implementation-evidence-2026-10-18.md) for the completed scope and checks.


Actual implementation screenshots, captured against isolated mock data:

- [Brightness mobile](implementation/brightness-mobile.png) · [desktop](implementation/brightness-desktop.png)
- [Temperature mobile](implementation/temperature-mobile.png) · [desktop](implementation/temperature-desktop.png)

These are browser screenshots of working views; the concept boards above remain
visual references for the broader entity families.
