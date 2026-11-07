# Spec: Positions viewer layout and playback legibility

ID: 057 | Status: Implemented (unreviewed) | Owner: unassigned | Updated: 2026-11-06
Constitution: 1.0.0 (proposed) | Components: frontend

## Problem, scope, exclusions

The device History → Positions viewer is hard to use on desktop and mobile:

1. **Selecting a recording loses your place.** The template swaps the whole body for
   its loading branch (`v-else-if="loading"` in `SpatialPositions.vue`) whenever the
   selected session's points are fetched. Canvas, recordings list and player unmount,
   page height collapses, and the browser clamps scroll to the top. The transport is
   last on the page, so the user must scroll back down for every selection.
2. **The timeline has no session context.** The scrubber is bounded to the selected
   session, but neither the session's identity nor its position inside the 1h/24h/7d
   window is shown. Playback looks like it might cover the whole range.
3. **Movement is easy to miss.** The path is a single polyline at `stroke-opacity: .35`;
   a target disappears from `tracks` 30 s after its last point; playback advances time
   in 200 ms steps with no motion cue. After a long gap the dot simply appears
   somewhere else, with no indication that it moved from the previous sample.
4. **Supporting defects.** `loadSessions()` auto-selects `sessions[0]` while the backend
   sorts sessions oldest-first (`backend/internal/spatialhistory/service.go`), so the
   oldest recording opens. `distanceMm` sums the entire session rather than the portion
   up to the playhead. The page has no media queries and the recordings list is a 14 rem
   inner scroller.

Facts above are read from the current source; no runtime measurements have been taken.

Outcome: a viewer where the map and its transport stay together and in view, the
selected recording is always identifiable, both time scopes (range and recording) are
visible, and movement cannot be missed.

In scope:

- desktop layout with a recordings rail and a transport anchored to the map;
- stable re-render when a recording is selected (no page scroll change, no unmount);
- a session-coverage timeline for the preset window plus a session scrubber;
- movement episodes, a next-movement action, recent-trail emphasis, hop connector,
  pulse cue, and honest "last known / no movement recorded" idle presentation;
- a collapsible recordings rail and a wide, fixed-height viewer with 1×/2×/3×/5×
  content zoom;
- one shared View panel above every spatial canvas (zoom, grid, field of view, display
  options) plus pointer/touch pan;
- played-distance readout through `utils/units.ts`;
- responsive behaviour at phone, tablet and desktop widths;
- loading, empty and failure states that do not destroy confirmed content;
- domain unit tests and browser coverage.

Exclusions:

- backend, recorder, sampling, retention or API changes; all new data is derived in the
  browser from `SpatialSession`/`SpatialPoint` already returned by
  `services/spatial-history.service.ts`;
- interpolating or inventing positions between samples;
- per-session heatmaps, clip creation, identity or cross-sensor path merging;
- deep-linking range/session/time in the URL (existing `?tab=history` behaviour is
  unchanged; recorded as a follow-up, not a requirement);
- changes to the live Presence surfaces that also use `SpatialCanvas.vue`.

## Scenarios and acceptance

### US-01 — Choosing a recording does not move the page (P1)

Independent verification: Playwright measurement of `window.scrollY` and mounted
controls before/after selection, plus failure injection of the points request.

- AC-01: Given the Positions player is scrolled so the transport is off-screen, when a
  recording row is selected, then `window.scrollY` changes by at most 2 px and the
  canvas, coverage timeline and transport remain mounted.
- AC-02: Given a session switch is in flight, when points load, then the previous
  content keeps its layout and is marked busy (`aria-busy`); the body is not replaced
  by a loading paragraph.
- AC-03: Given the recordings rail overflows, when the user selects a row, then the
  rail scrolls independently and brings the selected row into view inside the rail
  only.
- AC-04: Given the viewer opens, then the newest session in the range is selected, and
  the rail lists recordings newest-first with each local time range.
- AC-05: Given a session switch fails, then the last confirmed session stays rendered,
  the rail keeps the confirmed selection, and an alert announces the failure; the
  failed selection is not presented as loaded.

### US-02 — The timeline says what is playing (P1)

Independent verification: browser assertions on segment count, accessible names,
selection and seek; domain tests for proportional layout.

- AC-06: Given sessions inside the preset window, then the coverage timeline renders one
  proportional segment per session across the recorded envelope (the preset window when
  none are recorded), gaps between sessions are visibly empty, and the active session is
  distinguished by fill, outline and accessible state rather than colour alone.
- AC-07: Given a coverage segment, then activating it (pointer, Enter or Space) selects
  that session and seeks to its first recorded point; its accessible name includes the
  local time range, duration, point count and slots.
- AC-08: Given a selected session, then the scrubber is bounded to that session, shows
  the session start/end labels and elapsed/total time, and renders the played portion
  with both a fill and text.
- AC-09: Given playback or a seek, then the coverage timeline playhead, the scrubber and
  the canvas all reflect the same time on the next render.

### US-03 — Movement is visible and skippable (P1)

Independent verification: browser assertions on the next-movement action and cues;
domain tests for episode/span derivation.

- AC-10: Given a session with gaps between recorded samples, then movement episodes are
  derived and marked on the scrubber, and "Next movement" seeks to the first episode
  at or after the playhead and is disabled when none remain.
- AC-11: Given the playhead is inside a long gap, then the last known target stays
  visible, is labelled with its last sample time, and the transport states that no
  movement has been recorded for the elapsed duration.
- AC-12: Given the first sample after a gap (or a jump beyond the movement threshold),
  then a dashed connector is drawn from the previous sample and the target emits a
  pulse when a new sample becomes current.
- AC-13: Given `prefers-reduced-motion: reduce`, then pulse/connector animation is
  suppressed and a static connector plus text still communicate the movement.
- AC-14: Given a playhead position, then the distance readout reports distance travelled
  up to that playhead in metres/kilometres via `formatDistance`/`formatQuantity`, and
  the transport shows the whole-session total beside it.

### US-04 — Usable on phone, tablet and desktop (P2)

Independent verification: Playwright at 390×844, 768×1024 and 1440×900; keyboard pass.

- AC-15: Given a viewport narrower than the rail breakpoint, then recordings are
  reachable from a toolbar button that opens the existing bottom-sheet dialog, and
  selecting a recording closes the sheet without moving the page.
- AC-16: Given 390 px width, then no horizontal overflow occurs, interactive targets are
  at least 44×44 px, the transport sits in normal flow directly below the coverage
  timeline, and neither uses `position: fixed` or an overlay that hides the map or
  timeline; the canvas remains at least 40 dvh tall.
- AC-17: Given 768 px and 1440 px widths, then the viewer has no horizontal overflow,
  the coverage timeline and transport stay in flow and are reachable by ordinary
  vertical scrolling, and the Heatmap view plus the live Presence surfaces are
  unchanged.

### US-05 — Empty, loading and failure states (P2)

- AC-18: Given no sessions, then the coverage timeline shows an empty range, the
  transport is disabled, and the copy explains how to record movement.
- AC-19: Given a selected session with no points, then Play is disabled, no path or
  target is fabricated, and the state is distinguished from "no sessions".
- AC-20: Given sessions load but heatmap fails (or vice versa), then the failure is
  announced and the successfully loaded part remains usable.

### US-06 — Control how much space and zoom the viewer uses (P1)

Independent verification: Playwright bounding boxes, the SVG `viewBox`, and pressed-state
assertions on the controls.

- AC-21: Given a wide viewport with the recordings rail visible, when it is hidden, then
  the rail is removed, the toolbar offers “Show recordings (N)”, and restoring it keeps
  the same confirmed recording and playhead.
- AC-22: Given the Player or Heatmap, then the viewer spans its column width and its
  height stays within `clamp(16rem, 40vh, 24rem)`; 1× fits the room and the confirmed
  session's points; 2×/3×/5× magnify around the active sample captured when the level
  changes or the playhead is moved (the view stays fixed during playback), the target
  grows modestly, the device is pinned to the nearest view edge when it moves off-screen,
  and labels and markers stay readable; hiding the rail widens the viewer without
  changing its height.
- AC-23: Given the Field of view checkbox in Player or Heatmap, when the workspace FOV is
  above 0 it shows or hides that angle; when none is configured it shows or hides a
  clearly-labelled illustrative 90° cone. Neither path changes the saved angle.

### US-07 — One View panel and pan across spatial views (P1)

Independent verification: browser checks for the panel on Live and Positions, plus a
drag/reset assertion on the canvas.

- AC-24: Given Live, Draw (Zones/Room/Entry), Positions Player or Heatmap, then the canvas
  is preceded by the shared View panel with zoom, Grid and Field of view; Live and Draw
  additionally group Smoothing/Scale/Orientation/Raw in the same panel; Setup forms and
  the Home rail preview render no panel.
- AC-25: Given a wide viewport, the panel is expanded; on phones it starts collapsed and
  expands to a stacked layout with 44 px targets and no horizontal overflow.
- AC-26: Given a read-only spatial view, dragging with a mouse or one finger pans; in an
  editing view the Pan tool (or Space while the viewer is focused) suspends editing and
  pans; Reset view returns to zoom 1 and the fitted view; pan clears on view or session
  changes.

## Requirements and success

- FR-01: Session selection MUST update content in place, without unmounting the viewer
  or changing page scroll; acceptance: AC-01, AC-02.
- FR-02: Recordings MUST be navigation outside the map/transport flow, MUST default to
  the newest session, and MUST list newest-first; acceptance: AC-03, AC-04.
- FR-03: A failed session switch MUST retain the confirmed session and announce the
  failure; acceptance: AC-05, AC-20.
- FR-04: The recorded envelope within the preset window MUST be visualised as one
  proportional segment per session with explicit gaps and an operable active state;
  acceptance: AC-06, AC-07, AC-09.
- FR-05: The selected session MUST have its own bounded scrubber with start/end labels
  and non-colour progress; acceptance: AC-08, AC-09.
- FR-06: Movement episodes MUST be derivable from recorded samples and expose a
  next-movement seek; acceptance: AC-10.
- FR-07: A target MUST NOT disappear during gaps; the UI MUST say "no movement
  recorded" rather than claiming a measured still; acceptance: AC-11.
- FR-08: Sample changes after a gap MUST be signalled with a connector and pulse,
  suppressed under reduced motion; acceptance: AC-12, AC-13.
- FR-09: Distance MUST be scoped to the playhead and formatted through
  `utils/units.ts`, with the session total shown in the transport; acceptance: AC-14.
- FR-10: The layout MUST adapt at phone/tablet/desktop widths, use the existing
  bottom-sheet dialog below the rail breakpoint, and keep the player in normal flow
  without covering the map or coverage timeline; acceptance: AC-15, AC-16.
- FR-11: Heatmap, live Presence and other History modes MUST be unchanged;
  acceptance: AC-17.
- FR-12: Loading/empty/failure states MUST preserve confirmed content and never present
  missing samples as zero; acceptance: AC-02, AC-18, AC-19, AC-20.
- FR-13: New controls MUST be keyboard operable, named, and visibly focused;
  acceptance: AC-07, AC-13, AC-16.
- FR-14: The recordings rail MUST be collapsible on wide viewports and restorable from
  the toolbar without changing the confirmed recording or playhead; acceptance: AC-21.
- FR-15: The viewer MUST be full-width and height-clamped so the timeline and transport
  stay reachable; 1× MUST fit the room and session points; zoom MUST scale the map
  contents around an anchor captured on level or playhead changes, keep the view fixed
  during playback, grow the target modestly, keep the device visible at the edge, keep
  labels/markers readable, and leave the panel size unchanged; acceptance: AC-22.
- FR-16: The player MUST offer a display-only field-of-view toggle; it MUST use the
  configured angle when present, otherwise a labelled illustrative 90° fallback, and MUST
  NOT modify the saved angle; acceptance: AC-23.
- FR-17: One shared View panel MUST precede every full spatial canvas (Live, Draw,
  Positions, Heatmap) with zoom, Grid and Field of view; display controls MUST live in the
  same panel; Setup and preview surfaces MUST render none; acceptance: AC-24.
- FR-18: The panel MUST collapse and expand, start expanded on wide screens and collapsed
  on phones, be keyboard operable with 44 px targets, and never overflow horizontally;
  acceptance: AC-25.
- FR-19: Read-only viewers MUST pan with mouse/touch drag; editing viewers MUST pan only
  with the Pan tool or Space; Reset MUST restore zoom 1 and the fitted view; pan MUST
  clear on view or session changes; acceptance: AC-26.

- NFR-01: Selecting a recording changes page scroll by at most 2 px at 390, 768 and
  1440 px widths.
- NFR-02: No horizontal overflow at 390, 768 and 1440 px; the existing document
  overflow assertion remains green.
- NFR-03: Playback cadence stays at the existing 200 ms tick; derived helpers are O(points)
  and add no new dependency.
- NFR-04: Times use the viewer's local timezone consistently with the existing labels;
  distances convert millimetres to metres/kilometres through `utils/units.ts`, never by
  string concatenation.
- NFR-05: The coverage timeline must remain operable with up to 500 sessions and the
  scrubber with up to the API's 5,000-point session cap; segments below a minimum width
  get an expanded hit area instead of being dropped.
- NFR-06: No new browser persistence, logging or network calls; household positions stay
  in component memory only.
- NFR-07: The viewer width MUST equal its column (within 2 px) and its height MUST stay
  within 16–24 rem; changing zoom MUST NOT resize the panel or make labels/markers
  unreadable.

- SC-01: `frontend/e2e/devices.spec.ts` proves scroll stability, rail/sheet selection,
  coverage-segment selection, next-movement and the absence of overflow.
- SC-02: `frontend/src/__tests__/domain/spatial/playback.test.ts` covers proportional
  segments, movement episodes, played distance and gap handling.
- SC-03: `npm test -- --runInBand`, `npm run lint`, `npm run build` and
  `npm run type-check:e2e` pass with recorded results.
- SC-04: Keyboard and viewport verification is recorded for 390, 768, 1280 and 1440 px,
  including keyboard activation of the coverage segments and scrubber. Reduced-motion
  visual verification remains an explicit review item.
- SC-05: `frontend/e2e/devices.spec.ts` proves rail collapse/restore, the fixed-height
  wide viewer, 1×/2×/5× zoom with the active point in frame and the device pinned, and
  the enabled FOV toggle on desktop and mobile projects.
- SC-06: `frontend/e2e/devices.spec.ts` proves the shared View panel on Live and Positions,
  its zoom/Grid/FOV controls, drag panning with Reset, and no horizontal overflow at phone
  and desktop widths.

## Entities and external interfaces

Existing, unchanged:

- `SpatialSession { id, from, to, points, slots, capped }` and
  `SpatialPoint { at, slot, x, y }` (`frontend/src/services/spatial-history.service.ts`);
- `SpatialCanvas` props consumed by `SpatialPositions.vue`, `SpatialPanel.vue` and
  `PresenceRailTile.vue`.

New, browser-derived only:

- session segment: id, start/end ms, percentage placement inside the preset window;
- movement episode: sample time, slot, elapsed gap since the previous sample;
- played distance: per-slot sum of sample-to-sample distance up to a time;
- gap span: interval between consecutive recorded samples, presented as
  "no movement recorded", never as measured stillness.

No API, MQTT, WebSocket, persistence or configuration contract changes.

## Questions and assumptions

- Q-01: A movement episode starts at the first sample after a time gap. Gap threshold is
  assumed to be 10 s (the recorder only stores samples on ≥300 mm movement, so time is
  the discriminating signal). Assumption to validate against real recordings; it is a
  presentation constant, not a contract.
- Q-02: The rail lists newest-first while the coverage timeline stays chronological;
  the two orders serve different tasks and both show explicit times.
- Q-03: Range/session/time deep-linking is excluded from this increment.
- Q-04: Played progress uses accent tokens, not `--nh-danger`. Red is the error/exclusion
  token and FE-04 forbids colour-only status.
- Q-05: No interpolation between samples; motion is communicated by connectors, pulses
  and path emphasis only. Interpolation would depict movement the sensor did not report.
- Q-06: These are not device or API changes; the existing "latest point at or before
  `time`" rule is kept for rendering.
- Q-07: Direction was confirmed by the requester during design discussion on 2026-11-06
  (recordings rail plus a transport under the map; all four playback aids). Implementation
  dropped the initially discussed sticky transport when it covered the coverage timeline
  on short viewports; the transport now stays in normal flow. This is not a formal
  spec/plan review or maintainer approval.
