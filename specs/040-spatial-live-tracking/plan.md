# Plan: Spatial live tracking

Spec: [spec.md](spec.md) | Status: Implemented (unreviewed) | Author: unreviewed draft | Updated: 2026-11-03
Constitutions: parent 1.0.0 (proposed); backend 1.0.0 (proposed); frontend 1.0.0 (proposed)
Decision record: [ADR-001](adr-001-spatial-target-frames.md) | Contract:
[spatial-targets.md](contracts/spatial-targets.md)

## Context and design

### Observed current behavior

| Path | Today |
| --- | --- |
| `backend/internal/services/device_lifetime.go:149-206` | `Update` splits a report into live/delta/sampled, runs automation, metrics, then `publisher.Publish`, then `publishSpatialFreshness`, which forwards **each mapped axis separately**, at most one per axis per 250 ms |
| `backend/models/settings/cache.go:114-135` | `SpatialCoordinates` collects only mapped `x`/`y` names; active flags are excluded |
| `backend/internal/controllers/hub-controller.go:846` | Spatial freshness reuses the general `ws.DeviceUpdated` event |
| `backend/internal/services/observation_publisher.go:72-101` | Ordinary path is policy-gated; coordinates are ESPHome `diagnostic` entities, so the default live interval is 30 s (`backend/models/settings/reporting-defaults.json`) |
| `frontend/src/store/modules/ws/index.ts:57` | `deviceUpdated` → `hub/updateDevice` |
| `frontend/src/composables/spatial/useSpatialTargets.ts` | Tracks arrival time per mapped axis; recomputes targets every second |
| `frontend/src/domain/spatial/index.ts:95-120` | Renders a slot only when both axes are ≤5 s old and the mapped flag is `true`; no filtering |
| `frontend/src/components/device/spatial/SpatialCanvas.vue:187` | Draws raw slot coordinates |

The 30 s policy is already bypassed for mapped axes, so the visible problems are axis
skew plus unfiltered noise plus slot identity, not the diagnostic interval.

### Design — three layers

**1. Backend: complete frames instead of per-axis deltas.**

- New `backend/models/devices/spatial.go` holds the transport-free frame types so the
  device request events and the WebSocket layer can share them without an import cycle:

  ```go
  type SpatialAxis struct { Value float64; Unit string }
  type SpatialTarget struct { ID int; Active bool; X, Y *SpatialAxis }
  type SpatialFrame struct {
      Id      string
      At      time.Time
      Targets []SpatialTarget
  }
  ```

- New `backend/internal/services/spatial_frame.go`: a pure builder with no transport or
  device knowledge.

  ```go
  // BuildSpatialTargets takes the saved mapping and a lookup of accepted entity
  // values; it returns only rows with both axes finite and unit-bearing.
  func BuildSpatialTargets(mappings []settings.SpatialMapping, lookup func(string) (*devices.Entity, bool)) []devices.SpatialTarget
  ```

  Rules encoded here, not in the transport: skip incomplete rows, skip unknown length
  units, never emit axis values for an inactive slot, keep row numbers as slot ids.
- `backend/models/settings/cache.go`: add `SpatialMappings(id) []SpatialMapping` returning a
  copy under the read lock (`SpatialCoordinates` stays for the existing caller or is
  reimplemented on top of it).
- `device_lifetime.go`: replace `publishSpatialFreshness` with `publishSpatialTargets`:
  build the frame from `d.device` after `publisher.Publish`, compare a compact change key
  (slot/active/axis values) to the previous frame, and emit only when it changed.
  Rate limit `spatialFrameInterval = 100 * time.Millisecond`; when suppressed, schedule one
  trailing flush using the same owned-timer pattern as `ObservationPublisher.scheduleLocked`
  (`observation_publisher.go:141`), cancelled in `Close`/`ProviderOffline`. Emit once from
  `Seed` and from `OnConfigUpdated` so a fresh mapping or a new connection renders without
  waiting for the next device report.
- `models/devices/requests.go`: `OnSpatialTargets func(*SpatialFrame)`; remove
  `OnSpatialObserved` once nothing references it.
- `backend/internal/ws/eventhub.go`: `SpatialTargets = "spatialTargets"`;
  `hub-controller.go` wires the callback to `eventHub.Broadcast(ws.SpatialTargets, frame)`.

Ownership and ordering: frames are built only from `Seed`/`Update`, which run on the device
lane after automation and metrics. A saved mapping is picked up by the next report; the
configuration callback does not emit, because it is not serialized with device updates and
could read a torn mix of axes. The broadcast itself runs while the emitter lock is held, as
the observation publisher already does, so a slow socket write can delay the lane but never
reorder accepted state (NH-02, BE-02).

**2. Frontend: frames → filter → tracks → presentation.**

New pure modules under `frontend/src/domain/spatial/`, each with its own companion test:

- `frames.ts` — `parseSpatialFrame(payload)` validates the wire envelope and
  `frameSamples(frame)` converts per-axis units to millimetres using the existing length
  table. Source preference lives in the composable: frames own the view once received, and
  sampled data is the pre-frame fallback.
- `filter.ts` — per-track One Euro filter (Casiez et al., CHI 2012) applied to each axis,
  plus deadband, plausibility gate, gap reset, hold and jitter estimate.

  | Constant | Value | Purpose |
  | --- | --- | --- |
  | `SPATIAL_SMOOTHING_PRESETS` | off / low `{minCutoff 1.7 Hz, beta 0.02}` / medium `{1.0, 0.01}` (default) / high `{0.6, 0.005}` | three user-facing levels, no raw parameter UI |
  | `SPATIAL_FILTER_D_CUTOFF_HZ` | 1.0 | derivative smoothing |
  | `SPATIAL_DEADBAND_MM` | 35 | no visible movement below stationary jitter |
  | `SPATIAL_MAX_SPEED_MM_S` | 3000 | reject teleports; recover after two agreeing samples |
  | `SPATIAL_GAP_MS` | 600 | reset instead of dragging across a gap |
  | `SPATIAL_HOLD_MS` | 2000 | visibly held last position, then removal |

- `tracks.ts` — greedy association with slot stickiness and gate
  `700 mm + 1.5 m/s × dt`; duplicate suppression when two tracks stay within
  `SPATIAL_MERGE_MM = 350` for three frames (keeps the older id, exposes
  `duplicates` count); `SPATIAL_TRACK_RETAIN_MS = 8000`; monotonic `track-N` ids with a
  fixed palette from theme tokens. Labels stay `Track N`; never a person.
- `occupancy.ts` — zone hysteresis over filtered tracks: enter on two consecutive inside
  frames, exit after 800 ms outside, returns `{occupied, source}`. Reuses
  `zoneIsOccupied` semantics for exclusion masks (`workspace.ts:349-363`).
- `summary.ts` — one status line with a strict precedence: offline → mapping missing →
  idle → clear → `N tracks` plus the occupied zone name. The auto-fit viewport helper and the
  compact widget were removed with the always-on card (layout follow-up below).

Composable changes:

- New `frontend/src/composables/spatial/useSpatialTracking.ts` owns frame ingestion, filter
  instances, track association and the 2 s hold; it exposes `tracks`, `rawTargets`,
  `source`, `jitterMm`, `duplicates`, `holdExpired`. It resets on device change, mapping
  change, disconnect and offline.
- `useSpatialTargets.ts` keeps mapping/status ownership (`spatialTargetStatuses`) and
  remains the sampled fallback source; it no longer drives rendering alone.
- `useSpatialView.ts` composes both and passes display state down; drafts, saving and
  undo are untouched.
- `useSpatialPresence.ts` is the single live pipeline shared by the spatial tab (draft
  mapping/workspace) and the device-page summary (saved settings), so both surfaces cannot
  drift apart or run duplicate tracking state.

Presentation (revised after the layout review — the always-on card duplicated the workspace):

- `SpatialCanvas.vue` draws the supplied display tracks: state-dependent styling for
  `live`/`held`/`predicted`, a heading arrow from filtered velocity (short, faded for
  slow movement), and keeps every existing edit gesture, zone, object and trail path
  unchanged. Geometry stays in `domain/spatial/canvas.ts`.
- `DeviceTabIds.Presence` is added to the device tab set and withheld unless
  `supportsSpatialMapping` passes; `resolveDeviceTab` and `adjacentDeviceTab` take the
  available tabs so a withheld link or arrow key cannot land on an empty panel.
- New `components/device/spatial/SpatialPresenceChip.vue`: one `nh-chip` button in the
  device header stating offline / not set up / clear / N targets (from the frame snapshot
  only, never from the tracking pipeline) and opening the tab with a keyboard-operable,
  state-bearing accessible name.
- The compact radar widget and its always-on card are removed: they duplicated the tab's
  canvas. `viewport.ts` and its tests went with them; one canvas per page is now structural.
- `SpatialPanel.vue` becomes the tab body: sticky status + save header, view switcher,
  one display-control row (smoothing applied per sample, raw, grid, trails/room context,
  range, orientation), canvas hero with its readout and track list, then zone/room/entry/
  heatmap tools, with target mapping as the last section. Existing save flow, undo, session
  analytics and guidance are preserved.
- `SpatialHelp.vue` documents slot-versus-track, held/predicted/stale, the lag trade-off
  and the non-claims.

Deliberately not done in this increment: no persisted smoothing preference, no
subscription protocol, no backend filtering, no Home block, no history persistence, no
firmware writes. Each is listed in the spec's exclusions with its reason.

Costs and alternatives: ADR-001 compares per-entity override, per-connection stream,
refcounted live mode, frames and client-only pairing. Chosen: frames, because it removes
skew at the source with no connection bookkeeping and keeps the observation policy
untouched.

## Constitution Check

Results after implementation and the recorded checks; FE-04 records the one item verified by
inspection rather than automated tooling.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | Spec 040 records problem, scope, exclusions, AC-01–AC-14, failures and measurable SC-01–SC-03 before code; artifacts updated after implementation | PASS | PASS (spec/plan/tasks updated) |
| NH-02 | Frames are emitted after automation/metrics from accepted state, never replace `deviceUpdated`, never reorder or suppress automation inputs; held/predicted are labelled, never shown as fresh measurement (AC-05) | PASS (design) | PASS (`device_lifetime_test.go` ordering, frame tests) |
| NH-03 | Additive event on the existing authenticated socket to clients that already receive this device's updates; no new route, no secrets, positions stay in browser memory (NFR-03) | PASS (design) | PASS (review; no new route or storage) |
| NH-04 | Producer/consumer/compatibility in `contracts/spatial-targets.md`; pure builder and pure domain modules keep rules transport-free (BE-01); ADR-001 records the boundary decision and alternatives | PASS (design) | PASS (contract, ADR and code agree; runtime validation at ingress) |
| NH-05 | Every AC maps to a synthetic-sequence unit test or a recorded manual/e2e step; no device commands in routine tests | PASS (design) | PASS (checks above; real-device check unrun and stated) |
| NH-06 | Bounded rate/payload, change-driven emission, trailing timer cancelled on shutdown/offline, no persistence change, no invented history; filter memory is retention-only | PASS (design) | PASS (emitter close/discard tests, race run) |
| BE-01 | `spatial_frame.go` pure; transport wiring stays in controller/ws; frontend rules in `domain/spatial` | PASS (design) | PASS (package tests, no transport in builder) |
| BE-02 | `device_lifetime_test.go` covers accepted-state-before-output ordering, no-op when unmapped, and unchanged automation deltas | PASS (design) | PASS (existing ordering tests still pass; new frame tests) |
| BE-03 | Timer ownership in `DeviceLifetimeService`, cancelled on `Close`/`ProviderOffline`; deterministic fake clock in tests; race run on affected packages | PASS (design) | PASS (emitter close test; `-race` pass) |
| BE-04 | No storage, schema or retention change; `SpatialMappings` returns a copy under the read lock | N/A (no storage change) | PASS (`cache_test.go` copy test) |
| BE-05 | No route, auth or assistant change; payload contains no credentials | N/A | PASS (review) |
| FE-01 | Presentation in components, behaviour in `domain/spatial` + composables, state via existing Vuex modules; one presence pipeline inside the tab, the header chip reads only the frame snapshot | PASS (design) | PASS (review) |
| FE-02 | New event typed in `types/spatial.ts`, mapped in `store/modules/ws/index.ts`, runtime-checked at ingress before state mutation | PASS (design) | PASS (store tests for valid, malformed and unknown-device frames) |
| FE-03 | Source (`frames`/`sampled`), live/predicted/held/idle, offline and no-mapping states are all explicit; no silent "live" claim | PASS (design) | PASS (composable, summary and e2e tests) |
| FE-04 | Presence tab and chip are keyboard operable, the chip's name carries its state, and layouts are checked at 375/768/1440 px | PASS (design) | PARTIAL: chip behaviour is unit tested and the rewritten browser checks assert chip → tab, one canvas and 375 px fit, but that run is pending; real screen-reader behaviour inspected, not tested |
| FE-05 | No new persistence, no household positions in logs or bundles, no `VITE_*` secrets | PASS (design) | PASS (review) |

No rule exception is requested. If implementation needs one, it is recorded here with a
named maintainer approval before merge; agents cannot self-approve.

## Contracts, data, rollout

- Contract: [contracts/spatial-targets.md](contracts/spatial-targets.md). Producer
  `DeviceLifetimeService`, consumer `frontend/src/store/modules/ws/index.ts` →
  `hub/updateSpatialTargets`, payload typed in `frontend/src/types/spatial.ts`.
- Additive only: no persisted setting, schema, route, MQTT topic or MCP surface changes.
  `SpatialMappings` is a new read method on an existing cache.
- Runtime validation at ingress: unknown shapes, non-finite numbers, unknown units and
  more than three targets are dropped rather than committed to state (FE-02).
- Deployment order: backend first or frontend first, both safe. A new browser against an
  old backend reports `sampled`; an old browser ignores the event. No migration, no
  rollback step; reverting either side restores the previous behaviour.
- Recovery: reconnect resumes at the next frame; a device that goes offline ends emission
  and the UI shows offline rather than the last position as current.

## Traceability and verification

Implemented surface: `backend/models/devices/spatial.go`,
`backend/internal/services/spatial_frame.go`, `spatial_emitter.go`, changes in
`device_lifetime.go`, `models/settings/cache.go`, `internal/ws/eventhub.go`,
`internal/controllers/hub-controller.go`; frontend `domain/spatial/{frames,filter,
smoothing,tracks,occupancy,viewport,summary}.ts`, `composables/spatial/{useSpatialTracking,
useSpatialOccupancy,useSpatialPresence}.ts`, `SpatialCanvas.vue`, `SpatialPresenceChip.vue`,
`SpatialPanel.vue`, `DevicePage.vue`, `domain/devices/page.ts`, `useDeviceTabs.ts`, store
modules, help.

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| FR-01, AC-01, AC-06, NFR-01, SC-02 | backend spatial model, builder, emitter, lifetime, cache, ws, controller | `go test ./models/devices/ ./internal/services/ ./internal/ws/ ./internal/controllers/` — builder row/unit/inactive/distinct-name cases, emitter dedupe/rate/flush/revert/discard/close and newest-frame-cancels-pending, lifetime seed/interval/identical/clear/offline/revive, cache copy | pass |
| FR-02, AC-06 | `domain/spatial/frames.ts`, `store/modules/{ws,hub-state}` | `jest src/__tests__/domain/spatial/frames.test.ts`, `src/__tests__/store/modules/ws`, `src/__tests__/composables/spatial/useSpatialTracking.test.ts` — envelope rejection, unit conversion, duplicate slot, frames stay preferred after a sampled coordinate, sampled fallback before the first frame, malformed payload | pass |
| FR-03, AC-02–AC-05, SC-01 | `domain/spatial/filter.ts`, `composables/spatial/useSpatialTracking.ts` | `jest src/__tests__/domain/spatial/filter.test.ts` — ≥50% stationary jitter reduction, ≤150 mm lag at 1 m/s, no >0.5 m step, teleport reject/re-acquire, prediction cap, gap restart, reset, level change on existing state; composable predicted/idle/held tests including removal within 2 s | pass |
| FR-04, AC-07, AC-08 | `domain/spatial/tracks.ts` | `jest src/__tests__/domain/spatial/tracks.test.ts` — slot-swap identity, gate, retention, duplicate merge/reverse, reset | pass |
| FR-05, AC-09 | `domain/spatial/occupancy.ts`, `workspace.ts` | `jest src/__tests__/domain/spatial/occupancy.test.ts` — two-sample entry, exit grace, exclusion mask, inclusive edge, zone removal, "waiting for position data" before any sample | pass |
| FR-06, AC-10–AC-12, NFR-04, NFR-05 | `domain/devices/page.ts`, `useDeviceTabs.ts`, `SpatialPresenceChip.vue`, `SpatialPanel.vue`, `DevicePage.vue` | `jest src/__tests__/domain/devices/page.test.ts` — Presence tab offered only with mappable coordinates, withheld link falls back to Controls, arrows wrap over shown tabs; `jest .../summary.test.ts` — chip headline (not set up / offline / clear / N targets); `playwright e2e/devices.spec.ts` (desktop + mobile) — chip opens the tab, no canvas on Controls, one canvas, sticky header, 375 px fit, and the control row shares one baseline (select bottoms ≤1 px apart, checkbox centre ≤4 px) | pass |
| FR-07, AC-13 | `SpatialPanel.vue`, `SpatialCanvas.vue`, `filter.ts`, `useSpatialTracking.ts` | `filter.test.ts` proves a level change applies on existing filter state; `useSpatialTracking.test.ts` proves a level change replays recent history and produces a different position without a new frame; `e2e/devices.spec.ts` exercises the selector (including that it holds its value), raw/filtered toggle, jitter and duplicate readouts | pass |
| FR-08, AC-14 | `SpatialHelp.vue` | manual read-through against AC-14; states, trade-off and non-claims present, numbers from constants | pass |
| NFR-02 | `filter.ts`, `tracks.ts` | bounded maps (≤3 slots, retention prune, pair counters pruned) asserted in unit tests | pass |
| NFR-03 | store modules, composables, presence tab and chip | review: no persistence, no logging of positions, no new route; frames travel on the authenticated socket | pass |

Checks actually run (repository root, 2026-11-03, `GOCACHE` redirected to a writable
directory because the sandbox blocks the default build cache):

- `go build ./...` — pass.
- `go vet ./...` — pass.
- `go test ./... -count=1` — pass (one earlier run failed in
  `internal/controllers` while the Playwright suite was running a second Go server on the
  same machine; the package passed alone and in a clean full run, so the failure is
  contention, not a regression).
- `go test -race ./internal/services/... ./internal/ws/... ./models/devices/... ./models/settings/... -count=1` — pass.
- `npm test -- --runInBand` — 133 suites / 718 tests pass (final branch state, including
  spec 041 additions and the smoothing replay test).
- `npm run lint` — 0 errors, 28 pre-existing warnings elsewhere in the repository. The
  repository's ESLint config matches only `.js/.cjs/.mjs` and `.vue`, so the new `.ts`
  modules are type-checked by `vue-tsc` (build) and `type-check:e2e` but are not linted;
  that is a repository gap, recorded rather than claimed as coverage.
- `npm run build` — pass (includes `vue-tsc`).
- `npm run type-check:e2e` — pass.
- `npx playwright test --project=desktop` — 140 passed, 1 failed, 9 skipped (final run).
- `npx playwright test --project=mobile` — 149 passed, 1 failed (final run).
- The single failure is `ConnectionsPage.spec.ts` › "weather details retain five-day mobile
  pages and desktop scrolling" ("The weather data is invalid"). It reproduces on a clean
  `main` worktree with none of this branch's changes, so it is a pre-existing/environmental
  failure (date rollover), not a regression. The spatial and device specs pass.
- `npx playwright test e2e/devices.spec.ts` — 18 passed across both projects, including the
  rewritten spatial spec and its alignment assertion.
- Real-device check: not run; routine verification used synthetic frames only.

## Independent review corrections — 2026-11-03

An independent read-only review of the diff found two blockers and several defects. All were
fixed in the branch before completion; the numbered items below match that review.

- **Smoothing was inert (blocker).** The preset was captured when a slot's filter was created,
  so changing the level did nothing until the slot cleared. Smoothing is now resolved per
  sample (`createSpatialFilter()` + `apply(input, smoothing)`), with a regression test that
  applies a new level to existing filter state.
- **A routine sampled coordinate flipped the whole view (blocker).** The source was chosen by
  comparing hub frame time with browser sample arrival, so one throttled coordinate reset
  filters and identities, switched the status line, and could blank the dot. Frames now own
  the view for as long as they arrive, and sampled data is only the fallback before the first
  frame; a regression test asserts the source and label do not change.
- **Emitter could publish an older pending frame after a newer one.** A direct emission now
  cancels a queued flush, with a test that models a late timer.
- **Frames were built from the configuration callback, off the device lane.** That could read
  a torn mix of axes; the callback now only discards the previous frame and the next report
  emits, which the plan states explicitly.
- **The contract claimed a final empty `targets` array.** It now documents the cleared rows
  (`active: false`, no axes) that the builder actually sends.
- **Zone previews could claim "No target inside" before any data arrived.** Preview text now
  distinguishes "waiting for position data" using a `hasData` signal.
- **The offline browser assertion was vacuous.** It now asserts no `Track N (…)` row remains.
- **Age, hold and idle used the hub clock.** They now use browser arrival time, so clock skew
  cannot age positions; the hold window is about 2 s (at most 2.25 s with the 250 ms tick),
  and the tick pauses
  while the document is hidden.
- **Smaller corrections:** heading arrows only on live tracks; one mapping-completeness and one
  binary-flag rule shared by both paths; zone-occupancy text alternative and a state-summary
  accessible name on the then-current compact surface (since replaced by the tab and chip);
  jitter reported as recent and reset across a gap or a smoothing replay;
  diagnostic row copy and formatter; contract wording for distinct names and unreadable flags.

## Layout review follow-up — 2026-11-03

The device page showed a compact radar card above Controls for every mapped sensor, and the
Presence workspace then showed a nearly identical, larger radar. Both surfaces were correct in
isolation; together they looked like one view duplicated and made the radar compete with
readings. Resolution:

- Presence becomes its own device tab, offered only to devices that can map coordinates.
- The device header carries a state-bearing chip (offline / not set up / clear / N targets)
  that opens the tab; it reads the frame snapshot only, so it cannot disagree with the tab
  about tracks, hold or hysteresis.
- The always-on radar card, its widget component and the auto-fit viewport helper were
  removed rather than kept unused; the tab's canvas is the single radar.
- The tab body is laid out as a workspace: sticky status and save header, view switcher,
  one control row, canvas hero with readout, tools, and mapping last.

## Smoothing control follow-up — 2026-11-03

Changing the level appeared to do nothing. The select itself was binding correctly (browser
test asserts it holds its value) and the filter honoured a per-sample preset (unit test), but
frames are change-driven: with a still target there are no new samples to smooth, so the
control really was inert in that situation. The tracking composable now keeps a bounded recent
raw window per slot (3 s, 40 samples) and replays it, shifted to end now, through a fresh
filter when the level changes — the position updates immediately from recent motion. Jitter is
reported as "recent" rather than "since open" because the replay restarts that estimate.

## Implementation notes and deviations

- Initial design said `stale`; implemented as `idle` (aged) because frame silence is normal
  for a still target, and dropping the dot would misreport presence. Offline remains the
  only "not current" state.
- Source changes (frames ↔ sampled) reset filters and identities, otherwise a switch shows a
  phantom jump and a new label.
- Held tracks are drawn but excluded from zone occupancy and the status count: a cleared slot
  must not keep a zone or the summary reporting presence.
- Mapping statuses are driven by the slots a fresh frame draws (`liveSlots`), not by sampled
  coordinate arrival: the ordinary observation policy throttles those entities, so the old
  freshness check would label every working row "Waiting for fresh X/Y".
- Once a frame arrives the view stays on frames; a routine sampled coordinate cannot flip the
  source, relabel tracks or blank the dot. Sampled data remains the fallback for a server that
  never sends frames.
- Smoothing is resolved per sample, so changing the level affects the next frame instead of
  waiting for the slot to clear. A level change also replays the recent raw window (3 s,
  bounded), because a still sensor emits no new frames and the control would otherwise look
  inert.
- Age, hold and idle decisions use browser arrival time, not the hub's frame timestamp, so a
  clock-skewed server cannot make positions look stale or eternally fresh.
- `SpatialCoordinates` and `OnSpatialObserved` were replaced rather than kept alongside the
  new path; nothing else referenced them.

## Debt, risks, exceptions

- **Debt (pre-existing, not fixed):** ESLint does not cover `.ts` files in this repository,
  so lint evidence cannot speak for the new TypeScript modules; `vue-tsc` does.
- **Debt (pre-existing, touched, not fixed):** `measurementWhitelist` promotes the entity named
  `target_distance` to measurement while sibling coordinates stay diagnostic
  (`backend/models/devices/device.go:326-337`). Consequence: those entities refresh on
  different policies. This plan does not change the whitelist (it would alter bridge
  devices); frames bypass the policy for mapped axes instead. Follow-up: classify spatial
  coordinates from the explicit mapping rather than names.
- **Risk:** duplicate suppression can hide a genuine second person within 350 mm. Mitigation:
  reversible merge, visible duplicate count, and help text; never used for automation.
- **Risk:** smoothing level may suit one room and not another. Mitigation: three named
  levels, raw/filtered toggle, jitter readout.
- **Risk:** always-on frames for mapped devices add traffic. Mitigation: change-driven,
  10 Hz cap, ≤1 KB, no frames when idle; reconsider with ADR-001 if measurable.
- **Risk:** track retention may briefly carry a stale identity after someone leaves.
  Mitigation: 2 s visible hold, 8 s association memory, cleared on offline.
- **Scope note:** the smoothing preference is session-local by decision (Q-02); persisting
  it is an additive follow-up, not a hidden gap.
- **Scope note:** Home dashboard embedding is blocked by the Home block/source model and
  is deliberately deferred (Q-03).
- No constitutional exception is requested. No named maintainer approval is claimed.

## Completion

Implementation, checks and artifact updates are complete for this increment; independent
review is the remaining step. Deferred items from the spec's exclusion list remain out of scope and each needs
its own spec.

Changed documentation: `SpatialHelp.vue` (tracks, states, smoothing trade-off, limits),
contract and ADR in this directory. The repository map needs no change: no new area or
dependency was introduced.

Known limits stated in the UI and here: duplicate suppression can hide a second person
within 350 mm; smoothing adds lag proportional to the level; a still target ages to `idle`
instead of being removed, so a device that stays online while reporting nothing keeps showing
the last position as aged rather than cleared; the jitter and duplicate readouts describe the
last received data; positions are best-effort and never a person's identity.
