# Tasks: Positions player — continuous playback across recordings

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Status: in progress — T033
(transport below the fold at 900 px) and T048 (`StateHistory*`/ESPHome follow-ups) remain
open decisions, and a post-review remediation pass is recorded at the end.

Each task maps to requirement/acceptance IDs or a constitutional gate. Mark done only with
evidence.

## Preparation

- [x] T001 — **DONE**: additive `samples` contract recorded in
  `contracts/sessions-samples.md`; Q-01…Q-07 recorded in the spec. IDs: FR-05.
- [x] T002 — **DONE**: parent/backend/frontend rules checked in `plan.md`; no ADR needed
  (additive optional field, no storage/trust change). IDs: NH-04.

## Coverage and implementation

- [x] T003 — **DONE**: `service_test.go` covers sample derivation, the range filter,
  cross-session isolation, ordering and the cap; `spatial_history_test.go` asserts the
  `samples` key and value on the wire (decoded as raw JSON, not the domain struct).
  IDs: FR-05, AC-10, AC-12, NFR-02.
- [x] T004 — **DONE**: `Samples` on `spatialhistory.Session` plus `samplesFrom`, derived
  in the existing `Sessions()` pass from every stored timestamp inside the range, evenly
  reduced to `MaxSamplesPerSession` (240). `go test ./...` and `go build ./...` pass.
  IDs: FR-05.
- [x] T005 — **DONE**: `recordingOrder`, `recordingNeighbours`, `playbackRecordings`,
  `gapAfter`/`gapBefore`, `movementChapters`, `sanitiseSamples` and recorded-position marks
  on `sessionSegments`; 7 new domain tests (15 total). IDs: FR-01, FR-03, FR-06, FR-07.
- [x] T006 — **DONE**: continuous clock with `reachedEnd`, `awaitingCrossing`,
  `resumeCrossing`, `failCrossing`, seek cancellation, plus
  `useSpatialPlayback.test.ts` (8 tests). IDs: FR-01, FR-02, FR-04, AC-01…AC-05.
- [x] T007 — **DONE**: `PositionsTransport.vue` gains Previous/Next recording, "N of M" and
  local range, labelled chapter buttons, skipped-gap/end hints and a crossing error. IDs:
  FR-03, FR-07, FR-08, AC-06…AC-09, AC-14…AC-16.
- [x] T008 — **DONE**: `PositionsCoverageBar.vue` renders per-recording movement marks as
  sibling controls inside each segment and emits `play-from`. IDs: FR-06, AC-10…AC-13.
- [x] T009 — **DONE**: `SpatialPositions.vue` wiring: entry time on load, crossing watcher,
  play/pause-preserving navigation, Play all, play-from, Shift+Arrow, failure handling.
  IDs: FR-01…FR-04, FR-11, AC-01…AC-05, AC-20.
- [x] T010 — **DONE**: fixture derives `movements` with the hub's rule, adds a short
  `mock-session-night` for a fast boundary crossing, and freezes one epoch (the per-request
  `Date.now()` drift made a point fall outside its recording). IDs: NFR-05, SC-04.
- [x] T011 — **DONE**: `devices.spec.ts` covers resume, Next/Previous, disabled ends,
  Shift+Arrow, movement marks and activation, Play all crossing and the end of the filter;
  existing locators moved to the coverage segment/mark controls and counts to 4.
  IDs: AC-01…AC-19.

## Handoff

- [x] T012 — **DONE**: actual commands and results recorded in `plan.md`: backend `go test
  ./...`/`go build ./...`; frontend 154 suites / 1021 tests, lint 0 errors, build,
  e2e typecheck, `devices.spec.ts` 11/11 on both projects, full desktop 199 passed /
  9 skipped, full mobile 208 passed. Gate: NH-05.
- [x] T013 — **DONE**: failure, empty-range, session-without-points and 390 px scenarios all
  remain green; the crossing failure path is covered by the composable test and the
  existing browser failure scenario. IDs: AC-05, AC-17…AC-19.
- [x] T014 — **DONE**: contract updated, unrun checks listed, self-review removed a redundant
  prop/emit/wrapper and fixed two real defects found by the suite (stale `chapters`, fixture
  epoch drift); `loadPoints` now reports success so a failed load cannot start playback.
  Gates: NH-01/04/05. Review of the plan artifacts still requires a maintainer.

## Follow-up: visual polish (2026-11-20)

- [x] T015 — **DONE**: independent design critique plus full-page captures at 390/1024/1440 px
  drove a hierarchy pass. Canvas narrowed to `clamp(11rem, 26vh, 17rem)` (spec 057 AC-22
  superseded in FR-12); the range head now states the span it actually draws; played fill and
  the second playhead removed; selection is one solid accent block with an edge bar and dense
  recordings are outlines. IDs: AC-21, AC-24, FR-12.
- [x] T016 — **DONE**: `movementDensity()` measures the segment with a `ResizeObserver` and
  bands marks: individual above 8 px pitch, 3 vertical lanes between 2 and 8 px, density block
  below. 16 domain tests. At 24h the fixtures correctly band down; at 1280 px the morning
  recording keeps 4 individual marks. IDs: AC-22.
- [x] T017 — **DONE**: transport collapsed to three type tiers with the clock leading, Play the
  only primary, actions and metrics as two groups, one distance node, axis row as three times.
  Rails pluralise, show seconds under an hour, and print "1 sample" for degenerate rows;
  single-sample recordings say "Nothing to play or scrub" instead of drawing a dead scrubber;
  chevrons stay pinned to the title on phones. IDs: AC-23, AC-25, AC-18.
- [x] T018 — **DONE**: recorded checks after the pass: 154 suites / 1022 tests; lint 0 errors
  (23 pre-existing warnings); build and e2e typecheck pass; `devices.spec.ts` 11/11 desktop and
  mobile; full desktop 199 passed / 9 skipped; full mobile 208 passed. IDs: SC-03, SC-04.
- [x] T019 — **DONE**: fixture reshaped to a realistic day (a 2-minute recording with room per
  movement, a denser 3-minute one, an empty 21 s one and a single-sample one) so both the marks
  and density bands are exercised; the AC-11 long pause was preserved. IDs: AC-22, AC-11.

## Follow-up: review pass — bugs, performance, a11y, rules (2026-11-20)

- [x] T020 — **DONE (bug, high)**: a crossing superseded by a newer selection left
  `awaitingCrossing` true forever; every later tick returned early, so the clock froze while the
  transport still said "playing". Added `cancelCrossing()` and called it on both stale-response
  paths in `loadPoints`. Regression test `a superseded crossing is dropped so the clock cannot
  freeze mid-advance`. IDs: AC-01, FE-03.
- [x] T021 — **DONE (bug, high)**: the same test exposed that `advance()` trusted callers to
  have seeked, so a changed `bounds` could render a playhead outside the confirmed recording.
  `advance()` now clamps into range. IDs: AC-02, FE-03.
- [x] T022 — **DONE (perf, medium)**: `Sessions()` retained every `Point` of every session to
  derive episodes; a 7-day query spans hundreds of sessions. It now keeps only the timestamps
  (`movementsFrom`). Backend suite and build pass. IDs: NFR-02.
- [x] T023 — **DONE (perf, low)**: `sessionSegments()` sanitised each session's `movements`
  twice; validated once and passed to `movementMarks()`. IDs: NFR-01.
- [x] T024 — **DONE (a11y, high)**: the idle line sat in `aria-live="polite"` and recomputed every
  200 ms tick, re-announcing the same sentence five times a second. It is now bucketed to ten
  seconds under a minute and to minutes beyond it, so a long idle tail does not re-announce every
  second either. IDs: FE-04.
- [x] T025 — **DONE (a11y, medium)**: every movement mark is a focus stop, and the scrubber
  allowed 200 of them. Capped at 40, matching the range timeline, with sampling stated in the
  text. IDs: FE-04, AC-14.
- [x] T026 — **DONE (a11y, medium)**: a failed crossing was announced twice (page `role="alert"`
  plus a transport `role="alert"` with the same message). Removed the duplicate; the composable
  no longer owns an error string, and the page's single `pointsError` announcement covers it.
  IDs: FE-03, FE-04.
- [x] T027 — **DONE (dead code)**: `RenderedMovement.count` was always 1, so the
  "N movements around" label branch was unreachable; removed the field and branch. Documented
  the two undocumented exports and removed a stale `merged`/`clearError` comment trail.
  IDs: FR-07.
- [x] T028 — **DONE**: recorded checks after the pass — backend `go test ./...` and
  `go build ./...` pass; frontend 154 suites / 1023 tests; lint 0 errors (23 pre-existing
  warnings); build and e2e typecheck pass; `devices.spec.ts` 11/11 desktop and mobile;
  full desktop 199 passed / 9 skipped; full mobile 208 passed. IDs: SC-01…SC-05.

Known gap found but not changed (needs a decision, not a fix): if the range timeline is
mounted while hidden (`display: none`), `ResizeObserver` reports 0 and every recording bands
to "density" until the next size change. The player is only rendered in the Player view, so
this is currently unreachable; recording it rather than adding speculative code.

## Follow-up: proportion correction (2026-11-20)

- [x] T029 — **DONE (owner feedback)**: the polish pass shrank the recorded canvas to
  `clamp(11rem, 26vh, 17rem)` while leaving a full-size transport, which made the player the
  biggest thing on the page and the map small. Reverted the canvas to spec 057's
  `clamp(16rem, 40vh, 24rem)` and corrected FR-12/AC-23, which had documented the wrong sizes.
  A speculative `max-height` yield was written and then removed: it was the same guessing the
  measurement was meant to replace.
- [x] T030 — **DONE**: measured instead of guessing. The transport was 239 px tall against a
  256 px canvas at 620 px viewport because it stacked six rows. Collapsed idle / skipped-gap /
  movement-count / nothing-to-scrub into one status line, tightened padding and gaps, stepped
  the clock down from 1.5 rem to 1.15 rem (1 rem on phones), and sized Play like its siblings
  (`sm` rather than `md`). The `devices.spec.ts` guard now fails if the transport exceeds
  0.6 x the canvas height; it measures at 1440x800 and at 1280x600, the strictest case
  because the canvas has reached its 16 rem floor. Phones keep mandated 44 px targets.
- [x] T031 — **DONE (verified, no change needed)**: path rendering during motion is correct.
  Probed the polyline at 0/30/60/100 % of a recording (1/3/3/5 points, recent trail appears
  and clears) and sampled 12 frames during 4x playback: the path grew 1 -> 2 at the next
  recorded sample and never shrank. Unchecking Path removes both polylines.
- [x] T032 — **DONE**: after the correction — 154 suites / 1023 tests; lint 0 errors
  (23 pre-existing warnings); build and e2e typecheck pass; `devices.spec.ts` 11/11 on desktop
  and mobile.
- [ ] T033 — **OPEN (needs a decision)**: at a 900 px viewport the transport still sits ~220 px
  below the fold, because the device header, the always-expanded View panel and the range
  timeline sit above it. Options: start the View panel collapsed on short viewports, fold the
  range timeline next to the range buttons, or accept the scroll. Not changed without a
  decision.

## Follow-up: listed-but-unplayable recording (2026-11-20, owner report)

- [x] T034 — **DONE (bug, high)**: owner reported that clicking a listed recording disabled
  Play. Reproduced against the real store (`backend/data/spatial.db`, inspected from a copy
  because the live hub holds the lock): `Sessions()` looks back 30 minutes so a chunk that
  started before the filter but holds points inside it is listed, while `Points()` did not,
  so the detail returned `[]` and Play was disabled with no explanation. `Points()` now uses
  the same lookback. See `investigation-data-and-play.md`. IDs: FE-02, FE-03, AC-19.
- [x] T035 — **DONE (regression test)**: `TestPointsReturnWhatSessionsListedForAStraddlingChunk`
  in `service_test.go` builds the straddling chunk and asserts the two queries agree.
  IDs: SC-01.
- [x] T036 — **DONE (copy)**: when a recording's summary still reports points but its detail
  loads none, the page says "This recording lists N positions in this range, but none could be
  loaded" instead of "No recorded positions", and the browser test asserts it. IDs: FE-03.
- [x] T037 — **DONE (investigation, no code change)**: "most recordings barely have any
  motion" is the recorder's 30 cm movement gate, not a defect. 5 of 32 real recordings are
  single-sample (target holds still) and the rest hold 2–18 points over tens of seconds.
  No chunk is capped and volumes are far below `MaxPointsPerDay`. Tightening `movementGateMM`
  or `minSampleGap` would add detail for every device with history enabled, so it is left as a
  decision, not a change. Recorded in `investigation-data-and-play.md`.

## Follow-up: next-recording arrow and activity ticks (2026-11-20, owner report)

- [x] T038 — **DONE (bug)**: the "Next recording" chevron existed and worked but sat at the
  card's far edge, ~600 px from the title, where it read as part of the border. The label was
  `flex: 1 1 auto` and stretched. Measured the gap (title right 906 vs next left 910 of a
  684 px row), grouped prev/title/next into one cluster with the title `flex: 0 1 auto`
  ellipsising, and raised disabled-chevron opacity to 0.6. IDs: AC-08, FE-04.
- [x] T039 — **DONE (bug, root cause)**: the range timeline's ticks were *movement episode
  starts* — the first sample and each sample after a pause longer than 10 s — so a
  continuously moving recording showed almost none. Real data: 18 recorded samples produced
  10 ticks, and the operator could see motion where there was no tick. Replaced `movements`
  with `samples` in the payload: every recorded timestamp, evenly reduced to 240. Verified
  against the real store: reported samples now equal the stored points for every recording.
  IDs: AC-10, FR-05, NFR-02.
- [x] T040 — **DONE (design)**: replaced one 2 px tab stop per sample with one labelled
  control per recording plus an activity strip whose bars encode where samples cluster and
  how many, drawing nothing where nothing was recorded. Clicking anywhere in a segment plays
  from the nearest recorded position, so reaching a specific moment is one gesture instead of
  hunting an unhittable mark. Segments narrower than 28 px are drawn at 28 px and the row is
  rescaled to stay in the track; labels now state "N recorded positions". (Superseded by
  T077, which restores one focusable control per recorded position.) IDs: AC-22, AC-26.
- [x] T041 — **DONE**: removed the now-superseded `movementDensity` banding, its test and its
  CSS; contract replaced with `contracts/sessions-samples.md`; spec/plan updated (FR-05,
  AC-10, AC-22, AC-25, AC-26, NFR-02) rather than left describing the old field.
- [x] T042 — **DONE**: recorded checks — backend `go test ./...` and `go build ./...` pass;
  frontend 154 suites / 1022 tests; lint 0 errors (23 pre-existing warnings); build and e2e
  typecheck pass; `devices.spec.ts` 11/11 desktop and 11/11 mobile.

## Follow-up: clean-code remediation (2026-11-20, owner report)

Owner feedback: the produced code does not follow the repo's rules, instructions and skills —
large components, mixed responsibilities, single source of truth missing. Full record in
`clean-code-remediation.md`; two read-only audits plus measurement produced the findings.

- [x] T043 — **DONE**: `SpatialPositions.vue` 955 → 616 lines. Split into
  `usePositionsHistory` (filter, summary, confirmed points, heat cells, generation fences),
  `usePositionsPlayback` (clock ↔ loader coupling, crossings, gaps), `useElementSize`
  (`ResizeObserver` lifetime) and `domain/spatial/positions-playback.ts` (tracks, trails,
  hop links, latest sample, recorded envelope). Fit maths and the four duplicated bounds
  accumulators moved to `domain/spatial/positions.ts`. Rule: one responsibility per component.
- [x] T044 — **DONE**: history formatting violations. `HistoryStats` now formats through
  `formatWithUnit` (one missing-value sentinel, not two); `options.ts` has one `tooltipValue()`
  and prints a spread's unit once; `historyDisplayUnit()` is the single owner of the
  brightness/kelvin mapping; `formatHistoryRange()` replaced five hand-rolled period labels.
- [x] T045 — **DONE**: `useChartAppearance()` replaced the copy-pasted token reads, the
  undocumented `void settings.value` trick and the one-shot `matchMedia` read in both charts.
- [x] T046 — **DONE**: ESPHome copy consolidated into `ESPHOME_MESSAGES`; the diverged
  `isESPHomeDevice` copy replaced by the shared predicate; enrollment `status`/`reconnections`
  narrowed at ingress (the real vocabulary is `online`/`offline`, taken from the Go constants)
  and rendered through `connectionStatusLabel`; candidate ports validated; the add dialog now
  resets — including the encryption key — on a parent-driven close; hidden-tab polling gated;
  `close()`/`discardAndClose()` split so the guard is explicit rather than load-bearing order.
- [x] T047 — **DONE**: `junction` checks — 154 suites / 1023 tests, lint 0 errors (23
  pre-existing warnings), build and e2e typecheck pass, ESPHome browser specs 7/7. The one
  `devices.spec.ts` failure is the Live-view label assertion, reproduced on the untouched
  baseline and outside the changed files.
- [ ] T048 — **OPEN (needs a decision)**: `StateHistoryChart.vue` (290-line script, four
  responsibilities), the triple-duplicated `StateHistory*` labels/palette, `HistoryView.vue`'s
  dead CSS and unscoped `<style>`, and ESPHome's mutation `AbortController` scope + dialog
  split. Real findings with line numbers in the audit; not changed here because they alter
  behaviour and need their own browser coverage.

## Follow-up: second clean-code pass (2026-11-20)

Owner: components still mix responsibility; shared calculations repeat across components.
Record in `clean-code-remediation.md` § "Second pass".

- [x] T049 — **DONE**: extracted every user-facing sentence for the positions family into
  `domain/spatial/positions-labels.ts` (one pluraliser, one status ranking, one empty-state
  owner scoped per surface). The four components now only place text.
- [x] T050 — **DONE**: extracted strip geometry into `domain/spatial/positions-strip.ts`
  (bar grouping, minimum readable width, row rescale, nearest-click lookup, span placement).
  `PositionsCoverageBar.vue` no longer computes geometry; `PositionsTransport.vue` no longer
  computes percentage placement.
- [x] T051 — **DONE**: `formatMillimetres()` in `utils/units.ts` replaced
  `formatDistance(mm / 1000)` at the call site, which the formatting rules keep out of components.
- [x] T052 — **DONE (regression I caused)**: selecting a recording never seeked the playhead
  after the pass-1 split, so the canvas drew nothing. The seek now lives in
  `usePositionsPlayback.select()`, where the recording and the playhead change together.
- [x] T053 — **DONE**: `renderSegments` divided by a 1 px floor, expanding every recording to
  full width before the first measurement. Returns the proportional layout until measured.
- [x] T054 — **DONE**: 15 new domain tests over the extracted owners; two of my expectations were
  wrong and one exposed a real bug. Corrected three stale browser assertions that encoded
  assumptions these changes invalidated (the newest fixture holds one sample; the rail test now
  asserts "the same recording stays selected" rather than a fixture id).
- [x] T055 — **DONE**: 155 suites / 1038 tests, lint 0 errors, build and e2e typecheck pass,
  full browser suites green (desktop 199/9 skipped, mobile 208) — including the previously
  failing spatial test.

## Follow-up: independent review remediation (2026-11-22)

An independent review found two player failures, a contract regression and several rule
gaps. Each is fixed and covered below.

- [x] T056 — **DONE (bug, blocking)**: a failed crossing returned from `load` without
  settling the clock, leaving `awaitingCrossing` set so every later tick returned early,
  `playing` stayed true and Pause→Play could not recover. It now calls `failCrossing()`;
  the vanished-successor branch cancels and pauses instead of leaving the flag set.
  New `usePositionsPlayback.test.ts` reproduces it. IDs: AC-05, FE-03.
- [x] T057 — **DONE (bug, blocking)**: the first load never placed the playhead because
  `loadSessions` calls `loadPoints` directly and `time` stayed 0, so the canvas drew nothing
  and the clock showed "—" until another recording was selected. `useSpatialPlayback` now
  places the playhead inside changed bounds even while paused (a placement, not a seek, so
  it never stops a running clock). The e2e opens on the first recording instead of masking
  it, and asserts the clock is not "—". IDs: FE-03, AC-25.
- [x] T058 — **DONE (bug, blocking)**: `ESPHomeStatus` narrowed to `online|offline` while the
  hub also sends `connecting`, so an actively connecting device rendered "Not reported".
  `connecting` is now accepted and labelled; contract tests moved to
  `src/__tests__/contracts/esphome.test.ts`. IDs: FE-02.
- [x] T059 — **DONE (bug, blocking)**: the statistics chart gave Minimum–Maximum, Average and
  Last the same `style.color` while the legend advertised three tokens. `useChartAppearance`
  now publishes the three kind colours and the chart consumes them. IDs: FE-04.
- [x] T060 — **DONE (bug, high)**: `Heatmap` read `Chunks(device, from, to)` without the
  straddling-chunk lookback `Sessions`/`Points` use, so a chunk starting just before the
  window was listed with points but contributed no cells. New service regression test.
  IDs: BE-04.
- [x] T061 — **DONE (bug, high)**: doubling every short recording to 28 px while leaving
  left edges proportional let adjacent recordings overlap, making the later one
  unclickable. `renderSegments` now lays them sequentially (never overlapping) and
  rescales the row; the component uses the laid-out left. IDs: AC-22.
- [x] T062 — **DONE (bug, high)**: the fitted view box and zoom anchor mixed projected
  points with raw outline/heat-cell coordinates. Both functions project every input through
  `screenPoint`/`sensorProjectedPoint` before measuring. IDs: FE-03.
- [x] T063 — **DONE (bug, high)**: clicking a timeline segment emitted both `select` and
  `play-from`, issuing two identical points requests and letting the first reply supersede
  the second. It emits `play-from` only. IDs: AC-26, FE-03.
- [x] T064 — **DONE (bug, high)**: the window-level Shift+Arrow shortcut ignored editable
  targets, modifier chords and the open Recordings dialog. All three now guard it. IDs: FE-04.
- [x] T065 — **DONE (bug, high)**: `NumericHistory` passed `!!props.compact` once, so resizing
  across the compact breakpoint left stale axis/animation config. `useChartAppearance`
  accepts a getter and the chart passes one. IDs: FE-04.
- [x] T066 — **DONE (bug, medium)**: `formatMillimetres(350)` rendered "0 m"; sub-kilometre
  distances now keep two decimals. IDs: FE-03.
- [x] T067 — **DONE (bug, medium)**: the y-axis concatenated `" ${unit}"` by hand while
  tooltips used `formatWithUnit`, so axis and tooltip disagreed. Both use the shared
  formatter. IDs: FE-03.
- [x] T068 — **DONE (bug, medium)**: `skippedGap` was set on a crossing and only cleared by
  an explicit selection, masking every later status line. It now clears after a short notice.
  IDs: AC-03, FE-04.
- [x] T069 — **DONE (bug, medium)**: the zoom anchor was not re-captured after a recording
  change while zoomed. A post-flush watcher on the confirmed recording re-anchors it.
  IDs: AC-22.
- [x] T070 — **DONE (a11y, medium)**: the idle status line still changed every second inside
  its live region; it is bucketed to ten seconds under a minute and to minutes beyond.
  IDs: FE-04.
- [x] T071 — **DONE (contract, medium)**: `Slots []int` marshalled as `null` when empty while
  the browser declares `number[]`; `Sessions` now starts it as `[]`. IDs: BE-04, FE-02.
- [x] T072 — **DONE (tests)**: moved `positions-strip` coverage to
  `src/__tests__/domain/spatial/positions-strip.test.ts` and added companion tests for
  `usePositionsPlayback`, `positions-playback`, `usePositionsHistory` and
  `useChartAppearance`. IDs: NH-05.
- [x] T073 — **DONE (tooling)**: the flat ESLint config matched only JS and `.vue`, so
  `npm run lint` never linted TypeScript. A `**/*.{ts,tsx}` block now parses and lints it.
  IDs: NH-05.
- [x] T074 — **DONE (spec)**: corrected the `samples` definition (every stored timestamp,
  not episode starts), removed the stale `Movements`/`sanitiseMovements`/
  `POSITIONS_PREVIOUS_RESTART_MS`/`MAX_TRANSPORT_MARKS = 200` references, added BE-04/BE-05,
  and corrected the FR-12 measurement to the guarded 1280×600 case. (AC-11/AC-13/FR-06 were
  briefly re-scoped to the then-delivered per-recording design; T077 restores them.)
  IDs: NH-01, BE-04, BE-05.
- [x] T075 — **DONE (scoped)**: the full `devices.spec.ts` — the changed page/file — passes
  **22/22 on desktop and mobile**. The repo-wide browser suite (every spec file) was not rerun
  for this feature; that remains a PR-time check.
- [x] T076 — **DONE (checks)**: backend `go build ./...` and `go test ./...` pass; frontend
  Jest **161 suites / 1058 tests** pass; `npm run lint` **0 errors / 113 warnings** (now
  including TypeScript); `npm run build`, `vue-tsc` and `npm run type-check:e2e` pass; the
  changed Positions flow, `devices.spec.ts › spatial mappings save and show only current
  provider reports`, passes on desktop in 16.4 s. Full browser suites remain T075.

## Follow-up: per-movement marks on the range timeline (2026-11-22, owner request)

Owner asked to implement the review's per-movement-mark design after seeing a comparison
prototype.

- [x] T077 — **DONE**: `PositionsCoverageBar` renders one focusable `<button>` per recorded
  position, named by local time and recording via `movementMarkLabel`. The timeline splits one
  global `MAX_TIMELINE_MARKS = 120` budget across its recordings (per-recording caps alone let
  a 500-recording filter expose thousands of focus stops); a segment with no marks keeps one
  fallback control so an older hub's recording stays reachable by keyboard, and that control
  paints no tick (it is a selection target, not a measured position). `sampleMarks` keeps both
  endpoints. The head states
  `M marks shown` when the budget samples the true total. Colours use theme tokens
  (`--nh-text-muted`, `--nh-accent-contrast` on the selected accent fill, `--nh-focus`), and
  the selected segment's focus ring switches to the contrast token because `--nh-focus`
  resolves to the accent fill. Marks have a 24 px-tall centred hit area with a clamped
  left offset. Keyboard activation (Enter/Space) plays from the exact mark; pointer activation
  falls through to the segment's nearest-recorded-position handler, because dense mark hit
  areas overlap at the 28 px floor and would otherwise resolve to the wrong button. The decorative
  activity-bar geometry (`activityBars`, `RenderedSegment.activityBars`) was removed rather
  than left dead. The segment is a `role="group"` carrying the recording's accessible name,
  `aria-current` and `data-selected`; e2e selectors moved from `.nh-coverage-select` to
  `.nh-coverage-segment` / `.nh-coverage-mark`; the suite activates a mark (comparing the
  recorded instant) and covers the no-`samples` fallback. Spec AC-10/11/13/22/26 and FR-06 restored to the
  per-movement design. IDs: AC-10, AC-11, AC-13, AC-18, AC-22, AC-26, FR-06, FR-10, FE-04.
- [x] T078 — **DONE (a11y)**: the timeline uses a roving tabindex — exactly one mark is in the
  Tab order; ArrowLeft/ArrowRight move between marks and Enter/Space activates. That removes the
  "Tab 120 times to reach Play" problem on a large filter. The global `MAX_TIMELINE_MARKS = 120`
  budget still bounds how many arrow steps exist; lowering it or limiting marks to the selected
  recording remains a product decision, not a defect.
- [ ] T079 — **OPEN (needs a decision)**: `sessionSegments` documents clipping a recording to a
  narrower window, but its only caller passes the recordings' own envelope, so the clamp
  branches are unreachable and `activity[].pct` is relative to the unclipped span. Delete the
  clip contract or carry `clipFrom`/`clipTo` through both the mark and nearest-position maths.

## Follow-up: third review pass — bugs and rule gaps (2026-11-22)

Independent review of the branch. Findings fixed here; each is covered by a test in the source
owner's companion test file.

- [x] T080 — **DONE (bug, high)**: activating a recording's **last** recorded-position mark played
  the whole recording from its beginning. The last sample of a recording sits on its `to`, and
  `useSpatialPlayback.play()` rewound any playhead at the end to `from` — exactly the AC-11/AC-26
  case. `play()` now takes `rewindAtEnd` (default true) and `usePositionsPlayback.play()` passes
  false when the caller supplied an explicit `at`. Tests: `useSpatialPlayback.test.ts`
  (`play(false) keeps a playhead already at the end`) and `usePositionsPlayback.test.ts`
  (`play from the recording's final sample starts there`). IDs: AC-11, AC-26, FR-06.
- [x] T081 — **DONE (bug, high)**: a superseded crossing could stop a clock the operator had
  already started on another recording. The async crossing watcher awaited its load, then called
  `failCrossing()` unconditionally; a newer selection had already cancelled the crossing and
  begun playing. It now returns when the crossing is no longer pending. Test: `a superseded
  crossing reply cannot stop a clock started on another recording`. IDs: AC-05, FE-03.
- [x] T082 — **DONE (perf, medium)**: leaving the Player view (Heatmap) kept the 200 ms clock
  running, so crossings continued to issue points requests for an unmounted surface.
  `chooseView('heatmap')` now pauses. IDs: AC-19, NFR-03.
- [x] T083 — **DONE (dead code)**: removed the unused `.nh-transport-continuous` rule and the
  duplicated `.nh-coverage-segment { cursor }` declaration; dropped the `const total = trackWidth`
  alias in `renderSegments`.
- [x] T084 — **DONE (single source of truth)**: `PositionsTransport` and `PositionsCoverageBar`
  built user-facing sentences and unit conversions themselves (`Movement at … on slot …`,
  `T of T`, `"{d} of {d}"`, the fallback mark label). They now call `movementChapterLabel`,
  `playbackValueLabel`, `distanceProgressLabel` and `fallbackMarkLabel` in `positions-labels.ts`,
  which owns `formatMillimetres`. IDs: FR-08, FE-04.
- [x] T085 — **DONE (rule)**: `PositionsCoverageBar` hand-rolled a second `ResizeObserver` while
  `composables/positions/useElementSize` exists for exactly that; it now uses the composable.
  IDs: FE-01.
- [x] T086 — **DONE (bug, low)**: `useChartAppearance` called `window.matchMedia` twice for the
  same query; it now creates one `MediaQueryList` and observes it.
- [x] T087 — **DONE (T079 partial, contract)**: mark percentages are now placed over the span the
  segment is actually drawn over (and samples outside it are dropped), so a mark cannot point at
  a time the visible segment does not represent. `sampleCount` keeps the payload's true total,
  matching its existing contract. `nearestPositionAt` still maps the pointer ratio across the
  recording rather than the clipped span; that is identical for the only caller and T079 remains
  open for the general case.
- [x] T088 — **DONE (scope hygiene)**: `components.d.ts` had two PrimeVue entries removed as
  unrelated churn; regenerating with the dev server restored them, so the file is unchanged.
- [x] T089 — **DONE**: recorded checks after this pass — backend `go build ./...` and
  `go test ./internal/spatialhistory/... ./internal/api/...` pass; frontend Jest **161 suites /
  1065 tests**; `npm run lint` **0 errors / 113 warnings** (unchanged); `npm run build`
  (vue-tsc + vite) and `npm run type-check:e2e` pass. Browser (desktop): `devices.spec.ts`
  **11/11**; `DeviceMetrics.spec.ts` + `EntityViewDialog.spec.ts` **54 passed / 9 skipped / 0
  failed**; the spatial story also passes on **mobile**. Not rerun: the repo-wide browser suite
  and `devices.spec.ts` on mobile as a whole (the spatial case was run there).
- [x] T090 — **DONE (bug, low)**: `samplesFrom`'s even reduction used `index*len/Max`, which never
  selected the final timestamp (a capped 480-sample recording lost its last position). It now
  spreads over `len-1`, keeping both endpoints like the browser's `sampleMarks`; the backend test
  asserts both survive. `sanitiseSamples` keeps its first-240 ingress bound (a rogue over-long
  payload), which is a safety cap, not the sampling contract.
- [x] T091 — **DONE (dead code)**: removed the unused `recordingRangeLabel`, the `spatialSamples`
  identity wrapper in the mock server, and a stale "must band down" fixture comment (the banding
  was removed in T041).

Known, deliberately unchanged: "Play all recordings" is hidden when the filter holds exactly one
recording (`hasRecordings` is `length > 1`), which is a copy/product decision rather than a
defect; `positionedSpan`/`idleSpans` are invoked by both the page and the transport over the same
points (same helper, not duplicated logic).
