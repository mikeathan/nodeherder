# Plan: Distance-to-target labels in the spatial viewer

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft | Updated: 2026-11-09
Constitutions: parent 1.0.0 (proposed); frontend 1.0.0 (proposed)

## Context and design

Facts: the hub frame contract has no distance (`backend/models/devices/spatial.go`), the
viewer's tracks come from the filtered pipeline, `SpatialViewer.vue` already hosts Grid and
Field-of-view checkboxes, and `SpatialCanvas.vue` renders one `<text>` label per target
from its `renderedTracks` computed pass.

Design:

1. **Pure helper** in `src/domain/spatial/radar.ts`, beside the existing
   `formatSpatialDistance`:

   ```ts
   /** Label for one target: its name, plus a rounded distance when requested. */
   export function spatialTargetLabel(label: string, millimetres: number | null, showDistance: boolean): string
   ```

   - Returns `label` unchanged when `showDistance` is false.
   - Returns `label` unchanged when the distance is null/non-finite (AC-07) — never `— m`.
   - Otherwise rounds millimetres to the nearest 100 mm before formatting, so the text is
     quantised to 0.1 m and cannot flicker on filter noise (AC-05).
   - Formats through `formatQuantity(metres, 'distance', { digits: 1 })` so the unit comes
     from the shared formatter (FR-05).

2. **Canvas** (`SpatialCanvas.vue`): `renderedTracks` already resolves each track's drawn
   origin via `trackPoint()` (filtered, or raw when `showRaw`). Add the distance computed
   from that same origin — `Math.hypot(point.x, point.y)` where `point` is the projected
   `origin` — so the label can never disagree with the dot (AC-03, AC-06). A new
   `showDistances?: boolean` prop gates it. No new watcher or reactive source (NFR-01).

3. **Viewer** (`SpatialViewer.vue`): a third checkbox next to Grid and Field of view,
   `v-model:show-distances`, matching the existing markup so keyboard and label behaviour
   are identical (AC-08).

4. **Panel** (`SpatialPanel.vue`): `const showDistances = ref(false)` beside `showGrid` /
   `showFov`, passed to both canvas instances. Default off, so nothing changes until ticked
   (AC-01).

Why not `target_distance`: it is a single scalar for the firmware's own target, is not one
of the mapped coordinate entities, and can contradict the drawn position. Distances for
targets 2–3 would be impossible. Computing from the mapped (and already filtered) x/y keeps
the label honest and needs no contract change.

Files:

| Path | Change |
| --- | --- |
| `src/domain/spatial/radar.ts` | add `spatialTargetLabel` |
| `src/components/device/spatial/SpatialCanvas.vue` | `showDistances` prop; label from the drawn origin |
| `src/components/device/spatial/SpatialViewer.vue` | Distance checkbox |
| `src/components/device/spatial/SpatialPanel.vue` | `showDistances` state, wire to both canvases |
| `src/__tests__/domain/spatial/radar.test.ts` | helper cases |
| `e2e/devices.spec.ts` | toggle scenario in the existing spatial test |

Alternatives rejected:

- Reading `target_distance`: cannot describe more than one target and can disagree with the
  dot (see spec Q1).
- A separate overlay/SVG group with absolute positions: duplicates the projection maths and
  would drift from the dot on zoom.
- Persisting the preference: new storage and migration for a per-session display toggle;
  Grid and Field of view are not persisted either.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 060 with AC-01…AC-08 and resolved questions | PASS | PASS |
| NH-02 | Presentation only; no device command, no automation input, no change to accepted state | PASS | PASS |
| NH-03 | No new entry point, no secret or household data leaves the browser; positions stay local | PASS | PASS |
| NH-04 | Pure helper in the existing domain module; no new dependency or abstraction | PASS | PASS |
| NH-05 | Unit cases + browser toggle in the existing spatial test; results recorded | PASS | PASS |
| NH-06 | No persistence change; no timer, observer or subscription added | PASS | PASS |
| FE-01 | Canvas stays presentational; the derived label is a domain helper | PASS | PASS |
| FE-02 | Frame contract untouched; no new runtime validation needed | PASS | PASS |
| FE-03 | Distance is derived from the drawn position; unknown positions show no value | PASS | PASS |
| FE-04 | Labelled checkbox, keyboard operable, no colour-only signal, mobile label length checked | PASS | PASS |
| FE-05 | No secrets, no assistant text, no new persistence | PASS | PASS |

## Contracts, data, rollout

No HTTP/WS/MQTT/MCP or persistence change. `SpatialFrame` and the persisted workspace are
untouched, so no migration or compatibility period is required. Rollback is reverting the
component diff; no stored data refers to this feature.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, AC-02, AC-08 | `SpatialViewer.vue`, `SpatialPanel.vue` | browser: check/uncheck Distance, assert label text and keyboard focus | pass |
| FR-02, AC-03, AC-06 | `SpatialCanvas.vue` | browser: `hypot(1000, 2000)` renders `Track 1 · 2.2 m` | pass |
| FR-03, AC-05 | `radar.ts` | unit: 0.1 m quantisation; 2400–2444 mm share one label, 2468 mm steps | pass |
| FR-04, AC-07 | `radar.ts`, `SpatialCanvas.vue` | unit: null/undefined/NaN leave the label unchanged | pass |
| FR-05 | `radar.ts` | unit: output comes from `formatQuantity` | pass |
| AC-04 | `SpatialCanvas.vue` | label derives from the drawn origin; `Raw` shares the same source | pass |
| NFR-01, NFR-02 | all | diff review: no new reactive source, dependency or persistence | pass |

## Verification (recorded)

- `npm test -- --runInBand` → **775 passed / 140 suites** (4 new label cases).
- `npm run lint` → 0 errors (23 pre-existing warnings).
- `npm run build` → pass · `npm run type-check:e2e` → pass.
- `npx playwright test e2e/devices.spec.ts --project=desktop` → **9 passed**, including the
  live canvas, Positions replay and Heatmap in one test.
- `--project=mobile` → **9 passed**.
- Visual: label reads `Track 1 · 2.2 m` beside the dot and reverts to `Track 1` when the
  checkbox is cleared.

Selected commands (from `frontend/`): `npm test -- --runInBand`; `npm run lint`;
`npm run build`; `npm run type-check:e2e`;
`npx playwright test --project=desktop e2e/devices.spec.ts -g "spatial"` plus the same with
`--project=mobile`.

## Debt, risks, exceptions

- The distance is sensor-origin radial distance, not distance to the nearest zone edge. That
  matches what the axes mean; a zone-relative distance would be a separate feature.
- Quantising to 0.1 m trades precision for readability. Someone needing exact millimetres
  can use the existing accessible target list, which keeps 3-decimal output.
- `target_distance` remains visible as an ordinary device reading; it just is not what the
  canvas label shows. Recorded in the spec so the discrepancy is deliberate.

## Defect found while implementing (fixed)

`SpatialCanvas.vue` guarded its track label with `v-if="showTargetLabels !== false"`, but
the prop had no default. Vue casts an absent `Boolean` prop to `false`, so the condition was
never true and **track labels have never rendered** on any canvas. It went unnoticed because
the accessible target list (`Track 1 (live): X …`) still described each target, and no test
asserted on the SVG label.

Measured before the fix: a live target produced
`<g><circle class="nh-target is-live is-color-0"></circle><!----></g>` — the `<text>` had
compiled to a placeholder. After giving the prop `{ showTargetLabels: true }` through
`withDefaults`, the label renders and the preview surface
(`PresenceRailTile`) still opts out explicitly with `:show-target-labels="false"`.

This also restores labels on the Positions replay canvas, which passes real tracks. The
distance option builds on that same element, so it could not work without fixing the guard.

## Second review pass (fixed)

Reviewing my own diff turned up three defects, all in this change:

1. **Dead checkbox on the Positions viewer.** I hid the Distance checkbox with
   `v-if="showDistances !== undefined"`, but Vue casts an absent Boolean prop to `false`,
   never `undefined`. The replay viewer therefore showed a Distance checkbox that was not
   wired to its canvas — toggling it did nothing. Measured: the control count on the
   Positions view was `1` before the fix. Replaced with an explicit `distanceOption`
   opt-in that only `SpatialPanel` sets, and the browser test now asserts the control is
   absent there.
2. **Labels leaked into a tiny preview.** Flipping `showTargetLabels` to default `true`
   also enabled labels in `HomePresenceCard`'s preview, which (unlike `PresenceRailTile`)
   never opted out. At `scale: 4000` the label font is 200 view units inside an
   11 rem-high box, so it would have crowded the card. Added the same explicit
   `:show-target-labels="false"` the rail tile already used.
3. **Redundant parameter and wasted work.** `spatialTargetLabel` still took a
   `showDistance` flag even though the only caller had already branched on it, and
   `Math.hypot` ran for every track on every frame while the option was off. Dropped the
   parameter and moved the square root behind the enable check.

All six `SpatialCanvas` consumers were re-checked after the change: live panel (labels +
optional distance), Positions replay (labels), heatmap (no tracks), and the two Home
previews (labels explicitly off).

## Completion

Acceptance evidence recorded per task in `tasks.md`. No review approval is claimed.
