# Plan: Spatial view parity and presence card states

Spec: [spec.md](spec.md) | Status: Draft | Author: agent draft | Updated: 2026-11-10
Constitutions: parent 1.0.0 (proposed); frontend 1.0.0 (proposed)

## Context and design

Observed facts:

- The derived presence device is registered as soon as a mapping and at least one zone
  exist, but its entities have no data until the hub's engine receives a spatial frame.
  `useDerivedPresence` therefore returns `occupied: null`; the Home card currently maps
  that to **Unknown**, and the live panel maps it to **clear**. Both are wrong for missing
  data.
- The Positions player never wires `showDistances` or `distanceOption`; only the live
  panel does. `SpatialCanvas` already derives the label distance from the drawn origin.
- `SpatialPositions.baseViewBox` centers on the fitted content bounds. The live panel uses
  `radarViewBox(scale)`, centered on the sensor.

Design:

1. **One derived-presence display rule.** Add
   `derivedPresenceDisplay(values): 'not-derived' | 'occupied' | 'clear'` in
   `src/domain/spatial/derived.ts`, where `occupied === null` is `not-derived`, never
   clear. Use it in `HomePresenceCard.vue` and `PresenceRailTile.vue`. In
   `SpatialPanel.vue`, return "Server presence: not derived yet." when the derived device
   exists but `occupied` is null. The card shows an explanatory sentence instead of a
   count/zone claim for `not-derived`; known states keep the existing presentation.

2. **Replay distance parity.** `SpatialPositions.vue` gains
   `const showDistances = ref(false)`, binds `v-model:show-distances`, sets
   `distance-option` on the player `SpatialViewer`, and passes `:show-distances` to its
   `SpatialCanvas`. The Heatmap viewer stays unchanged because it renders no tracks.

3. **Sensor-centered replay fit.** Add
   `centeredSpatialView(bounds, center, aspect): { x, y, width, height }` to
   `src/domain/spatial/radar.ts`. It takes the maximum absolute offset from the sensor in
   each axis, chooses a view height that satisfies the canvas aspect, and returns a
   symmetric view box. `SpatialPositions.vue` uses `screenPoint(workspace.sensor,
   orientation)` as the center, replaces `fittedExtent` with the helper result, and uses
   the result's height for `symbolScale`/`labelStroke`. The fallback remains
   `radarViewBox(SPATIAL_DEFAULT_SCALE_MM)`. Zoom, focus, pan and reset keys are untouched.

4. **Compact, adaptive presence overview.** `PresenceRailTile.vue` becomes a compact row:
   a 3.75 rem 5:3 preview thumbnail (or an icon placeholder when there is no frame), name
   at 0.875 rem/600, state at 0.82 rem/500, metadata/links at 0.7 rem, and a bordered
   separator between sensors. `useSpatialPreview` exposes `hasFrameData` (any frame) so an
   empty frame still renders the room/sensor preview. `HomePresenceOverviewCard.vue` owns
   the filtered sensor list: one sensor renders the row as the card (no separate header,
   device name in the row); editing keeps a header for the remove control and hides the
   row name. Multiple sensors use a quiet section header ("Presence · N sensors") and named
   rows. `HomePresenceOverview.vue` receives that list and `showNames`, and the extra rail
   gap is removed. Values and links are unchanged (AC-12…AC-17).

5. **Presentation extraction (SRP/SSOT).** Presence copy, state/detail labels, zone chips,
   the overview model and the server readout move to `domain/spatial/presence.ts`. The
   Positions fit/symbol math moves to `domain/spatial/positions.ts`. `useInViewport` owns
   the shared observer lifecycle and `usePresenceTile` owns the compact-row view model.
   `HomePresenceCard`, `PresenceRailTile`, `HomePresenceOverviewCard`,
   `HomePresenceOverview`, `SpatialPanel` and `SpatialPositions` only render models; they
   no longer define presence labels or derive presence/positions.

6. **Tests.** Unit cases for the display rule, presence copy/models, geometry helper and
   position metrics; update the Positions browser case to expect a working Distance
   checkbox and a centered sensor marker; add a Home presence-card browser case for the
   not-derived copy and a tile-height assertion for the compact overview row.

Files:

| Path | Change |
| --- | --- |
| `frontend/src/domain/spatial/derived.ts` | `derivedPresenceDisplay` |
| `frontend/src/domain/spatial/presence.ts` | presence copy, labels, zone chips, overview and server models |
| `frontend/src/domain/spatial/positions.ts` | Positions fit/symbol metrics |
| `frontend/src/composables/useInViewport.ts` | shared observer lifecycle |
| `frontend/src/composables/spatial/usePresenceTile.ts` | compact presence row view model |
| `frontend/src/components/home/HomePresenceCard.vue` | render presence card model |
| `frontend/src/components/home/PresenceRailTile.vue` | compact row, thumbnail/placeholder, adaptive name |
| `frontend/src/components/home/HomePresenceOverviewCard.vue` | single/multiple header and sensor list |
| `frontend/src/components/home/HomePresenceOverview.vue` | receives sensors and `showNames` |
| `frontend/src/composables/spatial/useSpatialPreview.ts` | `hasFrameData` for an empty-frame preview |
| `frontend/src/components/device/spatial/SpatialPanel.vue` | null is not clear |
| `frontend/src/domain/spatial/radar.ts` | `centeredSpatialView` |
| `frontend/src/components/device/spatial/SpatialPositions.vue` | distance option, centered base view |
| `frontend/src/__tests__/domain/spatial/derived.test.ts` | display-state cases |
| `frontend/src/__tests__/domain/spatial/radar.test.ts` | geometry cases |
| `frontend/e2e/devices.spec.ts` | Positions distance + centering cases |
| `frontend/e2e/home.spec.ts` | presence card not-derived case (if fixture permits) |

Alternatives rejected:

- **Publish an initial clear from the presence engine:** the engine has no frame yet, so
  clear would be an invented measurement (NH-02). Unknown remains honest until a frame.
- **Fall back to firmware occupancy in the card:** mixes authorities and contradicts the
  spec 041 derived-presence contract. A future fallback must be explicitly labelled.
- **Keep the fitted-bounds center and offset the viewBox:** either the sensor stays
  off-center or one side of the content is clipped.
- **Use the live fixed `radarViewBox` for replay:** it would ignore the recorded extent
  and could clip a session that extends beyond the default range.
- **Add a Positions range selector:** unnecessary for this fix and adds persisted/UI state.
- **Show Distance on the Heatmap:** it has no target labels, so the control would be inert.

## Constitution Check

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 062 with AC-01…AC-11 and resolved Q1–Q3 | PASS | PASS |
| NH-02 | Missing derived values stay not-derived; no clear invention; no device command | PASS | PASS |
| NH-03 | No new entry point, data exposure, secret or fetch | PASS | PASS |
| NH-04 | Pure domain helpers; no new dependency or storage | PASS | PASS |
| NH-05 | Unit cases + desktop/mobile browser checks; results below | PASS (design) | PASS |
| NH-06 | No persistence, timer, subscription or retention change | PASS | PASS |
| FE-01 | Derived display rule in domain; components stay presentational | PASS | PASS |
| FE-02 | No API/WS payload or type contract change | PASS | PASS |
| FE-03 | Only confirmed values are shown; missing values keep an explicit state | PASS | PASS |
| FE-04 | Existing checkbox/button controls; keyboard and 375/768/1440 checks | PASS (design) | PASS |
| FE-05 | No secrets, assistant content or new persistence | PASS | PASS |

## Contracts, data, rollout

No wire, storage or configuration change. The new view state is in-memory like Grid and
Field of view; replay sessions and the live view are unaffected. Rollback is reverting the
component/domain diff.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01…AC-04 | `derived.ts`, card, rail, panel | Jest `derivedPresenceDisplay`; Home browser card transition | pass |
| FR-03, AC-05…AC-07 | `SpatialPositions.vue` | Positions browser: Distance present, label toggles, Heatmap has none | pass |
| FR-04, AC-08…AC-10 | `radar.ts`, `SpatialPositions.vue` | Jest `centeredSpatialView`; browser sensor-centre assertion | pass |
| AC-11 | `SpatialPositions.vue` | browser: 375/768/1440 overflow checks | pass |
| FR-10, AC-12…AC-17 | rail tile, overview card/list, preview | browser: single row is the card (no separate header); two sensors show one section header + two named rows; tile height < 100/110 px | pass |
| NFR-01…NFR-04 | all | diff review: no network, dependency, persistence or live-view change | pass |
| NFR-05, NFR-06 | domain/composable modules | unit tests for presence copy/models and position metrics; component diff has no labels or presence derivation | pass |

## Verification (recorded)

- `cd frontend && npm test -- --runInBand` → **146 suites / 821 tests passed**, including
  the new `domain/spatial/presence.test.ts` and `domain/spatial/positions.test.ts` suites.
- `cd frontend && npm run lint` → **0 errors, 23 pre-existing warnings** (unchanged set).
- `cd frontend && npm run build` → pass (`vue-tsc` + Vite).
- `cd frontend && npm run type-check:e2e` → pass.
- `npx playwright test e2e/devices.spec.ts e2e/home.spec.ts --project=desktop` →
  **24 passed**, including the Positions distance/centring, single-row presence, two-sensor
  section-header and not-derived card cases.
- The same command with `--project=mobile` → **24 passed**, including the 375 px checks.
- Backend untouched; no backend checks were required for this branch.

## Debt, risks, exceptions

- **Initial derived data:** the card will say “not derived yet” until the device reports a
  spatial frame. That is intentional; the engine must not invent clear.
- **Symmetric fit zooms out:** centering on the sensor can show more empty space than the
  previous bounds fit when the recorded path is asymmetric. This is the cost of parity and
  keeps all content visible.
- **Fixture coverage:** the Home presence card is now browser-driven: the test adds a
  derived device with no `occupied` value, sees the not-derived state, sends a
  `deviceUpdated` report and sees Clear. No side effects on a real hub.
- **No ADR:** additive presentation change with pure helpers; no boundary or storage
  decision.

## Completion

Implemented on `fix/spatial-replay-parity`. Acceptance evidence is recorded above and no
exceptions are requested. Missing derived presence is now labelled honestly, Positions
offers the same distance option as live, its fitted view is sensor-centred on desktop and
mobile, and the Home presence overview uses a compact row that matches the neighbouring
card type scale. A single sensor is one row (no separate card header); multiple sensors
get one section header plus named rows. Presence copy/view models and Positions view metrics
live in domain modules, and shared viewport/tile state lives in composables; the touched
components only render.
