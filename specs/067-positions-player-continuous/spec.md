# Spec: Positions player — continuous playback across recordings

ID: 067 | Status: Draft | Owner: agent draft for maintainer review | Updated: 2026-11-20
Constitution: 1.0.0 (Proposed) | Components: both (additive read contract + frontend)
Follows: [spec 057](../057-positions-viewer-ux/spec.md), [spec 056](../056-spatial-history/spec.md).
This spec states outcomes. It does not claim ratification, reviewer approval, or
implementation authorization.

## Problem, scope, exclusions

### Problem

The Positions player (device → History → Positions) plays **one recording at a time**.
`SpatialPositions.vue` bounds playback to the confirmed session
(`bounds = { from, to }` of that session) and `useSpatialPlayback.ts` stops at
`range.to`; `loadPoints()` then `pause()`s and seeks to the session start. A day of
movement is usually many short chunks in the rail (the recorder flushes every
5 minutes / 300 points — `backend/internal/spatialhistory/recorder.go`), so:

1. **There is no continuous play.** Pressing Play always restarts the selected chunk, and
   it stops after a few seconds. To watch an afternoon the operator must select each
   chunk and press Play again. Pause does not "resume from where it left off" across a
   recording boundary because playback never crosses one.
2. **There is no way to jump between recordings.** The transport offers "Next movement"
   *inside* the selected recording; there is no "next recording" / "previous recording".
   The rail and the range timeline can select another recording, but only by leaving
   the transport.
3. **The range timeline does not say where movement is.** `PositionsCoverageBar.vue`
   draws one proportional segment per recording from its `from`/`to` only. With a 24 h
   filter the bar says which chunks exist but nothing about where anything moved, so
   the operator clicks segments blindly and loads them one by one.
4. **The scrubber's own marks are unlabelled pixels.** `PositionsTransport.vue` renders
   gap stripes and episode ticks from `sampled(...)`, capped at 200 unlabelled
   `<span>`s. They show movement exists but cannot be read, hovered or activated.

Facts above are read from current source (`SpatialPositions.vue`, `useSpatialPlayback.ts`,
`PositionsCoverageBar.vue`, `PositionsTransport.vue`, `domain/spatial/playback.ts`,
`backend/internal/spatialhistory/{service,recorder}.go`). Research into comparable
players (Netflix/YouTube `Up next` + autoplay, chapter markers on the scrubber,
next/previous chapter) supports treating recordings as chapters of one timeline rather
than as unrelated files. No runtime measurements were taken.

### Outcome

The Positions player behaves as **one continuous, navigable timeline** over the
filtered recordings:

- Play continues from where it is through the remaining recordings in the filter, in
  chronological order, and Pause resumes at the same place.
- Next/Previous recording jump the player to the adjacent recording (and start playing
  if it was playing).
- The range timeline shows each recording's time span **and where movement was
  recorded inside it** as accessible marks; activating a mark plays from that moment, and
  clicking elsewhere in the recording plays from the nearest recorded position.
- The active recording's scrubber marks are labelled and keyboard-activatable; the
  current recording and its neighbours are identifiable without reading the rail.

### In scope

- Additive `samples` array on the existing `mode=sessions` payload (when the recording
  stored a position inside the requested range), produced by the existing service and
  reduced evenly to a bounded size like the other read models.
- A playback duration model: the union of the filtered recordings inside the preset
  window, ordered by start, with gaps addressed by an explicit "skipped" transition
  rather than silent acceleration.
- A continuous controller that advances across recording boundaries, keeps playing
  state and speed, and reports the recording it is in and the next/previous one.
- Transport additions: Play/Pause, Previous recording, Next recording, speed, Path,
  distance and the session scrubber (unchanged controls keep their behaviour).
- Range timeline: one accessible, focusable mark per recorded position (sampled to a
  stated bound when dense), plus click-to-play from the nearest position for the rest of
  the recording, the active recording, and the playhead; no per-session network fan-out.
- A visibly-labelled chapter model for the scrubber's movement ticks.
- Domain and composable unit tests plus browser coverage in `frontend/e2e/devices.spec.ts`.

Bundled on this branch but outside the positions behaviour above: the ESPHome
connection-status vocabulary (`connecting`), the history chart appearance/unit
formatting, and the ESPHome dialog lifecycle fixes. They are recorded in
`clean-code-remediation.md`, covered by their own tests, and do not change this spec's
acceptance criteria.

### Exclusions

- No recorder, sampling, retention, storage or routing change; `samples` is derived from
  chunks already read by `Sessions()`. No new endpoint or mode.
- No interpolated positions between samples (spec 057 Q-05 stands).
- No playback inside the Heatmap view; no per-movement thumbnails or previews.
- No change to the live Presence surfaces or to `SpatialCanvas` rendering contracts.
- No browser persistence of playhead, speed or selected recording.
- No autoplay-on-load; playback starts only from an explicit Play.
- No changes to spec 066 (wireless viewer) or the observation API.

## Scenarios and acceptance

### US-01 — Play the whole filtered history continuously (P1)

Independent verification: browser test against the multi-session mock fixture; composable
unit tests with a fake clock.

- AC-01: Given the newest recording is completed, when playback is playing, then the
  player loads the next recording in the filter (chronologically) and continues from
  its first movement without requiring a click; Play/Pause state stays "playing".
- AC-02: Given playback is paused inside a recording, when Play is pressed, then
  playback resumes at the paused instant and does not restart at the recording's first
  point.
- AC-03: Given playback crosses a recorded gap between two recordings, then the
  transport states the gap was skipped (start/end times and duration) and playback
  resumes at the next recording's first sample; the playhead is never shown inside an
  empty interval as if it were recording.
- AC-04: Given the last recording in the filter is completed, then playback stops and
  the transport says it reached the end; it does not loop or fabricate data.
- AC-20: Given the viewer opens on the newest recording (spec 057 AC-04), when
  “Play all” is activated, then playback starts at the earliest recording in the filter
  and runs continuously to the latest, so one action plays the whole filtered history.
- AC-05: Given a recording fails to load during continuous playback, then the previous
  confirmed recording stays rendered, the failure is announced, playback stops (no
  silent skip to a later recording) and the failed transition is not presented as played.

### US-02 — Jump between recordings (P1)

Independent verification: browser test on Previous/Next recording and boundary state;
composable unit tests for neighbour resolution.

- AC-06: Given a confirmed recording, when Next recording is activated, then the next
  recording in chronological order inside the filter becomes current at its first
  sample, and playback continues if it was running or stays paused if it was not.
- AC-07: Given a confirmed recording, when Previous recording is activated, then the
  previous recording in chronological order becomes current at its first sample; Previous
  and Next always move exactly one recording, so repeated presses stay predictable.
- AC-08: Given the first or last recording in the filter, then Previous or Next is
  disabled respectively and exposed as such to assistive technology.
- AC-09: Given Next/Previous from the keyboard, or the `Shift+Right`/`Shift+Left`
  shortcuts, then the same transition occurs and the transport control keeps focus.

### US-03 — The range timeline shows where movement is (P1)

Independent verification: browser assertions on the movement marks, their accessible
names and play-from; domain tests for proportional placement and capping.

- AC-10: Given the sessions response carries `samples`, then every recording segment
  on the range timeline renders a focusable mark for each recorded position, placed
  proportionally inside that segment (no per-session request), and movement inside a
  24 h filter is discoverable without selecting recordings.
- AC-11: Given a recorded-position mark, then its accessible name states the local time
  and the recording it belongs to; activating it (pointer, Enter or Space) selects that
  recording and starts playback at that mark's moment.
- AC-12: Given a recording has no recorded movement or the backend does not send
  `samples` (older hub), then the segment renders no marks and nothing is presented as
  measured; the segment label and selection behaviour are unchanged.
- AC-13: Given a segment in the range timeline, then its group accessible name states the
  time range, duration, points and slots, and additionally the recorded-position count;
  each mark's name states its own time.

### US-04 — Read the scrubber's movement marks (P2)

Independent verification: browser assertions on mark labels; domain tests for chapter
derivation and sampling.

- AC-14: Given movement episodes inside the active recording, then the scrubber renders
  labelled markers (not bare decorative spans) that are focusable, name their time and
  state they are movement, and seek when activated.
- AC-15: Given a recording with more episodes than the display cap, then the rendered
  marks are deterministically sampled and the count of episodes is stated in text, so a
  dense recording does not look empty or fabricate precision.
- AC-16: Given the playhead, then the current recording, its position among the filtered
  recordings (e.g. "2 of 5") and its local time range are always visible in the
  transport.

### US-05 — Unchanged behaviour, phones and failures (P2)

Independent verification: existing Positions browser assertions plus viewport checks.

- AC-17: Given no sessions or a failed sessions request, then the existing empty/error
  behaviour is unchanged and the new controls are absent or disabled, never enabled
  against no data.
- AC-18: Given 390 px width, then the transport controls wrap without horizontal
  overflow, the transport's buttons are at least 44×44 px, the range-timeline segment
  and movement marks remain keyboard operable with a visible focus ring (their width is
  bounded by the recording's proportional span, so they are not forced to 44 px wide),
  and the player stays in normal flow (no fixed overlay covering the map or timelines).
- AC-19: Given Heatmap or a live Presence view, then its behaviour is unchanged.

### US-06 — The player reads as a timeline, not a form (P1)

Independent verification: full-page screenshots at 390/1024/1440 px in dark and light; browser
assertions on the banded marks and the axis labels.

- AC-21: Given a filter whose recordings are much shorter than the filter window, then the
  range head states the recorded span it is actually drawn over and names the filter, so the
  axis and the 1h/24h/7d control cannot appear to contradict each other.
- AC-22: Given the current recording's samples, then its segment draws one focusable mark
  per recorded position, placed proportionally, and renders nothing where nothing was
  recorded. A segment narrower than 28 px is still drawn at 28 px so it can be read and
  its marks targeted. Dense recordings are sampled to a bound (stated in the visible count)
  rather than producing an unbounded tab order, and the segment's accessible name states
  the true sample count.
- AC-23: Given the transport, then exactly one element is visually primary (Play/Pause), the
  clock leads the row at one step above body size (not a monument), actions and metrics are two
  groups rather than one flat row, totals are stated once per scope (filter head vs "in this
  recording"), the transport is at most 0.6 × the canvas height, and idle / skipped-gap /
  movement-count / nothing-to-scrub states share one status line instead of stacking rows.
- AC-24: Given the range timeline, then the selected recording is distinguished by a leading
  edge bar plus an accent fill (not colour alone, and not the same treatment as a movement
  mark; `aria-current` carries the state to assistive technology), and no second playhead is
  drawn there.
- AC-25: Given a recording shorter than a second or holding one sample, then its row states
  "1 sample" instead of "0 s · 1 points", the transport says there is nothing to scrub rather
  than drawing a dead scrubber, and under an hour the rows include seconds so two recordings
  cannot print the same range. The newest recording is still selected by default (spec 057
  AC-04); such a recording explains itself instead of looking broken.
- AC-26: Given a recording row in the range timeline, then activating a recorded-position
  mark selects that recording and starts playback at that mark's moment; clicking elsewhere
  in the row starts playback at the recorded position nearest the click. Marks are real
  focus stops with padded targets, not decorative ticks.

## Requirements and success

- FR-01: Playback MUST advance chronologically across the filtered recordings and
  report each recording boundary explicitly; acceptance: AC-01, AC-03, AC-04.
- FR-02: Pause MUST preserve the playhead across recordings, and Play MUST resume from
  it; acceptance: AC-02.
- FR-03: Previous/Next recording MUST move to the adjacent filtered recording, be
  disabled at the ends, and be keyboard operable; acceptance: AC-06…AC-09.
- FR-04: A failure while loading the next recording MUST stop playback, keep the
  confirmed content and announce the failure; acceptance: AC-05.
- FR-05: `mode=sessions` MUST carry every recorded timestamp inside the requested range,
  exactly as `mode=points` does; it MUST be additive, evenly reduced to a bound, ordered
  ascending, and MUST NOT change existing fields; acceptance: AC-10, AC-12.
- FR-06: The range timeline MUST render an accessible, focusable mark per recorded
  position (sampled to a stated bound when dense), and activating one MUST select its
  recording and play from that moment; acceptance: AC-10, AC-11, AC-22.
- FR-07: The scrubber's movement marks MUST be labelled, focusable and activatable, and
  their sampling MUST be stated; acceptance: AC-14, AC-15.
- FR-08: The transport MUST always show the current recording's identity, local range
  and its position among the filtered recordings; acceptance: AC-16.
- FR-09: Existing empty/error/Heatmap/Presence behaviour MUST be unchanged;
  acceptance: AC-17, AC-19.
- FR-10: Controls MUST be keyboard operable, named, visibly focused, 44 px on phones,
  and MUST NOT overflow horizontally; acceptance: AC-09, AC-18.
- FR-11: One action MUST play the whole filtered history from its earliest recording;
  acceptance: AC-20.
- FR-12: The recorded canvas MUST keep spec 057's `clamp(16rem, 40vh, 24rem)` height (it is
  NOT narrowed; an earlier draft shrank it to `clamp(11rem, 26vh, 17rem)` and made the player,
  not the record, the biggest thing on the page). Instead the transport MUST stay proportionate
  to the canvas by collapsing its states into one status line and sizing its controls like the
  page's other dense controls: transport height ≤ 0.6 × canvas height. The browser guard
  measures at 1440×800 and at 1280×600, the latter being the strictest case because the
  canvas has reached its 16 rem floor while the transport keeps a fixed control height;
  acceptance: AC-23, SC-04.

- NFR-01: One sessions request must not grow into per-recording points requests;
  movement marks for the whole filter come from the single response.
- NFR-02: `samples` MUST be bounded (≤ 240 per session, sub-millisecond derivation
  from the points already iterated) and MUST NOT change the 500-session / 5,000-point
  caps.
- NFR-03: Playback cadence stays at the existing 200 ms tick; continuous transitions
  add no timers that outlive the composable scope.
- NFR-04: Times use the existing local-time labels; distances continue through
  `utils/units.ts`; a gap is labelled as skipped, never as measured zero movement.
- NFR-05: No new dependency, browser persistence, logging or household data in the
  bundle beyond the existing recording summary.

- SC-01: `backend/internal/spatialhistory/service_test.go` and
  `backend/internal/api/spatial_history_test.go` cover movement derivation, ordering,
  capping and payload compatibility.
- SC-02: `frontend/src/__tests__/domain/spatial/playback.test.ts` covers recording
  duration/neighbour/gap/chapter helpers.
- SC-03: `frontend/src/__tests__/composables/useSpatialPlayback.test.ts` covers
  continuous advance, resume, boundary flags and failure.
- SC-04: `frontend/e2e/devices.spec.ts` covers continuous play, resume, previous/next,
  the range timeline's per-movement marks and play-from, and no horizontal overflow.
- SC-05: `go test ./...`, `go build ./...`, `npm test -- --runInBand`, `npm run lint`,
  `npm run build`, `npm run type-check:e2e` and the browser suites pass with recorded
  results.

## Entities and external interfaces

Existing, unchanged: `SpatialSession { id, from, to, points, slots, capped }`,
`SpatialPoint { at, slot, x, y }`, `SpatialHeatCell`; `GET /api/spatial/history`
modes `sessions|points|heatmap`.

Additive field on `mode=sessions`:

```json
{ "id": "…", "from": 0, "to": 0, "points": 0, "slots": [1], "capped": false,
  "samples": [1737981234000, 1737981300000] }
```

- `samples`: ascending epoch-millisecond timestamps of **every** stored position inside
  `[from,to]` — the same points `mode=points` returns, evenly reduced to
  `MaxSamplesPerSession` (240). It deliberately does not carry only "episode starts": a
  continuously moving recording would show almost none, which reads as missing data.
  Absent or `[]` when none; an older hub omits it and the browser renders no activity.
  Optional in TypeScript (`samples?: number[]`).

New, browser-derived only: playback segment (recording placed on the unified timeline),
recording neighbour (previous/next), chapter (labelled movement marker).

No change to persistence, configuration, authentication or trust boundaries.

## Questions and assumptions

- Q-01: Movement episode definition is the existing presentation constant
  (`POSITIONS_IDLE_GAP_MS = 10 s`, spec 057 Q-01). Reusing it keeps one definition
  between the scrubber and the range timeline; the value is not a contract and may be
  tuned without a payload change. Assumption to validate against real recordings.
- Q-02: "Continuous" means chronological through the **filtered** recordings in the
  preset window, not only those already loaded; gaps shorter than a threshold are
  crossed as teleports, longer ones as explicit skipped spans (AC-03). Assumption:
  operators want the whole window, matching the rail's ordering intent (spec 057 Q-02);
  ordering is chronological for playback while the rail stays newest-first.
- Q-03: Previous/Next move exactly one filtered recording and are disabled at the ends.
  An earlier draft restarted a recording when Previous was pressed in its first seconds; that
  made a press right after Next unpredictable, so it was dropped. Moving one recording per
  press, with the scrubber for in-recording seeks, is the simpler rule.
- Q-04: The movement array is a read model over already-loaded chunks, not a second
  index or stored copy; it costs one iteration already present in `Sessions()`.
- Q-05: No autoplay on page load or on recording selection; only the explicit Play
  starts the clock, so a failed load cannot start unnoticed playback.
- Q-06: Playback does not loop and does not skip a failed recording silently; both are
  treated as end/failure states with text.
- Q-07: Direction was stated by the requester on 2026-11-20 (continuous play with
  resume, next recording, recorded activity rendered on the range timeline, and a UX check
  against comparable players). This is not a formal spec/plan review or maintainer
  approval.
