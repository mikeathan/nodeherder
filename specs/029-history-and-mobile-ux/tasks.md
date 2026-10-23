# Tasks: History, charts and mobile shell

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | ADR: [adr-006](adr-006-chart-interactions.md)
Status: Implemented; automated verification recorded in plan.md; physical-device/browser manual pass pending

Each task maps to requirement/acceptance IDs, names the paths it owns, and states its completion
evidence. New tests go in the companion file of the source they cover (for example
`frontend/src/__tests__/domain/history/stats.test.ts` for `src/domain/history/stats.ts`).

## Phase 1 — Preparation and proof of the baseline

- [x] T001 — Record the current baseline in the plan: chart heights (235/105), history selectors,
      bundle size from `npm run build`, and the existing history e2e list. Reproduce the three
      stated gaps (no zoom, no Yesterday, sheet dead end) before editing. IDs: NH-01, NH-05;
      done: baseline code findings/build sizes recorded in plan.md; acceptance screenshots are produced by the isolated browser tests (no PR created).
- [x] T002 — Prove the local interaction plugin against Chart.js 4.5.1 on the isolated fixture chart:
      mouse drag-select, wheel, two-pointer pinch, keyboard, `touch-action`, and single-finger drag
      scrubbing the readout **without** zooming (AC-05). Confirm that recomputing `historyOptions`
      after a range change moves the `afterBuildTicks` instants and the label format. IDs: FR-03,
      AC-04/05/06; depends: T001; done: findings recorded (including any browser that intercepts
      pinch), plugin registered after `Chart.register` in `HistoryCanvas.vue`.

## Phase 2 — Pure domain (no Vue, fully unit tested)

- [x] T010 — Extend `frontend/src/domain/history/range.ts` with the `yesterday` preset and exported
      clamp/step helpers (5 min…31 days, 20 % pan); IDs: FR-01, AC-01, AC-04; depends: T001;
      done: `src/__tests__/domain/history/range.test.ts` covers midnight boundaries, DST day
      lengths, custom ranges and clamping.
- [x] T011 — Add `frontend/src/domain/history/stats.ts`: enumerate every intersecting local day
      (no `MAX_BUCKETS` constant; step with `setDate(+1)` + `setHours(0,0,0,0)`), bucket the **raw
      decoded** samples, `null` for empty buckets, reuse `numericStats`, expose sample count, covered
      window and partial flag; IDs: FR-04, AC-07, AC-08, AC-13; depends: T010; done:
      `src/__tests__/domain/history/stats.test.ts` covers empty buckets, single sample, all-null
      series, partial input, decimated-input divergence, and DST including the pinned case
      `TZ=Europe/London`, 1 Mar 2026 23:30 GMT → 2 Apr 2026 00:30 BST (744 h, **33 buckets**, first
      and last flagged partial, a sample at 00:00 local counted once in the day starting there).
- [x] T012 — Add `frontend/src/domain/history/export.ts` with the two documented schemas
      (`timestamp,value,unit` for history; `bucket_start,bucket_end,min,average,max,last,samples,unit`
      for statistics), ISO-8601 timestamps with offset, empty cells for gaps and empty buckets, and a
      partial/empty guard; export always covers the selected range, never the zoomed viewport.
      IDs: FR-04, AC-08/09; depends: T011; done:
      `src/__tests__/domain/history/export.test.ts` covers both headers, gap/empty-bucket cells,
      escaping, offset formatting and partial refusal.
- [x] T013 — Add `frontend/src/domain/history/preferences.ts` for `nh-history-pref:<deviceId>` =
      `{ version: 1, preset, expose, custom?: { from, to } }` (versioned read/write, defensive parse,
      no throw on storage failure); IDs: FR-01, AC-02, FE-02, FE-05; depends: T010; done:
      `src/__tests__/domain/history/preferences.test.ts` covers a valid custom range restoring
      exactly, missing/unparsable/unordered/>31-day bounds falling back to `24h` with the stored
      expose kept, version mismatch, unreadable storage and unknown expose.

## Phase 3 — Chart component and interactions

- [x] T014 — Implement `frontend/src/components/chart/history/interactions.ts` as our own Chart.js
      plugin: mouse drag-select, wheel, two-pointer pinch, keyboard, `touch-action`, and single-finger
      scrub that changes only the readout; every range change goes through one `applyRange` that
      clamps and never leaves the loaded window, and the view recomputes `historyOptions` for the new
      range. No new dependency. IDs: FR-03, AC-04/05/06, NFR-01/02; depends: T010, T002; done:
      `src/__tests__/components/chart/history/interactions.test.ts` covers clamp mapping, keyboard
      steps, pinch scale maths, one-finger scrub leaving the range untouched, and no-op bounds; the
      build shows no new import in the chart chunk.
- [x] T015 — Add `HistoryChart.vue` (responsive height by breakpoint, focusable container with
      accessible name and range, readout slot, end marker, `touch-action`); IDs: FR-02, FR-03,
      AC-03/05/06, FE-04; depends: T014; done: component covered by domain/interaction Jest tests and real-input browser checks, and used by
      the view below.
- [x] T016 — Extract `HistoryToolbar.vue` (presets, expose chips, history/statistics switch,
      refresh/export) and `HistoryStats.vue`; IDs: FR-01, FR-04, AC-01/02/07; depends: T015, T011;
      done: companion domain/interaction Jest tests and `e2e/DeviceMetrics.spec.ts` verify state wiring; no Vue mounting tooling added.
- [x] T017 — Recompose `HistoryView.vue` from the parts, keep the `compact` prop and every existing
      loading/error/empty/partial message; IDs: FR-06, AC-13/14; depends: T016; done: existing
      history e2e selectors still resolve.
- [x] T018 — Move the samples table into `HistorySamples.vue` unchanged in behaviour; IDs: AC-14;
      depends: T017; done: table paging test still passes.

## Phase 4 — Pages, sheet and shell

- [x] T019 — `DeviceMetrics.vue`: wide layout for the history route, preference restore (including
      validated custom bounds with the `24h` fallback), statistics mode for numeric series only,
      export, keyboard hints, and a “narrow this range” action when the series is partial. Expose
      precedence is URL → saved preference → first chartable expose. IDs: FR-01/02/04,
      AC-01/02/03/07/08/09; depends: T017, T013; done: desktop e2e covers preset/expose switching,
      restore after reload, a stale/invalid custom range falling back to `24h`, a binary series
      without the switch, export, and a truncated series naming its covered window.
- [x] T020 — `EntityViewDialog.vue`: compact chart, "View full history" to
      `/devicepage/:id?tab=history&expose=<name>`, drag-to-dismiss (top-of-scroll, 120 px,
      reduced-motion aware), safe-area actions; IDs: FR-05, AC-11; depends: T015; done: sheet e2e
      covers drag, Escape, close, and the deep link landing on the same expose even when a different
      one is saved in preferences.
- [x] T021 — Shell: mark primary items in `navigation.ts`, filter them from the phone drawer with CSS, add
      `BottomNav.vue` (Home, Devices, Automations, Assistant, More→drawer), render it in
      `AppShell.vue` below 720 px, keep the drawer, add `--nh-safe-bottom`; IDs: FR-05, AC-10,
      NFR-03; depends: T001; done: layout e2e asserts 44 px targets, `aria-current` for tabs and the
      `includes` rule, More current on other routes with `aria-expanded` toggling the drawer, no
      horizontal scroll and safe-area padding at 390×844.
- [x] T022 — Platform: `index.html` gains `viewport-fit=cover` and `theme-color`;
      `public/manifest.webmanifest` plus icon; verify the app still boots when the manifest 404s;
      IDs: FR-05, AC-12, NFR-03; depends: T021; done: e2e asserts the meta tags and manifest fetch.

## Phase 5 — Documentation and supersession

- [x] T023 — Retire the root `charts_todo.md`: move anything still relevant into `plan.md` debt,
      delete the file, and note the supersession. IDs: NH-01; depends: T001; done: no dangling
      references (`rg charts_todo`).
- [x] T024 — Supersede **only the covered portions** of spec 007: T209 (history query correctness),
      T210 (pure history domain) and T211 (history rebuild) are delivered by 029; T212 stays open for
      Assistant layout, palette/contrast checks and legacy cleanup, with a note that its
      history-dependent parts move to 029; T213 stays open for the flows 029 does not cover. Update
      the docs that still show the select-based history flow. IDs: NH-01; depends: T023; done:
      links resolve, each superseded task names where its obligation went, and nothing unrelated is
      closed.

## Phase 6 — Verification and handoff

- [x] T030 — Component checks from `frontend/`: `npm test -- --runInBand`, `npm run lint`,
      `npm run build`, `npm run type-check:e2e`. IDs: NH-05, NFR-01; depends: T017–T022; done: pass
      with the recorded bundle delta, compared per chunk (initial entry does not grow, chart chunk
      within +4 kB gzip per NFR-01).
- [ ] T031 — Browser checks: migrate the existing `e2e/DeviceMetrics.spec.ts` Entity/Period
      `selectOption()` assertions to the new chips in this same change, then run
      `CI=true npx playwright test` on desktop and mobile projects, including new gesture tests driven
      by real input (mouse drag, wheel, CDP two-pointer pinch, keyboard) and the a11y assertions
      (names, focus ring, 44 px targets, no horizontal scroll). IDs: FR-03, AC-04…AC-12, AC-14,
      NFR-03; depends: T030; done: pass, no skipped selector, with the manual desktop/mobile pass
      recorded.
- [x] T032 — Prove no backend or contract drift: `backend/`: `go test ./...`; confirm no diff under
      `backend/` and no change to `contracts/history.ts` request shape. IDs: AC-13/14, BE-01…BE-05;
      depends: T030; done: command output recorded, diff shows frontend + docs only.
- [ ] T033 — Review pass: re-read the diff against every applicable NH/FE rule, confirm truthfulness
      cases (gaps, unknown, partial), update the plan's traceability table with actual results, and
      list anything unrun. IDs: NH-01…NH-06, FE-01…FE-05; depends: T031, T032; done: plan evidence
      complete and honest; no unapproved violation.


Execution results and design refinements are recorded in [plan.md](plan.md#execution-and-acceptance-record--2026-10-22).
T031's automated runs passed (full regression and final focused history/layout run); the remaining
physical Safari/safe-area/frame-timing and human acceptance pass are unrun. T033's code and
constitutional review is recorded; it stays open until those manual results are available.

CI follow-up: T031/T033 also cover deferred AppShell readiness in the shared `openNav` helper.
Add a controlled delayed-shell regression to existing `e2e/panel.spec.ts`, demonstrate the failure,
then verify both projects and the reported mobile shard before updating PR #74.

Review follow-up (2026-10-23): the reopened tasks also cover midnight-end ownership, unconditional
DST verification under a UTC launch environment, range-bar/marker readability, explicit export units,
scoped height variables, viewport shell skeleton, live theme metadata and crosshair reset/leave.
See plan.md for design; retain energy/event baselines and touch readout after pointer release.

Review correction results: UTC-launched Jest 53 suites/341 tests passed; lint/build/e2e types
passed. Focused checks 12/12 and corrected deferred-shell/default-view checks 6/6 passed. Final
full browser rerun: 201 passed, nine existing desktop phone-only skips, zero failures (210 cases).
The first full run exposed Vite query strings bypassing the old exact gate; the pathname gate
fix and rerun are recorded in plan.md. T031/T033 retain their physical/manual checks.
